package widgets

import (
	"fmt"

	"github.com/awesome-gocui/gocui"
)

type StatusWidget struct {
	name string
	x, y int
	w, h int
	body string
	v    *gocui.View
	ctx  ctx
}

func NewStatusWidget(name string, ctx ctx) *StatusWidget {

	return &StatusWidget{
		name: name,
		x:    0,
		y:    0,
		w:    5,
		h:    1,
		body: "",
		ctx:  ctx,
	}
}

func (w *StatusWidget) ConfigureKeys() error {
	// nothing to configure here
	return nil
}

func (w *StatusWidget) Layout(g *gocui.Gui) error {

	v, err := g.SetView(w.name, w.x, w.y, w.w, w.h, 0)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		_, _ = fmt.Fprintf(v, w.body)
	}
	v.Title = " Status "
	w.v = v
	return nil
}

func (w *StatusWidget) UpdateStatus(status string) {
	w.body = status
	w.ctx.RefreshView(w.name)
}

func (w *StatusWidget) RefreshView() {
	w.v.Clear()
	_, _ = fmt.Fprintf(w.v, w.body)
}
