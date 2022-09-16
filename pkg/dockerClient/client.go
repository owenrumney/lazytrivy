package dockerClient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
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
	client                *client.Client
	endpoint              string
	socketPath            string
	imageNames            []string
	trivyImagePresent     bool
	lazyTrivyImagePresent bool
}

func NewClient() *Client {

	endpoint, _, err := getHostEndpoint()
	if err != nil {
		logger.Errorf("Error getting docker context: %s", err)
	}

	logger.Debugf("Creating docker client")

	cli, err := client.NewClientWithOpts(client.WithHost(endpoint), client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	socketPath := strings.TrimPrefix(endpoint, "unix://")

	return &Client{
		client:     cli,
		endpoint:   endpoint,
		socketPath: socketPath,
	}
}

func (c *Client) IsDockerDesktop() bool {
	return strings.Contains(strings.ToLower(c.endpoint), "desktop")
}

func (c *Client) scan(ctx context.Context, command []string, scanTarget string, env []string, progress Progress, scanImageName string, additionalBinds ...string) (*output.Report, error) {

	switch scanImageName {
	case "lazytrivy:1.0.0":
		if !c.lazyTrivyImagePresent {
			report, err := c.buildScannerImage(ctx)
			if err != nil {
				return report, err
			}
		}
	case "aquasec/trivy:latest":
		if !c.trivyImagePresent {
			resp, err := c.client.ImagePull(ctx, scanImageName, types.ImagePullOptions{
				All: false,
			})
			if err != nil {
				return nil, err
			}

			_, _ = io.Copy(io.Discard, resp)
		}
	}
	logger.Debugf("Running trivy scan with command %s", command)

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		logger.Debugf("Error getting user home dir: %s", err)
		userHomeDir = os.TempDir()
	}

	cachePath := filepath.Join(userHomeDir, ".cache")

	binds := []string{
		// fmt.Sprintf("%s:/var/run/docker.sock", c.socketPath),
		"/var/run/docker.sock:/var/run/docker.sock",
		fmt.Sprintf("%s:/root/.cache", cachePath),
	}

	binds = append(binds, additionalBinds...)

	user := "root"
	if scanImageName == "lazytrivy:latest" {
		user = "trivy"
	}

	cont, err := c.client.ContainerCreate(ctx, &container.Config{
		Image:        scanImageName,
		Cmd:          command,
		Env:          env,
		AttachStdout: true,
		AttachStderr: true,
		User:         user,
	}, &container.HostConfig{
		Binds: binds,
	}, nil, nil, "")
	if err != nil {
		return nil, err
	}

	// make sure we kill the container
	// defer func() {
	// 	logger.Debugf("Removing container %s", cont.ID)
	// 	_ = c.client.ContainerRemove(ctx, cont.ID, types.ContainerRemoveOptions{})
	// }()

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

	out, err := c.client.ContainerLogs(ctx, cont.ID, types.ContainerLogsOptions{
		ShowStdout: true, ShowStderr: true, Follow: false,
	})
	if err != nil {
		return nil, err
	}

	content := ""
	buffer := bytes.NewBufferString(content)
	errContent := ""
	errBuffer := bytes.NewBufferString(errContent)

	_, _ = stdcopy.StdCopy(buffer, errBuffer, out)

	rep, err := output.FromJSON(scanTarget, buffer.String())
	if err != nil {
		logger.Errorf("Error parsing trivy output, response from container: %s", errBuffer.String())
		progress.UpdateStatus(fmt.Sprintf("Error scanning image %s", scanTarget))
		return nil, err
	}

	select {
	case <-ctx.Done():
		_ = c.client.ContainerRemove(ctx, cont.ID, types.ContainerRemoveOptions{})
		return nil, ctx.Err() // nolint
	default:

		progress.UpdateStatus(fmt.Sprintf("Scanning %s...done", scanTarget))
		return rep, nil
	}
}

func (c *Client) buildScannerImage(ctx context.Context) (*output.Report, error) {
	logger.Debugf("Creating the docker image, it isn't present")

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
		Tags:       []string{"lazytrivy:1.0.0"},
	})
	if err != nil {
		return nil, err
	}

	buffer := bytes.NewBufferString("")
	_, _ = io.Copy(buffer, resp.Body)
	if err := resp.Body.Close(); err != nil {
		return nil, err
	}
	return nil, nil
}
