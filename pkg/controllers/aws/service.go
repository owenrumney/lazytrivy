package aws

import (
	"context"
	"strings"

	"github.com/awesome-gocui/gocui"
	"github.com/owenrumney/lazytrivy/pkg/logger"
	"github.com/owenrumney/lazytrivy/pkg/output"
	"github.com/owenrumney/lazytrivy/pkg/widgets"
)

func (c *Controller) CacheDirectory() string {
	return c.cacheDirectory
}

func (c *Controller) SetSelected(selected string) {
	c.selectedService = strings.TrimSpace(selected)
}

func (c *Controller) UpdateCache(ctx context.Context) (*output.Report, error) {

	var cancellable context.Context
	c.Lock()
	defer c.Unlock()
	cancellable, c.ActiveCancel = context.WithCancel(ctx)
	report, err := c.DockerClient.ScanAccount(cancellable, c.Config.AWS.AccountNo, c.Config.AWS.Region, c)
	if err != nil {
		return nil, err
	}
	return report, nil
}

func (c *Controller) ScanService(_ context.Context, serviceName string) {

	report, err := c.state.getServiceReport(c.Config.AWS.AccountNo, c.Config.AWS.Region, serviceName)
	if err != nil {
		return
	}
	c.Cui.Update(func(gocui *gocui.Gui) error {
		if err := c.RenderAWSResultsReportSummary(report); err != nil {
			return err
		}
		_, err = c.Cui.SetCurrentView(widgets.Results)
		return err
	})
}

func (c *Controller) CancelCurrentScan(_ *gocui.Gui, _ *gocui.View) error {
	c.Lock()
	defer c.Unlock()
	if c.ActiveCancel != nil {
		logger.Debugf("Cancelling current scan")
		c.UpdateStatus("Current scan cancelled.")
		c.ActiveCancel()
		c.ActiveCancel = nil
	}
	return nil
}
