package widgets

import (
	"github.com/awesome-gocui/gocui"
	component "github.com/owenrumney/gocui-form"
	"github.com/owenrumney/lazytrivy/pkg/config"
)

type SettingsWidget struct {
	name string
	x, y int
	w, h int
	form *component.Form
	cfg  *config.Config
}

func NewSettingsWidget(name string, x, y, w, h int, gui *gocui.Gui, cfg *config.Config, closeFunc func(*gocui.Gui, *gocui.View) error) *SettingsWidget {
	form := component.NewForm(gui, name, x, y, w, h)

	form.AddCloseFunc(closeFunc)

	return &SettingsWidget{
		name: name,
		x:    x,
		y:    y,
		w:    w,
		h:    h,
		form: form,
		cfg:  cfg,
	}
}

func (w *SettingsWidget) Draw() {
	w.form.AddHeading("General Options")
	w.form.AddCheckBox("  Enable Debugging", 25, w.cfg.Debug)
	w.form.AddCheckBox("  Disable CA Verification", 25, w.cfg.Insecure)
	w.form.AddInputField("  Trivy Cache", 25, 30, w.cfg.CacheDirectory)
	w.form.AddHeading("Scanning Options")
	w.form.AddCheckBox("  Scan Vulnerabilities", 25, w.cfg.Scanner.ScanVulnerabilities)
	w.form.AddCheckBox("  Scan Misconfigs", 25, w.cfg.Scanner.ScanMisconfiguration)
	w.form.AddCheckBox("  Show Secrets", 25, w.cfg.Scanner.ScanSecrets)
	w.form.AddCheckBox("  Ignore Unfixed", 25, w.cfg.Scanner.IgnoreUnfixed)

	w.form.AddButton("Save", func(gui *gocui.Gui, view *gocui.View) error {

		w.cfg.Debug = w.form.GetCheckBoxState("  Enable Debugging")
		w.cfg.Insecure = w.form.GetCheckBoxState("  Disable CA Verification")
		w.cfg.CacheDirectory = w.form.GetFieldText("  Trivy Cache")
		w.cfg.Scanner.ScanVulnerabilities = w.form.GetCheckBoxState("  Scan Vulnerabilities")
		w.cfg.Scanner.ScanMisconfiguration = w.form.GetCheckBoxState("  Scan Misconfigs")
		w.cfg.Scanner.ScanSecrets = w.form.GetCheckBoxState("  Show Secrets")

		if err := w.cfg.Save(); err != nil {
			return err
		}

		return w.form.Close(gui, view)
	})

	w.form.Draw()
}
