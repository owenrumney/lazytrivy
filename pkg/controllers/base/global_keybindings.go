package base

import (
	"fmt"

	"github.com/awesome-gocui/gocui"
	"github.com/owenrumney/lazytrivy/pkg/logger"
)

func (c *Controller) ConfigureGlobalKeyBindings() error {
	logger.Debugf("Configuring global keyboard shortcuts")

	if err := c.Cui.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, Quit); err != nil {
		return fmt.Errorf("error setting keybinding for quitting with Ctrl+C: %w", err)
	}

	if err := c.Cui.SetKeybinding("", 'q', gocui.ModNone, Quit); err != nil {
		return fmt.Errorf("error setting keybinding for quitting: %w", err)
	}

	if err := c.Cui.SetKeybinding("", '?', gocui.ModNone, c.HelpFunc); err != nil {
		return fmt.Errorf("error setting keybinding for help: %w", err)
	}

	if err := c.Cui.SetKeybinding("", ',', gocui.ModNone, c.showSettings); err != nil {
		return fmt.Errorf("error setting keybinding for settings: %w", err)
	}

	if err := c.Cui.SetKeybinding("", 'l', gocui.ModNone, c.showLogOverlay); err != nil {
		return fmt.Errorf("error setting keybinding for log overlay: %w", err)
	}

	return nil
}
