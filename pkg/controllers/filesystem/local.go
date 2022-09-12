package filesystem

import (
	"context"
	"fmt"
	"strings"

	"github.com/awesome-gocui/gocui"
	"github.com/owenrumney/lazytrivy/pkg/logger"
	"github.com/owenrumney/lazytrivy/pkg/widgets"
)

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

func (c *Controller) moveViewLeft(*gocui.Gui, *gocui.View) error {
	if c.Cui.CurrentView().Name() == widgets.Results {
		_, err := c.Cui.SetCurrentView(widgets.Files)
		if err != nil {
			return fmt.Errorf("error getting the images view: %w", err)
		}
		if v, ok := c.Views[widgets.Images].(*widgets.ImagesWidget); ok {
			return v.SetSelectedImage(c.state.currentTarget)
		}
	}
	return nil
}

func (c *Controller) moveViewRight(*gocui.Gui, *gocui.View) error {
	if c.Cui.CurrentView().Name() == widgets.Services {
		_, err := c.Cui.SetCurrentView(widgets.Results)
		if err != nil {
			return fmt.Errorf("error getting the results view: %w", err)
		}
	}
	return nil
}

func (c *Controller) scanVulnerabilities() error {
	logger.Debugf("scanning for vulnerabilities")
	c.UpdateStatus("Scanning current working dir for vulnerabilities...")

	scanChecks := []string{"vuln", "config", "secret"}

	report, err := c.DockerClient.ScanFilesystem(context.TODO(), c.workingDireectory, scanChecks, c)
	if err != nil {
		return err
	}

	c.currentReport = report

	width := 20
	var targets []string

	for _, result := range report.Results {
		if result.HasIssues() {
			targets = append(targets, result.Target)
		}
	}
	root := createRootDir(targets)

	var lines []string
	lines = root.generateTree(lines, -1)

	for _, line := range lines {
		parts := strings.Split(line, "|")
		if len(parts[0]) > width {
			width = len(parts[0])
		}
	}

	logger.Debugf("Updating the files view with the identified services")
	if v, ok := c.Views[widgets.Files].(*widgets.FilesWidget); ok {
		if err := v.RefreshFiles(lines, width); err != nil {
			return err
		}
	}

	return nil
}

func (c *Controller) ShowTarget(ctx context.Context, target string) {
	c.currentTarget = target
	c.Cui.Update(func(gocui *gocui.Gui) error {
		if err := c.RenderFilesystemFileReport(); err != nil {
			return err
		}
		_, err := c.Cui.SetCurrentView(widgets.Results)
		return err
	})
}
