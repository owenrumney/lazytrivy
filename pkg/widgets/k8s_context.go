package widgets

import (
	"errors"
	"fmt"

	"github.com/awesome-gocui/gocui"
	"github.com/liamg/tml"
)

type K8sContextWidget struct {
	name           string
	x, y           int
	w, h           int
	currentContext string
	v              *gocui.View
	ctx            baseContext
}

func NewK8sContextWidget(name string, currentContext string, ctx baseContext) *K8sContextWidget {
	return &K8sContextWidget{
		name:           name,
		x:              1,
		y:              0,
		w:              5,
		h:              1,
		currentContext: currentContext,
		ctx:            ctx,
	}
}

func (w *K8sContextWidget) ConfigureKeys(*gocui.Gui) error {
	// nothing to configure here
	return nil
}

func (w *K8sContextWidget) Layout(g *gocui.Gui) error {
	v, err := g.SetView(w.name, w.x, w.y, w.w, w.h, 0)
	if err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return fmt.Errorf("%w", err)
		}
		_ = tml.Fprintf(v, " <blue>K8s Context: %s</blue>", w.currentContext)
	}

	v.Title = " K8s Context (c to change) "
	v.FrameRunes = []rune{'─', '│', '╭', '╮', '╰', '╯'}
	w.v = v
	return nil
}

func (w *K8sContextWidget) RefreshView() {
	w.v.Clear()
	_, _ = fmt.Fprintf(w.v, "K8s Context: %s", w.currentContext)
}

func (w *K8sContextWidget) UpdateContext(context string) {
	w.currentContext = context
	w.RefreshView()
}
