package gui

import (
	"fmt"

	"github.com/awesome-gocui/gocui"
	"github.com/owenrumney/lazytrivy/pkg/docker"
	"github.com/owenrumney/lazytrivy/pkg/output"
	"github.com/owenrumney/lazytrivy/pkg/widgets"
)

type widget interface {
	ConfigureKeys() error
	Layout(*gocui.Gui) error
}

type Gui struct {
	cui           *gocui.Gui
	dockerClient  *docker.DockerClient
	views         map[string]widget
	selectedImage string
}

func New() (*Gui, error) {
	cui, err := gocui.NewGui(gocui.OutputNormal, true)
	if err != nil {
		return nil, err
	}
	return &Gui{
		cui:          cui,
		dockerClient: docker.NewDockerClient(),
		views:        make(map[string]widget),
	}, nil
}

func (g *Gui) DockerClient() *docker.DockerClient {
	return g.dockerClient
}

func (g *Gui) CreateWidgets() error {
	maxX, maxY := g.cui.Size()

	g.views["images"] = widgets.NewImagesWidget("images", g)
	g.views["results"] = widgets.NewInfoWidget("results", g)
	g.views["menu"] = widgets.NewMenuWidget("menu", 0, maxY-3, maxX-1, maxY-1, g)

	g.SetManager()
	g.EnableMouse()

	return nil
}

func (g *Gui) SetManager() {
	views := make([]gocui.Manager, 0, len(g.views)+1)
	for _, v := range g.views {
		views = append(views, v)
	}

	views = append(views, gocui.ManagerFunc(flowLayout))
	g.cui.SetManager(views...)
}

func (g *Gui) Run() error {
	if err := g.cui.MainLoop(); err != nil && err != gocui.ErrQuit {
		return err
	}
	return nil
}

func (g *Gui) Initialise() {
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

func (g *Gui) SetKeyBinding(viewName string, key interface{}, mod gocui.Modifier, handler func(*gocui.Gui, *gocui.View) error) error {
	return g.cui.SetKeybinding(viewName, key, mod, handler)
}

func (g *Gui) Close() {
	g.cui.Close()
}

func (g *Gui) ShowCursor() {
	g.cui.Cursor = true
}

func (g *Gui) HideCursor() {
	g.cui.Cursor = false
}

func (g *Gui) EnableMouse() {
	g.cui.Mouse = true
}

func (g *Gui) DisableMouse() {
	g.cui.Mouse = false
}

func (g *Gui) SetSelectedImage(selected string) {
	g.selectedImage = selected
}

func (g *Gui) ScanImage(imageName string) {

	go func() {
		report := g.dockerClient.ScanImage(imageName)
		g.cui.Update(func(gocui *gocui.Gui) error {
			return g.renderResultsReport(imageName, report)
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

func (g *Gui) scanRemote(gui *gocui.Gui, _ *gocui.View) error {

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

func (g *Gui) renderResultsReport(imageName string, report output.Report) error {
	if v, ok := g.views["results"]; ok {
		v.(*widgets.InfoWidget).RenderReport(report, fmt.Sprintf(" %s ", imageName))
		_, err := g.cui.SetCurrentView("results")
		if err != nil {
			return err
		}
	}

	return nil
}

func (g *Gui) RefreshImages() error {
	images := g.dockerClient.ListImages()

	return g.views["images"].(*widgets.ImagesWidget).RefreshImages(images)
}
