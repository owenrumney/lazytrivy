package controller

import (
	"errors"
	"fmt"

	"github.com/awesome-gocui/gocui"
	"github.com/owenrumney/lazytrivy/pkg/widgets"
)

func (g *Controller) RefreshView(viewName string) {
	g.cui.Update(func(_ *gocui.Gui) error {
		if v, ok := g.views[viewName]; ok {
			v.RefreshView()
		}
		return nil
	})
}

func (g *Controller) RefreshWidget(widget widgets.Widget) {
	g.cui.Update(func(gui *gocui.Gui) error {
		return widget.Layout(gui)
	})
}

func setView(g *gocui.Gui, v *gocui.View) error {
	_, err := g.SetCurrentView(v.Name())
	return err
}

func flowLayout(g *gocui.Gui) error {
	imagesWidth := 0
	viewNames := []string{widgets.Images, widgets.Host, widgets.Results, widgets.Menu, widgets.Status}
	maxX, maxY := g.Size()
	x := 0
	for _, viewName := range viewNames {
		v, err := g.View(viewName)
		if err != nil {
			return fmt.Errorf("failed to get view %s: %w", viewName, err)
		}
		w, _ := v.Size()
		h := 1
		nextW := w
		nextH := maxY - 4
		nextX := x

		switch v.Name() {
		case widgets.Host:
			nextW = imagesWidth
			nextX = 0
			nextH = 3
		case widgets.Images:
			imagesWidth = w
			h = 4
		case widgets.Status:
			nextW = maxX - 1
			h = maxY - 6
		case widgets.Results:
			nextW = maxX - 1
			nextH = maxY - 7
		case widgets.Menu:
			nextX = 0
			h = maxY - 4
			nextH = maxY
		case widgets.Remote, widgets.Filter:
			continue
		}

		_, err = g.SetView(v.Name(), nextX, h, nextW, nextH, 0)
		if err != nil && errors.Is(err, gocui.ErrUnknownView) {
			return fmt.Errorf("%w", err)
		}
		if v.Name() == widgets.Images {
			x += nextW + 1
		}
	}
	return nil
}

func (g *Controller) cleanupResults() {
	if v, err := g.cui.View(widgets.Results); err == nil {
		v.Clear()
		v.Subtitle = ""
	}
	_ = g.cui.DeleteView(widgets.Filter)
}
