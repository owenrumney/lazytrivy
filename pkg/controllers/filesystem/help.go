package filesystem

import (
	"fmt"

	"github.com/awesome-gocui/gocui"
	"github.com/liamg/tml"
	"github.com/owenrumney/lazytrivy/pkg/widgets"
)

var helpCommands = []string{
	"",
	tml.Sprintf(" <blue>[s]</blue>can/update        Scan filesystem"),
	tml.Sprintf(" change <blue>[p]</blue>ath        Change target directory"),
	"",
}

func help(gui *gocui.Gui, _ *gocui.View) error {

	w, h := gui.Size()

	v := widgets.NewAnnouncementWidget("help", "Help", w, h, helpCommands, gui)

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
