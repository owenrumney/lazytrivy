package base

import (
	"context"
	"fmt"
	"sync"

	"github.com/awesome-gocui/gocui"
	"github.com/owenrumney/lazytrivy/pkg/config"
	"github.com/owenrumney/lazytrivy/pkg/dockerClient"
	"github.com/owenrumney/lazytrivy/pkg/widgets"
)

type Controller struct {
	sync.Mutex
	Cui          *gocui.Gui
	DockerClient *dockerClient.Client
	Views        map[string]widgets.Widget
	LayoutFunc   func(g *gocui.Gui) error
	HelpFunc     func(g *gocui.Gui, v *gocui.View) error
	Tab          widgets.Tab
	Config       *config.Config

	ActiveCancel context.CancelFunc
}

type ControllerView interface {
	CreateWidgets(manager Manager) error
	Initialise() error
	Tab() widgets.Tab
}

func (c *Controller) SetManager() {
	views := make([]gocui.Manager, 0, len(c.Views)+1)
	for _, v := range c.Views {
		views = append(views, v)
	}

	views = append(views, gocui.ManagerFunc(c.LayoutFunc))
	c.Cui.SetManager(views...)
}

func (c *Controller) RefreshView(viewName string) {
	c.Cui.Update(func(_ *gocui.Gui) error {
		if v, ok := c.Views[viewName]; ok {
			v.RefreshView()
		}
		return nil
	})
}

func (c *Controller) UpdateStatus(status string) {
	if v, ok := c.Views[widgets.Status].(*widgets.StatusWidget); ok {
		v.UpdateStatus(status)
		c.Cui.Update(func(_ *gocui.Gui) error {
			v.RefreshView()
			return nil
		})
	}
}

func (c *Controller) ClearStatus() {
	c.UpdateStatus("")
}

func Quit(_ *gocui.Gui, _ *gocui.View) error {
	return gocui.ErrQuit
}

func (c *Controller) SetKeyBinding(viewName string, key interface{}, mod gocui.Modifier, handler func(*gocui.Gui, *gocui.View) error) error {
	if err := c.Cui.SetKeybinding(viewName, key, mod, handler); err != nil {
		return fmt.Errorf("failed to set keybinding for %s: %w", key, err)
	}
	return nil
}
