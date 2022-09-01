package widgets

import (
	"errors"
	"fmt"

	"github.com/awesome-gocui/gocui"
)

type AddAccountWidget struct {
	name         string
	x, y         int
	w            int
	maxLength    int
	foundAccount string
}

func NewAddAccountWidget(name string, maxX, maxY, maxLength int, foundAccount string, ctx awsContext) (*AddAccountWidget, error) {
	x1 := maxX/2 - 50
	x2 := maxX/2 + 50
	y1 := maxY/2 - 1

	if err := ctx.SetKeyBinding(NewAccount, gocui.KeyEnter, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		if len(view.BufferLines()) > 0 {
			if account, _ := view.Line(0); account != "" {
				if err := ctx.UpdateAccount(account); err != nil {
					return err
				}
			}
		}
		gui.Mouse = true
		gui.Cursor = false

		if err := gui.DeleteView(NewAccount); err != nil {
			return fmt.Errorf("failed to delete view 'remote': %w", err)
		}
		if _, err := gui.SetCurrentView(Results); err != nil {
			return fmt.Errorf("failed to switch view to 'results': %w", err)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	if err := ctx.SetKeyBinding(NewAccount, gocui.KeyEsc, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		gui.Mouse = true
		gui.Cursor = false
		if _, err := gui.SetCurrentView(Images); err != nil {
			return err
		}
		return gui.DeleteView(NewAccount)
	}); err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return &AddAccountWidget{name: name, x: x1, y: y1, w: x2, maxLength: maxLength, foundAccount: foundAccount}, nil
}

func (i *AddAccountWidget) Layout(g *gocui.Gui) error {
	v, err := g.SetView(i.name, i.x, i.y, i.w, i.y+2, 0)
	if err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return fmt.Errorf("failed to create view: %w", err)
		}
		v.Title = " Enter Account Number "
		v.Editor = i
		v.Editable = true
		g.Cursor = true
		v.TitleColor = gocui.ColorGreen
		v.FrameColor = gocui.ColorGreen
		_, _ = fmt.Fprintf(v, i.foundAccount)
	}
	return nil
}

func (i *AddAccountWidget) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	cx, _ := v.Cursor()
	ox, _ := v.Origin()
	limit := ox+cx+1 > i.maxLength
	switch {
	case ch != 0 && mod == 0 && !limit:
		v.EditWrite(ch)
	case key == gocui.KeySpace && !limit:
		v.EditWrite(' ')
	case key == gocui.KeyBackspace || key == gocui.KeyBackspace2:
		v.EditDelete(true)
	}
}
