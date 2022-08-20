package gui

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/awesome-gocui/gocui"
	"github.com/owenrumney/lazytrivy/pkg/controllers/aws"
	"github.com/owenrumney/lazytrivy/pkg/controllers/base"
	"github.com/owenrumney/lazytrivy/pkg/controllers/vulnerabilities"
	"github.com/owenrumney/lazytrivy/pkg/docker"
	"github.com/owenrumney/lazytrivy/pkg/widgets"
)

type Controller struct {
	sync.Mutex
	cui              *gocui.Gui
	dockerClient     *docker.Client
	controllers      map[string]base.ControllerView
	activeController base.ControllerView
	runContext       context.Context
	runCancel        context.CancelFunc
	views            []gocui.Manager
}

func New() (*Controller, error) {
	cui, err := gocui.NewGui(gocui.OutputNormal, true)
	if err != nil {
		return nil, fmt.Errorf("failed to create gui: %w", err)
	}

	dockerClient := docker.NewClient()

	mainController := &Controller{
		cui:          cui,
		dockerClient: dockerClient,
		controllers: map[string]base.ControllerView{
			"vulnerabilities": vulnerabilities.NewVulnerabilityController(cui, dockerClient),
			"aws":             aws.NewAWSController(cui, dockerClient),
		},
	}
	mainController.activeController = mainController.controllers["vulnerabilities"]

	return mainController, nil
}

func (g *Controller) DockerClient() *docker.Client {
	return g.dockerClient
}

func (g *Controller) CreateWidgets() error {
	// maxX, _ := g.cui.Size()
	// tabs := widgets.NewTabWidget("tabs", 0, 0, maxX-1, 2)
	// tabs.SetActiveTab(g.activeController.Tab())
	//
	// g.cui.Update(func(gui *gocui.Gui) error {
	// 	if err := tabs.Layout(gui); err != nil {
	// 		return fmt.Errorf("failed to layout remote input: %w", err)
	// 	}
	// 	_, err := gui.SetCurrentView("tabs")
	// 	return err
	// })

	return g.activeController.CreateWidgets(g)
}

func (g *Controller) Initialise() error {

	if err := g.cui.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, base.Quit); err != nil {
		return err
	}

	if err := g.cui.SetKeybinding("", gocui.KeyTab, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {

		switch g.activeController.Tab() {
		case widgets.VulnerabilitiesTab:
			g.activeController = aws.NewAWSController(g.cui, g.dockerClient)
		case widgets.AWSTab:
			g.activeController = vulnerabilities.NewVulnerabilityController(g.cui, g.dockerClient)
		}

		if err := g.CreateWidgets(); err != nil {
			return err
		}

		if err := g.activeController.Initialise(); err != nil {
		}

		g.Initialise()
		return nil
	}); err != nil {
		return fmt.Errorf("error while creating Tabs navigation key binding: %w", err)
	}

	return g.activeController.Initialise()
}

func (g *Controller) Run() error {
	for {
		if err := g.cui.MainLoop(); err != nil {
			if errors.Is(err, gocui.ErrQuit) {
				return nil
			}
		}
	}
	return nil
}

func (g *Controller) SetKeyBinding(viewName string, key interface{}, mod gocui.Modifier, handler func(*gocui.Gui, *gocui.View) error) error {
	if err := g.cui.SetKeybinding(viewName, key, mod, handler); err != nil {
		return fmt.Errorf("failed to set keybinding for %s: %w", key, err)
	}
	return nil
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

func (g *Controller) AddViews(w ...gocui.Manager) {
	for _, widget := range w {
		g.views = append(g.views, widget)
	}
}

func (g *Controller) RefreshManager() {

	views := make([]gocui.Manager, 0, len(g.views)+1)
	for _, v := range g.views {
		views = append(views, v)
	}

	g.cui.SetManager(views...)

}
