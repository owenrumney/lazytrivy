package widgets

import "github.com/awesome-gocui/gocui"

const (
	Filter   = "filter"
	Host     = "host"
	Images   = "images"
	Menu     = "menu"
	Remote   = "remote"
	Results  = "results"
	Status   = "status"
	Summary  = "summary"
	Services = "services"
	Account  = "account"
)

type Widget interface {
	ConfigureKeys() error
	Layout(*gocui.Gui) error
	RefreshView()
}

type Tab string

const (
	VulnerabilitiesTab Tab = "Vulnerabilities"
	AWSTab             Tab = "AWS"
)
