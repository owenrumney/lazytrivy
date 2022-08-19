package vulnerabilities

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/awesome-gocui/gocui"
	"github.com/owenrumney/lazytrivy/pkg/widgets"
)

func (g *VulnerabilityController) SetSelectedImage(selected string) {
	g.setSelected(strings.TrimSpace(selected))
}

func (g *VulnerabilityController) ScanImage(ctx context.Context, imageName string) {
	// g.cleanupResults()

	var cancellable context.Context
	g.Lock()
	defer g.Unlock()
	cancellable, g.ActiveCancel = context.WithCancel(ctx)
	go func() {
		report, err := g.DockerClient.ScanImage(cancellable, imageName, g)
		if err != nil {
			return
		}
		g.Cui.Update(func(gocui *gocui.Gui) error {
			return g.RenderResultsReport(report)
		})
	}()
}

func (g *VulnerabilityController) scanRemote(gui *gocui.Gui, _ *gocui.View) error {
	maxX, maxY := gui.Size()

	gui.Cursor = true
	remote, err := widgets.NewInputWidget(widgets.Remote, maxX, maxY, 150, g)
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

func (g *VulnerabilityController) ScanAllImages(ctx context.Context) {
	// g.cleanupResults()

	var cancellable context.Context
	g.Lock()
	defer g.Unlock()
	cancellable, g.ActiveCancel = context.WithCancel(ctx)
	go func() {
		reports, err := g.DockerClient.ScanAllImages(cancellable, g)
		if err != nil {
			return
		}
		if err := g.RenderResultsReportSummary(reports); err != nil {
			g.UpdateStatus(err.Error())
		}
		g.UpdateStatus("All images scanned.")
	}()
}

func (g *VulnerabilityController) RefreshImages() error {
	g.UpdateStatus("Refreshing images")
	defer g.ClearStatus()

	images := g.DockerClient.ListImages()
	g.updateImages(images)

	if v, ok := g.Views[widgets.Images].(*widgets.ImagesWidget); ok {
		return v.RefreshImages(g.images, g.imageWidth)
	}
	return errors.New("error getting the images view") //nolint:goerr113
}

func (g *VulnerabilityController) CancelCurrentScan(_ *gocui.Gui, _ *gocui.View) error {
	g.Lock()
	defer g.Unlock()
	if g.ActiveCancel != nil {
		g.UpdateStatus("Current scan cancelled.")
		g.ActiveCancel()
		g.ActiveCancel = nil
	}
	return nil
}
