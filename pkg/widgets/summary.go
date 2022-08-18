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
	ctx  ctx
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

	if vulnerability.Title != "" {
		titleLines := i.prettyLines(vulnerability.Title, i.w-i.x-20)
		first := true
		for _, line := range titleLines {
			if first {
				lines = append(lines, tml.Sprintf("\n<bold>Title:</bold>              %s", line))
				first = false
			} else {
				lines = append(lines, tml.Sprintf("                     %s", line))
			}
		}

		lines = append(lines, "\n")
	}

	if vulnerability.Description != "" {
		descriptionLines := i.prettyLines(vulnerability.Description, i.w-i.x-20)
		first := true
		for _, line := range descriptionLines {
			if first {
				lines = append(lines, tml.Sprintf("<bold>Description:</bold>        %s", line))
				first = false
			} else {
				lines = append(lines, tml.Sprintf("                    %s", line))
			}
		}
		lines = append(lines, "\n")
	}
	lines = append(lines, tml.Sprintf("<bold>Vulnerability ID:</bold>   %s\n", vulnerability.VulnerabilityID))
	lines = append(lines, tml.Sprintf("<bold>Severity:</bold>           %s\n", vulnerability.Severity))
	lines = append(lines, tml.Sprintf("<bold>Package Name:</bold>       %s\n", vulnerability.PkgName))
	lines = append(lines, tml.Sprintf("<bold>Installed Version:</bold>  %s\n", vulnerability.InstalledVersion))
	if vulnerability.FixedVersion != "" {
		lines = append(lines, tml.Sprintf("<bold>Fixed Version:</bold>      %s\n", vulnerability.FixedVersion))
	}
	lines = append(lines, tml.Sprintf("<bold>More Info:</bold>          <blue>%s</blue>\n", vulnerability.PrimaryURL))

	v, err := g.SetView(i.name, i.x, i.y, i.w, i.h, 0)
	if err != nil {
		if !errors.Is(err, gocui.ErrUnknownView) {
			return fmt.Errorf("failed to create view: %w", err)
		}
	}

	fmt.Fprintln(v, strings.Join(lines, "\n"))
	v.Title = fmt.Sprintf(" Summary for %s ", i.vuln.VulnerabilityID)
	v.Wrap = true
	v.TitleColor = gocui.ColorGreen
	v.FrameColor = gocui.ColorGreen
	return nil
}

func (i *SummaryWidget) prettyLines(input string, maxLength int) []string {
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
	return lines
}
