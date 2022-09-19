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

type AWSResultWidget struct {
	ResultsWidget
	name string
	x, y int
	w, h int

	ctx           awsContext
	currentResult *output.Result
	results       []*output.Result
	issues        []output.Issue
}

func NewAWSResultWidget(name string, g awsContext) *AWSResultWidget {
	widget := &AWSResultWidget{
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

func (w *AWSResultWidget) ConfigureKeys(gui *gocui.Gui) error {
	if err := w.configureListWidgetKeys(w.name); err != nil {
		return err
	}

	if err := w.ctx.SetKeyBinding(w.name, gocui.KeyEnter, gocui.ModNone, w.diveDeeper); err != nil {
		return fmt.Errorf("failed to set keybinding: %w", err)
	}

	if err := w.ctx.SetKeyBinding(w.name, 'b', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if w.results != nil && len(w.results) > 0 {
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

func (w *AWSResultWidget) diveDeeper(g *gocui.Gui, _ *gocui.View) error {
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
		x, y, wi, h := w.v.Dimensions()

		var issue output.Issue
		id := w.CurrentItemPosition()
		if id >= 0 && id < len(w.issues) {
			issue = w.issues[id]
		} else {
			return nil
		}

		summary, err := NewSummaryWidget("summary", x+2, y+(h/2), wi-2, h-1, w.ctx, issue)

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

func (w *AWSResultWidget) Layout(g *gocui.Gui) error {
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
	v.SelBgColor = gocui.ColorGreen
	v.SelFgColor = gocui.ColorBlack | gocui.AttrBold
	if g.CurrentView() == v {
		v.FrameColor = gocui.ColorGreen
	} else {
		v.FrameColor = gocui.ColorDefault
	}
	return nil
}

func (w *AWSResultWidget) Reset() {
	w.v.Clear()
	w.v.Title = " Results "

	w.v.Subtitle = ""
	if err := w.v.SetOrigin(0, 0); err != nil {
		panic(err)
	}
}

func (w *AWSResultWidget) UpdateResultsTable(reports []*output.Report, g *gocui.Gui) {
	if len(reports) == 0 {
		return
	}

	w.mode = SummaryResultMode
	w.currentReport = reports[0]
	w.v.Clear()
	w.body = []string{}

	if w.currentReport == nil || !w.currentReport.HasIssues() {
		width, height := w.v.Size()

		lines := []string{
			"Great News!",
			"",
			"No misconfigurations found!",
		}

		announcement := NewAnnouncementWidget(Announcement, "No Results", width, height, lines, g, Services)
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

func (w *AWSResultWidget) RenderReport(report *output.Report, severity string) {
	w.currentReport = report

	w.GenerateFilteredReport(severity, nil)
}

func (w *AWSResultWidget) GenerateFilteredReport(severity string, g *gocui.Gui) {
	if w.currentResult == nil || len(w.currentResult.Issues) == 0 {
		width, height := w.v.Size()

		lines := []string{
			"Great News!",
			"",
			"No misconfigurations found!",
		}

		announcement := NewAnnouncementWidget(Announcement, "No Results", width, height, lines, g)
		_ = announcement.Layout(g)
		_, _ = g.SetCurrentView(Announcement)

		return
	}

	w.mode = DetailsResultMode
	w.issues = []output.Issue{}

	var severities []string
	if w.results != nil && len(w.results) > 0 {
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
	issues := w.currentResult.GetIssuesForSeverity(severity)

	sort.Slice(issues, func(i, j int) bool {
		return severityAsInt(issues[i].GetSeverity()) < severityAsInt(issues[j].GetSeverity()) //nolint:scopelint
	})

	for _, issue := range issues {

		f, b := colouredSeverity(issue.GetSeverity())
		toPrint := fmt.Sprintf("**%d***  %s % -16s %s", misconfigurationCount, tml.Sprintf(f+"% -10s"+b, issue.GetSeverity()),
			issue.GetID(), issue.GetTitle())

		bodyContent = append(bodyContent, toPrint)
		w.issues = append(w.issues, issue)
		misconfigurationCount++
	}

	w.body = bodyContent

	w.ctx.RefreshView(w.name)

	w.SetStartPosition(3)
	w.v.Subtitle = fmt.Sprintf(" %s ", strings.Join(severities, " | "))
}

func (w *AWSResultWidget) RefreshView() {
	w.refreshView()
}

func (w *AWSResultWidget) CurrentReport() *output.Report {
	return w.currentReport
}
