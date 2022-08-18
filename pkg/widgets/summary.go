package widgets

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/awesome-gocui/gocui"
	"github.com/liamg/tml"
	"github.com/owenrumney/lazytrivy/pkg/output"
)

type SummaryWidget struct {
	name string
	x, y int
	w, h int
	vuln output.Vulnerability
}

func NewSummaryWidget(name string, x, y, w, h int, ctx ctx, vulnerability output.Vulnerability) (*SummaryWidget, error) {
	if err := ctx.SetKeyBinding(Remote, gocui.KeyEnter, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		if len(view.BufferLines()) > 0 {
			if image, _ := view.Line(0); image != "" {
				ctx.ScanImage(context.Background(), image)
			}
		}
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

	if err := ctx.SetKeyBinding("summary", gocui.KeyArrowDown, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		_, oy := view.Origin()
		_ = view.SetOrigin(0, oy+1)

		return nil
	}); err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	if err := ctx.SetKeyBinding("summary", gocui.KeyArrowUp, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		_, oy := view.Origin()
		if oy > 0 {
			_ = view.SetOrigin(0, oy-1)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	if err := ctx.SetKeyBinding("summary", gocui.KeyEsc, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		gui.Mouse = true
		gui.Cursor = false
		if _, err := gui.SetCurrentView(Results); err != nil {
			return err
		}
		return gui.DeleteView("summary")
	}); err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	if err := ctx.SetKeyBinding("summary", 'q', gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		gui.Mouse = true
		gui.Cursor = false
		if _, err := gui.SetCurrentView(Results); err != nil {
			return err
		}
		return gui.DeleteView("summary")
	}); err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return &SummaryWidget{name: name, x: x, y: y, w: w, h: h, vuln: vulnerability}, nil
}

func (i *SummaryWidget) Layout(g *gocui.Gui) error {
	vulnerability := i.vuln
	var lines []string

	lines = printMultiline(vulnerability.Title, "Title", lines, i.w-i.x-20)
	lines = printMultiline(vulnerability.Title, "Description", lines, i.w-i.x-20)
	lines = printSingleLine(vulnerability.VulnerabilityID, "Vulnerability ID", lines)

	if vulnerability.DataSource != nil && vulnerability.DataSource.Name != "" {
		lines = append(lines, tml.Sprintf("<green>DataSource:</green>\n  %s\n", vulnerability.DataSource.Name))
	}
	lines = printSingleLine(vulnerability.Severity, "Severity", lines)
	lines = printSingleLine(vulnerability.SeveritySource, "Severity Source", lines)
	lines = printSingleLine(vulnerability.PkgName, "Package Name", lines)
	lines = printSingleLine(vulnerability.PkgPath, "Package Path", lines)
	lines = printSingleLine(vulnerability.InstalledVersion, "Installed Version", lines)
	lines = printSingleLine(vulnerability.FixedVersion, "Fixed Version", lines)
	if vulnerability.CVSS != nil {
		for cvss, vals := range vulnerability.CVSS {
			lines = append(lines, tml.Sprintf("<green>%s:</green>", cvss))
			if valsMap, ok := vals.(map[string]interface{}); ok {
				for k, v := range valsMap {
					lines = append(lines, tml.Sprintf("  %s: %v", k, v))
				}
			}
			lines = append(lines, "")
		}
	}

	if vulnerability.PrimaryURL != "" {
		lines = append(lines, tml.Sprintf("<green>More Info:</green>\n  <blue>%s</blue>\n", vulnerability.PrimaryURL))
	}
	if len(vulnerability.References) > 0 {
		lines = append(lines, tml.Sprintf("<green>References:</green>"))
		for _, reference := range vulnerability.References {
			lines = append(lines, tml.Sprintf("  <blue>%s</blue>", reference))
		}
	}

	v, err := g.SetView(i.name, i.x, i.y, i.w, i.h, 0)
	if err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return fmt.Errorf("failed to create view: %w", err)
		}
	}

	_, _ = fmt.Fprintln(v, strings.Join(lines, "\n"))
	v.Title = fmt.Sprintf(" Summary for %s ", i.vuln.VulnerabilityID)
	v.Wrap = true
	v.TitleColor = gocui.ColorGreen
	v.FrameColor = gocui.ColorGreen
	return nil
}

func printSingleLine(source string, heading string, lines []string) []string {
	if source != "" {
		lines = append(lines, tml.Sprintf("<green>%s:</green>\n  %s\n", heading, source))
	}
	return lines
}

func printMultiline(source string, heading string, lines []string, maxLength int) []string {
	if source != "" {
		titleLines := prettyLines(source, maxLength)
		first := true
		for _, line := range titleLines {
			if first {
				lines = append(lines, tml.Sprintf("\n<green>%s:</green>\n  %s", heading, line))
				first = false
			} else {
				lines = append(lines, tml.Sprintf("  %s", line))
			}
		}
		lines = append(lines, "")
	}
	return lines
}

func prettyLines(input string, maxLength int) []string {
	var lines []string
	words := strings.Split(input, " ")

	line := ""

	for _, w := range words {
		if len(line)+len(w)+1 < maxLength {
			line += w + " "
		} else {
			lines = append(lines, line)
			line = w + " "
		}
	}
	if len(lines) == 0 && line != "" {
		lines = append(lines, line)
	}

	return lines
}
