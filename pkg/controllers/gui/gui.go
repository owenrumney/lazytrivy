package gui

import (
	"errors"
	"fmt"
	"sync"

	"github.com/awesome-gocui/gocui"
	"github.com/owenrumney/lazytrivy/pkg/controllers/base"
	"github.com/owenrumney/lazytrivy/pkg/controllers/vulnerabilities"
	"github.com/owenrumney/lazytrivy/pkg/docker"
)

type Controller struct {
	sync.Mutex
	cui              *gocui.Gui
	dockerClient     *docker.Client
	controllers      map[string]base.ControllerView
	activeController base.ControllerView
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
		},
	}
	mainController.activeController = mainController.controllers["vulnerabilities"]

	return mainController, nil
}

func (g *Controller) DockerClient() *docker.Client {
	return g.dockerClient
}

func (g *Controller) CreateWidgets() error {
	return g.activeController.CreateWidgets()
}

func (g *Controller) Initialise() error {
	return g.activeController.Initialise()
}

func (g *Controller) Run() error {
	if err := g.cui.MainLoop(); err != nil && !errors.Is(err, gocui.ErrQuit) {
		return fmt.Errorf("error occurred during the run main loop: %w", err)
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
