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
	ListWidget
	name            string
	x, y            int
	w, h            int
	body            []string
	v               *gocui.View
	ctx             awsContext
	currentReport   *output.Report
	currentResult   *output.Result
	results         []*output.Result
	vulnerabilities []output.Misconfiguration
	resultIndex     int
	mode            ResultsMode
	imageWidth      int
	yPos            int
	yOrigin         int
	page            int
}

func NewAWSResultWidget(name string, g awsContext) *AWSResultWidget {
	widget := &AWSResultWidget{
		ListWidget: ListWidget{
			ctx: g,
		},
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

func (w *AWSResultWidget) ConfigureKeys() error {
	if err := w.configureListWidgetKeys(w.name); err != nil {
		panic(err)
	}

	if err := w.ctx.SetKeyBinding(w.name, gocui.KeyEnter, gocui.ModNone, w.diveDeeper); err != nil {
		return fmt.Errorf("failed to set keybinding: %w", err)
	}

	if err := w.ctx.SetKeyBinding(w.name, 's', gocui.ModNone, w.diveDeeper); err != nil {
		return fmt.Errorf("failed to set keybinding: %w", err)
	}

	if err := w.ctx.SetKeyBinding(w.name, 'b', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if w.results != nil && len(w.results) > 0 {
			w.UpdateResultsTable(w.currentReport)
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

func (w *AWSResultWidget) addFilteringKeyBinding(key rune, severity string) error {
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

func (w *AWSResultWidget) addFilteringKeyBindings() error {
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

func (w *AWSResultWidget) diveDeeper(g *gocui.Gui, v *gocui.View) error {
	switch w.mode {
	case SummaryResultMode:
		_, y := w.v.Cursor()
		w.currentResult = w.results[y-3]
		logger.Debug("Diving deeper into result: %s", w.currentResult.Target)
		w.GenerateFilteredReport("ALL")
	case DetailsResultMode:
		// x, y, wi, h := v.Dimensions()
		//
		// var result *output.Result
		// if w.resultIndex >= 0 && w.resultIndex < len(w.results) {
		// 	result = w.results[w.resultIndex]
		// } else {
		// 	return nil
		// }
		//
		// summary, err := NewSummaryWidget("summary", x+2, y+(h/2), wi-2, h-1, w.ctx, result)
		// if err != nil {
		// 	return err
		// }
		// g.Update(func(g *gocui.Gui) error {
		// 	if err := summary.Layout(g); err != nil {
		// 		return fmt.Errorf("failed to layout remote input: %w", err)
		// 	}
		// 	_, err := g.SetCurrentView("summary")
		// 	if err != nil {
		// 		return fmt.Errorf("failed to set current view: %w", err)
		// 	}
		// 	return nil
		// })
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

func (w *AWSResultWidget) UpdateResultsTable(report *output.Report) {
	w.mode = SummaryResultMode
	w.currentReport = report

	width, _ := w.v.Size()
	imageWidth := width - 50
	w.imageWidth = imageWidth
	var bodyContent []string //nolint:prealloc

	headers := []string{
		fmt.Sprintf("\n ARN% *s", width-53, ""),
		"   Critical",
		"   High",
		"   Medium",
		"   Low",
		"   Unknown ",
	}

	bodyContent = append(bodyContent, strings.Join(headers, ""))
	bodyContent = append(bodyContent, strings.Repeat("─", width))

	for _, result := range report.Results {
		severities := result.GetSeverityCounts()

		target, _ := truncateANSIString(result.Target, width-50)
		row := []string{
			fmt.Sprintf(" % -*s", width-50, target),
			tml.Sprintf("<bold><red>% 11d</red></bold>", severities["CRITICAL"]),
			tml.Sprintf("<red>% 7d</red>", severities["HIGH"]),
			tml.Sprintf("<yellow>% 9d</yellow>", severities["MEDIUM"]),
			tml.Sprintf("% 6d", severities["LOW"]),
			tml.Sprintf("% 10d ", severities["UNKNOWN"]),
		}
		bodyContent = append(bodyContent, strings.Join(row, ""))
		w.results = append(w.results, result)
	}

	w.body = bodyContent

	w.ctx.RefreshView(w.name)

	w.topMost = 3
	w.bottomMost = len(w.body)
	w.v.Subtitle = ""
}

func (w *AWSResultWidget) RenderReport(report *output.Report, severity string) {
	w.currentReport = report

	w.GenerateFilteredReport(severity)
}

func (w *AWSResultWidget) GenerateFilteredReport(severity string) {
	if w.currentResult == nil || len(w.currentResult.Misconfigurations) == 0 {
		return
	}

	w.mode = SummaryResultMode
	w.vulnerabilities = []output.Misconfiguration{}

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

	bodyContent := []string{""} //nolint:prealloc

	misconfigurations := w.currentResult.Misconfigurations

	sort.Slice(misconfigurations, func(i, j int) bool {
		return severityAsInt(misconfigurations[i].Severity) < severityAsInt(misconfigurations[j].Severity) //nolint:scopelint
	})

	for _, misconfig := range misconfigurations {

		f, b := colouredSeverity(misconfig.Severity)
		toPrint := fmt.Sprintf("  %s % -16s %s", tml.Sprintf(f+"% -10s"+b, misconfig.Severity),
			misconfig.ID, misconfig.Title)

		bodyContent = append(bodyContent, toPrint)
		w.vulnerabilities = append(w.vulnerabilities, misconfig)
	}

	w.body = bodyContent
	w.topMost = 3

	w.ctx.RefreshView(w.name)

	_ = w.v.SetCursor(0, w.topMost)
	w.page = 0
	w.yPos = 3
	w.yOrigin = 0
	w.resultIndex = 0
	w.v.Subtitle = fmt.Sprintf(" %s ", strings.Join(severities, " | "))
}

func severityAsInt(severity string) int {
	switch severity {
	case "CRITICAL":
		return 0
	case "HIGH":
		return 1
	case "MEDIUM":
		return 2
	case "LOW":
		return 3
	case "UNKNOWN":
		return 5
	default:
		return -1
	}
}

func colouredSeverity(severity string) (string, string) {
	switch severity {
	case "CRITICAL":
		return "<bold><red>", "</red></bold>"
	case "HIGH":
		return "<red>", "</red>"
	case "MEDIUM":
		return "<yellow>", "</yellow>"
	case "LOW":
		return "<blue>", "</blue>"
	default:
		return "", ""
	}
}

func (w *AWSResultWidget) RefreshView() {
	width, _ := w.v.Size()

	w.v.Clear()
	for _, line := range w.body {
		truncated, unencodedLength := truncateANSIString(line, width-1)
		if strings.HasPrefix(line, "─") {
			truncated = line
		}
		printer := fmt.Sprintf("%s%s", truncated, strings.Repeat(" ", width-unencodedLength))
		_, _ = fmt.Fprintln(w.v, printer)

	}

	w.v.SetCursor(0, w.topMost)
}

func (w *AWSResultWidget) CurrentReport() *output.Report {
	return w.currentReport
}
