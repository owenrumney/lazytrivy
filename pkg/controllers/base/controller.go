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

func (g *Controller) SetManager() {
	views := make([]gocui.Manager, 0, len(g.Views)+1)
	for _, v := range g.Views {
		views = append(views, v)
	}

	views = append(views, gocui.ManagerFunc(g.LayoutFunc))
	g.Cui.SetManager(views...)
}

func (g *Controller) RefreshView(viewName string) {
	g.Cui.Update(func(_ *gocui.Gui) error {
		if v, ok := g.Views[viewName]; ok {
			v.RefreshView()
		}
		return nil
	})
}

func (g *Controller) RefreshWidget(widget widgets.Widget) {
	g.Cui.Update(func(gui *gocui.Gui) error {
		return widget.Layout(gui)
	})
}

func (g *Controller) RenderResultsReport(report *output.Report) error {
	if v, ok := g.Views[widgets.Results].(*widgets.ResultsWidget); ok {
		v.RenderReport(report, "ALL")
		_, err := g.Cui.SetCurrentView(widgets.Results)
		if err != nil {
			return fmt.Errorf("failed to set current view: %w", err)
		}
	}
	return nil
}

func (g *Controller) RenderAWSResultsReport(report *output.Report) error {
	if v, ok := g.Views[widgets.Results].(*widgets.AWSResultWidget); ok {
		v.RenderReport(report, "ALL")
		_, err := g.Cui.SetCurrentView(widgets.Results)
		if err != nil {
			return fmt.Errorf("failed to set current view: %w", err)
		}
	}
	return nil
}

func (g *Controller) RenderAWSResultsReportSummary(report *output.Report) error {
	if v, ok := g.Views[widgets.Results].(*widgets.AWSResultWidget); ok {
		v.UpdateResultsTable(report)
		_, err := g.Cui.SetCurrentView(widgets.Results)
		if err != nil {
			return fmt.Errorf("error setting current view: %w", err)
		}
	}
	return errors.New("failed to render results report summary") //nolint:goerr113
}

func (g *Controller) RenderResultsReportSummary(reports []*output.Report) error {
	if v, ok := g.Views[widgets.Results].(*widgets.ResultsWidget); ok {
		v.UpdateResultsTable(reports)
		_, err := g.Cui.SetCurrentView(widgets.Results)
		if err != nil {
			return fmt.Errorf("error setting current view: %w", err)
		}
	}
	return errors.New("failed to render results report summary") //nolint:goerr113
}

func (g *Controller) UpdateStatus(status string) {
	if v, ok := g.Views[widgets.Status].(*widgets.StatusWidget); ok {
		v.UpdateStatus(status)
		v.RefreshView()
	}
}

func (g *Controller) ClearStatus() {
	g.UpdateStatus("")
}

func Quit(_ *gocui.Gui, _ *gocui.View) error {
	return gocui.ErrQuit
}
