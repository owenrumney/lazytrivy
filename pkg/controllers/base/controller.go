package base

import (
	"context"
	"fmt"
	"sync"

	"github.com/awesome-gocui/gocui"
	"github.com/owenrumney/lazytrivy/pkg/config"
	"github.com/owenrumney/lazytrivy/pkg/engine"
	"github.com/owenrumney/lazytrivy/pkg/logger"
	"github.com/owenrumney/lazytrivy/pkg/widgets"
)

type Controller struct {
	sync.Mutex
	Cui        *gocui.Gui
	Engine     *engine.Client
	Views      map[string]widgets.Widget
	LayoutFunc func(g *gocui.Gui) error
	HelpFunc   func(g *gocui.Gui, v *gocui.View) error
	Tab        widgets.Tab
	Config     *config.Config

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
	if status != "" {
		logger.Debugf("%s", status)
	}

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

func (c *Controller) SetSelected(selected string) {
	// Base implementation - controllers can override this
	c.UpdateStatus(fmt.Sprintf("Selected: %s", selected))
}

func (c *Controller) showLogOverlay(g *gocui.Gui, v *gocui.View) error {
	// Capture the current view name to return focus later
	var previousView string
	if currentView := g.CurrentView(); currentView != nil {
		previousView = currentView.Name()
	}

	logOverlay := widgets.NewLogOverlayWidget("logOverlay", c)
	logOverlay.SetPreviousView(previousView)

	if err := logOverlay.Layout(g); err != nil {
		return fmt.Errorf("failed to layout log overlay: %w", err)
	}

	if err := logOverlay.ConfigureKeys(g); err != nil {
		return fmt.Errorf("failed to configure log overlay keys: %w", err)
	}

	_, err := g.SetCurrentView("logOverlay")
	if err != nil {
		return fmt.Errorf("failed to set current view to log overlay: %w", err)
	}

	return nil
}
