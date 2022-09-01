package aws

import (
	"context"
	"fmt"
	"strings"

	"github.com/awesome-gocui/gocui"
	"github.com/owenrumney/lazytrivy/pkg/widgets"
)

func (c *Controller) CacheDirectory() string {
	return c.cacheDirectory
}

func (c *Controller) SetSelected(selected string) {
	c.selectedService = strings.TrimSpace(selected)
}

func (c *Controller) ScanService(ctx context.Context, serviceName string) {
	// c.cleanupResults()

	// var cancellable context.Context
	// c.Lock()
	// defer c.Unlock()
	// cancellable, c.ActiveCancel = context.WithCancel(ctx)
	// go func() {
	// report, err := c.DockerClient.ScanService(cancellable, c.selectedService, c)
	// if err != nil {
	// 	return
	// }
	report, err := c.state.getServiceReport(c.Config.AWS.AccountNo, c.Config.AWS.Region, serviceName)
	if err != nil {
		return
	}
	c.Cui.Update(func(gocui *gocui.Gui) error {
		return c.RenderAWSResultsReportSummary(report)
	})
	// }()
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

func (c *Controller) addNewAccount(gui *gocui.Gui, _ *gocui.View) error {
	maxX, maxY := gui.Size()

	account, _, err := c.discoverAccount()
	if err != nil {
		account = ""
	}

	gui.Cursor = true
	remote, err := widgets.NewAddAccountWidget(widgets.NewAccount, maxX, maxY, 150, account, c)
	if err != nil {
		return fmt.Errorf("failed to create add account input: %w", err)
	}
	gui.Update(func(g *gocui.Gui) error {
		gui.Mouse = false
		if err := remote.Layout(gui); err != nil {
			return fmt.Errorf("failed to layout add account input: %w", err)
		}
		_, err := gui.SetCurrentView(widgets.NewAccount)
		if err != nil {
			return fmt.Errorf("failed to set current view: %w", err)
		}
		return nil
	})
	return nil
}

func (c *Controller) addNewRegion(gui *gocui.Gui, _ *gocui.View) error {
	maxX, maxY := gui.Size()

	_, region, err := c.discoverAccount()
	if err != nil {
		region = ""
	}
	gui.Cursor = true
	remote, err := widgets.NewAddAccountWidget(widgets.NewAccount, maxX, maxY, 150, region, c)
	if err != nil {
		return fmt.Errorf("failed to create add account input: %w", err)
	}
	gui.Update(func(g *gocui.Gui) error {
		gui.Mouse = false
		if err := remote.Layout(gui); err != nil {
			return fmt.Errorf("failed to layout add account input: %w", err)
		}
		_, err := gui.SetCurrentView(widgets.NewAccount)
		if err != nil {
			return fmt.Errorf("failed to set current view: %w", err)
		}
		return nil
	})
	return nil
}
