package controller

import (
	"context"
	"fmt"

	"github.com/awesome-gocui/gocui"
	"github.com/owenrumney/lazytrivy/pkg/widgets"
)

func (g *Controller) configureGlobalKeys() error {
	if err := g.cui.SetKeybinding("", gocui.MouseLeft, gocui.ModNone, setView); err != nil {
		return fmt.Errorf("error setting keybinding for view switching: %w", err)
	}

	if err := g.cui.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return fmt.Errorf("error setting keybinding for quitting with Ctrl+C: %w", err)
	}

	if err := g.cui.SetKeybinding("", 'c', gocui.ModNone, g.CancelCurrentScan); err != nil {
		return fmt.Errorf("error setting keybinding for cancelling current scan: %w", err)
	}

	if err := g.cui.SetKeybinding("", 's', gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		g.ScanImage(context.Background(), g.selectedImage)
		return nil
	}); err != nil {
		return fmt.Errorf("error setting keybinding for scanning image: %w", err)
	}

	if err := g.cui.SetKeybinding("", 'a', gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		g.ScanAllImages(context.Background())
		return nil
	}); err != nil {
		return fmt.Errorf("error setting keybinding for scanning all images: %w", err)
	}

	if err := g.cui.SetKeybinding("", 'q', gocui.ModNone, quit); err != nil {
		return fmt.Errorf("error setting keybinding for quitting: %w", err)
	}

	if err := g.cui.SetKeybinding("", 'r', gocui.ModNone, g.scanRemote); err != nil {
		return fmt.Errorf("error setting keybinding for scanning remote: %w", err)
	}

	if err := g.cui.SetKeybinding("", 'i', gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		return g.RefreshImages()
	}); err != nil {
		return fmt.Errorf("error setting keybinding for refreshing images: %w", err)
	}

	if err := g.cui.SetKeybinding("", gocui.KeyArrowLeft, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		if g.cui.CurrentView().Name() == widgets.Results {
			_, err := g.cui.SetCurrentView(widgets.Images)
			if err != nil {
				return fmt.Errorf("error getting the images view: %w", err)
			}
			if v, ok := g.views[widgets.Images].(*widgets.ImagesWidget); ok {
				return v.SetSelectedImage(g.selectedImage)
			}
		}
		return nil
	}); err != nil {
		return fmt.Errorf("error setting keybinding for moving left: %w", err)
	}

	if err := g.cui.SetKeybinding("", gocui.KeyArrowRight, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		if g.cui.CurrentView().Name() == widgets.Images {
			_, err := g.cui.SetCurrentView(widgets.Results)
			return fmt.Errorf("error getting the results view: %w", err)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("error setting keybinding for moving right: %w", err)
	}
	return nil
}
