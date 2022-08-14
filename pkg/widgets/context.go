package widgets

import (
	"github.com/awesome-gocui/gocui"
	"github.com/owenrumney/lazytrivy/pkg/docker"
)

type ctx interface {
	ScanImage(imageName string)
	DockerClient() *docker.DockerClient
	SetKeyBinding(viewName string, key interface{}, mod gocui.Modifier, handler func(*gocui.Gui, *gocui.View) error) error
	SetSelectedImage(imageName string)
	RefreshView(viewName string)
}
