package widgets

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/awesome-gocui/gocui"
)

type FilesWidget struct {
	ListWidget
	name string
	x, y int
	w, h int
	body []string

	ctx fsContext
	v   *gocui.View
}

func NewFilesWidget(name string, g fsContext) *FilesWidget {
	w := 28

	widget := &FilesWidget{
		ListWidget: ListWidget{
			ctx:                 g,
			selectionChangeFunc: g.SetSelected,
		},
		name: name,
		x:    0,
		y:    0,
		w:    w,
		h:    1,
		ctx:  g,
		body: []string{" Press 's' to scan path "},
	}

	return widget
}

func (w *FilesWidget) ConfigureKeys(*gocui.Gui) error {
	if err := w.ctx.SetKeyBinding(w.name, gocui.KeyArrowUp, gocui.ModNone, w.previousItem); err != nil {
		return fmt.Errorf("failed to set the previous image %w", err)
	}

	if err := w.ctx.SetKeyBinding(w.name, 's', gocui.ModNone, w.ctx.ScanVulnerabilities); err != nil {

	}

	if err := w.ctx.SetKeyBinding(w.name, gocui.KeyArrowDown, gocui.ModNone, w.nextItem); err != nil {
		return fmt.Errorf("failed to set the next image %w", err)
	}

	if err := w.ctx.SetKeyBinding(w.name, gocui.KeyEnter, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		w.ctx.ShowTarget(context.Background(), w.SelectTarget())
		return nil
	}); err != nil {
		return fmt.Errorf("error setting keybinding for scanning image: %w", err)
	}

	return nil
}

func (w *FilesWidget) Layout(g *gocui.Gui) error {
	v, err := g.SetView(w.name, w.x, w.y, w.w, w.h, 0)
	if err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return fmt.Errorf("%w", err)
		}
		_, _ = fmt.Fprint(v, w.body)
		_ = v.SetCursor(0, 0)
		v.Highlight = false
	}
	v.Title = " Files "

	v.SelBgColor = gocui.ColorGreen | gocui.AttrDim
	v.SelFgColor = gocui.ColorBlack | gocui.AttrBold
	if g.CurrentView() == v {
		v.FrameColor = gocui.ColorGreen
	} else {
		v.FrameColor = gocui.ColorDefault
	}

	w.v = v
	return nil
}

func (w *FilesWidget) RefreshFiles(files []string, fileWidth int) error {
	w.w = fileWidth + 4

	if len(files) == 0 {
		files = append(files, "No issues found         ")
	} else {

		fileList := make([]string, len(files))
		for i, file := range files {
			fileList[i] = fmt.Sprintf("% -*s", fileWidth+1, file)
		}
		w.bottomMost = len(fileList)
	}
	w.body = files
	w.v.Highlight = true
	w.RefreshView()
	_ = w.v.SetCursor(0, 0)
	return nil

}

func (w *FilesWidget) SelectTarget() string {
	_, y := w.v.Cursor()
	if y >= len(w.body) {
		return ""
	}
	target := strings.TrimSpace(w.body[y])
	parts := strings.Split(target, "|")
	if len(parts) > 1 {
		return strings.TrimSpace(parts[1])
	}
	return parts[0]
}

func (w *FilesWidget) RefreshView() {
	w.v.Clear()
	for _, line := range w.body {
		parts := strings.Split(line, "|")
		_, _ = fmt.Fprintln(w.v, parts[0])
	}
}
