package widgets

import (
	"context"

	"github.com/awesome-gocui/gocui"
)

type baseContext interface {
	SetKeyBinding(viewName string, key interface{}, mod gocui.Modifier, handler func(*gocui.Gui, *gocui.View) error) error
	SetSelected(selected string)
	RefreshView(viewName string)
}

type vulnerabilityContext interface {
	baseContext
	ScanImage(ctx context.Context, imageName string)
}

type awsContext interface {
	baseContext
	ScanService(ctx context.Context, serviceName string)
	UpdateAccount(account string) error
	UpdateRegion(region string) error
}

type fsContext interface {
	baseContext
	ShowTarget(ctx context.Context, target string)
	ScanVulnerabilities(gui *gocui.Gui, view *gocui.View) error
	SetWorkingDirectory(dir string)
}
