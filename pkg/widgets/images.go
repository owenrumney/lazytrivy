package widgets

import (
	"fmt"
	"strings"

	"github.com/awesome-gocui/gocui"
)

type ImagesWidget struct {
	name string
	x, y int
	w, h int
	body string

	changed    bool
	imageCount int
	ctx        ctx
	v          *gocui.View
}

func NewImagesWidget(name string, g ctx) *ImagesWidget {

	images := g.DockerClient().ListImages()
	w := 0

	for _, image := range images {
		if len(image) > w {
			w = len(image) + 4
		}
	}

	widget := &ImagesWidget{
		name:       name,
		x:          0,
		y:          0,
		w:          w,
		h:          1,
		body:       strings.Join(images, "\n"),
		imageCount: len(images),
		ctx:        g,
	}

	return widget
}

func (w *ImagesWidget) ConfigureKeys() error {

	if err := w.ctx.SetKeyBinding(w.name, gocui.KeyArrowUp, gocui.ModNone, w.PreviousImage); err != nil {
		return err
	}

	if err := w.ctx.SetKeyBinding(w.name, gocui.KeyArrowDown, gocui.ModNone, w.NextImage); err != nil {
		return err
	}

	if err := w.ctx.SetKeyBinding(w.name, 's', gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		if image := w.SelectedImage(); image != "" {
			w.ctx.ScanImage(image)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (w *ImagesWidget) Layout(g *gocui.Gui) error {
	v, err := g.SetView(w.name, w.x, w.y, w.w, w.h, 0)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		_, _ = fmt.Fprint(v, w.body)
	}
	v.Title = " Images "
	v.Highlight = true
	v.Autoscroll = true
	v.Highlight = false
	v.SelFgColor = gocui.ColorGreen
	if g.CurrentView() == v {
		v.FrameColor = gocui.ColorGreen
	} else {
		v.FrameColor = gocui.ColorDefault
	}

	if !w.changed {
		v.SetCursor(0, 0)
		v.SetHighlight(0, true)
	}

	w.v = v
	return nil
}

func (w *ImagesWidget) PreviousImage(_ *gocui.Gui, view *gocui.View) error {
	_, y := view.Cursor()

	if y > 0 {
		view.SetHighlight(y, false)
		view.SetHighlight(y-1, true)
		_ = view.SetCursor(0, y-1)
	}

	if image, err := w.v.Line(y); err == nil {
		w.ctx.SetSelectedImage(image)
	}

	w.changed = true
	return nil
}

func (w *ImagesWidget) NextImage(_ *gocui.Gui, view *gocui.View) error {
	_, y := view.Cursor()

	if y <= w.imageCount-1 {
		view.SetHighlight(y, false)
		view.SetHighlight(y+1, true)
		_ = view.SetCursor(0, y+1)
	}
	if image, err := w.v.Line(y); err == nil {
		w.ctx.SetSelectedImage(image)
	}

	w.changed = true
	return nil
}

func (w *ImagesWidget) RefreshImages(images []string) error {
	for _, image := range images {
		if len(image) > w.w {
			w.w = len(image) + 4
		}
	}
	w.body = strings.Join(images, "\n")
	return nil
}

func (w *ImagesWidget) SelectedImage() string {
	_, y := w.v.Cursor()
	if image, err := w.v.Line(y); err == nil {
		return image
	}
	return ""
}
