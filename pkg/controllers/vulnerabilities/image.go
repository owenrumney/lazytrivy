package vulnerabilities

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/awesome-gocui/gocui"
	"github.com/owenrumney/lazytrivy/pkg/widgets"
)

func (c *Controller) SetSelected(selected string) {
	c.setSelected(strings.TrimSpace(selected))
}

func (c *Controller) ScanImage(ctx context.Context, imageName string) {
	// c.cleanupResults()

	var cancellable context.Context
	c.Lock()
	defer c.Unlock()
	cancellable, c.ActiveCancel = context.WithCancel(ctx)
	go func() {
		report, err := c.DockerClient.ScanImage(cancellable, imageName, c)
		if err != nil {
			return
		}
		c.Cui.Update(func(gocui *gocui.Gui) error {
			return c.RenderResultsReport(report)
		})
	}()
}

func (c *Controller) scanRemote(gui *gocui.Gui, _ *gocui.View) error {
	maxX, maxY := gui.Size()

	gui.Cursor = true
	remote, err := widgets.NewInputWidget(widgets.Remote, maxX, maxY, 150, c)
	if err != nil {
		return fmt.Errorf("failed to create remote input: %w", err)
	}
	gui.Update(func(g *gocui.Gui) error {
		gui.Mouse = false
		if err := remote.Layout(gui); err != nil {
			return fmt.Errorf("failed to layout remote input: %w", err)
		}
		_, err := gui.SetCurrentView(widgets.Remote)
		if err != nil {
			return fmt.Errorf("failed to set current view: %w", err)
		}
		return nil
	})
	return nil
}

func (c *Controller) ScanAllImages(ctx context.Context) {
	// c.cleanupResults()

	var cancellable context.Context
	c.Lock()
	defer c.Unlock()
	cancellable, c.ActiveCancel = context.WithCancel(ctx)
	go func() {
		reports, err := c.DockerClient.ScanAllImages(cancellable, c)
		if err != nil {
			return
		}
		if err := c.RenderResultsReportSummary(reports); err != nil {
			c.UpdateStatus(err.Error())
		}
		c.UpdateStatus("All images scanned.")
	}()
}

func (c *Controller) RefreshImages() error {
	c.UpdateStatus("Refreshing images")
	defer c.ClearStatus()

	images := c.DockerClient.ListImages()
	c.updateImages(images)

	if v, ok := c.Views[widgets.Images].(*widgets.ImagesWidget); ok {
		return v.RefreshImages(c.images, c.imageWidth)
	}
	return errors.New("error getting the images view") //nolint:goerr113
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