package widgets

import (
	"errors"
	"fmt"
	"strings"

	"github.com/awesome-gocui/gocui"
	"github.com/liamg/tml"
	"github.com/owenrumney/lazytrivy/pkg/logger"
	"github.com/owenrumney/lazytrivy/pkg/output"
)

type ResultsMode int

const (
	SummaryResultMode ResultsMode = iota
	DetailsResultMode
)

type ResultsWidget struct {
	ListWidget
	name string

	generateReportFunc     func(severity string, gui *gocui.Gui)
	updateResultsTableFunc func(reports []*output.Report, g *gocui.Gui)
	ctx                    baseContext
	currentReport          *output.Report
	mode                   ResultsMode
	v                      *gocui.View

	// Common fields for deep dive functionality
	issues []output.Issue // Common issue storage
}

func NewResultsWidget(name string, generateReportFunc func(severity string, gui *gocui.Gui),
	updateResultsTableFunc func(reports []*output.Report, g *gocui.Gui), g baseContext) ResultsWidget {
	widget := ResultsWidget{
		ListWidget: ListWidget{
			ctx:                 g,
			body:                []string{},
			selectionChangeFunc: g.SetSelected,
		},
		name:                   name,
		generateReportFunc:     generateReportFunc,
		updateResultsTableFunc: updateResultsTableFunc,
		ctx:                    g,
	}

	return widget
}

func (w *ResultsWidget) addFilteringKeyBinding(key rune, severity string, _ *gocui.Gui) error {
	if err := w.ctx.SetKeyBinding(w.name, key, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		if w.currentReport == nil {
			return nil
		}
		switch severity {
		case "ALL":
			w.generateReportFunc(severity, gui)
		default:
			if w.currentReport.SeverityCount[severity] > 0 {
				w.generateReportFunc(severity, gui)
			}
		}

		return nil
	}); err != nil {
		return fmt.Errorf("failed to set keybinding: %w", err)
	}
	return nil
}

func (w *ResultsWidget) addFilteringKeyBindings(gui *gocui.Gui) error {
	logger.Debugf("adding filtering keybindings")
	if err := w.addFilteringKeyBinding('e', "ALL", gui); err != nil {
		return err
	}
	if err := w.addFilteringKeyBinding('c', "CRITICAL", nil); err != nil {
		return err
	}
	if err := w.addFilteringKeyBinding('h', "HIGH", nil); err != nil {
		return err
	}
	if err := w.addFilteringKeyBinding('m', "MEDIUM", nil); err != nil {
		return err
	}
	if err := w.addFilteringKeyBinding('l', "LOW", nil); err != nil {
		return err
	}
	if err := w.addFilteringKeyBinding('u', "UNKNOWN", nil); err != nil {
		return err
	}

	return nil
}

func (w *ResultsWidget) layout(g *gocui.Gui, x int, y int, wi int, h int) error {

	v, err := g.View(w.name)
	if err != nil {
		v, err = g.SetView(w.name, x, y, wi, h, 0)
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

func (w *ResultsWidget) refreshView() {
	width, _ := w.v.Size()

	w.v.Clear()
	for _, line := range w.body {
		line := stripIdentifierPrefix(line)
		truncated, unencodedLength := truncateANSIString(line, width-1)
		if strings.HasPrefix(line, "─") {
			truncated = line
		}
		printer := fmt.Sprintf("%s%s", truncated, strings.Repeat(" ", width-unencodedLength))
		_, _ = fmt.Fprintln(w.v, printer)

	}

	_ = w.v.SetCursor(0, w.topMost)
}

func (w *ResultsWidget) CurrentReport() *output.Report {
	return w.currentReport
}

// Common deep dive implementation for DetailsResultMode
func (w *ResultsWidget) DiveDeeper(g *gocui.Gui, v *gocui.View) error {
	logger.Debugf("Base DiveDeeper called, mode: %d", w.mode)
	if w.mode == DetailsResultMode {
		id := w.CurrentItemPosition()
		logger.Debugf("Base DiveDeeper: CurrentItemPosition returned: %d, issues length: %d", id, len(w.issues))
		if id >= 0 && id < len(w.issues) {
			issue := w.issues[id]
			logger.Debugf("Base DiveDeeper: Opening summary for issue: %s", issue.GetID())
			x, y, wi, h := w.v.Dimensions()

			summary, err := NewSummaryWidget("summary", x+2, y+2, wi-4, h-4, w.ctx, issue)
			if err != nil {
				logger.Debugf("Base DiveDeeper: Error creating summary widget: %v", err)
				return err
			}

			g.Update(func(g *gocui.Gui) error {
				if err := summary.Layout(g); err != nil {
					return fmt.Errorf("failed to layout summary widget: %w", err)
				}
				_, err := g.SetCurrentView("summary")
				if err != nil {
					return fmt.Errorf("failed to set current view: %w", err)
				}
				return nil
			})
		} else {
			logger.Debugf("Base DiveDeeper: Invalid issue index %d - returning early", id)
		}
	}
	return nil
}

// Common issue line formatting
func (w *ResultsWidget) FormatIssueLine(index int, issue output.Issue) string {
	f, b := colouredSeverity(issue.GetSeverity())
	return fmt.Sprintf("**%d***  %s %-16s %s", index,
		tml.Sprintf(f+"%-12s"+b, issue.GetSeverity()), issue.GetID(), issue.GetTitle())
}

func (w *ResultsWidget) RefreshView() {

}

func (w *ResultsWidget) Layout(*gocui.Gui) error {
	return nil
}

func (w *ResultsWidget) ConfigureKeys(*gocui.Gui) error {
	return nil
}
