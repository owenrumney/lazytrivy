package widgets

import "github.com/awesome-gocui/gocui"

const (
	Account      = "account"
	Announcement = "announcement"
	Filter       = "filter"
	Host         = "host"
	Images       = "images"
	Menu         = "menu"
	Remote       = "remote"
	Results      = "results"
	Services     = "services"
	Status       = "status"
	Summary      = "summary"
	Files        = "files"
	ScanPath     = "scanpath"
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
