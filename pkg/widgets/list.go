package widgets

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/awesome-gocui/gocui"
)

type ListWidget struct {
	ctx                 baseContext
	topMost             int
	bottomMost          int
	currentPos          int
	body                []string
	selectionChangeFunc func(selection string)
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
	if w.currentPos <= w.topMost {
		return nil
	}
	v.MoveCursor(0, -1)

	_, y := v.Cursor()
	_, oy := v.Origin()
	if w.selectionChangeFunc != nil {
		if selected, err := v.Line(y + oy); err == nil {
			w.selectionChangeFunc(selected)
		}
	}
	w.currentPos = y + oy
	return nil
}

func (w *ListWidget) nextItem(_ *gocui.Gui, v *gocui.View) error {
	if w.currentPos >= w.bottomMost {
		return nil
	}
	v.MoveCursor(0, 1)

	_, oy := v.Origin()
	_, y := v.Cursor()
	if _, h := v.Size(); y == h {
		if err := v.SetOrigin(0, oy+1); err != nil {
			return err
		}
		v.MoveCursor(0, -1)
		y--
		oy++
	}

	if w.selectionChangeFunc != nil {
		if selected, err := v.Line(y + oy); err == nil {
			w.selectionChangeFunc(selected)
		}
	}
	w.currentPos = y + oy

	return nil
}

func (w *ListWidget) CurrentItemPosition() int {
	if len(w.body) == 0 {
		return -1
	}

	currentLine := w.body[w.currentPos]
	if strings.HasPrefix(currentLine, "**") {
		idString := strings.TrimPrefix(strings.Split(currentLine, "***")[0], "**")
		id, err := strconv.Atoi(idString)
		if err == nil {
			return id
		}
	}
	return -1
}

func (w *ListWidget) SetStartPosition(pos int) {
	w.currentPos = pos
	w.topMost = pos
}
