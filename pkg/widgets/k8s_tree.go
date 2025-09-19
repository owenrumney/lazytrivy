package widgets

import (
	"context"
	"errors"
	"fmt"

	"github.com/awesome-gocui/gocui"
)

type K8sTreeWidget struct {
	ListWidget
	name string
	x, y int
	w, h int
	body []string

	ctx k8sContext
	v   *gocui.View
}

type k8sContext interface {
	baseContext
	SetSelected(string)
	ScanCluster(context.Context)
	NavigateBack()
}

func NewK8sTreeWidget(name string, g k8sContext) *K8sTreeWidget {
	w := 35

	widget := &K8sTreeWidget{
		ListWidget: ListWidget{
			ctx:                 g,
			selectionChangeFunc: func(string) {}, // Don't auto-select on navigation
		},
		name: name,
		x:    0,
		y:    0,
		w:    w,
		h:    1,
		ctx:  g,
		body: []string{" Press 's' to scan cluster "},
	}

	return widget
}

func (w *K8sTreeWidget) ConfigureKeys(*gocui.Gui) error {
	if err := w.ctx.SetKeyBinding(w.name, gocui.KeyArrowUp, gocui.ModNone, w.previousItem); err != nil {
		return fmt.Errorf("failed to set the previous item %w", err)
	}

	if err := w.ctx.SetKeyBinding(w.name, gocui.KeyArrowDown, gocui.ModNone, w.nextItem); err != nil {
		return fmt.Errorf("failed to set the next item %w", err)
	}

	if err := w.ctx.SetKeyBinding(w.name, gocui.KeyArrowLeft, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		w.ctx.NavigateBack()
		return nil
	}); err != nil {
		return fmt.Errorf("failed to set left arrow key binding: %w", err)
	}

	if err := w.ctx.SetKeyBinding(w.name, gocui.KeyEnter, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		if line, err := view.Line(w.currentPos); err == nil {
			w.ctx.SetSelected(line)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to set enter key binding: %w", err)
	}

	if err := w.ctx.SetKeyBinding(w.name, 's', gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		w.body = []string{" Scanning cluster... "}
		w.RefreshView()
		w.ctx.ScanCluster(context.Background())
		return nil
	}); err != nil {
		return fmt.Errorf("failed to set scan binding: %w", err)
	}

	return nil
}

func (w *K8sTreeWidget) Layout(g *gocui.Gui) error {
	v, err := g.SetView(w.name, w.x, w.y, w.w, w.h, 0)
	if err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return fmt.Errorf("%w", err)
		}
		w.v = v
		w.RefreshView()
	}
	v.Title = " K8s Resources "
	v.Highlight = true
	v.SelBgColor = gocui.ColorDefault
	v.SelFgColor = gocui.ColorBlue | gocui.AttrBold
	if g.CurrentView() == v {
		v.FrameColor = gocui.ColorBlue
	} else {
		v.FrameColor = gocui.ColorDefault
	}
	v.FrameRunes = []rune{'─', '│', '╭', '╮', '╰', '╯'}

	return nil
}

func (w *K8sTreeWidget) RefreshView() {
	w.v.Clear()
	for _, line := range w.body {
		_, _ = fmt.Fprintln(w.v, stripIdentifierPrefix(line))
	}
}

func (w *K8sTreeWidget) UpdateTree(items []string) {
	w.body = items
	w.bottomMost = len(items) - 1
	w.currentPos = 0 // Reset to first item
	w.topMost = 0    // Reset top position
	w.RefreshView()
}

func (w *K8sTreeWidget) SetTitle(title string) {
	if w.v != nil {
		w.v.Title = fmt.Sprintf(" %s ", title)
	}
}
