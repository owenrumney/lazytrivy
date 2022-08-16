package widgets

import "github.com/awesome-gocui/gocui"

const (
	Host    = "host"
	Images  = "images"
	Menu    = "menu"
	Remote  = "remote"
	Results = "results"
	Status  = "status"
)

type Widget interface {
	ConfigureKeys() error
	Layout(*gocui.Gui) error
	RefreshView()
}
