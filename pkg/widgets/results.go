package widgets

import (
	"fmt"
	"sort"
	"strings"

	"github.com/aquasecurity/table"
	"github.com/awesome-gocui/gocui"
	"github.com/liamg/tml"
	"github.com/owenrumney/lazytrivy/pkg/output"
)

type InfoWidget struct {
	name          string
	x, y          int
	w, h          int
	body          string
	v             *gocui.View
	ctx           ctx
	currentReport *output.Report
}

func NewInfoWidget(name string, g ctx) *InfoWidget {
	widget := &InfoWidget{
		name: name,
		x:    0,
		y:    0,
		w:    10,
		h:    10,
		ctx:  g,
	}

	return widget
}

func (w *InfoWidget) ConfigureKeys() error {
	if err := w.ctx.SetKeyBinding(w.name, gocui.MouseWheelDown, gocui.ModNone, w.ScrollDown); err != nil {
		return err
	}

	if err := w.ctx.SetKeyBinding(w.name, gocui.MouseWheelUp, gocui.ModNone, w.ScrollUp); err != nil {
		return err
	}

	if err := w.ctx.SetKeyBinding(w.name, 'f', gocui.ModNone, w.CreateFilterView); err != nil {
		return err
	}

	return nil
}

func (w *InfoWidget) Layout(g *gocui.Gui) error {
	v, err := g.SetView(w.name, w.x, w.y, w.w, w.h, 0)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		_ = tml.Fprintf(v, w.body)
	}
	v.Title = " Results "
	v.Wrap = true
	if g.CurrentView() == v {
		v.FrameColor = gocui.ColorGreen
	} else {
		v.FrameColor = gocui.ColorDefault
	}
	w.v = v
	return nil
}

func (w *InfoWidget) Reset() {
	w.v.Clear()
	w.v.Title = " Results "

	w.v.Subtitle = ""
	if err := w.v.SetOrigin(0, 0); err != nil {
		panic(err)
	}
}

func (w *InfoWidget) ScrollDown(_ *gocui.Gui, view *gocui.View) error {
	if view != nil {
		view.Autoscroll = false
		_, h := view.Size()
		ox, oy := view.Origin()
		newPos := oy + 3
		if newPos > w.v.LinesHeight()-(h+100) {
			return nil
		}
		if err := view.SetOrigin(ox, newPos); err != nil {
			return nil
		}
	}
	return nil
}

func (w *InfoWidget) ScrollUp(_ *gocui.Gui, view *gocui.View) error {
	if view != nil {
		view.Autoscroll = false
		ox, oy := view.Origin()
		newPos := oy - 3
		if newPos < 0 {
			return nil
		}
		if err := view.SetOrigin(ox, newPos); err != nil {
			return nil
		}
	}
	return nil
}

func (w *InfoWidget) UpdateResultsTable(reports []*output.Report) {

	w.v.Clear()

	t := table.New(w.v)
	t.AddHeaders("Image", "Critical", "High", "Medium", "Low", "Unknown")
	t.SetHeaderAlignment(table.AlignLeft)
	t.SetAlignment(table.AlignLeft, table.AlignCenter, table.AlignCenter, table.AlignCenter, table.AlignCenter, table.AlignCenter, table.AlignCenter)

	for _, report := range reports {

		t.AddRow(report.ImageName,
			tml.Sprintf("<bold><red>%d</red></bold>", (report.SeverityCount["CRITICAL"])),
			tml.Sprintf("<red>%d</red>", (report.SeverityCount["HIGH"])),
			tml.Sprintf("<yellow>%d</yellow>", (report.SeverityCount["MEDIUM"])),
			tml.Sprintf("%d", (report.SeverityCount["LOW"])),
			tml.Sprintf("%d", (report.SeverityCount["UNKNOWN"])))

	}

	_ = tml.Fprintf(w.v, `
Trivy has scanned <green>%d</green> images for vulnerabilities.

`, len(reports))

	t.Render()

	_, _ = fmt.Fprintf(w.v, `

For more information about an images vulnerabilities', select from the image list on the left to do a more specific scan.


`)
	w.ctx.RefreshView(w.name)
}

