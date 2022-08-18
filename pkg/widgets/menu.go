package widgets

import (
	"errors"
	"fmt"
	"strings"

	"github.com/awesome-gocui/gocui"
	"github.com/liamg/tml"
)

type MenuWidget struct {
	name string
	x, y int
	w, h int
	body []string
	v    *gocui.View
	ctx  ctx
}

func (w *MenuWidget) RefreshView() {
	panic("unimplemented")
}

func NewMenuWidget(name string, x, y, w, h int, ctx ctx) *MenuWidget {
	menuItems := []string{
		"<blue>[enter] or [s]</blue>can", "scan <blue>[a]</blue>ll", "<blue>[r]</blue>emote",
		"<green>[i]</green>mage refresh", "<red>[t]</red>erminate scan", "<red>[q]</red>uit",
		"\n\n<yellow>Navigation: Use arrow keys to navigate and ESC to exit screens</yellow>",
	}

	return &MenuWidget{
		name: name,
		x:    x,
		y:    y,
		w:    w,
		h:    h,
		body: menuItems,
		v:    nil,
		ctx:  ctx,
	}
}

func (w *MenuWidget) ConfigureKeys() error {
	// nothing to configure here
	return nil
}

func (w *MenuWidget) Layout(g *gocui.Gui) error {
	v, err := g.SetView(w.name, w.x, w.y, w.w, w.h, 0)
	if err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return fmt.Errorf("%w", err)
		}
	}
	v.Clear()
	_ = tml.Fprintf(v, strings.Join(w.body, " | "))
	v.Frame = false
	w.v = v
	return nil
}
