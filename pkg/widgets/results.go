package widgets

import (
	"fmt"
	"sort"
	"strings"

	"github.com/awesome-gocui/gocui"
	"github.com/liamg/tml"
	"github.com/owenrumney/lazytrivy/pkg/output"
)

type InfoWidget struct {
	name string
	x, y int
	w, h int
	body string
	v    *gocui.View
	ctx  ctx
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
		ox, oy := view.Origin()
		if err := view.SetOrigin(ox, oy+3); err != nil {
			return nil
		}
	}
	return nil
}

func (w *InfoWidget) ScrollUp(_ *gocui.Gui, view *gocui.View) error {
	if view != nil {
		view.Autoscroll = false
		ox, oy := view.Origin()
		if err := view.SetOrigin(ox, oy-3); err != nil {
			return nil
		}
	}
	return nil
}

func (w *InfoWidget) RenderReport(report output.Report, imageName string) {

	w.Reset()

	var blocks []string

	for _, result := range report.Results {
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
		blocks = append(blocks, fmt.Sprintf("\n  <bold>%s</bold>\n%s", result.Target, strings.Join(vulnerabilities, "\n")))
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
