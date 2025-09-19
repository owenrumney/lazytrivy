package widgets

import "github.com/awesome-gocui/gocui"

const (
	Announcement = "announcement"
	Files        = "files"
	Filter       = "filter"
	Host         = "host"
	Images       = "images"
	K8sContext   = "k8scontext"
	K8sTree      = "k8stree"
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
	ImagesTab     Tab = "Image"
	FileSystemTab Tab = "FileSystem"
	K8sTab        Tab = "Kubernetes"
)
