package widgets

import "github.com/awesome-gocui/gocui"

type Widget interface {
	ConfigureKeys() error
	Layout(*gocui.Gui) error
	RefreshView()
}
