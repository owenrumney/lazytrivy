package widgets

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/awesome-gocui/gocui"
)

type ImagesWidget struct {
	ListWidget
	name string
	x, y int
	w, h int

	imageCount int
	ctx        vulnerabilityContext
	v          *gocui.View
}

func NewImagesWidget(name string, g vulnerabilityContext) *ImagesWidget {
	w := 25

	widget := &ImagesWidget{
		ListWidget: ListWidget{
			ctx:                 g,
			selectionChangeFunc: g.SetSelected,
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

func (w *ImagesWidget) ConfigureKeys() error {
	if err := w.ctx.SetKeyBinding(w.name, gocui.KeyArrowUp, gocui.ModNone, w.previousItem); err != nil {
		return fmt.Errorf("failed to set the previous image %w", err)
	}

	if err := w.ctx.SetKeyBinding(w.name, gocui.KeyArrowDown, gocui.ModNone, w.nextItem); err != nil {
		return fmt.Errorf("failed to set the next image %w", err)
	}

	if err := w.ctx.SetKeyBinding(w.name, gocui.KeyEnter, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		if image := w.SelectedImage(); image != "" {
			w.ctx.ScanImage(context.Background(), image)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("error setting keybinding for scanning image: %w", err)
	}

	if err := w.ctx.SetKeyBinding(w.name, 's', gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		if image := w.SelectedImage(); image != "" {
			w.ctx.ScanImage(context.Background(), image)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

func (w *ImagesWidget) Layout(g *gocui.Gui) error {
	v, err := g.SetView(w.name, w.x, w.y, w.w, w.h, 0)
	if err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return fmt.Errorf("%w", err)
		}
		w.v = v
		w.RefreshView()
	}
	v.Title = " Images "
	v.Highlight = true
	v.Autoscroll = true
	v.Highlight = true
	v.SelBgColor = gocui.ColorGreen | gocui.AttrDim
	v.SelFgColor = gocui.ColorBlack | gocui.AttrBold
	if g.CurrentView() == v {
		v.FrameColor = gocui.ColorGreen
	} else {
		v.FrameColor = gocui.ColorDefault
	}

	return nil
}

func (w *ImagesWidget) RefreshImages(images []string, imageWidth int) error {
	w.w = imageWidth + 4

	imageList := make([]string, len(images))
	for i, image := range images {
		imageList[i] = fmt.Sprintf("**%d*** % -*s", i, imageWidth+1, image)
	}

	w.bottomMost = len(imageList)
	w.body = imageList
	w.RefreshView()
	_ = w.v.SetCursor(0, 0)
	return nil
}

func (w *ImagesWidget) SetSelectedImage(image string) error {
	for i, line := range w.body {
		line = strings.TrimPrefix(line, "** ")
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

func (w *ImagesWidget) SelectedImage() string {
	_, y := w.v.Cursor()
	if image, err := w.v.Line(y); err == nil {
		return stripIdentifierPrefix(image)
	}
	return ""
}

func (w *ImagesWidget) RefreshView() {
	w.v.Clear()
	for _, line := range w.body {
		_, _ = fmt.Fprintln(w.v, stripIdentifierPrefix(line))
	}
}
