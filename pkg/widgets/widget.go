package widgets

import "github.com/awesome-gocui/gocui"

const (
	Filter     = "filter"
	Host       = "host"
	Images     = "images"
	Menu       = "menu"
	Remote     = "remote"
	NewAccount = "new_account"
	Results    = "results"
	Status     = "status"
	Summary    = "summary"
	Services   = "services"
	Account    = "account"
)

type Widget interface {
	ConfigureKeys() error
	Layout(*gocui.Gui) error
	RefreshView()
}
