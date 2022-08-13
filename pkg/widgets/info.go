package widgets

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/awesome-gocui/gocui"
	"github.com/liamg/tml"
	"github.com/owenrumney/lazytrivy/pkg/docker"
	"github.com/owenrumney/lazytrivy/pkg/output"
)

type InfoWidget struct {
	name string
	x, y int
	w, h int
	body string
	v    *gocui.View
	ctx  context.Context
	cli  *docker.DockerClient
}

func NewInfoWidget(name string, cli *docker.DockerClient, x, y, w, h int, body string) *InfoWidget {
	lines := strings.Split(body, "\n")

	for _, l := range lines {
		if len(l) > w {
			w = len(l)
		}
	}

	return &InfoWidget{
		name: name,
		x:    x,
		y:    y,
		w:    w,
		h:    h,
		body: body,
		ctx:  context.Background(),
		cli:  cli,
	}
}

func (w *InfoWidget) ViewName() string {
	return w.v.Name()
}

func (w *InfoWidget) Layout(g *gocui.Gui) error {
	v, err := g.SetView(w.name, w.x, w.y, w.w, w.h, 0)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		tml.Fprintf(v, w.body)
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

func (w *InfoWidget) SetSubTitle(subTitle string) {
	w.v.Subtitle = subTitle
}

func (w *InfoWidget) RenderReport(g *gocui.Gui, report output.Report) {

	var blocks []string

	for _, result := range report.Results {
		var vulns []string

		sort.Slice(result.Vulnerabilities, func(i, j int) bool {
			return result.Vulnerabilities[i].Severity < result.Vulnerabilities[j].Severity
		})

		for _, vuln := range result.Vulnerabilities {
			severityOpener, severityCloser := colouredSeverity(vuln.Severity)

			vulns = append(vulns, fmt.Sprintf(`  
  %[1]s┌[%[3]s]%[2]s
  %[1]s│%[2]s ID:        %[4]s
  %[1]s│%[2]s Title:     %[5]s
  %[1]s│%[2]s Package:   %[6]s
  %[1]s│%[2]s More Info: <blue>%[7]s</blue>
  %[1]s└─%[2]s`, severityOpener, severityCloser, vuln.Severity, vuln.VulnerabilityID, vuln.Title, vuln.PkgName, vuln.PrimaryURL))
		}
		blocks = append(blocks, fmt.Sprintf("\n  <bold>%s</bold>\n%s", result.Target, strings.Join(vulns, "\n")))
	}

	tml.Fprintf(w.v, strings.Join(blocks, "\n"))
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
