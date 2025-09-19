package widgets

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/awesome-gocui/gocui"
)

type PathBrowserWidget struct {
	name        string
	x, y        int
	w, h        int
	maxLength   int
	currentPath string
	entries     []string
	selectedIdx int
	inInputMode bool // true when editing text, false when navigating list
	ctx         fsContext
}

func NewPathBrowserWidget(name string, maxX, maxY, maxLength int, currentPath string, ctx fsContext) (*PathBrowserWidget, error) {
	x1 := maxX/2 - 60
	x2 := maxX/2 + 60
	y1 := maxY/2 - 15
	y2 := maxY/2 + 15

	widget := &PathBrowserWidget{
		name:        name,
		x:           x1,
		y:           y1,
		w:           x2,
		h:           y2,
		maxLength:   maxLength,
		currentPath: currentPath,
		selectedIdx: 0,
		inInputMode: true,
		ctx:         ctx,
	}

	// Load initial directory contents
	if err := widget.loadDirectoryContents(); err != nil {
		// If we can't load the current path, try parent directory
		widget.currentPath = filepath.Dir(currentPath)
		if err := widget.loadDirectoryContents(); err != nil {
			// Fallback to home directory
			if home, err := os.UserHomeDir(); err == nil {
				widget.currentPath = home
				widget.loadDirectoryContents()
			}
		}
	}

	if err := widget.setupKeyBindings(); err != nil {
		return nil, fmt.Errorf("failed to setup key bindings: %w", err)
	}

	return widget, nil
}

func (w *PathBrowserWidget) loadDirectoryContents() error {
	entries, err := os.ReadDir(w.currentPath)
	if err != nil {
		return err
	}

	w.entries = []string{}

	// Add parent directory option (unless we're at root)
	if w.currentPath != "/" && w.currentPath != "" {
		w.entries = append(w.entries, "..")
	}

	// Only add directories (not files)
	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name()+"/")
		}
	}

	// Sort and add directories
	sort.Strings(dirs)
	w.entries = append(w.entries, dirs...)

	// Reset selection to first item
	w.selectedIdx = 0
	return nil
}

func (w *PathBrowserWidget) setupKeyBindings() error {
	// Enter key - confirm selection or navigate
	if err := w.ctx.SetKeyBinding(w.name, gocui.KeyEnter, gocui.ModNone, w.handleEnter); err != nil {
		return err
	}

	// Escape - cancel
	if err := w.ctx.SetKeyBinding(w.name, gocui.KeyEsc, gocui.ModNone, w.handleEscape); err != nil {
		return err
	}

	// Tab - switch between input and list
	if err := w.ctx.SetKeyBinding(w.name, gocui.KeyTab, gocui.ModNone, w.handleTab); err != nil {
		return err
	}

	// Arrow keys - navigate list
	if err := w.ctx.SetKeyBinding(w.name, gocui.KeyArrowUp, gocui.ModNone, w.handleArrowUp); err != nil {
		return err
	}

	if err := w.ctx.SetKeyBinding(w.name, gocui.KeyArrowDown, gocui.ModNone, w.handleArrowDown); err != nil {
		return err
	}

	return nil
}

func (w *PathBrowserWidget) handleEnter(gui *gocui.Gui, view *gocui.View) error {
	if w.inInputMode {
		// User pressed enter in text input - validate and use the typed path
		var pathToUse string
		if len(view.BufferLines()) > 0 {
			if line, err := view.Line(0); err == nil {
				pathToUse = strings.TrimSpace(line)
			}
		}

		if pathToUse == "" {
			pathToUse = w.currentPath
		}

		// Expand ~ to home directory
		if strings.HasPrefix(pathToUse, "~/") {
			if home, err := os.UserHomeDir(); err == nil {
				pathToUse = filepath.Join(home, pathToUse[2:])
			}
		}

		// Check if path exists and is a directory
		if stat, err := os.Stat(pathToUse); err == nil {
			if stat.IsDir() {
				w.ctx.SetWorkingDirectory(pathToUse)
				return w.close(gui)
			} else {
				w.ctx.UpdateStatus(fmt.Sprintf("Path must be a directory: %s", pathToUse))
				return nil
			}
		} else {
			w.ctx.UpdateStatus(fmt.Sprintf("Directory does not exist: %s", pathToUse))
			return nil
		}
	} else {
		// User pressed enter on a list item
		if w.selectedIdx >= 0 && w.selectedIdx < len(w.entries) {
			selectedEntry := w.entries[w.selectedIdx]

			if selectedEntry == ".." {
				// Navigate to parent directory
				w.currentPath = filepath.Dir(w.currentPath)
				w.loadDirectoryContents()
				w.updateView(gui)
			} else if strings.HasSuffix(selectedEntry, "/") {
				// Navigate into directory or select it
				dirPath := filepath.Join(w.currentPath, strings.TrimSuffix(selectedEntry, "/"))

				// For now, let's navigate into the directory.
				// User can press Enter again in input mode to select the current directory
				w.currentPath = dirPath
				w.loadDirectoryContents()
				w.updateView(gui)
			}
		}
	}
	return nil
}

