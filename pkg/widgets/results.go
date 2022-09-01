package widgets

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/awesome-gocui/gocui"
	"github.com/liamg/tml"
	"github.com/owenrumney/lazytrivy/pkg/output"
)

type ResultsMode int

const (
	ArnResultMode = iota
	SummaryResultMode
	DetailsResultMode
)

type ResultsWidget struct {
	name               string
	x, y               int
	w, h               int
	body               []string
	v                  *gocui.View
	ctx                vulnerabilityContext
	currentReport      *output.Report
	reports            []*output.Report
	vulnerabilities    []output.Vulnerability
	vulnerabilityIndex int
	mode               ResultsMode
	imageWidth         int
	yPos               int
	yOrigin            int
	page               int
}

func NewResultsWidget(name string, g vulnerabilityContext) *ResultsWidget {
	widget := &ResultsWidget{
		name: name,
		x:    0,
		y:    0,
		w:    10,
		h:    10,
		v:    nil,
		body: []string{},
		ctx:  g,
	}

	return widget
}

func (w *ResultsWidget) ConfigureKeys() error {
	if err := w.ctx.SetKeyBinding(w.name, gocui.KeyArrowDown, gocui.ModNone, w.nextResult); err != nil {
		return fmt.Errorf("failed to set keybinding: %w", err)
	}

	if err := w.ctx.SetKeyBinding(w.name, gocui.KeyArrowUp, gocui.ModNone, w.previousResult); err != nil {
		return fmt.Errorf("failed to set keybinding: %w", err)
	}

	if err := w.ctx.SetKeyBinding(w.name, gocui.KeyEnter, gocui.ModNone, w.diveDeeper); err != nil {
		return fmt.Errorf("failed to set keybinding: %w", err)
	}

	if err := w.ctx.SetKeyBinding(w.name, 's', gocui.ModNone, w.diveDeeper); err != nil {
		return fmt.Errorf("failed to set keybinding: %w", err)
	}

	if err := w.ctx.SetKeyBinding(w.name, 'b', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if w.reports != nil && len(w.reports) > 0 {
			w.UpdateResultsTable(w.reports)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to set keybinding: %w", err)
	}

	err := w.addFilteringKeyBindings()
	if err != nil {
		return err
	}

	return nil
}

func (w *ResultsWidget) addFilteringKeyBinding(key rune, severity string) error {
	if err := w.ctx.SetKeyBinding(w.name, key, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		if w.currentReport == nil {
			return nil
		}
		switch severity {
		case "ALL":
			w.GenerateFilteredReport(severity)
		default:
			if w.currentReport.SeverityCount[severity] > 0 {
				w.GenerateFilteredReport(severity)
			}
		}

		return nil
	}); err != nil {
		return fmt.Errorf("failed to set keybinding: %w", err)
	}
	return nil
}

func (w *ResultsWidget) addFilteringKeyBindings() error {
	if err := w.addFilteringKeyBinding('e', "ALL"); err != nil {
		return err
	}
	if err := w.addFilteringKeyBinding('c', "CRITICAL"); err != nil {
		return err
	}
	if err := w.addFilteringKeyBinding('h', "HIGH"); err != nil {
		return err
	}
	if err := w.addFilteringKeyBinding('m', "MEDIUM"); err != nil {
		return err
	}
	if err := w.addFilteringKeyBinding('l', "LOW"); err != nil {
		return err
	}
	if err := w.addFilteringKeyBinding('u', "UNKNOWN"); err != nil {
		return err
	}

	return nil
}

