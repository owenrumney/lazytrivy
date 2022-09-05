package vulnerabilities

import (
	"fmt"

	"github.com/awesome-gocui/gocui"
	base2 "github.com/owenrumney/lazytrivy/pkg/controllers/base"
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

	if err := c.Cui.SetKeybinding("", 'a', gocui.ModNone, c.ScanAllImages); err != nil {
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

	if err := c.Cui.SetKeybinding("", gocui.KeyArrowLeft, gocui.ModNone, c.moveViewLeft); err != nil {
		return fmt.Errorf("error setting keybinding for moving left: %w", err)
	}

	if err := c.Cui.SetKeybinding("", gocui.KeyArrowRight, gocui.ModNone, c.moveViewRight); err != nil {
		return fmt.Errorf("error setting keybinding for moving right: %w", err)
	}

	return nil
}
