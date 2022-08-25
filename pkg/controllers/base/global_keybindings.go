package base

import (
	"fmt"

	"github.com/awesome-gocui/gocui"
)

func (c *Controller) ConfigureGlobalKeyBindings() error {

	if err := c.Cui.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, Quit); err != nil {
		return fmt.Errorf("error setting keybinding for quitting with Ctrl+C: %w", err)
	}

	if err := c.Cui.SetKeybinding("", 'q', gocui.ModNone, Quit); err != nil {
		return fmt.Errorf("error setting keybinding for quitting: %w", err)
	}

	if err := c.Cui.SetKeybinding("", '?', gocui.ModNone, c.HelpFunc); err != nil {
		return fmt.Errorf("error setting keybinding for help: %w", err)
	}

	return nil
}
