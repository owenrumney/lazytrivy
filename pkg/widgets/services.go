package widgets

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/awesome-gocui/gocui"
)

type ServicesWidget struct {
	ListWidget
	name string
	x, y int
	w, h int
	body string

	ctx awsContext
	v   *gocui.View
}

func NewServicesWidget(name string, g awsContext) *ServicesWidget {
	w := 28

	widget := &ServicesWidget{
		ListWidget: ListWidget{
			ctx: g,
		},
		name: name,
		x:    0,
		y:    0,
		w:    w,
		h:    1,
		ctx:  g,
	}

	return widget
}

func (w *ServicesWidget) ConfigureKeys() error {
	if err := w.ctx.SetKeyBinding(w.name, gocui.KeyArrowUp, gocui.ModNone, w.previousItem); err != nil {
		return fmt.Errorf("failed to set the previous image %w", err)
	}

	if err := w.ctx.SetKeyBinding(w.name, gocui.KeyArrowDown, gocui.ModNone, w.nextItem); err != nil {
		return fmt.Errorf("failed to set the next image %w", err)
	}

	if err := w.ctx.SetKeyBinding(w.name, gocui.KeyEnter, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		w.ctx.ScanService(context.Background(), w.SelectedService())
		return nil
	}); err != nil {
		return fmt.Errorf("error setting keybinding for scanning image: %w", err)
	}

	if err := w.ctx.SetKeyBinding(w.name, 's', gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		if image := w.SelectedService(); image != "" {
			w.ctx.ScanService(context.Background(), image)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

func (w *ServicesWidget) Layout(g *gocui.Gui) error {
	v, err := g.SetView(w.name, w.x, w.y, w.w, w.h, 0)
	if err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return fmt.Errorf("%w", err)
		}
		_, _ = fmt.Fprint(v, w.body)
		v.SetCursor(0, 0)
	}
	v.Title = " Services "
	v.Highlight = true
	v.SelBgColor = gocui.ColorGreen | gocui.AttrDim
	v.SelFgColor = gocui.ColorBlack | gocui.AttrBold
	if g.CurrentView() == v {
		v.FrameColor = gocui.ColorGreen
	} else {
		v.FrameColor = gocui.ColorDefault
	}

	w.v = v
	return nil
}

func (w *ServicesWidget) RefreshServices(services []string, serviceWidth int) error {
	// w.w = serviceWidth + 4

	serviceList := make([]string, len(services))
	for i, service := range services {
		serviceList[i] = fmt.Sprintf(" % -*s", serviceWidth+1, service)
	}

	w.body = strings.Join(serviceList, "\n")
	w.v.Clear()
	w.bottomMost = len(serviceList)
	_, _ = fmt.Fprintf(w.v, w.body)
	_ = w.v.SetCursor(0, 0)
	return nil
}

func (w *ServicesWidget) SetSelectedImage(image string) error {
	for i, line := range strings.Split(w.body, "\n") {
		if strings.TrimSpace(line) == image {
			y := i + 1
			if err := w.v.SetCursor(0, y); err != nil {
				return fmt.Errorf("%w", err)
			}
			break
		}
	}
	return nil
}

func (w *ServicesWidget) SelectedService() string {
	_, y := w.v.Cursor()
	if service, err := w.v.Line(y); err == nil {
		return strings.TrimSpace(service)
	}
	return ""
}

func (w *ServicesWidget) RefreshView() {
	w.v.Clear()
	_, _ = fmt.Fprintf(w.v, w.body)
}
