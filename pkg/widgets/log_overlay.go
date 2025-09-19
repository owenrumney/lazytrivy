package widgets

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/awesome-gocui/gocui"
)

type LogOverlayWidget struct {
	name         string
	x, y         int
	w, h         int
	v            *gocui.View
	ctx          baseContext
	body         []string
	previousView string
}

func NewLogOverlayWidget(name string, g baseContext) *LogOverlayWidget {
	return &LogOverlayWidget{
		name: name,
		ctx:  g,
		body: []string{},
	}
}

func (w *LogOverlayWidget) SetPreviousView(viewName string) {
	w.previousView = viewName
}

func (w *LogOverlayWidget) ConfigureKeys(gui *gocui.Gui) error {
	// ESC and 'q' to close the overlay
	if err := w.ctx.SetKeyBinding(w.name, gocui.KeyEsc, gocui.ModNone, w.closeOverlay); err != nil {
		return fmt.Errorf("failed to set ESC keybinding: %w", err)
	}

	if err := w.ctx.SetKeyBinding(w.name, 'q', gocui.ModNone, w.closeOverlay); err != nil {
		return fmt.Errorf("failed to set 'q' keybinding: %w", err)
	}

	// Arrow keys for scrolling
	if err := w.ctx.SetKeyBinding(w.name, gocui.KeyArrowUp, gocui.ModNone, w.scrollUp); err != nil {
		return fmt.Errorf("failed to set up arrow keybinding: %w", err)
	}

	if err := w.ctx.SetKeyBinding(w.name, gocui.KeyArrowDown, gocui.ModNone, w.scrollDown); err != nil {
		return fmt.Errorf("failed to set down arrow keybinding: %w", err)
	}

	// Page up/down for faster scrolling
	if err := w.ctx.SetKeyBinding(w.name, gocui.KeyPgup, gocui.ModNone, w.pageUp); err != nil {
		return fmt.Errorf("failed to set page up keybinding: %w", err)
	}

	if err := w.ctx.SetKeyBinding(w.name, gocui.KeyPgdn, gocui.ModNone, w.pageDown); err != nil {
		return fmt.Errorf("failed to set page down keybinding: %w", err)
	}

	return nil
}

func (w *LogOverlayWidget) Layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	// Make overlay slightly smaller than full screen
	margin := 2
	w.x = margin
	w.y = margin
	w.w = maxX - margin*2 - 1
	w.h = maxY - margin*2 - 1

	v, err := g.SetView(w.name, w.x, w.y, w.w, w.h, 0)
	if err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return fmt.Errorf("failed to create log overlay view: %w", err)
		}
	}

	w.v = v
	v.Title = " LazyTrivy Logs (ESC or 'q' to close, ↑↓ to scroll, PgUp/PgDn for fast scroll) "
	v.FrameColor = gocui.ColorYellow
	v.FrameRunes = []rune{'═', '║', '╔', '╗', '╚', '╝'}
	v.Wrap = true
	v.Autoscroll = false

	// Load and display log content
	if err := w.loadLogContent(); err != nil {
		w.body = []string{fmt.Sprintf("Error loading logs: %v", err)}
	}

	w.refreshView()
	return nil
}

func (w *LogOverlayWidget) loadLogContent() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	logFile := filepath.Join(home, ".lazytrivy", "logs", "lazytrivy.log")

	file, err := os.Open(logFile)
	if err != nil {
		return fmt.Errorf("failed to open log file %s: %w", logFile, err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)

	// Read all lines first
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read log file: %w", err)
	}

	// Get the last 200 lines (tail behavior)
	tailLines := 200
	if len(lines) > tailLines {
		w.body = lines[len(lines)-tailLines:]
	} else {
		w.body = lines
	}

	if len(w.body) == 0 {
		w.body = []string{"No log entries found"}
	}

	return nil
}

func (w *LogOverlayWidget) refreshView() {
	if w.v == nil {
		return
	}

	w.v.Clear()
	for _, line := range w.body {
		_, _ = fmt.Fprintln(w.v, line)
	}

	// Auto-scroll to bottom initially
	if len(w.body) > 0 {
		maxY := len(w.body) - 1
		_, viewHeight := w.v.Size()
		if maxY > viewHeight {
			w.v.SetOrigin(0, maxY-viewHeight+1)
		}
	}
}

func (w *LogOverlayWidget) closeOverlay(g *gocui.Gui, v *gocui.View) error {
	if err := g.DeleteView(w.name); err != nil {
		return fmt.Errorf("failed to delete log overlay: %w", err)
	}

	// Return focus to the previous view if we have one
	if w.previousView != "" {
		_, err := g.SetCurrentView(w.previousView)
		if err != nil {
			// If we can't set the previous view, don't return an error
			// just let the GUI handle it
		}
	}

	return nil
}

func (w *LogOverlayWidget) scrollUp(g *gocui.Gui, v *gocui.View) error {
	ox, oy := v.Origin()
	if oy > 0 {
		if err := v.SetOrigin(ox, oy-1); err != nil {
			return err
		}
	}
	return nil
}

func (w *LogOverlayWidget) scrollDown(g *gocui.Gui, v *gocui.View) error {
	ox, oy := v.Origin()
	_, viewHeight := v.Size()
	maxY := len(w.body) - viewHeight

	if oy < maxY {
		if err := v.SetOrigin(ox, oy+1); err != nil {
			return err
		}
	}
	return nil
}

func (w *LogOverlayWidget) pageUp(g *gocui.Gui, v *gocui.View) error {
	ox, oy := v.Origin()
	_, viewHeight := v.Size()

	newY := oy - viewHeight + 1
	if newY < 0 {
		newY = 0
	}

	if err := v.SetOrigin(ox, newY); err != nil {
		return err
	}
	return nil
}

func (w *LogOverlayWidget) pageDown(g *gocui.Gui, v *gocui.View) error {
	ox, oy := v.Origin()
	_, viewHeight := v.Size()
	maxY := len(w.body) - viewHeight

	newY := oy + viewHeight - 1
	if newY > maxY {
		newY = maxY
	}

	if err := v.SetOrigin(ox, newY); err != nil {
		return err
	}
	return nil
}
