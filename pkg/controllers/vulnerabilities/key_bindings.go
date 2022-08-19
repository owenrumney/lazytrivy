package vulnerabilities

import (
	"context"
	"fmt"

	"github.com/awesome-gocui/gocui"
	base2 "github.com/owenrumney/lazytrivy/pkg/controllers/base"
	"github.com/owenrumney/lazytrivy/pkg/widgets"
)

func (g *VulnerabilityController) configureKeyBindings() error {
	if err := g.Cui.SetKeybinding("", gocui.MouseLeft, gocui.ModNone, base2.SetView); err != nil {
		return fmt.Errorf("error setting keybinding for view switching: %w", err)
	}

	if err := g.Cui.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, base2.Quit); err != nil {
		return fmt.Errorf("error setting keybinding for quitting with Ctrl+C: %w", err)
	}

	if err := g.Cui.SetKeybinding("", 't', gocui.ModNone, g.CancelCurrentScan); err != nil {
		return fmt.Errorf("error setting keybinding for cancelling current scan: %w", err)
	}

	if err := g.Cui.SetKeybinding("", 'a', gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		g.ScanAllImages(context.Background())
		return nil
	}); err != nil {
		return fmt.Errorf("error setting keybinding for scanning all images: %w", err)
	}

	if err := g.Cui.SetKeybinding("", 'q', gocui.ModNone, base2.Quit); err != nil {
		return fmt.Errorf("error setting keybinding for quitting: %w", err)
	}

	if err := g.Cui.SetKeybinding("", 'r', gocui.ModNone, g.scanRemote); err != nil {
		return fmt.Errorf("error setting keybinding for scanning remote: %w", err)
	}

	if err := g.Cui.SetKeybinding("", 'i', gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		return g.RefreshImages()
	}); err != nil {
		return fmt.Errorf("error setting keybinding for refreshing images: %w", err)
	}

	if err := g.Cui.SetKeybinding("", gocui.KeyArrowLeft, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		if g.Cui.CurrentView().Name() == widgets.Results {
			_, err := g.Cui.SetCurrentView(widgets.Images)
			if err != nil {
				return fmt.Errorf("error getting the images view: %w", err)
			}
			if v, ok := g.Views[widgets.Images].(*widgets.ImagesWidget); ok {
				return v.SetSelectedImage(g.state.selectedImage)
			}
		}
		return nil
	}); err != nil {
		return fmt.Errorf("error setting keybinding for moving left: %w", err)
	}

	if err := g.Cui.SetKeybinding("", gocui.KeyArrowRight, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		if g.Cui.CurrentView().Name() == widgets.Images {
			_, err := g.Cui.SetCurrentView(widgets.Results)
			if err != nil {
				return fmt.Errorf("error getting the results view: %w", err)
			}
		}
		return nil
	}); err != nil {
		return fmt.Errorf("error setting keybinding for moving right: %w", err)
	}

	return nil
}