func (w *ResultsWidget) diveDeeper(g *gocui.Gui, v *gocui.View) error {
	switch w.mode {
	case SummaryResultMode:
		_, y := w.v.Cursor()
		w.currentReport = w.reports[y-3]
		w.GenerateFilteredReport("ALL")
	case DetailsResultMode:
		x, y, wi, h := v.Dimensions()

		var vuln output.Vulnerability
		if w.vulnerabilityIndex >= 0 && w.vulnerabilityIndex < len(w.vulnerabilities) {
			vuln = w.vulnerabilities[w.vulnerabilityIndex]
		} else {
			return nil
		}

		summary, err := NewSummaryWidget("summary", x+2, y+(h/2), wi-2, h-1, w.ctx, vuln)
		if err != nil {
			return err
		}
		g.Update(func(g *gocui.Gui) error {
			if err := summary.Layout(g); err != nil {
				return fmt.Errorf("failed to layout remote input: %w", err)
			}
			_, err := g.SetCurrentView("summary")
			if err != nil {
				return fmt.Errorf("failed to set current view: %w", err)
			}
			return nil
		})
	}

	return nil
}

func (w *ResultsWidget) previousResult(_ *gocui.Gui, v *gocui.View) error {
	for {
		if w.yPos+w.yOrigin > 3 {
			v.MoveCursor(0, -1)
			_, y := v.Cursor()
			currentLine, err := v.Line(y + w.yOrigin)
			if err != nil {
				return err
			}
			if strings.TrimSpace(currentLine) != "" && !strings.HasPrefix(strings.TrimSpace(currentLine), "Target: ") {
				break
			}
		} else {
			return nil
		}
	}

	_, height := v.Size()
	if w.yPos == 0 {
		// we're at the bottom of the list
		w.page--
		w.yOrigin = w.page * height
		w.yPos = 0
		if w.page == 0 {
			w.yPos = height - 1
		}
	} else {
		_, w.yPos = v.Cursor()
	}
	w.vulnerabilityIndex--
	return nil
}

func (w *ResultsWidget) nextResult(_ *gocui.Gui, v *gocui.View) error {
	_, lastY := v.Cursor()
	for {
		v.MoveCursor(0, 1)
		_, y := v.Cursor()
		if y == lastY {
			return nil
		}
		lastY = y
		currentLine, err := v.Line(y + w.yOrigin)
		if err != nil {
			return nil //nolint:nilerr
		}
		if strings.TrimSpace(currentLine) != "" && !strings.HasPrefix(strings.TrimSpace(currentLine), "Target: ") {
			break
		}
	}

	_, w.yPos = v.Cursor()
	_, h := v.Size() //nolint:ifshort
	if w.yPos >= h {
		// we're at the bottom of the list
		w.page++
		w.yOrigin = w.page * h
		w.yPos = 0
		if w.page == 0 {
			w.yPos = 3
		}
	}

	w.vulnerabilityIndex++

	return nil
}

func (w *ResultsWidget) Layout(g *gocui.Gui) error {
	v, err := g.View(w.name)
	if err != nil {
		v, err = g.SetView(w.name, w.x, w.y, w.w, w.h, 0)
		if err != nil {
			if !errors.Is(err, gocui.ErrUnknownView) {
				return fmt.Errorf("%w", err)
			}
		}
	}

	v.Clear()
	v.Title = " Results "
	v.Highlight = true
	v.SelBgColor = gocui.ColorGreen
	v.SelFgColor = gocui.ColorBlack | gocui.AttrBold
	if g.CurrentView() == v {
		v.FrameColor = gocui.ColorGreen
	} else {
		v.FrameColor = gocui.ColorDefault
	}

	width, _ := v.Size()
	if w.v != nil {
		width, _ = w.v.Size()
	}

	w.v = v
	for _, line := range w.body {
		switch w.mode {
		case DetailsResultMode:
			truncated, unencodedLength := truncateANSIString(line, width-1)
			printer := fmt.Sprintf("%s%s", truncated, strings.Repeat(" ", width-unencodedLength))
			_, _ = fmt.Fprintln(v, printer)
		case SummaryResultMode, ArnResultMode:
			_, _ = fmt.Fprintln(v, line)
		}
	}

	_ = v.SetCursor(0, w.yPos)
	_ = v.SetOrigin(0, w.yOrigin)

	return nil
}

func (w *ResultsWidget) Reset() {
	w.v.Clear()
	w.v.Title = " Results "

	w.v.Subtitle = ""
	if err := w.v.SetOrigin(0, 0); err != nil {
		panic(err)
	}
}

