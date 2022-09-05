package vulnerabilities

import (
	"github.com/awesome-gocui/gocui"
	"github.com/liamg/tml"
	"github.com/owenrumney/lazytrivy/pkg/widgets"
)

var helpCommands = []string{
	"",
	tml.Sprintf(" <blue>[s]</blue>can              Scan selected image"),
	tml.Sprintf(" scan <blue>[a]</blue>ll          Scan all images"),
	tml.Sprintf(" <blue>[r]</blue>emote            Scan remote image"),
	tml.Sprintf(" <blue>[i]</blue>mage refresh     Refresh image list"),
	"",
}

func help(gui *gocui.Gui, _ *gocui.View) error {

	w, h := gui.Size()

	v := widgets.NewHelpWidget("help", w/2-22, h/2-4, w/2+22, h/2+4, helpCommands)

	if err := gui.SetKeybinding("help", gocui.KeyEsc, gocui.ModNone, func(gui *gocui.Gui, _ *gocui.View) error {
		if _, err := gui.SetCurrentView("images"); err != nil {
			return err
		}
		return gui.DeleteView("help")
	}); err != nil {
		return err
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
