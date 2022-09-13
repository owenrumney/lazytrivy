package widgets

import "github.com/awesome-gocui/gocui"

const (
	Account      = "account"
	Announcement = "announcement"
	Files        = "files"
	Filter       = "filter"
	Host         = "host"
	Images       = "images"
	Menu         = "menu"
	PathChange   = "pathchange"
	Remote       = "remote"
	Results      = "results"
	ScanPath     = "scanpath"
	Services     = "services"
	Status       = "status"
	Summary      = "summary"
)

type Widget interface {
	ConfigureKeys(*gocui.Gui) error
	Layout(*gocui.Gui) error
	RefreshView()
}

type Tab string

const (
	VulnerabilitiesTab Tab = "Vulnerabilities"
	AWSTab             Tab = "AWS"
	FileSystemTab      Tab = "FileSystem"
)
