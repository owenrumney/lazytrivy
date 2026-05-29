package widgets

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/awesome-gocui/gocui"
)

type ImagesWidget struct {
	ListWidget
	name string
	x, y int
	w, h int

	allImages     []string
	visibleImages []string
	filterQuery   string
	filterMode    bool
	imageWidth    int

	ctx vulnerabilityContext
	v   *gocui.View
}

func NewImagesWidget(name string, g vulnerabilityContext) *ImagesWidget {
	w := 25

	widget := &ImagesWidget{
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
	}

	return widget
}

func (w *ImagesWidget) ConfigureKeys(*gocui.Gui) error {
	if err := w.ctx.SetKeyBinding(w.name, gocui.KeyArrowUp, gocui.ModNone, w.previousItem); err != nil {
		return fmt.Errorf("failed to set the previous image %w", err)
	}

	if err := w.ctx.SetKeyBinding(w.name, gocui.KeyArrowDown, gocui.ModNone, w.nextItem); err != nil {
		return fmt.Errorf("failed to set the next image %w", err)
	}

	if err := w.ctx.SetKeyBinding(w.name, gocui.KeyEnter, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		if w.filterMode {
			w.exitFilterMode(false)
			return nil
		}
		return w.scanSelectedImage()
	}); err != nil {
		return fmt.Errorf("error setting keybinding for scanning image: %w", err)
	}

	if err := w.ctx.SetKeyBinding(w.name, 's', gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		return w.scanSelectedImage()
	}); err != nil {
		return fmt.Errorf("%w", err)
	}

	if err := w.ctx.SetKeyBinding(w.name, '/', gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		w.enterFilterMode(view)
		return nil
	}); err != nil {
		return fmt.Errorf("error setting keybinding for filtering images: %w", err)
	}

	if err := w.ctx.SetKeyBinding(w.name, gocui.KeyEsc, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		if w.filterMode || w.filterQuery != "" {
			w.exitFilterMode(true)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("error setting keybinding for clearing image filter: %w", err)
	}

	if err := w.ctx.SetKeyBinding(w.name, gocui.KeyBackspace, gocui.ModNone, w.removeFilterCharacter); err != nil {
		return fmt.Errorf("error setting keybinding for image filter backspace: %w", err)
	}

	if err := w.ctx.SetKeyBinding(w.name, gocui.KeyBackspace2, gocui.ModNone, w.removeFilterCharacter); err != nil {
		return fmt.Errorf("error setting keybinding for image filter backspace: %w", err)
	}

	return nil
}

func (w *ImagesWidget) Layout(g *gocui.Gui) error {
	v, err := g.SetView(w.name, w.x, w.y, w.w, w.h, 0)
	if err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return fmt.Errorf("%w", err)
		}
		w.v = v
		w.RefreshView()
	}
	v.Title = w.title()
	v.Highlight = true
	v.Editor = w
	v.Editable = w.filterMode
	v.Highlight = true
	v.SelBgColor = gocui.ColorDefault | gocui.AttrDim
	v.SelFgColor = gocui.ColorBlue | gocui.AttrBold
	if g.CurrentView() == v {
		v.FrameColor = gocui.ColorBlue
	} else {
		v.FrameColor = gocui.ColorDefault
	}
	v.FrameRunes = []rune{'─', '│', '╭', '╮', '╰', '╯'}

	return nil
}

func (w *ImagesWidget) RefreshImages(images []string, imageWidth int) error {
	w.allImages = append([]string(nil), images...)
	w.imageWidth = imageWidth
	w.w = imageWidth + 4

	if len(images) == 0 {
		w.visibleImages = nil
		w.body = nil
		w.currentPos = 0
		w.bottomMost = 0
		w.RefreshView()
		w.updateTitle()
		w.ctx.UpdateStatus("No local images found")
		return nil
	}

	w.applyFilter(w.SelectedImage())
	return nil
}

func (w *ImagesWidget) SetSelectedImage(image string) error {
	for i, line := range w.body {
		if stripIdentifierPrefix(line) == strings.TrimSpace(image) {
			w.setCursorToPosition(i)
			break
		}
	}
	return nil
}

func (w *ImagesWidget) SelectedImage() string {
	if w.v == nil {
		return ""
	}
	_, y := w.v.Cursor()
	_, oy := w.v.Origin()
	if image, err := w.v.Line(y + oy); err == nil {
		return stripIdentifierPrefix(image)
	}
	return ""
}

