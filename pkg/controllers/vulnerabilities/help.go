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

	v := widgets.NewAnnouncementWidget("help", "Help", helpCommands, gui)

	gui.Update(func(g *gocui.Gui) error {
		if err := v.Layout(g); err != nil {
			return err
		}
		_, err := g.SetCurrentView("help")
		return err
	})

	return nil
}
