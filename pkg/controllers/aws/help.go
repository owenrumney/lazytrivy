package aws

import (
	"github.com/awesome-gocui/gocui"
	"github.com/liamg/tml"
	"github.com/owenrumney/lazytrivy/pkg/widgets"
)

var helpCommands = []string{
	"",
	tml.Sprintf(" <blue>[u]</blue>pdate            Rescan account/region"),
	tml.Sprintf(" <blue>[n]</blue>ew entry         Add an account/region"),
	tml.Sprintf(" switch <blue>[a]</blue>ccount    Switch account"),
	tml.Sprintf(" switch <blue>[r]</blue>egion     Switch region"),
	"",
}

func help(gui *gocui.Gui, _ *gocui.View) error {

	w, h := gui.Size()

	v := widgets.NewHelpWidget("help", w/2-22, h/2-4, w/2+22, h/2+3, helpCommands)

	gui.SetKeybinding("help", gocui.KeyEsc, gocui.ModNone, func(gui *gocui.Gui, _ *gocui.View) error {
		gui.SetCurrentView("services")
		return gui.DeleteView("help")
	})

	gui.Update(func(g *gocui.Gui) error {
		v.Layout(g)
		_, err := g.SetCurrentView("help")
		return err
	})

	return nil
}