func (w *ImagesWidget) RefreshView() {
	if w.v == nil {
		return
	}
	w.v.Clear()
	for _, line := range w.body {
		_, _ = fmt.Fprintln(w.v, stripIdentifierPrefix(line))
	}
}

func (w *ImagesWidget) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	if !w.filterMode {
		return
	}

	previousSelection := w.SelectedImage()
	switch {
	case ch != 0 && mod == gocui.ModNone:
		w.filterQuery += string(ch)
	case key == gocui.KeySpace:
		w.filterQuery += " "
	default:
		return
	}
	w.applyFilter(previousSelection)
}

func (w *ImagesWidget) enterFilterMode(v *gocui.View) {
	if v != nil {
		w.v = v
	}
	w.filterMode = true
	if w.v != nil {
		w.v.Editor = w
		w.v.Editable = true
	}
	w.updateTitle()
	w.ctx.UpdateStatus("Filtering images. Press Enter to keep the filter or Esc to clear it.")
}

func (w *ImagesWidget) exitFilterMode(clear bool) {
	previousSelection := w.SelectedImage()
	w.filterMode = false
	if w.v != nil {
		w.v.Editable = false
	}

	if clear {
		w.filterQuery = ""
		w.applyFilter(previousSelection)
		w.ctx.UpdateStatus("Image filter cleared")
		return
	}

	w.updateTitle()
}

func (w *ImagesWidget) removeFilterCharacter(_ *gocui.Gui, _ *gocui.View) error {
	if !w.filterMode || w.filterQuery == "" {
		return nil
	}

	previousSelection := w.SelectedImage()
	runes := []rune(w.filterQuery)
	w.filterQuery = string(runes[:len(runes)-1])
	w.applyFilter(previousSelection)
	return nil
}

func (w *ImagesWidget) scanSelectedImage() error {
	image := w.SelectedImage()
	if image == "" {
		return nil
	}

	w.ctx.SetSelected(image)
	w.ctx.ScanImage(context.Background())
	return nil
}

func (w *ImagesWidget) applyFilter(previousSelection string) {
	query := strings.ToLower(strings.TrimSpace(w.filterQuery))
	if query == "" {
		w.visibleImages = append([]string(nil), w.allImages...)
	} else {
		w.visibleImages = w.visibleImages[:0]
		for _, image := range w.allImages {
			if strings.Contains(strings.ToLower(image), query) {
				w.visibleImages = append(w.visibleImages, image)
			}
		}
	}

	w.renderImageList()
	w.RefreshView()
	w.selectBestImage(previousSelection)
	w.updateTitle()

	if w.filterQuery != "" && len(w.visibleImages) == 0 {
		w.ctx.UpdateStatus(fmt.Sprintf("No images match filter: %s", w.filterQuery))
	}
}

func (w *ImagesWidget) renderImageList() {
	w.body = make([]string, len(w.visibleImages))
	for i, image := range w.visibleImages {
		w.body[i] = fmt.Sprintf("**%d*** % -*s", i, w.imageWidth+1, image)
	}

	w.currentPos = 0
	w.bottomMost = len(w.body) - 1
	if len(w.body) == 0 {
		w.bottomMost = 0
	}
}

func (w *ImagesWidget) selectBestImage(previousSelection string) {
	if len(w.visibleImages) == 0 {
		if w.v != nil {
			_ = w.v.SetOrigin(0, 0)
			_ = w.v.SetCursor(0, 0)
		}
		w.currentPos = 0
		return
	}

	selectedIndex := 0
	for i, image := range w.visibleImages {
		if image == previousSelection {
			selectedIndex = i
			break
		}
	}

	w.setCursorToPosition(selectedIndex)
	w.ctx.SetSelected(w.visibleImages[selectedIndex])
}

func (w *ImagesWidget) setCursorToPosition(pos int) {
	w.currentPos = pos
	if w.v == nil {
		return
	}

	height := 0
	_, height = w.v.Size()
	origin := 0
	cursorY := pos
	if height > 0 && pos >= height {
		origin = pos - height + 1
		cursorY = height - 1
	}

	_ = w.v.SetOrigin(0, origin)
	_ = w.v.SetCursor(0, cursorY)
}

func (w *ImagesWidget) updateTitle() {
	if w.v != nil {
		w.v.Title = w.title()
	}
}

func (w *ImagesWidget) title() string {
	if w.filterMode || w.filterQuery != "" {
		return fmt.Sprintf(" Images /%s %d/%d ", w.filterQuery, len(w.visibleImages), len(w.allImages))
	}
	return " Images "
}
