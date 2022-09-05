package gui

import (
	"errors"
	"fmt"
	"sync"

	"github.com/awesome-gocui/gocui"
	"github.com/owenrumney/lazytrivy/pkg/logger"

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
	views            []gocui.Manager
	config           *config.Config
}

func New() (*Controller, error) {
	logger.Debug("Creating GUI")
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

func (c *Controller) DockerClient() *docker.Client {
	return c.dockerClient
}

func (c *Controller) CreateWidgets() error {
	return c.activeController.CreateWidgets(c)
}

func (c *Controller) Initialise() error {
	if c.config.Debug == true {
		logger.EnableDebugging()
	}

	if err := c.cui.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, base.Quit); err != nil {
		return err
	}

	if err := c.cui.SetKeybinding("", 'm', gocui.ModNone, c.switchMode); err != nil {
		return fmt.Errorf("error while creating Tabs navigation key binding: %w", err)
	}

	return c.activeController.Initialise()
}

func (c *Controller) Run() error {
	for {
		if err := c.cui.MainLoop(); err != nil {
			if errors.Is(err, gocui.ErrQuit) {
				return nil
			}
		}
	}
}

func (c *Controller) SetKeyBinding(viewName string, key interface{}, mod gocui.Modifier, handler func(*gocui.Gui, *gocui.View) error) error {
	if err := c.cui.SetKeybinding(viewName, key, mod, handler); err != nil {
		return fmt.Errorf("failed to set keybinding for %s: %w", key, err)
	}
	return nil
}

func (c *Controller) Close() {
	c.cui.Close()
}

func (c *Controller) ShowCursor() {
	c.cui.Cursor = true
}

func (c *Controller) HideCursor() {
	c.cui.Cursor = false
}

func (c *Controller) EnableMouse() {
	c.cui.Mouse = true
}

func (c *Controller) DisableMouse() {
	c.cui.Mouse = false
}

func (c *Controller) AddViews(w ...gocui.Manager) {
	c.views = append(c.views, w...)
}

func (c *Controller) RefreshManager() {

	views := make([]gocui.Manager, 0, len(c.views)+1)
	views = append(views, c.views...)

	c.cui.SetManager(views...)

}

func (c *Controller) switchMode(gui *gocui.Gui, _ *gocui.View) error {

	switch c.activeController.Tab() {
	case widgets.VulnerabilitiesTab:
		c.activeController = aws.NewAWSController(c.cui, c.dockerClient, c.config)
	case widgets.AWSTab:
		c.activeController = vulnerabilities.NewVulnerabilityController(c.cui, c.dockerClient, c.config)
	}

	if err := c.CreateWidgets(); err != nil {
		return err
	}

	return c.Initialise()
}
