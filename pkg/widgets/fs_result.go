package widgets

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/awesome-gocui/gocui"
	"github.com/liamg/tml"
	"github.com/owenrumney/lazytrivy/pkg/logger"
	"github.com/owenrumney/lazytrivy/pkg/output"
)

type FSResultWidget struct {
	ResultsWidget
	name string
	x, y int
	w, h int

	ctx           fsContext
	currentResult *output.Result
	results       []*output.Result
}

func NewFSResultWidget(name string, g fsContext) *FSResultWidget {
	widget := &FSResultWidget{
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

func (w *FSResultWidget) ConfigureKeys(gui *gocui.Gui) error {
	if err := w.configureListWidgetKeys(w.name); err != nil {
		return err
	}

	if err := w.ctx.SetKeyBinding(w.name, gocui.KeyEnter, gocui.ModNone, w.diveDeeper); err != nil {
		return fmt.Errorf("failed to set keybinding: %w", err)
	}

	if err := w.ctx.SetKeyBinding(w.name, 'b', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if len(w.results) > 0 {
			w.UpdateResultsTable([]*output.Report{w.currentReport}, g)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to set keybinding: %w", err)
	}

	if err := w.addFilteringKeyBindings(gui); err != nil {
		return err
	}

	return nil
}

func (w *FSResultWidget) diveDeeper(g *gocui.Gui, v *gocui.View) error {
	switch w.mode {
	case SummaryResultMode:
		id := w.CurrentItemPosition()
		if id < 0 || id > len(w.results) {
			return nil
		}
		w.currentResult = w.results[id]
		logger.Debugf("Diving deeper into result: %s", w.currentResult.Target)
		w.GenerateFilteredReport("ALL", g)
	case DetailsResultMode:
		// Use the common deep dive implementation from base ResultsWidget
		return w.DiveDeeper(g, v)
	}

	return nil
}

func (w *FSResultWidget) Layout(g *gocui.Gui) error {
	v, err := g.View(w.name)
	if err != nil {
		v, err = g.SetView(w.name, w.x, w.y, w.w, w.h, 0)
		if err != nil {
			if !errors.Is(err, gocui.ErrUnknownView) {
				return fmt.Errorf("%w", err)
			}
		}
	}

	w.v = v
	v.Title = " Results "
	v.Highlight = true
	v.SelBgColor = gocui.ColorDefault
	v.SelFgColor = gocui.ColorBlue | gocui.AttrBold
	if g.CurrentView() == v {
		v.FrameColor = gocui.ColorBlue
	} else {
		v.FrameColor = gocui.ColorDefault
	}
	v.FrameRunes = []rune{'─', '│', '╭', '╮', '╰', '╯'}
	return nil
}

func (w *FSResultWidget) Reset() {
	w.v.Clear()
	w.v.Title = " Results "

	w.v.Subtitle = ""
	if err := w.v.SetOrigin(0, 0); err != nil {
		panic(err)
	}
}

func (w *FSResultWidget) UpdateResultsTable(reports []*output.Report, g *gocui.Gui) {
	if len(reports) == 0 {
		return
	}

	w.mode = SummaryResultMode
	w.currentReport = reports[0]
	w.currentReport.Process()
	w.v.Clear()
	w.body = []string{}

	if w.currentReport == nil || !w.currentReport.HasIssues() {

		lines := []string{
			"Great News!",
			"",
			"No misconfigurations found!",
		}

		announcement := NewAnnouncementWidget(Announcement, "No Results", lines, g)
		_ = announcement.Layout(g)
		_, _ = g.SetCurrentView(Announcement)

		return
	}

	width, _ := w.v.Size()

	var bodyContent []string //nolint:prealloc

	headers := []string{
		fmt.Sprintf("\n ARN% *s", width-53, ""),
		"   Critical",
		"   High",
		"   Medium",
		"   Low",
		"   Unknown ",
	}

	bodyContent = append(bodyContent, "")
	bodyContent = append(bodyContent, strings.Join(headers, ""))
	bodyContent = append(bodyContent, strings.Repeat("─", width))
	resultCount := 0
	for _, result := range w.currentReport.Results {
		severities := result.GetSeverityCounts()

		target, _ := truncateANSIString(result.Target, width-50)
		row := []string{
			fmt.Sprintf("**%d*** % -*s", resultCount, width-50, target),
			tml.Sprintf("<bold><red>% 11d</red></bold>", severities["CRITICAL"]),
			tml.Sprintf("<red>% 7d</red>", severities["HIGH"]),
			tml.Sprintf("<yellow>% 9d</yellow>", severities["MEDIUM"]),
			tml.Sprintf("% 6d", severities["LOW"]),
			tml.Sprintf("% 10d ", severities["UNKNOWN"]),
		}
		bodyContent = append(bodyContent, strings.Join(row, ""))
		w.results = append(w.results, result)
		resultCount++
	}

	w.body = bodyContent

	_, _ = g.SetCurrentView(Results)
	w.ctx.RefreshView(w.name)

	w.SetStartPosition(3)
	w.bottomMost = len(w.body)
	w.v.Subtitle = ""
}

func (w *FSResultWidget) RenderReport(result *output.Result, report *output.Report, severity string) {
	w.currentResult = result
	w.currentReport = report

	w.GenerateFilteredReport(severity, nil)
}

func (w *FSResultWidget) GenerateFilteredReport(severity string, _ *gocui.Gui) {

	w.mode = DetailsResultMode
	w.issues = []output.Issue{}

	var severities []string
	if len(w.results) > 0 {
		severities = append(severities, "[B]ack")
	}
	severities = append(severities, "[E]verything")
	resultSevs := w.currentResult.GetSeverityCounts()

	for _, sev := range []string{"CRITICAL", "HIGH", "MEDIUM", "LOW", "UNKNOWN"} {
		if count, ok := resultSevs[sev]; ok {
			if count == 0 {
				continue
			}
			severities = append(severities, fmt.Sprintf("[%s]%s", sev[:1], strings.ToLower(sev[1:])))
		}
	}

	width, _ := w.v.Size()

	var bodyContent []string //nolint:prealloc

	headers := []string{
		fmt.Sprintf(" %s", w.currentResult.Target),
	}

	bodyContent = append(bodyContent, "")
	bodyContent = append(bodyContent, strings.Join(headers, ""))
	bodyContent = append(bodyContent, strings.Repeat("─", width))

	misconfigurationCount := 0
	misconfigurations := w.currentResult.GetIssuesForSeverity(severity)

	// Deduplicate issues by ID
	seenIssues := make(map[string]bool)
	var uniqueIssues []output.Issue
	for _, issue := range misconfigurations {
		if !seenIssues[issue.GetID()] {
			seenIssues[issue.GetID()] = true
			uniqueIssues = append(uniqueIssues, issue)
		}
	}

	sort.Slice(uniqueIssues, func(i, j int) bool {
		return severityAsInt(uniqueIssues[i].GetSeverity()) < severityAsInt(uniqueIssues[j].GetSeverity()) //nolint:scopelint
	})

	for _, issue := range uniqueIssues {
		// Use the common formatting method
		line := w.FormatIssueLine(misconfigurationCount, issue)
		bodyContent = append(bodyContent, line)
		w.issues = append(w.issues, issue)
		misconfigurationCount++
	}

	w.body = bodyContent

	w.ctx.RefreshView(w.name)
	w.bottomMost = len(w.body)
	w.SetStartPosition(3)
	w.v.Subtitle = fmt.Sprintf(" %s ", strings.Join(severities, " | "))
}

func (w *FSResultWidget) RefreshView() {
	w.refreshView()
}

func (w *FSResultWidget) CurrentReport() *output.Report {
	return w.currentReport
}
