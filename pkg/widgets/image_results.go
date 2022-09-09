package widgets

import (
	"fmt"
	"sort"
	"strings"

	"github.com/awesome-gocui/gocui"
	"github.com/liamg/tml"
	"github.com/owenrumney/lazytrivy/pkg/output"
)

type ImageResultWidget struct {
	ResultsWidget
	name string
	x, y int
	w, h int

	ctx             vulnerabilityContext
	reports         []*output.Report
	vulnerabilities []output.Vulnerability
}

func NewImageResultWidget(name string, g vulnerabilityContext) *ImageResultWidget {
	widget := &ImageResultWidget{
		name: name,
		x:    0,
		y:    0,
		w:    10,
		h:    10,
		ctx:  g,
	}

	widget.ResultsWidget = NewResultsWidget(name, widget.GenerateFilteredReport, widget.UpdateResultsTable, g)
	return widget
}

func (w *ImageResultWidget) ConfigureKeys(g *gocui.Gui) error {
	if err := w.configureListWidgetKeys(w.name); err != nil {
		return err
	}

	if err := w.ctx.SetKeyBinding(w.name, gocui.KeyEnter, gocui.ModNone, w.diveDeeper); err != nil {
		return fmt.Errorf("failed to set keybinding: %w", err)
	}

	if err := w.ctx.SetKeyBinding(w.name, 's', gocui.ModNone, w.diveDeeper); err != nil {
		return fmt.Errorf("failed to set keybinding: %w", err)
	}

	if err := w.ctx.SetKeyBinding(w.name, 'b', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if w.reports != nil && len(w.reports) > 0 {
			w.UpdateResultsTable(w.reports, g)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to set keybinding: %w", err)
	}

	if err := w.addFilteringKeyBindings(nil); err != nil {
		return err
	}

	return nil
}

func (w *ImageResultWidget) diveDeeper(g *gocui.Gui, v *gocui.View) error {
	switch w.mode {
	case SummaryResultMode:
		id := w.CurrentItemPosition()
		if id >= 0 && id < len(w.reports) {
			w.currentReport = w.reports[id]
		} else {
			return nil
		}

		w.GenerateFilteredReport("ALL", g)
	case DetailsResultMode:
		x, y, wi, h := v.Dimensions()

		var vuln output.Vulnerability
		id := w.CurrentItemPosition()
		if id >= 0 && id < len(w.vulnerabilities) {
			vuln = w.vulnerabilities[id]
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

func (w *ImageResultWidget) UpdateResultsTable(reports []*output.Report, g *gocui.Gui) {
	w.mode = SummaryResultMode
	w.reports = reports

	width, _ := w.v.Size()
	var bodyContent []string //nolint:prealloc
	var reportCount int

	headers := []string{
		fmt.Sprintf("\n Image% *s", width-55, ""),
		"   Critical",
		"   High",
		"   Medium",
		"   Low",
		"   Unknown ",
	}

	bodyContent = append(bodyContent, " ")
	bodyContent = append(bodyContent, strings.Join(headers, ""))
	bodyContent = append(bodyContent, strings.Repeat("─", width))

	for _, report := range reports {
		row := []string{
			fmt.Sprintf("**%d*** % -*s", reportCount, width-50, report.ImageName),
			tml.Sprintf("<bold><red>% 11d</red></bold>", report.SeverityCount["CRITICAL"]),
			tml.Sprintf("<red>% 7d</red>", report.SeverityCount["HIGH"]),
			tml.Sprintf("<yellow>% 9d</yellow>", report.SeverityCount["MEDIUM"]),
			tml.Sprintf("% 6d", report.SeverityCount["LOW"]),
			tml.Sprintf("% 10d ", report.SeverityCount["UNKNOWN"]),
		}
		bodyContent = append(bodyContent, strings.Join(row, ""))
		reportCount++
	}

	w.body = bodyContent

	w.ctx.RefreshView(w.name)

	w.SetStartPosition(3)
	w.bottomMost = len(w.body)
	w.v.Subtitle = ""
}

func (w *ImageResultWidget) RenderReport(report *output.Report, severity string, cui *gocui.Gui) {
	w.currentReport = report
	w.v.Clear()
	w.body = []string{}

	if w.currentReport == nil || !w.currentReport.HasIssues() {
		width, height := w.v.Size()

		lines := []string{
			"Great News!",
			"",
			"No vulnerabilities found!",
		}

		announcement := NewAnnouncementWidget(Announcement, "No Results", width, height, lines, cui)
		_ = announcement.Layout(cui)
		_, _ = cui.SetCurrentView(Announcement)

		return
	}

	w.GenerateFilteredReport(severity, cui)

	_, _ = cui.SetCurrentView(Results)
}

func (w *ImageResultWidget) GenerateFilteredReport(severity string, g *gocui.Gui) {
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
	vulnCounter := 0

	for _, result := range results {
		if len(result.Vulnerabilities) == 0 {
			continue
		}

		bodyContent = append(bodyContent, "")
		bodyContent = append(bodyContent, tml.Sprintf("<bold>Target:</bold> <blue>%s</blue>", result.Target))
		bodyContent = append(bodyContent, "")

		sort.Slice(result.Vulnerabilities, func(i, j int) bool {
			return severityAsInt(result.Vulnerabilities[i].Severity) < severityAsInt(result.Vulnerabilities[j].Severity) //nolint:scopelint
		})

		for _, v := range result.Vulnerabilities {
			f, b := colouredSeverity(v.Severity)
			toPrint := fmt.Sprintf("**%d***  %s % -16s %s", vulnCounter, tml.Sprintf(f+"% -10s"+b, v.Severity),
				v.VulnerabilityID, v.Title)

			bodyContent = append(bodyContent, toPrint)
			w.vulnerabilities = append(w.vulnerabilities, v)
			vulnCounter++
		}
	}

	w.body = bodyContent
	w.SetStartPosition(3)
	w.bottomMost = len(w.body)

	w.ctx.RefreshView(w.name)
	_ = w.v.SetCursor(0, w.topMost)
	w.v.Subtitle = fmt.Sprintf(" %s ", strings.Join(severities, " | "))
}

func (w *ImageResultWidget) RefreshView() {
	w.refreshView()
}

func (w *ImageResultWidget) Layout(g *gocui.Gui) error {
	return w.layout(g, w.x, w.y, w.w, w.h)
}
