package aws

import (
	"context"
	"strings"

	"github.com/awesome-gocui/gocui"
)

func (c *Controller) CacheDirectory() string {
	return c.cacheDirectory
}

func (c *Controller) SetSelected(selected string) {
	c.selectedService = strings.TrimSpace(selected)
}

func (c *Controller) ScanService(ctx context.Context, imageName string) {
	// c.cleanupResults()

	var cancellable context.Context
	c.Lock()
	defer c.Unlock()
	cancellable, c.ActiveCancel = context.WithCancel(ctx)
	go func() {
		report, err := c.DockerClient.ScanService(cancellable, c.selectedService, c)
		if err != nil {
			return
		}
		c.Cui.Update(func(gocui *gocui.Gui) error {
			return c.RenderAWSResultsReportSummary(report)
		})
	}()
}

func (c *Controller) CancelCurrentScan(_ *gocui.Gui, _ *gocui.View) error {
	c.Lock()
	defer c.Unlock()
	if c.ActiveCancel != nil {
		c.UpdateStatus("Current scan cancelled.")
		c.ActiveCancel()
		c.ActiveCancel = nil
	}
	return nil
}
