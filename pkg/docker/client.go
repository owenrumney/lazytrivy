package docker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
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

type DockerClient struct {
	client            *client.Client
	ctx               context.Context
	imageNames        []string
	trivyImagePresent bool
}

func NewDockerClient() *DockerClient {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	return &DockerClient{
		ctx:    context.Background(),
		client: cli,
	}
}

func (c *DockerClient) ListImages() []string {

	images, err := c.client.ImageList(c.ctx, types.ImageListOptions{
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

func (c *DockerClient) ScanImage(imageName string) (*output.Report, error) {

	ctx := context.Background()

	resp, _ := c.client.ImagePull(ctx, "aquasec/trivy:latest", types.ImagePullOptions{
		All: false,
	})

	defer func() { _ = resp.Close() }()
	_, _ = io.Copy(ioutil.Discard, resp)

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	cachePath := filepath.Join(os.TempDir(), "trivycache")

	cont, err := cli.ContainerCreate(ctx, &container.Config{
		Image:        "aquasec/trivy",
		Cmd:          []string{"image", "-f=json", imageName},
		AttachStdout: true,
		AttachStderr: true,
	}, &container.HostConfig{
		Binds: []string{
			"/var/run/docker.sock:/var/run/docker.sock",
			fmt.Sprintf("%s:/root/.cache", cachePath),
		},
	}, nil, nil, "")
	if err != nil {
		return nil, err
	}

	if err := cli.ContainerStart(ctx, cont.ID, types.ContainerStartOptions{}); err != nil {
		return nil, err
	}

	statusCh, errCh := cli.ContainerWait(ctx, cont.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			panic(err)
		}
	case <-statusCh:
	}

	out, err := cli.ContainerLogs(ctx, cont.ID, types.ContainerLogsOptions{ShowStdout: true, Follow: false})
	if err != nil {
		return nil, err
	}

	content := ""
	buffer := bytes.NewBufferString(content)
	_, _ = stdcopy.StdCopy(buffer, buffer, out)

	rep, err := output.FromJson(buffer.String())
	if err != nil {
		return nil, err
	}

	// clean up
	err = cli.ContainerRemove(ctx, cont.ID, types.ContainerRemoveOptions{Force: true})
	if err != nil {
		return nil, err
	}

	return rep, nil
}

func (c *DockerClient) ScanAllImages(ctx context.Context) ([]*output.Report, error) {

	var reports []*output.Report

	for _, imageName := range c.imageNames {
		if report, err := c.ScanImage(imageName); err != nil {
			return reports, err
		} else {
			reports = append(reports, report)
		}
		switch {
		case <-ctx.Done():
			return nil, ctx.Err()
		}

	}

	return reports, nil
}
