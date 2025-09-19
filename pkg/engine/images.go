package engine

import (
	"context"
	"fmt"

	"github.com/owenrumney/lazytrivy/pkg/config"
	"github.com/owenrumney/lazytrivy/pkg/logger"
	"github.com/owenrumney/lazytrivy/pkg/output"
)

func (c *Client) ScanAllImages(ctx context.Context, cfg *config.Config, progress Progress, reportComplete func(report *output.Report) error) error {

	for _, imageName := range c.imageNames {
		report, err := c.ScanImage(ctx, imageName, cfg, progress)
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

func (c *Client) ScanImage(ctx context.Context, imageName string, cfg *config.Config, progress Progress) (*output.Report, error) {
	progress.UpdateStatus(fmt.Sprintf("Scanning image %s...", imageName))
	command := []string{"image", "-f=json"}
	command = c.updateCommand(command, cfg)
	command = append(command, imageName)

	if report, err := c.scan(ctx, command, imageName, []string{}, progress, "aquasec/trivy:latest", EngineDocker); err == nil {
		return report, nil
	}
	return c.scan(ctx, command, imageName, []string{}, progress, "aquasec/trivy:latest", EnginePodman)
}
