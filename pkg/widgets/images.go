package widgets

import (
	"fmt"
	"strings"

	"github.com/awesome-gocui/gocui"
	"github.com/owenrumney/lazytrivy/pkg/docker"
)

type ImagesWidget struct {
	name         string
	x, y         int
	w, h         int
	body         string
	v            *gocui.View
	changed      bool
	imageCount   int
	dockerClient *docker.DockerClient
}

func NewImagesWidget(name string, cli *docker.DockerClient, x, y, w, h int) *ImagesWidget {

	images := cli.ListImages()

	for _, image := range images {
		if len(image) > w {
			w = len(image) + 4
		}
	}

	widget := &ImagesWidget{
		name:         name,
		x:            x,
		y:            y,
		w:            w,
		h:            h,
		body:         strings.Join(images, "\n"),
		imageCount:   len(images),
		dockerClient: cli,
	}

	return widget
}

func (w *ImagesWidget) ViewName() string {
	return w.v.Name()
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

	w.changed = true
	return nil
}

func (w *ImagesWidget) RefreshImages() error {
	images := w.dockerClient.ListImages()

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
