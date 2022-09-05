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
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/owenrumney/lazytrivy/pkg/logger"
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
	logger.Debug("Creating docker client")
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
			if strings.HasPrefix(imageName, "lazytrivy:") {
				logger.Debug("Found trivy image %s", imageName)
				c.trivyImagePresent = true

				continue
			}
			imageNames = append(imageNames, imageName)
		}
	}

	sort.Strings(imageNames)
	c.imageNames = imageNames

	logger.Debug("Found %d images", len(imageNames))
	return c.imageNames
}

func (c *Client) ScanAccount(ctx context.Context, accountNo, region string, progress Progress) (*output.Report, error) {
	return c.ScanService(ctx, "", accountNo, region, progress)
}

func (c *Client) ScanService(ctx context.Context, serviceName string, accountNo, region string, progress Progress) (*output.Report, error) {
	var env []string

	var updateCache bool
	target := accountNo
	additionalInfo := " maybe make a cuppa"
	if serviceName != "" {
		updateCache = true
		target = serviceName
		additionalInfo = ""
	}

	progress.UpdateStatus(fmt.Sprintf("Scanning %s...%s", target, additionalInfo))
	for _, envVar := range os.Environ() {
		if strings.HasPrefix(envVar, "AWS_") {
			env = append(env, envVar)
		}
	}

	command := []string{
		"aws", "--region", region, "-f=json",
	}

	if serviceName != "" {
		logger.Debug("Scan will target service %s", serviceName)
		command = append(command, "--services", serviceName)
	}
	if updateCache {
		logger.Debug("Cache will be updated for %s", serviceName)
		command = append(command, "--update-cache")
	}
	return c.scan(ctx, command, target, env, progress)
}

func (c *Client) ScanImage(ctx context.Context, imageName string, progress Progress) (*output.Report, error) {
	logger.Debug("Scanning image %s", imageName)
	progress.UpdateStatus(fmt.Sprintf("Scanning image %s...", imageName))
	command := []string{"image", "-f=json", imageName}

	return c.scan(ctx, command, imageName, []string{}, progress)
}

func (c *Client) scan(ctx context.Context, command []string, scanTarget string, env []string, progress Progress) (*output.Report, error) {
	if !c.trivyImagePresent {
		logger.Debug("Creating the docker image, it isn't present")

		dockerfile := createDockerFile()
		tempDir, err := os.MkdirTemp("", "lazytrivy")
		dockerFilePath := filepath.Join(tempDir, "Dockerfile")

		defer func() { _ = os.RemoveAll(tempDir) }()

		if err := os.WriteFile(dockerFilePath, []byte(dockerfile), 0644); err != nil {
			return nil, err
		}

		tar, err := archive.TarWithOptions(tempDir, &archive.TarOptions{})
		if err != nil {
			return nil, err
		}

		resp, err := c.client.ImageBuild(ctx, tar, types.ImageBuildOptions{
			PullParent: true,
			Dockerfile: "Dockerfile",
			Tags:       []string{"lazytrivy:latest"},
		})

		if err != nil {
			return nil, err
		}

		_, _ = io.Copy(io.Discard, resp.Body)
		if err := resp.Body.Close(); err != nil {
			return nil, err
		}

	}

	logger.Debug("Running trivy scan with command %s", command)

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		logger.Debug("Error getting user home dir: %s", err)
		userHomeDir = os.TempDir()
	}

	cachePath := filepath.Join(userHomeDir, ".cache")
	awsPath := filepath.Join(userHomeDir, ".aws")

	cont, err := c.client.ContainerCreate(ctx, &container.Config{
		Image:        "lazytrivy:latest",
		Cmd:          command,
		Env:          env,
		AttachStdout: true,
		AttachStderr: false,
		User:         "trivy",
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

	//  make sure we kill the container
	defer func() {
		logger.Debug("Removing container %s", cont.ID)
		_ = c.client.ContainerRemove(ctx, cont.ID, types.ContainerRemoveOptions{})
	}()

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
		_ = c.client.ContainerRemove(ctx, cont.ID, types.ContainerRemoveOptions{})
		return nil, ctx.Err() // nolint
	default:

		progress.UpdateStatus(fmt.Sprintf("Scanning of %s...done", scanTarget))
		return rep, nil
	}
}

func (c *Client) ScanAllImages(ctx context.Context, progress Progress) ([]*output.Report, error) {
	var reports []*output.Report // nolint

	for _, imageName := range c.imageNames {
		progress.UpdateStatus(fmt.Sprintf("Scanning image %s...", imageName))
		logger.Debug("Scanning image %s", imageName)

		report, err := c.ScanImage(ctx, imageName, progress)
		if err != nil {
			return nil, err
		}
		progress.UpdateStatus(fmt.Sprintf("Scanning image %s...done", imageName))
		logger.Debug("Scanning image %s...done", imageName)
		reports = append(reports, report)
		select {
		case <-ctx.Done():
			logger.Debug("Context cancelled")
			return nil, ctx.Err() // nolint
		default:
		}
	}

	return reports, nil
}
