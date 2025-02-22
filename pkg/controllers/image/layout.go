package image

import (
	"errors"
	"fmt"

	"github.com/awesome-gocui/gocui"
	"github.com/owenrumney/lazytrivy/pkg/widgets"
)

func vulnerabilityLayout(g *gocui.Gui) error {

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
		y := 0
		nextW := w
		nextH := maxY - 4
		nextX := x

		switch v.Name() {
		case widgets.Host:
			nextW = imagesWidth
			nextW = maxX - 1
			nextX = 0
			nextH = 2
		case widgets.Images:
			imagesWidth = w
			y = 3
		case widgets.Status:
			nextW = maxX - 1
			y = maxY - 6
		case widgets.Results:
			nextW = maxX - 1
			nextH = maxY - 7
			y = 3
		case widgets.Menu:
			nextX = 0
			y = maxY - 4
			nextH = maxY
		case widgets.Remote, widgets.Filter:
			continue
		}

		_, err = g.SetView(v.Name(), nextX, y, nextW, nextH, 0)
		if err != nil && errors.Is(err, gocui.ErrUnknownView) {
			return fmt.Errorf("%w", err)
		}
		if v.Name() == widgets.Images {
			x += nextW + 1
		}
	}
	return nil
}
