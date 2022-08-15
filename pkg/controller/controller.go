package controller

import (
	"context"
	"fmt"
	"sync"

	"github.com/awesome-gocui/gocui"
	"github.com/owenrumney/lazytrivy/pkg/docker"
	"github.com/owenrumney/lazytrivy/pkg/output"
	"github.com/owenrumney/lazytrivy/pkg/widgets"
)

type Controller struct {
	sync.Mutex
	cui           *gocui.Gui
	dockerClient  *docker.DockerClient
	views         map[string]widgets.Widget
	selectedImage string
	activeCancel  context.CancelFunc
}

func New() (*Controller, error) {
	cui, err := gocui.NewGui(gocui.OutputNormal, true)
	if err != nil {
		return nil, err
	}
	return &Controller{
		cui:          cui,
		dockerClient: docker.NewDockerClient(),
		views:        make(map[string]widgets.Widget),
	}, nil
}

func (g *Controller) DockerClient() *docker.DockerClient {
	return g.dockerClient
}

func (g *Controller) CreateWidgets() error {
	maxX, maxY := g.cui.Size()

	g.views["images"] = widgets.NewImagesWidget("images", g)
	g.views["results"] = widgets.NewInfoWidget("results", g)
	g.views["menu"] = widgets.NewMenuWidget("menu", 0, maxY-3, maxX-1, maxY-1, g)
	g.views["status"] = widgets.NewStatusWidget("status", g)
	g.views["host"] = widgets.NewHostWidget("host", g)

	g.SetManager()
	g.EnableMouse()

	return nil
}

func (g *Controller) SetManager() {
	views := make([]gocui.Manager, 0, len(g.views)+1)
	for _, v := range g.views {
		views = append(views, v)
	}

	views = append(views, gocui.ManagerFunc(flowLayout))
	g.cui.SetManager(views...)
}

func (g *Controller) Run() error {
	if err := g.cui.MainLoop(); err != nil && err != gocui.ErrQuit {
		return err
	}
	return nil
}

func (g *Controller) Initialise() {
	g.cui.Update(func(gui *gocui.Gui) error {
		if err := g.configureGlobalKeys(); err != nil {
			return err
		}

		for _, v := range g.views {
			if err := v.ConfigureKeys(); err != nil {
				return err
			}
		}

		_, err := gui.SetCurrentView("images")
		return err
	})
}

func (g *Controller) SetKeyBinding(viewName string, key interface{}, mod gocui.Modifier, handler func(*gocui.Gui, *gocui.View) error) error {
	return g.cui.SetKeybinding(viewName, key, mod, handler)
}

func (g *Controller) Close() {
	g.cui.Close()
}

func (g *Controller) ShowCursor() {
	g.cui.Cursor = true
}

func (g *Controller) HideCursor() {
	g.cui.Cursor = false
}

func (g *Controller) EnableMouse() {
	g.cui.Mouse = true
}

func (g *Controller) DisableMouse() {
	g.cui.Mouse = false
}

func (g *Controller) CancelCurrentScan(_ *gocui.Gui, _ *gocui.View) error {
	g.Lock()
	defer g.Unlock()
	if g.activeCancel != nil {
		g.UpdateStatus("Current scan cancelled.")
		g.activeCancel()
		g.activeCancel = nil
	}
	return nil
}

func (g *Controller) SetSelectedImage(selected string) {
	g.selectedImage = selected
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
			return // TODO: Need to do something here
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
			return // TODO: Need to do something here
		}
		if err := g.renderResultsReportSummary(reports); err != nil {
			g.UpdateStatus(err.Error())
		}
		g.UpdateStatus("All images scanned.")
	}()

}

func flowLayout(g *gocui.Gui) error {

	imagesWidth := 0
	viewNames := []string{"images", "host", "results", "menu", "status"}
	maxX, maxY := g.Size()
	x := 0
	for _, viewName := range viewNames {
		v, err := g.View(viewName)
		w, _ := v.Size()
		h := 1
		nextW := w
		nextH := maxY - 3
		nextX := x

		switch v.Name() {
		case "host":
			nextW = imagesWidth
			nextX = 0
			nextH = 3
		case "images":
			imagesWidth = w
			h = 4
		case "status":
			nextW = maxX - 1
			h = maxY - 5
		case "results":
			nextW = maxX - 1
			nextH = maxY - 6
		case "menu", "remote", "critical":
			continue
		}

		_, err = g.SetView(v.Name(), nextX, h, nextW, nextH, 0)
		if err != nil && err != gocui.ErrUnknownView {
			return err
		}
		if v.Name() == "images" {
			x += nextW + 1
		}
		h = 1
	}
	return nil
}

func quit(_ *gocui.Gui, _ *gocui.View) error {
	return gocui.ErrQuit
}

func setView(g *gocui.Gui, v *gocui.View) error {
	_, err := g.SetCurrentView(v.Name())
	return err
}

func (g *Controller) scanRemote(gui *gocui.Gui, _ *gocui.View) error {

	maxX, maxY := gui.Size()

	g.ShowCursor()
	remote, err := widgets.NewInput("remote", maxX, maxY, 150, g)
	if err != nil {
		return err
	}
	gui.Update(func(g *gocui.Gui) error {
		gui.Mouse = false
		if err := remote.Layout(gui); err != nil {
			return err
		}
		_, err := gui.SetCurrentView("remote")
		return err
	})
	return nil
}

func (g *Controller) renderResultsReport(imageName string, report *output.Report) error {
	if v, ok := g.views["results"]; ok {
		infoWidget := v.(*widgets.InfoWidget)
		infoWidget.RenderReport(report, fmt.Sprintf(" %s ", imageName), "ALL")
		_, err := g.cui.SetCurrentView("results")
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *Controller) renderResultsReportSummary(reports []*output.Report) error {
	if v, ok := g.views["results"]; ok {
		v.(*widgets.InfoWidget).UpdateResultsTable(reports)
		_, err := g.cui.SetCurrentView("results")
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *Controller) UpdateStatus(status string) {
	if v, ok := g.views["status"]; ok {
		v.(*widgets.StatusWidget).UpdateStatus(status)
		v.RefreshView()
	}
}

func (g *Controller) ClearStatus() {
	g.UpdateStatus("")
}

func (g *Controller) RefreshImages() error {
	g.UpdateStatus("Refreshing images")
	images := g.dockerClient.ListImages()
	g.ClearStatus()
	return g.views["images"].(*widgets.ImagesWidget).RefreshImages(images)
}

func (g *Controller) RefreshView(viewName string) {
	g.cui.Update(func(_ *gocui.Gui) error {
		if v, ok := g.views[viewName]; ok {
			v.RefreshView()
		}
		return nil
	})
}

func (g *Controller) cleanupResults() {
	if v, err := g.cui.View("results"); err == nil {
		v.Clear()
		v.Subtitle = ""
	}
	g.cui.DeleteView("filter")
}
