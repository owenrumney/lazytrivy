package widgets

import (
	"errors"
	"fmt"

	"github.com/awesome-gocui/gocui"
	"github.com/liamg/tml"
)

type Tab string

const (
	VulnerabilitiesTab Tab = "Vulnerabilities"
	AWSTab             Tab = "AWS"
)

type TabWidget struct {
	name      string
	x, y      int
	w, h      int
	body      []Tab
	v         *gocui.View
	ActiveTab Tab
}

func (w *TabWidget) RefreshView() {
	panic("unimplemented")
}

func NewTabWidget(name string, x, y, w, h int) *TabWidget {
	menuItems := []Tab{
		VulnerabilitiesTab, AWSTab,
	}

	return &TabWidget{
		name: name,
		x:    x,
		y:    y,
		w:    w,
		h:    h,
		body: menuItems,
		v:    nil,
	}
}

func (w *TabWidget) ConfigureKeys() error {
	// nothing to configure here
	return nil
}

func (w *TabWidget) SetActiveTab(tab Tab) {
	w.ActiveTab = tab
}

func (w *TabWidget) Layout(g *gocui.Gui) error {
	v, err := g.SetView(w.name, w.x, w.y, w.w, w.h, 0)
	if err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return fmt.Errorf("%w", err)
		}
	}
	v.Clear()

	for _, tab := range w.body {
		if tab == w.ActiveTab {
			_ = tml.Fprintf(v, " <blue>%s</blue>", tab)
		} else {
			_ = tml.Fprintf(v, " %s ", tab)
		}
	}

	w.v = v
	return nil
}
