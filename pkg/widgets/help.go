package widgets

import (
	"errors"
	"fmt"
	"strings"

	"github.com/awesome-gocui/gocui"
)

type HelpWidget struct {
	name string
	x, y int
	w, h int
	body []string
	v    *gocui.View
}

func (w *HelpWidget) RefreshView() {
	panic("unimplemented")
}

func NewHelpWidget(name string, x, y, w, h int, helpItems []string) *HelpWidget {
	// TODO update to accept parent size and calculate own size based on helpItems

	return &HelpWidget{
		name: name,
		x:    x,
		y:    y,
		w:    w,
		h:    h,
		body: helpItems,
		v:    nil,
	}
}

func (w *HelpWidget) ConfigureKeys() error {
	// nothing to configure here
	return nil
}

func (w *HelpWidget) Layout(g *gocui.Gui) error {
	v, err := g.SetView(w.name, w.x, w.y, w.w, w.h, 0)
	if err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return fmt.Errorf("%w", err)
		}
	}
	v.Clear()
	_, _ = fmt.Fprint(v, strings.Join(w.body, "\n"))
	v.Title = " Help "
	v.Subtitle = " ESC to exit"
	v.Wrap = true
	v.FrameColor = gocui.ColorGreen
	w.v = v
	return nil
}
