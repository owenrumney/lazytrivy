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
}

func NewMenuWidget(name string, x, y, w, h int, commands []string) *MenuWidget {

	return &MenuWidget{name: name, x: x, y: y, w: w, h: h, body: commands}
}

func (w *MenuWidget) ViewName() string {
	return w.v.Name()
}

func (w *MenuWidget) Layout(g *gocui.Gui) error {

	v, err := g.SetView(w.name, w.x, w.y, w.w, w.h, 0)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		_, _ = fmt.Fprint(v, fmt.Sprintf("Help: %s", strings.Join(w.body, " | ")))
	}
	v.Frame = false
	w.v = v
	return nil
}
