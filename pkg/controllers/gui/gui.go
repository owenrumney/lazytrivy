package gui

import (
	"errors"
	"fmt"
	"sync"

	"github.com/awesome-gocui/gocui"

	"github.com/owenrumney/lazytrivy/pkg/config"
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
	activeController base.ControllerView
	// runContext       context.Context
	// runCancel        context.CancelFunc
	views  []gocui.Manager
	config *config.Config
}

func New() (*Controller, error) {
	cui, err := gocui.NewGui(gocui.OutputNormal, true)
	if err != nil {
		return nil, fmt.Errorf("failed to create gui: %w", err)
	}

	dockerClient := docker.NewClient()

	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	mainController := &Controller{
		cui:          cui,
		dockerClient: dockerClient,

		config: cfg,
	}
	mainController.activeController = vulnerabilities.NewVulnerabilityController(cui, dockerClient, cfg)

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
			g.activeController = aws.NewAWSController(g.cui, g.dockerClient, g.config)
		case widgets.AWSTab:
			g.activeController = vulnerabilities.NewVulnerabilityController(g.cui, g.dockerClient, g.config)
		}

		if err := g.CreateWidgets(); err != nil {
			return err
		}

		if err := g.activeController.Initialise(); err != nil {
			return err
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
	g.views = append(g.views, w...)
}

func (g *Controller) RefreshManager() {

	views := make([]gocui.Manager, 0, len(g.views)+1)
	views = append(views, g.views...)

	g.cui.SetManager(views...)

}
