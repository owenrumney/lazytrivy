package dockerClient

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/owenrumney/lazytrivy/pkg/logger"
	"github.com/owenrumney/lazytrivy/pkg/output"
)

func (c *Client) ListImages() []string {
	images, err := c.client.ImageList(context.Background(), types.ImageListOptions{
		All:     false,
		Filters: filters.Args{},
	})
	if err != nil {
		logger.Errorf("Error listing images: %s", err)
		return nil
	}

	var imageNames []string

	for _, image := range images {
		if image.RepoTags != nil {
			imageName := image.RepoTags[0]

			if strings.HasPrefix(imageName, "aquasec/trivy") {
				logger.Debugf("Found trivy image %s", imageName)
				c.trivyImagePresent = true
				continue
			} else if strings.HasPrefix(imageName, "lazytrivy") {
				logger.Debugf("Found lazy trivy image %s", imageName)
				c.lazyTrivyImagePresent = true
				continue
			} else if strings.HasPrefix(imageName, "<none>:") {
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

func (c *Client) ScanAllImages(ctx context.Context, progress Progress, reportComplete func(report *output.Report) error) error {

	for _, imageName := range c.imageNames {
		progress.UpdateStatus(fmt.Sprintf("Scanning image %s...", imageName))
		logger.Debugf("Scanning image %s", imageName)

		report, err := c.ScanImage(ctx, imageName, progress)
		if err != nil {
			return err
		}
		progress.UpdateStatus(fmt.Sprintf("Scanning image %s...done", imageName))
		logger.Debugf("Scanning image %s...done", imageName)
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

func (c *Client) ScanImage(ctx context.Context, imageName string, progress Progress) (*output.Report, error) {
	logger.Debugf("Scanning image %s", imageName)
	progress.UpdateStatus(fmt.Sprintf("Scanning image %s...", imageName))
	command := []string{"image", "-f=json", imageName}

	return c.scan(ctx, command, imageName, []string{}, progress, "aquasec/trivy:latest")
}
