package widgets

import (
	"fmt"
	"os"

	"github.com/awesome-gocui/gocui"
)

type HostWidget struct {
	name string
	x, y int
	w, h int
	body string
	v    *gocui.View
	ctx  ctx
}

func NewHostWidget(name string, ctx ctx) *HostWidget {

	hostName, err := os.Hostname()
	if err != nil {
		hostName = "unknown"
	}
	return &HostWidget{
		name: name,
		x:    1,
		y:    0,
		w:    5,
		h:    1,
		body: hostName,
		ctx:  ctx,
	}
}

func (w *HostWidget) ConfigureKeys() error {
	// nothing to configure here
	return nil
}

func (w *HostWidget) Layout(g *gocui.Gui) error {

	v, err := g.SetView(w.name, w.x, w.y, w.w, w.h, 0)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		_, _ = fmt.Fprintf(v, w.body)
	}
	v.Title = " Host "
	w.v = v
	return nil
}

func (w *HostWidget) RefreshView() {
	w.v.Clear()
	_, _ = fmt.Fprintf(w.v, w.body)
}