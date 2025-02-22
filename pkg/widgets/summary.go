package widgets

import (
	"errors"
	"fmt"
	"strings"

	"github.com/awesome-gocui/gocui"
	"github.com/liamg/tml"
	"github.com/owenrumney/lazytrivy/pkg/output"
)

type SummaryWidget struct {
	name  string
	x, y  int
	w, h  int
	issue output.Issue
}

func NewSummaryWidget(name string, x, y, w, h int, ctx baseContext, issue output.Issue) (*SummaryWidget, error) {

	// override the default keybindings
	_ = ctx.SetKeyBinding(Summary, 'a', gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error { return nil })

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

	return &SummaryWidget{name: name, x: x, y: y, w: w, h: h, issue: issue}, nil
}

func (i *SummaryWidget) Layout(g *gocui.Gui) error {
	if i.issue == nil {
		return fmt.Errorf("no issue to display")
	}

	switch i.issue.GetType() {
	case output.IssueTypeVulnerability:
		return i.layoutVulnerability(g)
	case output.IssueTypeMisconfiguration:
		return i.layoutMisconfiguration(g)
	case output.IssueTypeSecret:
		return i.layoutSecret(g)
	}

	return nil
}

func (i *SummaryWidget) layoutSecret(g *gocui.Gui) error {
	secret := i.issue
	var lines []string

	lines = printMultiline(secret.GetTitle(), "Title", lines, i.w-i.x-20)
	lines = printMultiline(secret.GetDescription(), "Description", lines, i.w-i.x-20)
	lines = printSingleLine(secret.GetSeverity(), "Severity", lines)
	lines = printSingleLine(secret.GetID(), "Misconfiguration ID", lines)
	lines = printSingleLine(secret.GetMatch(), "Match", lines)
	lines = printSingleLine(secret.GetDeleted(), "Deleted?", lines)

	v, err := g.SetView(i.name, i.x, i.y, i.w, i.h, 0)
	if err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return fmt.Errorf("failed to create view: %w", err)
		}
	}

	_, _ = fmt.Fprintln(v, strings.Join(lines, "\n"))
	v.Title = fmt.Sprintf(" Summary for %s ", secret.GetID())
	v.Subtitle = " Escape or 'q' to close "
	v.Wrap = true
	v.TitleColor = gocui.ColorGreen
	v.FrameColor = gocui.ColorGreen
	return nil
}

func (i *SummaryWidget) layoutMisconfiguration(g *gocui.Gui) error {
	misconfig := i.issue
	var lines []string
	cause := misconfig.GetCauseMetadata()

	lines = printMultiline(misconfig.GetTitle(), "Title", lines, i.w-i.x-20)
	lines = printMultiline(misconfig.GetDescription(), "Description", lines, i.w-i.x-20)
	if cause.Resource != "" {
		lines = printSingleLine(cause.Resource, "Resource", lines)
	}
	lines = printSingleLine(misconfig.GetSeverity(), "Severity", lines)
	lines = printMultiline(misconfig.GetMessage(), "Message", lines, i.w-i.x-20)
	lines = printMultiline(misconfig.GetResolution(), "Resolution", lines, i.w-i.x-20)
	lines = printSingleLine(misconfig.GetID(), "Misconfiguration ID", lines)

	if len(misconfig.GetReferences()) > 0 {
		lines = append(lines, tml.Sprintf("<green> References:</green>"))
		for _, reference := range misconfig.GetReferences() {
			lines = append(lines, tml.Sprintf("   <blue>%s</blue>", reference))
		}
	}

	if len(cause.Code.Lines) > 0 {
		lines = append(lines, tml.Sprintf("<green> Code:</green>"))
		for _, line := range cause.Code.Lines {
			if line.Truncated {
				lines = append(lines, tml.Sprintf("   %d: ...", line.Number))
			} else if line.IsCause {
				lines = append(lines, tml.Sprintf("   %d: <red>%s</red>", line.Number, line.Content))
			} else {
				lines = append(lines, tml.Sprintf("   %d: %s", line.Number, line.Content))
			}

		}
	}

	v, err := g.SetView(i.name, i.x, i.y, i.w, i.h, 0)
	if err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return fmt.Errorf("failed to create view: %w", err)
		}
	}

	_, _ = fmt.Fprintln(v, strings.Join(lines, "\n"))
	v.Title = fmt.Sprintf(" Summary for %s ", misconfig.GetID())
	v.Subtitle = " Escape or 'q' to close "
	v.Wrap = true
	v.TitleColor = gocui.ColorGreen
	v.FrameColor = gocui.ColorGreen
	return nil
}

func (i *SummaryWidget) layoutVulnerability(g *gocui.Gui) error {
	vulnerability := i.issue
	var lines []string

	lines = printMultiline(vulnerability.GetTitle(), "Title", lines, i.w-i.x-20)
	lines = printMultiline(vulnerability.GetTitle(), "Description", lines, i.w-i.x-20)
	lines = printSingleLine(vulnerability.GetID(), "Vulnerability ID", lines)

	if vulnerability.GetDatasourceName() != "" {
		lines = append(lines, tml.Sprintf("<green> DataSource:</green>\n   %s\n", vulnerability.GetDatasourceName()))
	}
	lines = printSingleLine(vulnerability.GetSeverity(), "Severity", lines)
	lines = printSingleLine(vulnerability.GetSeveritySource(), "Severity Source", lines)
	lines = printSingleLine(vulnerability.GetPackageName(), "Package Name", lines)
	lines = printSingleLine(vulnerability.GetPackagePath(), "Package Path", lines)
	lines = printSingleLine(vulnerability.GetInstalledVersion(), "Installed Version", lines)
	lines = printSingleLine(vulnerability.GetFixedVersion(), "Fixed Version", lines)
	if vulnerability.GetCVSS() != nil {
		for cvss, vals := range vulnerability.GetCVSS() {
			lines = append(lines, tml.Sprintf("<green> %s:</green>", cvss))
			if valsMap, ok := vals.(map[string]interface{}); ok {
				for k, v := range valsMap {
					lines = append(lines, tml.Sprintf("   %s: %v", k, v))
				}
			}
			lines = append(lines, "")
		}
	}

	if vulnerability.GetPrimaryURL() != "" {
		lines = append(lines, tml.Sprintf("<green> More Info:</green>\n   <blue>%s</blue>\n", vulnerability.GetPrimaryURL()))
	}
	if len(vulnerability.GetReferences()) > 0 {
		lines = append(lines, tml.Sprintf("<green> References:</green>"))
		for _, reference := range vulnerability.GetReferences() {
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
	v.Title = fmt.Sprintf(" Summary for %s ", vulnerability.GetID())
	v.Subtitle = " Escape or 'q' to close "
	v.Wrap = true
	v.TitleColor = gocui.ColorGreen
	v.FrameColor = gocui.ColorGreen
	return nil
}

func printSingleLine(source string, heading string, lines []string) []string {
	if source != "" {
		lines = append(lines, tml.Sprintf("<green> %s:</green>\n   %s\n", heading, source))
	}
	return lines
}

func printMultiline(source string, heading string, lines []string, maxLength int) []string {
	if source != "" {
		titleLines := prettyLines(source, maxLength)
		first := true
		for _, line := range titleLines {
			if first {
				lines = append(lines, tml.Sprintf("\n<green> %s:</green>\n   %s", heading, line))
				first = false
			} else {
				lines = append(lines, tml.Sprintf("  %s", line))
			}
		}
		lines = append(lines, "")
	}
	return lines
}
