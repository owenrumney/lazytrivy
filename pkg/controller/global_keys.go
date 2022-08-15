package controller

import (
	"context"

	"github.com/awesome-gocui/gocui"
	"github.com/owenrumney/lazytrivy/pkg/widgets"
)

func (g *Controller) configureGlobalKeys() error {

	if err := g.cui.SetKeybinding("", gocui.MouseLeft, gocui.ModNone, setView); err != nil {
		return err
	}

	if err := g.cui.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}

	if err := g.cui.SetKeybinding("", 'c', gocui.ModNone, g.CancelCurrentScan); err != nil {
		return err
	}

	if err := g.cui.SetKeybinding("", 's', gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		g.ScanImage(context.Background(), g.selectedImage)
		return nil
	}); err != nil {
		return err
	}

	if err := g.cui.SetKeybinding("", 'a', gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		g.ScanAllImages(context.Background())
		return nil
	}); err != nil {
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
			if err != nil {
				return err
			}
			return g.views["images"].(*widgets.ImagesWidget).SetSelectedImage(g.selectedImage)
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
