package image

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/awesome-gocui/gocui"
	"github.com/owenrumney/lazytrivy/pkg/logger"
	"github.com/owenrumney/lazytrivy/pkg/output"
	"github.com/owenrumney/lazytrivy/pkg/widgets"
)

func (c *Controller) SetSelected(selected string) {
	if selected == "" {
		return
	}
	c.UpdateStatus(fmt.Sprintf("Selected: %s", selected))
	c.setSelected(strings.TrimSpace(selected))
}

func (c *Controller) ScanImage(ctx context.Context) {
	var cancellable context.Context
	c.Lock()
	defer c.Unlock()
	cancellable, c.ActiveCancel = context.WithCancel(ctx)
	go func() {
		report, err := c.Engine.ScanImage(cancellable, c.selectedImage, c.Config, c)
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
	remote, err := widgets.NewRemoteImageWidget(widgets.Remote, maxX, maxY, 150, c)
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

func (c *Controller) ScanAllImages(gui *gocui.Gui, _ *gocui.View) error {

	_, err := gui.SetCurrentView(widgets.Status)
	if err != nil {
		return nil
	}

	var cancellable context.Context
	c.Lock()
	defer c.Unlock()
	cancellable, c.ActiveCancel = context.WithCancel(context.Background())

	if len(c.images) > 10 {
		lines := []string{
			"", fmt.Sprintf("Scanning %d images may take a while", len(c.images)), "",
			"Press 't' to terminate if you get bored", "",
		}

		announce := widgets.NewAnnouncementWidget(widgets.Announcement, "Caution", lines, c.Cui, widgets.Status)
		if err := announce.Layout(gui); err != nil {
			return err
		}

		_, err := gui.SetCurrentView(widgets.Announcement)
		if err != nil {
			return nil
		}
	}

	go func() {
		var reports []*output.Report

		err := c.Engine.ScanAllImages(cancellable, c.Config, c, func(report *output.Report) error {
			reports = append(reports, report)
			if err := c.RenderResultsReportSummary(reports); err != nil {
				c.UpdateStatus(err.Error())
				c.returnToResults()
				return err
			}
			return nil
		})
		if err != nil {
			c.returnToResults()
			return
		}
		if err := c.RenderResultsReportSummary(reports); err != nil {
			c.returnToResults()
			c.UpdateStatus(err.Error())
		}
		c.UpdateStatus("All images scanned.")
		c.returnToResults()

	}()
	return nil
}

func (c *Controller) returnToResults() {
	_ = c.Cui.DeleteView(widgets.Announcement)

	_, err := c.Cui.SetCurrentView(widgets.Results)
	if err != nil {
		logger.Errorf("failed to set current view: %v", err)
	}
}

func (c *Controller) RefreshImages() error {
	c.UpdateStatus("Refreshing images")
	defer c.ClearStatus()

	images, err := c.Engine.ListImages()
	if err != nil {
		c.UpdateStatus(fmt.Sprintf("Error listing images: %v", err))
		return err
	}

	w, _ := c.Cui.Size()

	c.updateImages(images, w/4)

	if v, ok := c.Views[widgets.Images].(*widgets.ImagesWidget); ok {
		return v.RefreshImages(c.images, c.imageWidth)
	}
	return errors.New("error getting the images view") //nolint:goerr113
}

func (c *Controller) CancelCurrentScan(gui *gocui.Gui, _ *gocui.View) error {
	c.Lock()
	defer c.Unlock()
	if c.ActiveCancel != nil {
		logger.Debugf("Cancelling current scan")
		c.UpdateStatus("Current scan cancelled.")
		c.ActiveCancel()
		c.ActiveCancel = nil
		_, _ = gui.SetCurrentView(widgets.Results)
	}
	return nil
}
