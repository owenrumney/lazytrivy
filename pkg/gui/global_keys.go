package gui

import (
	"github.com/awesome-gocui/gocui"
)

func (g *Gui) configureGlobalKeys() error {

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
		return g.RefreshImages()
	}); err != nil {
		return err
	}

	if err := g.cui.SetKeybinding("", gocui.KeyArrowLeft, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		if g.cui.CurrentView().Name() == "results" {
			_, err := g.cui.SetCurrentView("images")
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	if err := g.cui.SetKeybinding("", gocui.KeyArrowRight, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		if g.cui.CurrentView().Name() == "images" {
			_, err := g.cui.SetCurrentView("results")
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}