func (w *PathBrowserWidget) handleEscape(gui *gocui.Gui, view *gocui.View) error {
	return w.close(gui)
}

func (w *PathBrowserWidget) handleTab(gui *gocui.Gui, view *gocui.View) error {
	w.inInputMode = !w.inInputMode
	w.updateView(gui)
	return nil
}

func (w *PathBrowserWidget) handleArrowUp(gui *gocui.Gui, view *gocui.View) error {
	if !w.inInputMode && len(w.entries) > 0 {
		if w.selectedIdx > 0 {
			w.selectedIdx--
		}
		w.updateView(gui)
	}
	return nil
}

func (w *PathBrowserWidget) handleArrowDown(gui *gocui.Gui, view *gocui.View) error {
	if !w.inInputMode && len(w.entries) > 0 {
		if w.selectedIdx < len(w.entries)-1 {
			w.selectedIdx++
		}
		w.updateView(gui)
	}
	return nil
}

func (w *PathBrowserWidget) close(gui *gocui.Gui) error {
	gui.Mouse = true
	gui.Cursor = false

	if err := gui.DeleteView(w.name); err != nil {
		return fmt.Errorf("failed to delete view '%s': %w", w.name, err)
	}
	if _, err := gui.SetCurrentView(Files); err != nil {
		return fmt.Errorf("failed to switch view to 'files': %w", err)
	}
	return nil
}

func (w *PathBrowserWidget) updateView(gui *gocui.Gui) {
	gui.Update(func(g *gocui.Gui) error {
		if v, err := g.View(w.name); err == nil {
			w.refreshView(v)
		}
		return nil
	})
}

func (w *PathBrowserWidget) refreshView(v *gocui.View) {
	// Clear the view
	v.Clear()

	// Set input field text
	fmt.Fprint(v, w.currentPath)

	// Add separator
	fmt.Fprintln(v, "")

	// Get the actual view width and create separator line
	width, _ := v.Size()
	if width > 0 {
		fmt.Fprintln(v, strings.Repeat("─", width))
	} else {
		fmt.Fprintln(v, strings.Repeat("─", 80)) // Fallback
	}

	// Add directory listing (directories only)
	for i, entry := range w.entries {
		prefix := "  "
		if !w.inInputMode && i == w.selectedIdx {
			prefix = "> "
		}

		// All entries are either ".." or directories with trailing slash
		fmt.Fprintf(v, "%s%s\n", prefix, entry)
	}

	// Position cursor appropriately
	if w.inInputMode {
		v.SetCursor(len(w.currentPath), 0)
	}
}

func (w *PathBrowserWidget) Layout(g *gocui.Gui) error {
	v, err := g.SetView(w.name, w.x, w.y, w.w, w.h, 0)
	if err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return fmt.Errorf("failed to create view: %w", err)
		}

		v.Title = " Directory Browser - Tab to switch between input/list, Enter to navigate/select "
		v.Editor = &pathBrowserEditor{widget: w}
		v.Editable = true
		g.Cursor = true

		v.TitleColor = gocui.ColorGreen
		v.FrameColor = gocui.ColorBlue
		v.FrameRunes = []rune{'─', '│', '╭', '╮', '╰', '╯'}

		w.refreshView(v)
	}
	return nil
}

// Custom editor for the path browser input field
type pathBrowserEditor struct {
	widget *PathBrowserWidget
}

func (e *pathBrowserEditor) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	if !e.widget.inInputMode {
		return // Don't edit when in list navigation mode
	}

	cx, _ := v.Cursor()
	ox, _ := v.Origin()
	limit := ox+cx+1 > e.widget.maxLength

	switch {
	case ch != 0 && mod == 0 && !limit:
		v.EditWrite(ch)
		e.updatePath(v)
	case key == gocui.KeySpace && !limit:
		v.EditWrite(' ')
		e.updatePath(v)
	case key == gocui.KeyBackspace || key == gocui.KeyBackspace2:
		v.EditDelete(true)
		e.updatePath(v)
	}
}

func (e *pathBrowserEditor) updatePath(v *gocui.View) {
	// Get current text from input
	if len(v.BufferLines()) > 0 {
		if line, err := v.Line(0); err == nil {
			newPath := strings.TrimSpace(line)

			// Expand ~ to home directory
			if strings.HasPrefix(newPath, "~/") {
				if home, err := os.UserHomeDir(); err == nil {
					newPath = filepath.Join(home, newPath[2:])
				}
			}

			// If it's a valid directory, update the listing
			if stat, err := os.Stat(newPath); err == nil && stat.IsDir() {
				e.widget.currentPath = newPath
				e.widget.loadDirectoryContents()
			}
		}
	}
}
