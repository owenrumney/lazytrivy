package base

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/awesome-gocui/gocui"
	"github.com/owenrumney/lazytrivy/pkg/config"
	"github.com/owenrumney/lazytrivy/pkg/docker"
	"github.com/owenrumney/lazytrivy/pkg/output"
	"github.com/owenrumney/lazytrivy/pkg/widgets"
)

type Controller struct {
	sync.Mutex
	Cui          *gocui.Gui
	DockerClient *docker.Client
	Views        map[string]widgets.Widget
	LayoutFunc   func(g *gocui.Gui) error
	HelpFunc     func(g *gocui.Gui, v *gocui.View) error
	Tab          widgets.Tab
	Config       *config.Config

	ActiveCancel context.CancelFunc
}

type ControllerView interface {
	CreateWidgets(manager Manager) error
	Initialise() error
	Tab() widgets.Tab
}

func (c *Controller) SetManager() {
	views := make([]gocui.Manager, 0, len(c.Views)+1)
	for _, v := range c.Views {
		views = append(views, v)
	}

	views = append(views, gocui.ManagerFunc(c.LayoutFunc))
	c.Cui.SetManager(views...)
}

func (c *Controller) RefreshView(viewName string) {
	c.Cui.Update(func(_ *gocui.Gui) error {
		if v, ok := c.Views[viewName]; ok {
			v.RefreshView()
		}
		return nil
	})
}

func (c *Controller) RefreshWidget(widget widgets.Widget) {
	c.Cui.Update(func(gui *gocui.Gui) error {
		return widget.Layout(gui)
	})
}

func (c *Controller) RenderResultsReport(report *output.Report) error {
	if v, ok := c.Views[widgets.Results].(*widgets.ResultsWidget); ok {
		v.RenderReport(report, "ALL")
		_, err := c.Cui.SetCurrentView(widgets.Results)
		if err != nil {
			return fmt.Errorf("failed to set current view: %w", err)
		}
	}
	return nil
}

func (c *Controller) RenderAWSResultsReport(report *output.Report) error {
	if v, ok := c.Views[widgets.Results].(*widgets.AWSResultWidget); ok {
		v.RenderReport(report, "ALL")
		_, err := c.Cui.SetCurrentView(widgets.Results)
		if err != nil {
			return fmt.Errorf("failed to set current view: %w", err)
		}
	}
	return nil
}

func (c *Controller) RenderAWSResultsReportSummary(report *output.Report) error {
	if v, ok := c.Views[widgets.Results].(*widgets.AWSResultWidget); ok {
		v.UpdateResultsTable(report)
		// _, err := c.Cui.SetCurrentView(widgets.Results)
		// if err != nil {
		// 	return fmt.Errorf("error setting current view: %w", err)
		// }
	}
	return errors.New("failed to render results report summary") //nolint:goerr113
}

func (c *Controller) RenderResultsReportSummary(reports []*output.Report) error {
	if v, ok := c.Views[widgets.Results].(*widgets.ResultsWidget); ok {
		v.UpdateResultsTable(reports)
		_, err := c.Cui.SetCurrentView(widgets.Results)
		if err != nil {
			return fmt.Errorf("error setting current view: %w", err)
		}
	}
	return errors.New("failed to render results report summary") //nolint:goerr113
}

func (c *Controller) UpdateStatus(status string) {
	if v, ok := c.Views[widgets.Status].(*widgets.StatusWidget); ok {
		v.UpdateStatus(status)
		v.RefreshView()
	}
}

func (c *Controller) ClearStatus() {
	c.UpdateStatus("")
}

func Quit(_ *gocui.Gui, _ *gocui.View) error {
	return gocui.ErrQuit
}

func (c *Controller) SetKeyBinding(viewName string, key interface{}, mod gocui.Modifier, handler func(*gocui.Gui, *gocui.View) error) error {
	if err := c.Cui.SetKeybinding(viewName, key, mod, handler); err != nil {
		return fmt.Errorf("failed to set keybinding for %s: %w", key, err)
	}
	return nil
}
