package widgets

import (
	"errors"
	"fmt"
	"strings"

	"github.com/awesome-gocui/gocui"
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
	v.SelBgColor = gocui.ColorGreen
	v.SelFgColor = gocui.ColorBlack | gocui.AttrBold
	if g.CurrentView() == v {
		v.FrameColor = gocui.ColorGreen
	} else {
		v.FrameColor = gocui.ColorDefault
	}
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

func (w *ResultsWidget) RefreshView() {

}

func (w *ResultsWidget) Layout(*gocui.Gui) error {
	return nil
}

func (w *ResultsWidget) ConfigureKeys(*gocui.Gui) error {
	return nil
}