func (w *ResultsWidget) UpdateResultsTable(reports []*output.Report) {
	w.mode = SummaryResultMode
	w.reports = reports

	width, _ := w.v.Size()
	imageWidth := width - 50
	w.imageWidth = imageWidth
	var bodyContent []string //nolint:prealloc

	headers := []string{
		fmt.Sprintf("\n Image% *s", width-55, ""),
		"   Critical",
		"   High",
		"   Medium",
		"   Low",
		"   Unknown ",
	}

	bodyContent = append(bodyContent, strings.Join(headers, ""))
	bodyContent = append(bodyContent, strings.Repeat("─", width))

	for _, report := range reports {
		row := []string{
			fmt.Sprintf(" % -*s", width-50, report.ImageName),
			tml.Sprintf("<bold><red>% 11d</red></bold>", report.SeverityCount["CRITICAL"]),
			tml.Sprintf("<red>% 7d</red>", report.SeverityCount["HIGH"]),
			tml.Sprintf("<yellow>% 9d</yellow>", report.SeverityCount["MEDIUM"]),
			tml.Sprintf("% 6d", report.SeverityCount["LOW"]),
			tml.Sprintf("% 10d ", report.SeverityCount["UNKNOWN"]),
		}
		bodyContent = append(bodyContent, strings.Join(row, ""))
	}

	w.body = bodyContent
	w.ctx.RefreshView(w.name)

	w.yPos = 3
	w.page = 0
	w.yOrigin = 0
	w.v.Subtitle = ""
}

func (w *ResultsWidget) RenderReport(report *output.Report, severity string) {
	w.currentReport = report

	w.GenerateFilteredReport(severity)
}

func (w *ResultsWidget) GenerateFilteredReport(severity string) {
	if w.currentReport == nil {
		return
	}
	w.mode = DetailsResultMode
	w.vulnerabilities = []output.Vulnerability{}

	var severities []string
	if w.reports != nil && len(w.reports) > 0 {
		severities = append(severities, "[B]ack")
	}
	severities = append(severities, "[E]verything")

	for _, sev := range []string{"CRITICAL", "HIGH", "MEDIUM", "LOW", "UNKNOWN"} {
		if count, ok := w.currentReport.SeverityCount[sev]; ok {
			if count == 0 {
				continue
			}
			severities = append(severities, fmt.Sprintf("[%s]%s", sev[:1], strings.ToLower(sev[1:])))
		}
	}

	var results []*output.Result
	if severity != "ALL" {
		results = w.currentReport.SeverityMap[severity]
	} else {
		results = w.currentReport.Results
	}

	var bodyContent []string //nolint:prealloc

	for _, result := range results {
		if len(result.Vulnerabilities) == 0 {
			continue
		}

		bodyContent = append(bodyContent, tml.Sprintf("\n  <bold>Target:</bold> <blue>%s</blue>\n", result.Target))

		sort.Slice(result.Vulnerabilities, func(i, j int) bool {
			return severityAsInt(result.Vulnerabilities[i].Severity) < severityAsInt(result.Vulnerabilities[j].Severity) //nolint:scopelint
		})

		for _, v := range result.Vulnerabilities {
			f, b := colouredSeverity(v.Severity)
			toPrint := fmt.Sprintf("  %s % -16s %s", tml.Sprintf(f+"% -10s"+b, v.Severity),
				v.VulnerabilityID, v.Title)

			bodyContent = append(bodyContent, toPrint)
			w.vulnerabilities = append(w.vulnerabilities, v)
		}
	}

	w.body = bodyContent

	w.ctx.RefreshView(w.name)
	w.page = 0
	w.yPos = 3
	w.yOrigin = 0
	w.vulnerabilityIndex = 0
	w.v.Subtitle = fmt.Sprintf(" %s ", strings.Join(severities, " | "))
}

func (w *ResultsWidget) RefreshView() {
}

func (w *ResultsWidget) CurrentReport() *output.Report {
	return w.currentReport
}
