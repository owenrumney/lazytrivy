package widgets

import (
	"errors"
	"fmt"

	"github.com/awesome-gocui/gocui"
	"github.com/liamg/tml"
)

type ScanPathWidget struct {
	name string
	x, y int
	w, h int
	body string
	v    *gocui.View
	ctx  baseContext
}

func NewScanPathWidget(name string, workingDir string, ctx baseContext) *ScanPathWidget {

	return &ScanPathWidget{
		name: name,
		x:    1,
		y:    0,
		w:    5,
		h:    1,
		body: workingDir,
		ctx:  ctx,
	}
}

func (w *ScanPathWidget) ConfigureKeys(*gocui.Gui) error {
	// nothing to configure here
	return nil
}

func (w *ScanPathWidget) Layout(g *gocui.Gui) error {
	v, err := g.SetView(w.name, w.x, w.y, w.w, w.h, 0)
	if err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return fmt.Errorf("%w", err)
		}
		_ = tml.Fprintf(v, " <blue>%s</blue>", w.body)
	}

	v.Title = " Current Scan Path "
	w.v = v
	return nil
}

func (w *ScanPathWidget) UpdateWorkingDir(workingDir string) {
	w.body = workingDir
	w.RefreshView()
}

func (w *ScanPathWidget) RefreshView() {
	w.v.Clear()
	_, _ = fmt.Fprintf(w.v, "%s", w.body)
}
