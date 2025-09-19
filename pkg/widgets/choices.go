package widgets

import (
	"errors"
	"fmt"

	"github.com/awesome-gocui/gocui"
)

type ChoiceWidget struct {
	ListWidget
	name         string
	x, y         int
	w, h         int
	title        string
	body         []string
	ctx          baseContext
	v            *gocui.View
	updateAction func(string) error
}

func (w *ChoiceWidget) RefreshView() {
	panic("unimplemented")
}

func NewChoiceWidget(name string, width, height int, title string, choices []string, updateAction func(string) error, ctx baseContext) *ChoiceWidget {
	maxLength := 0

	for _, item := range choices {
		if len(item) > maxLength {
			maxLength = len(item)
		}
	}

	maxLength += 2
	maxHeight := len(choices) + 2

	x := width/2 - maxLength/2
	w := width/2 + maxLength/2

	y := height/2 - maxHeight/2
	h := height/2 + maxHeight/2

	return &ChoiceWidget{
		ListWidget: ListWidget{
			ctx:        ctx,
			bottomMost: len(choices) - 1,
		},
		name:         name,
		x:            x,
		y:            y,
		w:            w,
		h:            h,
		title:        title,
		body:         choices,
		ctx:          ctx,
		updateAction: updateAction,
		v:            nil,
	}
}

func (w *ChoiceWidget) ConfigureKeys(*gocui.Gui) error {
	if err := w.configureListWidgetKeys(w.name); err != nil {
		return err
	}

	if err := w.ctx.SetKeyBinding(w.name, gocui.KeyEsc, gocui.ModNone, exitModal); err != nil {
		return fmt.Errorf("error will setting choices key binding to escape: %w", err)
	}

	if err := w.ctx.SetKeyBinding(w.name, 'q', gocui.ModNone, exitModal); err != nil {
		return fmt.Errorf("error will setting choices key binding to q: %w", err)
	}

	if err := w.ctx.SetKeyBinding(w.name, gocui.KeyEnter, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		line := w.body[w.currentPos]
		line = stripIdentifierPrefix(line)
		if err := w.updateAction(line); err != nil {
			return err
		}

		return exitModal(gui, view)
	}); err != nil {
		return fmt.Errorf("error while setting key binding for enter: %w", err)
	}
	return nil
}

func (w *ChoiceWidget) Layout(g *gocui.Gui) error {
	v, err := g.SetView(w.name, w.x, w.y, w.w, w.h, 0)
	if err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return fmt.Errorf("%w", err)
		}
		v.Clear()
		for _, line := range w.body {
			line = stripIdentifierPrefix(line)
			_, _ = fmt.Fprintf(v, " %s \n", line)
		}

		w.SetStartPosition(0)
		w.v = v
		if err := w.ConfigureKeys(nil); err != nil {
			return err
		}

	}
	v.Title = fmt.Sprintf(" %s ", w.title)
	v.Highlight = true
	v.Autoscroll = false
	v.SelBgColor = gocui.ColorDefault
	v.SelFgColor = gocui.ColorBlue | gocui.AttrBold
	v.Wrap = true
	v.FrameColor = gocui.ColorBlue
	v.FrameRunes = []rune{'─', '│', '╭', '╮', '╰', '╯'}

	return nil
}
