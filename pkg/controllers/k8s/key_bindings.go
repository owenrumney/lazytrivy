package k8s

import (
	"context"
	"fmt"

	"github.com/awesome-gocui/gocui"
)

func (c *Controller) configureKeyBindings() error {
	// Configure global key bindings first (includes 'q' for quit)
	if err := c.ConfigureGlobalKeyBindings(); err != nil {
		return fmt.Errorf("error configuring global keybindings: %w", err)
	}

	// Global key bindings for K8s mode
	if err := c.SetKeyBinding("", gocui.KeyCtrlC, gocui.ModNone, c.CancelCurrentScan); err != nil {
		return fmt.Errorf("failed to set cancel key binding: %w", err)
	}

	if err := c.Cui.SetKeybinding("", 't', gocui.ModNone, c.CancelCurrentScan); err != nil {
		return fmt.Errorf("error setting keybinding for cancelling current scan: %w", err)
	}

	// K8s specific key bindings
	if err := c.SetKeyBinding("", 's', gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		c.ScanCluster(context.Background())
		return nil
	}); err != nil {
		return fmt.Errorf("failed to set scan key binding: %w", err)
	}

	if err := c.SetKeyBinding("", gocui.KeyEsc, gocui.ModNone, c.BackToParent); err != nil {
		return fmt.Errorf("failed to set back key binding: %w", err)
	}

	// Context switching
	if err := c.SetKeyBinding("", 'c', gocui.ModNone, c.showContextChoice); err != nil {
		return fmt.Errorf("failed to set context choice key binding: %w", err)
	}

	// Left arrow to go back from results to tree
	if err := c.Cui.SetKeybinding("", gocui.KeyArrowLeft, gocui.ModNone, c.moveViewLeft); err != nil {
		return fmt.Errorf("failed to set left arrow key binding: %w", err)
	}

	// Right arrow to go back from tree to results
	if err := c.Cui.SetKeybinding("", gocui.KeyArrowRight, gocui.ModNone, c.moveViewRight); err != nil {
		return fmt.Errorf("failed to set left arrow key binding: %w", err)
	}

	return nil
}
