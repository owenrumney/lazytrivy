package vulnerabilities

import (
	"context"
	"fmt"

	"github.com/awesome-gocui/gocui"
	base2 "github.com/owenrumney/lazytrivy/pkg/controllers/base"
	"github.com/owenrumney/lazytrivy/pkg/widgets"
)

func (c *Controller) configureKeyBindings() error {

	if err := c.ConfigureGlobalKeyBindings(); err != nil {
		return fmt.Errorf("error configuring global keybindings: %w", err)
	}

	if err := c.Cui.SetKeybinding("", gocui.MouseLeft, gocui.ModNone, base2.SetView); err != nil {
		return fmt.Errorf("error setting keybinding for view switching: %w", err)
	}

	if err := c.Cui.SetKeybinding("", 't', gocui.ModNone, c.CancelCurrentScan); err != nil {
		return fmt.Errorf("error setting keybinding for cancelling current scan: %w", err)
	}

	if err := c.Cui.SetKeybinding("", 'a', gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		c.ScanAllImages(context.Background())
		return nil
	}); err != nil {
		return fmt.Errorf("error setting keybinding for scanning all images: %w", err)
	}

	if err := c.Cui.SetKeybinding("", 'r', gocui.ModNone, c.scanRemote); err != nil {
		return fmt.Errorf("error setting keybinding for scanning remote: %w", err)
	}

	if err := c.Cui.SetKeybinding("", 'i', gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		return c.RefreshImages()
	}); err != nil {
		return fmt.Errorf("error setting keybinding for refreshing images: %w", err)
	}

	if err := c.Cui.SetKeybinding("", gocui.KeyArrowLeft, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		if c.Cui.CurrentView().Name() == widgets.Results {
			_, err := c.Cui.SetCurrentView(widgets.Images)
			if err != nil {
				return fmt.Errorf("error getting the images view: %w", err)
			}
			if v, ok := c.Views[widgets.Images].(*widgets.ImagesWidget); ok {
				return v.SetSelectedImage(c.state.selectedImage)
			}
		}
		return nil
	}); err != nil {
		return fmt.Errorf("error setting keybinding for moving left: %w", err)
	}

	if err := c.Cui.SetKeybinding("", gocui.KeyArrowRight, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		if c.Cui.CurrentView().Name() == widgets.Images {
			_, err := c.Cui.SetCurrentView(widgets.Results)
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
