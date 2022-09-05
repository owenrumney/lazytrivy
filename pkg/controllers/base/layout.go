package base

import (
	"github.com/awesome-gocui/gocui"
)

func SetView(g *gocui.Gui, v *gocui.View) error {
	_, err := g.SetCurrentView(v.Name())
	return err
}
