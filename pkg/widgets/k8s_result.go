package widgets

import (
	"fmt"
	"strings"

	"github.com/awesome-gocui/gocui"
	"github.com/liamg/tml"
	"github.com/owenrumney/lazytrivy/pkg/logger"
	"github.com/owenrumney/lazytrivy/pkg/output"
)

type K8sResultWidget struct {
	ResultsWidget
	name string
	x, y int
	w, h int

	ctx           k8sContext
	currentResult *output.Result
	reports       []*output.Report
	targets       []string // Store the actual targets for deep dive
}

func NewK8sResultWidget(name string, g k8sContext) *K8sResultWidget {
	widget := &K8sResultWidget{
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

func (w *K8sResultWidget) ConfigureKeys(gui *gocui.Gui) error {
	if err := w.configureListWidgetKeys(w.name); err != nil {
		return err
	}

	if err := w.ctx.SetKeyBinding(w.name, gocui.KeyEnter, gocui.ModNone, w.diveDeeper); err != nil {
		return fmt.Errorf("failed to set keybinding: %w", err)
	}

	if err := w.ctx.SetKeyBinding(w.name, 'b', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		if len(w.reports) > 0 {
			w.UpdateResultsTable(w.reports, g)
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

func (w *K8sResultWidget) diveDeeper(g *gocui.Gui, v *gocui.View) error {
	logger.Debugf("K8s diveDeeper called, mode: %d", w.mode)

	switch w.mode {
	case SummaryResultMode:
		id := w.CurrentItemPosition()
		logger.Debugf("K8s diveDeeper: CurrentItemPosition returned: %d", id)
		logger.Debugf("K8s diveDeeper: targets length: %d", len(w.targets))
		logger.Debugf("K8s diveDeeper: body length: %d", len(w.body))

		if len(w.body) > 0 {
			logger.Debugf("K8s diveDeeper: body content (first 5 lines):")
			for i, line := range w.body {
				if i >= 5 {
					break
				}
				logger.Debugf("  [%d]: %s", i, line)
			}
		}

		if id < 0 || id >= len(w.targets) {
			logger.Debugf("K8s diveDeeper: Invalid ID %d, targets length: %d - returning early", id, len(w.targets))
			return nil
		}

		// Use the stored target to find the correct result
		targetToFind := w.targets[id]
		logger.Debugf("K8s diveDeeper: Looking for target: %s", targetToFind)

		// Find the report containing this target
		for reportIdx, report := range w.reports {
			for resultIdx, result := range report.Results {
				if result.Target == targetToFind {
					logger.Debugf("K8s diveDeeper: Found matching result at report[%d].result[%d]: %s", reportIdx, resultIdx, result.Target)
					w.currentResult = result
					w.GenerateFilteredReport("ALL", g)
					return nil
				}
			}
		}
		logger.Debugf("K8s diveDeeper: No matching result found for target: %s", targetToFind)

	case DetailsResultMode:
		logger.Debugf("K8s diveDeeper: DetailsResultMode, calling base DiveDeeper")
		// Use the common deep dive implementation from base ResultsWidget
		return w.DiveDeeper(g, v)
	}
	return nil
}

func (w *K8sResultWidget) RefreshView() {
	w.refreshView()
}

func (w *K8sResultWidget) GenerateFilteredReport(severity string, _ *gocui.Gui) {
	logger.Debugf("K8s GenerateFilteredReport called with severity: %s", severity)
	w.mode = DetailsResultMode
	w.issues = []output.Issue{}

	if w.currentResult == nil {
		logger.Debugf("K8s GenerateFilteredReport: currentResult is nil")
		w.body = []string{"No results to display"}
		w.RefreshView()
		return
	}

	logger.Debugf("K8s GenerateFilteredReport: currentResult target: %s", w.currentResult.Target)

	var severities []string
	severities = append(severities, "[B]ack", "[E]verything")
	resultSevs := w.currentResult.GetSeverityCounts()

	for _, sev := range []string{"CRITICAL", "HIGH", "MEDIUM", "LOW", "UNKNOWN"} {
		if count, ok := resultSevs[sev]; ok && count > 0 {
			severities = append(severities, fmt.Sprintf("[%s]%s", sev[:1], strings.ToLower(sev[1:])))
		}
	}

	width, _ := w.v.Size()
	var bodyContent []string

	// Header
	headers := []string{
		fmt.Sprintf(" %s", w.currentResult.Target),
	}

	bodyContent = append(bodyContent, "")
	bodyContent = append(bodyContent, strings.Join(headers, ""))
	bodyContent = append(bodyContent, strings.Repeat("─", width))

	// Get issues for the selected severity
	issues := w.currentResult.GetIssuesForSeverity(severity)
	logger.Debugf("K8s GenerateFilteredReport: Got %d issues for severity %s", len(issues), severity)

	// Deduplicate issues by ID
	seenIssues := make(map[string]bool)
	var uniqueIssues []output.Issue
	for _, issue := range issues {
		if !seenIssues[issue.GetID()] {
			seenIssues[issue.GetID()] = true
			uniqueIssues = append(uniqueIssues, issue)
		}
	}

	logger.Debugf("K8s GenerateFilteredReport: After deduplication: %d unique issues", len(uniqueIssues))
	issueCount := 0

	for _, issue := range uniqueIssues {
		f, b := colouredSeverity(issue.GetSeverity())
		toPrint := fmt.Sprintf("**%d***  %s %-16s %s", issueCount,
			tml.Sprintf(f+"%-12s"+b, issue.GetSeverity()), issue.GetID(), issue.GetTitle())

		bodyContent = append(bodyContent, toPrint)
		w.issues = append(w.issues, issue)
		issueCount++
	}

	logger.Debugf("K8s GenerateFilteredReport: Final issues array length: %d", len(w.issues))
	w.body = bodyContent
	w.bottomMost = len(w.body)
	w.SetStartPosition(3)
	if w.v != nil {
		w.v.Subtitle = fmt.Sprintf(" %s ", strings.Join(severities, " | "))
	}

	w.RefreshView()
}

func (w *K8sResultWidget) RenderReport(result *output.Result, report *output.Report, severity string) {
	w.currentResult = result
	w.currentReport = report
	w.GenerateFilteredReport(severity, nil)
}

func (w *K8sResultWidget) Layout(g *gocui.Gui) error {
	return w.layout(g, w.x, w.y, w.w, w.h)
}

func (w *K8sResultWidget) Reset() {
	w.v.Clear()
	w.v.Title = " K8s Results "
	w.v.Subtitle = ""
	w.body = []string{}
	w.currentPos = 0
	w.topMost = 0
	w.bottomMost = 0
}

func (w *K8sResultWidget) UpdateResultsTable(reports []*output.Report, _ *gocui.Gui) {
	logger.Debugf("K8s UpdateResultsTable called with %d reports", len(reports))

	w.reports = reports
	w.body = nil
	w.targets = nil
	w.mode = SummaryResultMode

	resultIndex := 0
	for reportIdx, report := range reports {
		logger.Debugf("K8s UpdateResultsTable: Processing report %d with %d results", reportIdx, len(report.Results))
		for resultIdx, result := range report.Results {
			if len(result.Issues) > 0 {
				// Format the line with **index*** prefix for CurrentItemPosition() to work
				formattedLine := fmt.Sprintf("**%d*** %s", resultIndex, result.Target)
				w.body = append(w.body, formattedLine)
				w.targets = append(w.targets, result.Target)
				logger.Debugf("K8s UpdateResultsTable: Added result[%d] -> targets[%d] = %s", resultIdx, resultIndex, result.Target)
				resultIndex++
			} else {
				logger.Debugf("K8s UpdateResultsTable: Skipping result[%d] %s (no issues)", resultIdx, result.Target)
			}
		}
	}

	logger.Debugf("K8s UpdateResultsTable: Final state - %d body items, %d targets", len(w.body), len(w.targets))
	w.RefreshView()
}

// Override CurrentItemPosition to handle K8s-specific logic
func (w *K8sResultWidget) CurrentItemPosition() int {
	logger.Debugf("K8s CurrentItemPosition: mode=%d, currentPos=%d, body length=%d", w.mode, w.currentPos, len(w.body))

	if w.mode == SummaryResultMode {
		// In summary mode, use the ListWidget's position directly but validate against targets
		if w.currentPos >= 0 && w.currentPos < len(w.targets) {
			logger.Debugf("K8s CurrentItemPosition: SummaryMode, returning currentPos: %d", w.currentPos)
			return w.currentPos
		} else {
			logger.Debugf("K8s CurrentItemPosition: SummaryMode, invalid currentPos %d for targets length %d", w.currentPos, len(w.targets))
			return -1
		}
	} else {
		// In details mode, use the parent logic
		parentPos := w.ListWidget.CurrentItemPosition()
		logger.Debugf("K8s CurrentItemPosition: DetailsMode, parent returned: %d", parentPos)
		return parentPos
	}
}
