package base

import (
	"github.com/awesome-gocui/gocui"
	"github.com/owenrumney/lazytrivy/pkg/widgets"
)

func (c *Controller) showSettings(_ *gocui.Gui, _ *gocui.View) error {

	w, h := c.Cui.Size()

	settings := widgets.NewSettingsWidget("settings", w/2-30, h/2-15, 60, 20, c.Cui, c.Config, func(gui *gocui.Gui, view *gocui.View) error {
		_, err := gui.SetCurrentView(widgets.Results)
		return err
	})

	settings.Draw()

	return nil
}
