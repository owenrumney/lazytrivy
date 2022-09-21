package aws

import (
	"fmt"

	"github.com/awesome-gocui/gocui"
	"github.com/liamg/tml"
	"github.com/owenrumney/lazytrivy/pkg/widgets"
)

var helpCommands = []string{
	"",
	tml.Sprintf(" <blue>[s]</blue>can/update        Scan/Rescan"),
	tml.Sprintf(" switch <blue>[a]</blue>ccount     Switch account"),
	tml.Sprintf(" switch <blue>[r]</blue>egion      Switch region"),
	"",
}

func help(gui *gocui.Gui, _ *gocui.View) error {
	v := widgets.NewAnnouncementWidget("help", "Help", helpCommands, gui)

	if err := gui.SetKeybinding("help", gocui.KeyEsc, gocui.ModNone, func(gui *gocui.Gui, _ *gocui.View) error {
		if _, err := gui.SetCurrentView("services"); err != nil {
			return err
		}
		return gui.DeleteView("help")
	}); err != nil {
		return fmt.Errorf("error setting keybinding for help: %w", err)
	}

	gui.Update(func(g *gocui.Gui) error {
		if err := v.Layout(g); err != nil {
			return err
		}
		_, err := g.SetCurrentView("help")
		return err
	})

	return nil
}
