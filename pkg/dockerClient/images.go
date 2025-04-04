package dockerClient

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/owenrumney/lazytrivy/pkg/logger"
	"github.com/owenrumney/lazytrivy/pkg/output"
)

func (c *Client) ListImages() []string {
	images, err := c.client.ImageList(context.Background(), image.ListOptions{
		All:     false,
		Filters: filters.Args{},
	})
	if err != nil {
		logger.Errorf("Error listing images: %s", err)
		return nil
	}

	var imageNames []string

	for _, image := range images {
		if image.RepoTags != nil && len(image.RepoTags) > 0 {
			imageName := image.RepoTags[0]

			if strings.Contains(imageName, "aquasec/trivy") {
				logger.Debugf("Found trivy image %s", imageName)
				c.trivyImagePresent = true
				continue
			} else if strings.Contains(imageName, "lazytrivy") {
				logger.Debugf("Found lazy trivy image %s", imageName)
				c.lazyTrivyImagePresent = true
				continue
			} else if strings.Contains(imageName, "<none>:") {
				// we don't need to be showing these
				continue
			}
			imageNames = append(imageNames, imageName)
		}
	}
	sort.Strings(imageNames)
	c.imageNames = imageNames

	logger.Debugf("Found %d images", len(imageNames))
	return c.imageNames
}

func (c *Client) ScanAllImages(ctx context.Context, insecure bool, progress Progress, reportComplete func(report *output.Report) error) error {

	for _, imageName := range c.imageNames {
		report, err := c.ScanImage(ctx, imageName, insecure, progress)
		if err != nil {
			return err
		}
		if err := reportComplete(report); err != nil {
			logger.Errorf("Error reporting scan results: %s", err)
			ctx.Done()
		}
		select {
		case <-ctx.Done():
			logger.Debugf("Context cancelled")
			return ctx.Err() // nolint
		default:
		}
	}

	return nil
}

func (c *Client) ScanImage(ctx context.Context, imageName string, insecure bool, progress Progress) (*output.Report, error) {
	progress.UpdateStatus(fmt.Sprintf("Scanning image %s...", imageName))
	command := []string{"image", "-f=json"}
	if insecure {
		command = append(command, "--insecure")
	}
	command = append(command, imageName)

	if report, err := c.scan(ctx, command, imageName, []string{}, progress, "aquasec/trivy:latest", EngineDocker); err == nil {
		return report, nil
	}
	return c.scan(ctx, command, imageName, []string{}, progress, "aquasec/trivy:latest", EnginePodman)
}
