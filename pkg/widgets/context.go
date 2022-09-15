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
	ScanImage(ctx context.Context)
}

type awsContext interface {
	baseContext
	ScanService(ctx context.Context)
	UpdateAccount(account string) error
	UpdateRegion(region string) error
}

type fsContext interface {
	baseContext
	ShowTarget(ctx context.Context)
	ScanVulnerabilities(gui *gocui.Gui, view *gocui.View) error
	SetWorkingDirectory(dir string)
}
