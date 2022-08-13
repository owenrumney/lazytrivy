package gui

import (
	"fmt"

	"github.com/awesome-gocui/gocui"
)

func (g *Gui) configureKeyBindings() error {

	if err := g.configureGlobalKeyBindings(); err != nil {
		return fmt.Errorf("error configuring global keybindings: %w", err)
	}

	if err := g.configureImagesKeyBindings(); err != nil {
		return fmt.Errorf("error configuring images keybindings: %w", err)
	}

	if err := g.configureResultsKeyBindings(); err != nil {
		return fmt.Errorf("error configuring results keybindings: %w", err)
	}

	return nil
}

func (g *Gui) configureGlobalKeyBindings() error {

	if err := g.cui.SetKeybinding("", gocui.MouseLeft, gocui.ModNone, setView); err != nil {
		return err
	}

	if err := g.cui.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}

	if err := g.cui.SetKeybinding("", 'q', gocui.ModNone, quit); err != nil {
		return err
	}

	if err := g.cui.SetKeybinding("", 'r', gocui.ModNone, g.scanRemote); err != nil {
		return err
	}

	if err := g.cui.SetKeybinding("", 'i', gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		return g.images.RefreshImages()
	}); err != nil {
		return err
	}

	if err := g.cui.SetKeybinding("", gocui.KeyArrowLeft, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		if g.cui.CurrentView().Name() == "main" {
			_, err := g.cui.SetCurrentView("images")
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	if err := g.cui.SetKeybinding("", gocui.KeyArrowRight, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		if g.cui.CurrentView().Name() == "images" {
			_, err := g.cui.SetCurrentView("main")
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (g *Gui) configureImagesKeyBindings() error {
	if err := g.cui.SetKeybinding(g.images.ViewName(), gocui.KeyArrowUp, gocui.ModNone, g.images.PreviousImage); err != nil {
		return err
	}

	if err := g.cui.SetKeybinding(g.images.ViewName(), gocui.KeyArrowDown, gocui.ModNone, g.images.NextImage); err != nil {
		return err
	}

	if err := g.cui.SetKeybinding(g.images.ViewName(), 's', gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		if image := g.SelectedImage(); image != "" {
			g.ScanImage(image)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (g *Gui) configureResultsKeyBindings() error {

	if err := g.cui.SetKeybinding(g.results.ViewName(), gocui.MouseWheelDown, gocui.ModNone, g.results.ScrollDown); err != nil {
		return err
	}

	if err := g.cui.SetKeybinding(g.results.ViewName(), gocui.MouseWheelUp, gocui.ModNone, g.results.ScrollUp); err != nil {
		return err
	}
	return nil
}
