package widgets

import (
	"fmt"
	"strings"

	"github.com/awesome-gocui/gocui"
)

type MenuWidget struct {
	name string
	x, y int
	w, h int
	body []string
	v    *gocui.View
	ctx  ctx
}

// RefreshView implements Widget
func (w *MenuWidget) RefreshView() {
	panic("unimplemented")
}

func NewMenuWidget(name string, x, y, w, h int, ctx ctx) *MenuWidget {

	menuItems := []string{
		"[s]scan", "[r]emote", "[i]mage refresh", "[q]uit",
	}

	return &MenuWidget{name: name, x: x, y: y, w: w, h: h, body: menuItems, ctx: ctx}
}

func (w *MenuWidget) ConfigureKeys() error {
	// nothing to configure here
	return nil
}

func (w *MenuWidget) Layout(g *gocui.Gui) error {

	v, err := g.SetView(w.name, w.x, w.y, w.w, w.h, 0)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		_, _ = fmt.Fprintf(v, "Help: %s", strings.Join(w.body, " | "))
	}
	v.Frame = false
	w.v = v
	return nil
}
