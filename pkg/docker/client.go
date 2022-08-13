package docker

import (
	"bytes"
	"context"
	"fmt"
	"log"
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
	client *client.Client
	ctx    context.Context
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
				continue
			}
			imageNames = append(imageNames, imageName)
		}
	}

	sort.Strings(imageNames)
	return imageNames
}

func (c *DockerClient) ScanImage(imageName string) output.Report {

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
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
		panic(err)
	}

	if err := cli.ContainerStart(ctx, cont.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
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
		return output.Report{}
	}

	content := ""
	buffer := bytes.NewBufferString(content)
	_, _ = stdcopy.StdCopy(buffer, buffer, out)

	rep, err := output.FromJson(buffer.String())
	if err != nil {
		log.Panicln(err)
	}

	// clean up
	err = cli.ContainerRemove(ctx, cont.ID, types.ContainerRemoveOptions{Force: true})
	if err != nil {
		return rep
	}

	return rep
}
