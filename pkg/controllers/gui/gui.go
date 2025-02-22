package gui

import (
	"errors"
	"fmt"
	"sync"

	"github.com/awesome-gocui/gocui"
	"github.com/owenrumney/lazytrivy/pkg/controllers/filesystem"
	"github.com/owenrumney/lazytrivy/pkg/logger"

	"github.com/owenrumney/lazytrivy/pkg/config"
	"github.com/owenrumney/lazytrivy/pkg/controllers/base"
	"github.com/owenrumney/lazytrivy/pkg/controllers/image"
	"github.com/owenrumney/lazytrivy/pkg/dockerClient"
	"github.com/owenrumney/lazytrivy/pkg/widgets"
)

type Controller struct {
	sync.Mutex
	cui              *gocui.Gui
	dockerClient     *dockerClient.Client
	activeController base.ControllerView
	views            []gocui.Manager
	config           *config.Config
}

func New(tab widgets.Tab, cfg *config.Config) (*Controller, error) {

	logger.Debugf("Creating GUI")
	cui, err := gocui.NewGui(gocui.OutputNormal, true)
	if err != nil {
		return nil, fmt.Errorf("failed to create gui: %w", err)
	}

	dkrClient, err := dockerClient.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	mainController := &Controller{
		cui:          cui,
		dockerClient: dkrClient,
		config:       cfg,
	}

	switch tab {
	case widgets.FileSystemTab:
		mainController.activeController = filesystem.NewFilesystemController(cui, dkrClient, cfg)
	default:
		mainController.activeController = image.NewVulnerabilityController(cui, dkrClient, cfg)

	}

	return mainController, nil
}

func (c *Controller) DockerClient() *dockerClient.Client {
	return c.dockerClient
}

func (c *Controller) CreateWidgets() error {
	return c.activeController.CreateWidgets(c)
}

func (c *Controller) Initialise() error {
	if c.config.Debug {
		logger.Configure()
	}
	if c.config.Trace {
		logger.EnableTracing()
	}

	if err := c.cui.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, base.Quit); err != nil {
		return err
	}

	if err := c.cui.SetKeybinding("", 'w', gocui.ModNone, c.switchMode); err != nil {
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

func (c *Controller) switchMode(gui *gocui.Gui, _ *gocui.View) error {

	choices := []string{
		"**2*** File System",
		"**2*** Image",
	}

	w, y := c.cui.Size()

	choiceWidget := widgets.NewChoiceWidget("mode", w, y, "Scanning Mode", choices, func(selectedMode string) error {
		switch selectedMode {
		case "File System":
			c.activeController = filesystem.NewFilesystemController(c.cui, c.dockerClient, c.config)
		default:
			c.activeController = image.NewVulnerabilityController(c.cui, c.dockerClient, c.config)
		}

		if err := c.CreateWidgets(); err != nil {
			return err
		}

		if err := c.Initialise(); err != nil {
			return err
		}
		return nil
	}, c)

	_ = choiceWidget.Layout(gui)
	_, err := gui.SetCurrentView("mode")
	return err

}

func (c *Controller) UpdateStatus(status string) {
	// TODO implement me
	panic("implement me")
}

func (c *Controller) SetSelected(_ string) {
	// TODO implement me
	panic("implement me")
}

func (c *Controller) RefreshView(_ string) {
	// TODO implement me
	panic("implement me")
}
