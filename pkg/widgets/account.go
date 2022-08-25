package widgets

import (
	"errors"
	"fmt"

	"github.com/awesome-gocui/gocui"
	"github.com/liamg/tml"
)

type AccountWidget struct {
	name string
	x, y int
	w, h int
	body string
	v    *gocui.View
}

func NewAccountWidget(name, accountNumber, region string) *AccountWidget {
	accountRegion := fmt.Sprintf("%s (%s)", accountNumber, region)

	return &AccountWidget{
		name: name,
		x:    1,
		y:    0,
		w:    5,
		h:    1,
		body: accountRegion,
	}
}

func (w *AccountWidget) ConfigureKeys() error {
	// nothing to configure here
	return nil
}

func (w *AccountWidget) Layout(g *gocui.Gui) error {
	v, err := g.SetView(w.name, w.x, w.y, w.w, w.h, 0)
	if err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return fmt.Errorf("%w", err)
		}
		_ = tml.Fprintf(v, " <blue>%s</blue>", w.body)
	}

	v.Title = " AWS Account "
	w.v = v
	return nil
}

func (w *AccountWidget) RefreshView() {
	w.v.Clear()
	_, _ = fmt.Fprintf(w.v, w.body)
}

func (w *AccountWidget) UpdateAccount() {
	w.v.Clear()
	_, _ = fmt.Fprintf(w.v, w.body)
}