func (w *InfoWidget) RenderReport(report *output.Report, imageName string, severity string) {
	w.currentReport = report

	w.GenerateFilteredReport(imageName, severity)
}

func (w *InfoWidget) GenerateFilteredReport(imageName string, severity string) {

	w.Reset()

	var blocks []string

	var results []output.Result
	if severity != "ALL" {
		results = w.currentReport.SeverityMap[severity]
	} else {
		results = w.currentReport.Results
	}

	for _, result := range results {
		if len(result.Vulnerabilities) == 0 {
			continue
		}

		var vulnerabilities []string

		sort.Slice(result.Vulnerabilities, func(i, j int) bool {
			return result.Vulnerabilities[i].Severity < result.Vulnerabilities[j].Severity
		})

		for _, v := range result.Vulnerabilities {
			severityOpener, severityCloser := colouredSeverity(v.Severity)

			vulnerabilities = append(vulnerabilities, fmt.Sprintf(`  
  %[1]s┌[%[3]s]%[2]s
  %[1]s│%[2]s ID:        %[4]s
  %[1]s│%[2]s Title:     %[5]s
  %[1]s│%[2]s Package:   %[6]s
  %[1]s│%[2]s More Info: <blue>%[7]s</blue>
  %[1]s└─%[2]s`, severityOpener, severityCloser, v.Severity, v.VulnerabilityID, v.Title, v.PkgName, v.PrimaryURL))
		}
		blocks = append(blocks, fmt.Sprintf("\n  <bold><blue>%s</blue></bold>\n%s", result.Target, strings.Join(vulnerabilities, "\n")))
	}

	w.v.Subtitle = imageName
	_ = tml.Fprintf(w.v, strings.Join(blocks, "\n"))
}

func colouredSeverity(severity string) (string, string) {
	switch severity {
	case "CRITICAL":
		return "<bold><red>", "</red></bold>"
	case "HIGH":
		return "<red>", "</red>"
	case "MEDIUM":
		return "<yellow>", "</yellow>"
	default:
		return "", ""
	}
}

func (w *InfoWidget) RefreshView() {
}

func (w *InfoWidget) CreateFilterView(gui *gocui.Gui, view *gocui.View) error {
	if w.currentReport == nil {
		return nil
	}

	colourSevs := []string{"Click severity to apply filter:", "<green>ALL</green>"}
	sevs := []string{"Click severity to apply filter:", "ALL"}

	for _, sev := range []string{"CRITICAL", "HIGH", "MEDIUM", "LOW", "UNKNOWN"} {
		if count, ok := w.currentReport.SeverityCount[sev]; ok {
			if count == 0 {
				continue
			}
			left, right := colouredSeverity(sev)
			colourSevs = append(colourSevs, left+sev+right)
			sevs = append(sevs, sev)
		}
	}
	x, _, width, h := view.Dimensions()

	v, err := gui.SetView("filter", x+1, h-3, width-1, h-1, 0)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		_ = tml.Fprintf(v, strings.Join(colourSevs, " | "))
	}

	if err := gui.SetKeybinding("filter", gocui.MouseLeft, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		x, _ := view.Cursor()
		pos := 0
		start := 0
		selectedSeverity := ""
		for _, s := range sevs {
			pos = pos + len(s)
			if start < x && x <= pos {
				selectedSeverity = s
				break
			}
			start = pos + 3
			pos = start
		}
		if selectedSeverity != "" && selectedSeverity != "Click severity to apply filter:" {
			w.GenerateFilteredReport(w.currentReport.ImageName, selectedSeverity)
			_ = gui.DeleteView("filter")
		}

		return nil
	}); err != nil {
		return err
	}

	v.Frame = true
	v.FrameColor = gocui.ColorYellow
	return nil
}
