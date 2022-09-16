package widgets

import (
	"errors"
	"fmt"
	"strings"

	"github.com/awesome-gocui/gocui"
)

type AnnouncementWidget struct {
	name         string
	x, y         int
	w, h         int
	body         []string
	v            *gocui.View
	title        string
	ctx          *gocui.Gui
	nextViewName string
}

func (w *AnnouncementWidget) RefreshView() {
	panic("unimplemented")
}

func NewAnnouncementWidget(name, title string, width, height int, lines []string, ctx *gocui.Gui, nextView ...string) *AnnouncementWidget {
	nextViewName := Results
	if len(nextView) > 0 {
		nextViewName = nextView[0]
	}

	maxLength := 0

	for _, item := range lines {
		if len(item) > maxLength {
			maxLength = len(item)
		}
	}

	maxLength += 2
	maxHeight := len(lines) + 2

	x := width/2 - maxLength/2
	w := width/2 + maxLength/2

	y := height/2 - maxHeight/2
	h := height/2 + maxHeight/2

	return &AnnouncementWidget{
		name:         name,
		title:        title,
		x:            x,
		y:            y,
		w:            w,
		h:            h,
		body:         lines,
		v:            nil,
		nextViewName: nextViewName,
		ctx:          ctx,
	}
}

func (w *AnnouncementWidget) ConfigureKeys(*gocui.Gui) error {
	if err := w.ctx.SetKeybinding(w.name, gocui.KeyEsc, gocui.ModNone, w.close); err != nil {
		return err
	}
	if err := w.ctx.SetKeybinding(w.name, gocui.KeyEnter, gocui.ModNone, w.close); err != nil {
		return err
	}

	if err := w.ctx.SetKeybinding(w.name, 'q', gocui.ModNone, w.close); err != nil {
		return err
	}
	return nil
}

func (w *AnnouncementWidget) close(gui *gocui.Gui, _ *gocui.View) error {
	if _, err := gui.SetCurrentView(w.nextViewName); err != nil {
		return err
	}
	return gui.DeleteView(w.name)

}

func (w *AnnouncementWidget) Layout(g *gocui.Gui) error {
	v, err := g.SetView(w.name, w.x, w.y, w.w, w.h, 0)
	if err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return fmt.Errorf("%w", err)
		}
	}
	v.Clear()
	_, _ = fmt.Fprint(v, strings.Join(w.body, "\n"))
	v.Title = fmt.Sprintf(" %s ", w.title)
	v.Subtitle = " ESC to close "
	v.Wrap = true
	v.FrameColor = gocui.ColorGreen
	w.v = v

	return w.ConfigureKeys(g)
}
