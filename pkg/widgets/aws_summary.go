package widgets

import (
	"errors"
	"fmt"
	"strings"

	"github.com/awesome-gocui/gocui"
	"github.com/liamg/tml"
	"github.com/owenrumney/lazytrivy/pkg/output"
)

type AWSSummaryWidget struct {
	name             string
	x, y             int
	w, h             int
	misconfiguration output.Misconfiguration
}

func NewAWSSummaryWidget(name string, x, y, w, h int, ctx awsContext, vulnerability output.Misconfiguration) (*AWSSummaryWidget, error) {
	if err := ctx.SetKeyBinding(Remote, gocui.KeyEnter, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {

		gui.Mouse = true
		gui.Cursor = false

		if err := gui.DeleteView(Remote); err != nil {
			return fmt.Errorf("failed to delete view 'remote': %w", err)
		}
		if _, err := gui.SetCurrentView(Results); err != nil {
			return fmt.Errorf("failed to switch view to 'results': %w", err)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	if err := ctx.SetKeyBinding(Summary, gocui.KeyArrowDown, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		_, oy := view.Origin()
		_ = view.SetOrigin(0, oy+1)

		return nil
	}); err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	if err := ctx.SetKeyBinding(Summary, gocui.KeyArrowUp, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		_, oy := view.Origin()
		if oy > 0 {
			_ = view.SetOrigin(0, oy-1)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	if err := ctx.SetKeyBinding(Summary, gocui.KeyEsc, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		gui.Mouse = true
		gui.Cursor = false
		if _, err := gui.SetCurrentView(Results); err != nil {
			return err
		}
		return gui.DeleteView(Summary)
	}); err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	if err := ctx.SetKeyBinding(Summary, 'q', gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		gui.Mouse = true
		gui.Cursor = false
		if _, err := gui.SetCurrentView(Results); err != nil {
			return err
		}
		return gui.DeleteView(Summary)
	}); err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return &AWSSummaryWidget{name: name, x: x, y: y, w: w, h: h, misconfiguration: vulnerability}, nil
}

func (i *AWSSummaryWidget) Layout(g *gocui.Gui) error {
	misconfig := i.misconfiguration
	var lines []string

	lines = printMultiline(misconfig.Title, "Title", lines, i.w-i.x-20)
	lines = printMultiline(misconfig.Description, "Description", lines, i.w-i.x-20)
	lines = printSingleLine(misconfig.Severity, "Severity", lines)
	lines = printMultiline(misconfig.Message, "Message", lines, i.w-i.x-20)
	lines = printMultiline(misconfig.Resolution, "Resolution", lines, i.w-i.x-20)
	lines = printSingleLine(misconfig.ID, "Misconfiguration ID", lines)

	if len(misconfig.References) > 0 {
		lines = append(lines, tml.Sprintf("<green> References:</green>"))
		for _, reference := range misconfig.References {
			lines = append(lines, tml.Sprintf("   <blue>%s</blue>", reference))
		}
	}
	v, err := g.SetView(i.name, i.x, i.y, i.w, i.h, 0)
	if err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return fmt.Errorf("failed to create view: %w", err)
		}
	}

	_, _ = fmt.Fprintln(v, strings.Join(lines, "\n"))
	v.Title = fmt.Sprintf(" Summary for %s ", i.misconfiguration.ID)
	v.Subtitle = " Escape or 'q' to close "
	v.Wrap = true
	v.TitleColor = gocui.ColorGreen
	v.FrameColor = gocui.ColorGreen
	return nil
}
