package docker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/owenrumney/lazytrivy/pkg/output"
)

type Progress interface {
	UpdateStatus(status string)
	ClearStatus()
}

type Client struct {
	client            *client.Client
	imageNames        []string
	trivyImagePresent bool
}

func NewClient() *Client {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	return &Client{
		client: cli,
	}
}

func (c *Client) ListImages() []string {
	images, err := c.client.ImageList(context.Background(), types.ImageListOptions{
		All:     false,
		Filters: filters.Args{},
	})
	if err != nil {
		panic(err)
	}

	var imageNames []string

	for _, image := range images {
		if image.RepoTags != nil {
			imageName := image.RepoTags[0]
			if strings.HasPrefix(imageName, "aquasec/trivy:") {
				c.trivyImagePresent = true

				continue
			}
			imageNames = append(imageNames, imageName)
		}
	}

	sort.Strings(imageNames)
	c.imageNames = imageNames

	return c.imageNames
}

func (c *Client) ScanService(ctx context.Context, serviceName string, progress Progress) (*output.Report, error) {
	var env []string

	progress.UpdateStatus(fmt.Sprintf("Scanning service %s...", serviceName))
	for _, envVar := range os.Environ() {
		if strings.HasPrefix(envVar, "AWS_") {
			env = append(env, envVar)
		}
	}

	env = append(env, "AWS_REGION=us-east-1")

	return c.scan(ctx, []string{
		"aws", "--service", serviceName, "--region", "us-east-1", "-f=json",
	}, serviceName, env, progress)
}

func (c *Client) ScanImage(ctx context.Context, imageName string, progress Progress) (*output.Report, error) {

	progress.UpdateStatus(fmt.Sprintf("Scanning image %s...", imageName))
	command := []string{"image", "-f=json", imageName}

	return c.scan(ctx, command, imageName, []string{}, progress)
}

func (c *Client) scan(ctx context.Context, command []string, scanTarget string, env []string, progress Progress) (*output.Report, error) {
	if !c.trivyImagePresent {
		progress.UpdateStatus("Pulling latest Trivy image...")

		resp, _ := c.client.ImagePull(ctx, "aquasec/trivy:latest", types.ImagePullOptions{
			All: false,
		})
		defer func() { _ = resp.Close() }()
		_, _ = io.Copy(io.Discard, resp)
	}
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		userHomeDir = os.TempDir()
	}

	cachePath := filepath.Join(userHomeDir, ".cache")
	awsPath := filepath.Join(userHomeDir, ".aws")

	cont, err := c.client.ContainerCreate(ctx, &container.Config{
		Image:        "aquasec/trivy",
		Cmd:          command,
		Env:          env,
		AttachStdout: true,
		AttachStderr: false,
		User:         "root",
	}, &container.HostConfig{
		Binds: []string{
			"/var/run/docker.sock:/var/run/docker.sock",
			fmt.Sprintf("%s:/root/.cache", cachePath),
			fmt.Sprintf("%s:/root/.aws", awsPath),
		},
	}, nil, nil, "")
	if err != nil {
		return nil, err
	}

	// make sure we kill the container
	// defer func() { _ = c.client.ContainerRemove(ctx, cont.ID, types.ContainerRemoveOptions{}) }()

	if err := c.client.ContainerStart(ctx, cont.ID, types.ContainerStartOptions{}); err != nil {
		return nil, err
	}

	statusCh, errCh := c.client.ContainerWait(ctx, cont.ID, container.WaitConditionNotRunning)
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-errCh:
		if err != nil {
			return nil, err
		}
	case <-statusCh:
	}

	out, err := c.client.ContainerLogs(ctx, cont.ID, types.ContainerLogsOptions{ShowStdout: true, Follow: false})
	if err != nil {
		return nil, err
	}

	content := ""
	buffer := bytes.NewBufferString(content)
	_, _ = stdcopy.StdCopy(buffer, buffer, out)

	rep, err := output.FromJSON(scanTarget, buffer.String())
	if err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err() // nolint
	default:
		progress.UpdateStatus(fmt.Sprintf("Scanning image %s...done", scanTarget))
		return rep, nil
	}
}

func (c *Client) ScanAllImages(ctx context.Context, progress Progress) ([]*output.Report, error) {
	var reports []*output.Report // nolint

	for _, imageName := range c.imageNames {
		progress.UpdateStatus(fmt.Sprintf("Scanning image %s...", imageName))

		report, err := c.ScanImage(ctx, imageName, progress)
		if err != nil {
			return nil, err
		}
		progress.UpdateStatus(fmt.Sprintf("Scanning image %s...done", imageName))
		reports = append(reports, report)
		select {
		case <-ctx.Done():
			return nil, ctx.Err() // nolint
		default:
		}
	}

	return reports, nil
}
