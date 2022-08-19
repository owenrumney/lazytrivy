package widgets

import (
	"context"

	"github.com/awesome-gocui/gocui"
)

type vulnerabilityContext interface {
	ScanImage(ctx context.Context, imageName string)
	SetKeyBinding(viewName string, key interface{}, mod gocui.Modifier, handler func(*gocui.Gui, *gocui.View) error) error
	SetSelectedImage(imageName string)
	RefreshView(viewName string)
	RefreshWidget(widget Widget)
}
