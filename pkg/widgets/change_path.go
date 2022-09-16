package widgets

import (
	"errors"
	"fmt"
	"os"

	"github.com/awesome-gocui/gocui"
)

type PathChangeWidget struct {
	name        string
	x, y        int
	w           int
	maxLength   int
	currentPath string
}

func NewPathChangeWidget(name string, maxX, maxY, maxLength int, currentPath string, ctx fsContext) (*PathChangeWidget, error) {
	x1 := maxX/2 - 50
	x2 := maxX/2 + 50
	y1 := maxY/2 - 1

	if err := ctx.SetKeyBinding(PathChange, gocui.KeyEnter, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		if len(view.BufferLines()) > 0 {
			if workingDirectory, _ := view.Line(0); workingDirectory != "" {
				if _, err := os.Stat(workingDirectory); err == nil {
					ctx.SetWorkingDirectory(workingDirectory)
				} else {
					if os.IsNotExist(err) {
						ctx.UpdateStatus(fmt.Sprintf("Nope, %s does not exist", workingDirectory))
					}
				}

			}
		}
		gui.Mouse = true
		gui.Cursor = false

		if err := gui.DeleteView(PathChange); err != nil {
			return fmt.Errorf("failed to delete view 'remote': %w", err)
		}
		if _, err := gui.SetCurrentView(Files); err != nil {
			return fmt.Errorf("failed to switch view to 'results': %w", err)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	if err := ctx.SetKeyBinding(PathChange, gocui.KeyEsc, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		gui.Mouse = true
		gui.Cursor = false
		if _, err := gui.SetCurrentView(Files); err != nil {
			return err
		}
		return gui.DeleteView(PathChange)
	}); err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return &PathChangeWidget{name: name, x: x1, y: y1, w: x2, maxLength: maxLength, currentPath: currentPath}, nil
}

func (w *PathChangeWidget) Layout(g *gocui.Gui) error {
	v, err := g.SetView(w.name, w.x, w.y, w.w, w.y+2, 0)
	if err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return fmt.Errorf("failed to create view: %w", err)
		}
		v.Title = " Enter path to scan "
		v.Editor = w
		v.Editable = true
		g.Cursor = true

		v.TitleColor = gocui.ColorGreen
		v.FrameColor = gocui.ColorGreen
		fmt.Fprintln(v, w.currentPath)

		v.SetCursor(len(w.currentPath), 0)
	}
	return nil
}

func (w *PathChangeWidget) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	cx, _ := v.Cursor()
	ox, _ := v.Origin()
	limit := ox+cx+1 > w.maxLength
	switch {
	case ch != 0 && mod == 0 && !limit:
		v.EditWrite(ch)
	case key == gocui.KeySpace && !limit:
		v.EditWrite(' ')
	case key == gocui.KeyBackspace || key == gocui.KeyBackspace2:
		v.EditDelete(true)
	}
}
