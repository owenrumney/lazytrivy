package gui

import (
	"fmt"

	"github.com/awesome-gocui/gocui"
	"github.com/owenrumney/lazytrivy/pkg/docker"
	"github.com/owenrumney/lazytrivy/pkg/widgets"
)

type Gui struct {
	cui          *gocui.Gui
	dockerClient *docker.DockerClient
	images       *widgets.ImagesWidget
	results      *widgets.InfoWidget
	menu         *widgets.MenuWidget
}

func New() (*Gui, error) {
	cui, err := gocui.NewGui(gocui.OutputNormal, false)
	if err != nil {
		return nil, err
	}
	return &Gui{
		cui:          cui,
		dockerClient: docker.NewDockerClient(),
	}, nil
}

func (g *Gui) CreateWidgets() error {
	maxX, maxY := g.cui.Size()

	g.images = widgets.NewImagesWidget("images", g.dockerClient, 1, 2, 0, maxY-3)
	g.results = widgets.NewInfoWidget("main", g.dockerClient, 0, 2, maxX-1, maxY-3, "")
	g.menu = widgets.NewMenuWidget("menu", 0, maxY-3, maxX-1, maxY-1, []string{
		"[s]scan", "[r]emote", "[i]mage refresh", "[q]uit",
	})

	fl := gocui.ManagerFunc(flowLayout)
	g.cui.SetManager(g.images, g.results, g.menu, fl)

	return nil
}

func (g *Gui) Run() error {
	if err := g.cui.MainLoop(); err != nil && err != gocui.ErrQuit {
		return err
	}
	return nil
}

func (g *Gui) Initialise() {
	g.cui.Update(func(gui *gocui.Gui) error {
		if err := g.configureKeyBindings(); err != nil {
			return err
		}
		_, err := gui.SetCurrentView("images")
		return err
	})
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

func (g *Gui) ShowMouse() {
	g.cui.Mouse = true
}

func (g *Gui) HideMouse() {
	g.cui.Mouse = false
}

func (g *Gui) SelectedImage() string {
	return g.images.SelectedImage()
}

func (g *Gui) ScanImage(imageName string) {
	g.results.Reset()

	go func() {
		report := g.dockerClient.ScanImage(imageName)
		g.cui.Update(func(gocui *gocui.Gui) error {
			g.results.Reset()
			g.results.SetSubTitle(fmt.Sprintf(" %s ", imageName))
			g.results.RenderReport(gocui, report)
			return nil
		})
	}()

}

func flowLayout(g *gocui.Gui) error {
	views := g.Views()
	maxX, _ := g.Size()
	x := 0
	for _, v := range views {
		w, h := v.Size()
		nextW := w

		switch v.Name() {
		case "scanning", "menu", "remote":
			continue
		case "main":
			nextW = maxX - 1
		}

		_, err := g.SetView(v.Name(), x, 1, nextW, h+1, 0)
		if err != nil && err != gocui.ErrUnknownView {
			return err
		}
		x += nextW + 2
	}
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func setView(g *gocui.Gui, v *gocui.View) error {
	_, err := g.SetCurrentView(v.Name())
	return err

}

func (g *Gui) scanRemote(gui *gocui.Gui, view *gocui.View) error {

	maxX, maxY := gui.Size()

	x1 := maxX/2 - 50
	x2 := maxX/2 + 50
	y1 := maxY/2 - 1
	// y2 := (maxY/2 + 1)

	gui.Cursor = true
	remote := widgets.NewInput("remote", x1, y1, x2, 100)
	gui.Update(func(g *gocui.Gui) error {
		gui.Mouse = false
		remote.Layout(gui)
		_, err := gui.SetCurrentView("remote")
		return err
	})

	if err := gui.SetKeybinding("remote", gocui.KeyEnter, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		if len(view.BufferLines()) > 0 {
			if image, _ := view.Line(0); image != "" {
				g.ScanImage(image)
			}
		}
		gui.Mouse = true
		gui.Cursor = false
		return gui.DeleteView("remote")
	}); err != nil {
		return err
	}

	return gui.SetKeybinding("remote", gocui.KeyEsc, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		gui.Mouse = true
		gui.Cursor = false
		return gui.DeleteView("remote")
	})
}
