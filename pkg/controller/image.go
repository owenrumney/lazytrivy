package controller

import (
	"context"
	"errors"
	"strings"

	"github.com/awesome-gocui/gocui"
	"github.com/owenrumney/lazytrivy/pkg/widgets"
)

func (g *Controller) SetSelectedImage(selected string) {
	g.setSelected(strings.TrimSpace(selected))
}

func (g *Controller) ScanImage(ctx context.Context, imageName string) {
	g.cleanupResults()

	var cancellable context.Context
	g.Lock()
	defer g.Unlock()
	cancellable, g.activeCancel = context.WithCancel(ctx)
	go func() {
		report, err := g.dockerClient.ScanImage(cancellable, imageName, g)
		if err != nil {
			return
		}
		g.cui.Update(func(gocui *gocui.Gui) error {
			return g.renderResultsReport(imageName, report)
		})
	}()
}

func (g *Controller) ScanAllImages(ctx context.Context) {
	g.cleanupResults()

	var cancellable context.Context
	g.Lock()
	defer g.Unlock()
	cancellable, g.activeCancel = context.WithCancel(ctx)
	go func() {
		reports, err := g.dockerClient.ScanAllImages(cancellable, g)
		if err != nil {
			return
		}
		if err := g.renderResultsReportSummary(reports); err != nil {
			g.UpdateStatus(err.Error())
		}
		g.UpdateStatus("All images scanned.")
	}()
}

func (g *Controller) RefreshImages() error {
	g.UpdateStatus("Refreshing images")
	defer g.ClearStatus()

	images := g.dockerClient.ListImages()
	g.updateImages(images)

	if v, ok := g.views[widgets.Images].(*widgets.ImagesWidget); ok {
		return v.RefreshImages(g.images, g.imageWidth)
	}
	return errors.New("error getting the images view") //nolint:goerr113
}
