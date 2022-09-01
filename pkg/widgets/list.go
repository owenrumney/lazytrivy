package widgets

import (
	"fmt"

	"github.com/awesome-gocui/gocui"
)

type ListWidget struct {
	ctx        baseContext
	topMost    int
	bottomMost int
	currentPos int
}

func (w *ListWidget) configureListWidgetKeys(name string) error {
	if err := w.ctx.SetKeyBinding(name, gocui.KeyArrowDown, gocui.ModNone, w.nextItem); err != nil {
		return fmt.Errorf("failed to set keybinding: %w", err)
	}

	if err := w.ctx.SetKeyBinding(name, gocui.KeyArrowUp, gocui.ModNone, w.previousItem); err != nil {
		return fmt.Errorf("failed to set keybinding: %w", err)
	}
	return nil
}

func (w *ListWidget) previousItem(_ *gocui.Gui, v *gocui.View) error {
	if w.currentPos == w.topMost {
		return nil
	}
	v.MoveCursor(0, -1)

	_, y := v.Cursor()
	_, oy := v.Origin()
	if selected, err := v.Line(y + oy); err == nil {
		w.ctx.SetSelected(selected)
	}
	w.currentPos = y + oy
	return nil
}

func (w *ListWidget) nextItem(_ *gocui.Gui, v *gocui.View) error {
	if w.currentPos == w.bottomMost {
		return nil
	}
	v.MoveCursor(0, 1)

	_, h := v.Size()
	_, oy := v.Origin()
	_, y := v.Cursor()
	if y == h-1 {
		if err := v.SetOrigin(0, oy+1); err != nil {
			return err
		}
		v.MoveCursor(0, -1)
		y--
		oy++
	}

	if selected, err := v.Line(y + oy); err == nil {
		w.ctx.SetSelected(selected)
	}
	w.currentPos = y + oy

	return nil
}
