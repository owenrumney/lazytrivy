package controller

import (
	"fmt"

	"github.com/awesome-gocui/gocui"
	"github.com/owenrumney/lazytrivy/pkg/docker"
	"github.com/owenrumney/lazytrivy/pkg/output"
	"github.com/owenrumney/lazytrivy/pkg/widgets"
)

type Controller struct {
	cui           *gocui.Gui
	dockerClient  *docker.DockerClient
	views         map[string]widgets.Widget
	selectedImage string
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

func (g *Controller) SetSelectedImage(selected string) {
	g.selectedImage = selected
}

func (g *Controller) ScanImage(imageName string) {

	go func() {
		report, err := g.dockerClient.ScanImage(imageName)
		if err != nil {
			return //TODO: Need to do something here
		}
		g.cui.Update(func(gocui *gocui.Gui) error {
			return g.renderResultsReport(imageName, *report)
		})
	}()

}

func flowLayout(g *gocui.Gui) error {
	views := g.Views()
	maxX, maxY := g.Size()
	x := 0
	for _, v := range views {
		w, _ := v.Size()
		nextW := w

		switch v.Name() {
		case "scanning", "menu", "remote":
			continue
		case "results":
			nextW = maxX - 1
		}

		_, err := g.SetView(v.Name(), x, 1, nextW, maxY-3, 0)
		if err != nil && err != gocui.ErrUnknownView {
			return err
		}
		x += nextW + 2
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

func (g *Controller) renderResultsReport(imageName string, report output.Report) error {
	if v, ok := g.views["results"]; ok {
		v.(*widgets.InfoWidget).RenderReport(report, fmt.Sprintf(" %s ", imageName))
		_, err := g.cui.SetCurrentView("results")
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *Controller) RefreshImages() error {
	images := g.dockerClient.ListImages()

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
