package filesystem

import (
	"fmt"

	"github.com/awesome-gocui/gocui"
	"github.com/owenrumney/lazytrivy/pkg/config"
	"github.com/owenrumney/lazytrivy/pkg/controllers/base"
	"github.com/owenrumney/lazytrivy/pkg/dockerClient"
	"github.com/owenrumney/lazytrivy/pkg/logger"
	"github.com/owenrumney/lazytrivy/pkg/output"
	"github.com/owenrumney/lazytrivy/pkg/widgets"
)

type Controller struct {
	*base.Controller
	*state
}

func (c *Controller) SetWorkingDirectory(dir string) {
	c.workingDirectory = dir
	c.Config.Filesystem.WorkingDirectory = dir

	if v, ok := c.Views[widgets.ScanPath]; ok {
		if sp, ok := v.(*widgets.ScanPathWidget); ok {
			sp.UpdateWorkingDir(dir)
		}
	}
}

func NewFilesystemController(cui *gocui.Gui, dockerClient *dockerClient.Client, cfg *config.Config) *Controller {

	return &Controller{
		&base.Controller{
			Cui:          cui,
			DockerClient: dockerClient,
			Views:        make(map[string]widgets.Widget),
			LayoutFunc:   layout,
			HelpFunc:     help,
			Config:       cfg,
		},
		&state{
			workingDirectory: cfg.Filesystem.WorkingDirectory,
		},
	}
}

func (c *Controller) CreateWidgets(manager base.Manager) error {
	logger.Debugf("Creating file system view widgets")

	maxX, maxY := c.Cui.Size()
	c.Views[widgets.Files] = widgets.NewFilesWidget(widgets.Files, c)
	c.Views[widgets.Results] = widgets.NewFSResultWidget(widgets.Results, c)
	c.Views[widgets.Menu] = widgets.NewMenuWidget(widgets.Menu, 0, maxY-3, maxX-1, maxY-1)
	c.Views[widgets.Status] = widgets.NewStatusWidget(widgets.Status)
	c.Views[widgets.ScanPath] = widgets.NewScanPathWidget(widgets.Host, c.workingDirectory, c)

	for _, v := range c.Views {
		_ = v.Layout(c.Cui)
		manager.AddViews(v)
	}
	manager.AddViews(gocui.ManagerFunc(c.LayoutFunc))
	c.SetManager()

	return nil
}

func (c *Controller) Initialise() error {
	logger.Debugf("Initialising Filesystem controller")
	var outerErr error

	c.Cui.Update(func(gui *gocui.Gui) error {

		logger.Debugf("Configuring keyboard shortcuts")
		if err := c.configureKeyBindings(); err != nil {
			return fmt.Errorf("failed to configure global keys: %w", err)
		}

		for _, v := range c.Views {
			if err := v.ConfigureKeys(gui); err != nil {
				return fmt.Errorf("failed to configure view keys: %w", err)
			}
		}

		_, err := gui.SetCurrentView(widgets.Files)
		if err != nil {
			outerErr = fmt.Errorf("failed to set current view: %w", err)
		}

		return err
	})

	return outerErr
}

func (c *Controller) Tab() widgets.Tab {
	return widgets.FileSystemTab
}

func (c *Controller) SetSelected(selected string) {
	if v, ok := c.Views[widgets.Files].(*widgets.FilesWidget); ok {
		c.currentTarget = v.SelectTarget()
	}

	c.currentTarget = selected
}

func (c *Controller) RenderFilesystemFileReport() error {
	if v, ok := c.Views[widgets.Results].(*widgets.FSResultWidget); ok {
		logger.Tracef("Rendering filesystem report for %s", c.currentTarget)
		result, err := c.currentReport.GetResultForTarget(c.currentTarget)
		if err != nil {
			return err
		}
		c.currentResult = result

		if result.HasIssues() {

			v.RenderReport(result, c.currentReport, "ALL")
			if _, err := c.Cui.SetCurrentView(widgets.Results); err != nil {
				return fmt.Errorf("failed to set current view: %w", err)
			}
		} else {

			lines := []string{
				"Great News!",
				"",
				"No Issues found!",
			}

			announcement := widgets.NewAnnouncementWidget(widgets.Announcement, "No Results", lines, c.Cui)
			_ = announcement.Layout(c.Cui)
			_, _ = c.Cui.SetCurrentView(widgets.Announcement)

		}
	}
	return nil
}

func (c *Controller) RenderFilesystemFileReportd(report *output.Report) error {
	if v, ok := c.Views[widgets.Results].(*widgets.FSResultWidget); ok {
		v.UpdateResultsTable([]*output.Report{report}, c.Cui)

	}
	return fmt.Errorf("failed to render results report summary") //nolint:goerr113
}

func (c *Controller) ScanVulnerabilities(g *gocui.Gui, _ *gocui.View) error {
	return c.scanVulnerabilities()
}
