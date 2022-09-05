package aws

import (
	"fmt"

	"github.com/awesome-gocui/gocui"
	"github.com/owenrumney/lazytrivy/pkg/controllers/base"
	"github.com/owenrumney/lazytrivy/pkg/logger"
)

func (c *Controller) configureKeyBindings() error {
	logger.Debug("Configuring global AWS Controller keyboard shortcuts")
	if err := c.ConfigureGlobalKeyBindings(); err != nil {
		return fmt.Errorf("error configuring global keybindings: %w", err)
	}

	if err := c.Cui.SetKeybinding("", gocui.MouseLeft, gocui.ModNone, base.SetView); err != nil {
		return fmt.Errorf("error setting keybinding for view switching: %w", err)
	}

	if err := c.Cui.SetKeybinding("", 't', gocui.ModNone, c.CancelCurrentScan); err != nil {
		return fmt.Errorf("error setting keybinding for cancelling current scan: %w", err)
	}

	if err := c.Cui.SetKeybinding("", gocui.KeyArrowLeft, gocui.ModNone, c.moveViewLeft); err != nil {
		return fmt.Errorf("error setting keybinding for moving left: %w", err)
	}

	if err := c.Cui.SetKeybinding("", gocui.KeyArrowRight, gocui.ModNone, c.moveViewRight); err != nil {
		return fmt.Errorf("error setting keybinding for moving right: %w", err)
	}

	if err := c.Cui.SetKeybinding("", 'a', gocui.ModNone, c.switchAccount); err != nil {
		return fmt.Errorf("error setting keybinding for switching account %w", err)
	}

	if err := c.Cui.SetKeybinding("", 'r', gocui.ModNone, c.switchRegion); err != nil {
		return fmt.Errorf("error setting keybinding for switching region %w", err)
	}

	if err := c.Cui.SetKeybinding("", 's', gocui.ModNone, c.scanAccount); err != nil {
		return fmt.Errorf("error setting keybinding for scanning account %w", err)
	}

	if err := c.Cui.SetKeybinding("", 's', gocui.ModNone, c.scanAccount); err != nil {
		return fmt.Errorf("error setting keybinding for scanning account %w", err)
	}

	return nil
}
