package controller

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/awesome-gocui/gocui"
	"github.com/owenrumney/lazytrivy/pkg/docker"
	"github.com/owenrumney/lazytrivy/pkg/output"
	"github.com/owenrumney/lazytrivy/pkg/widgets"
)

type Controller struct {
	sync.Mutex
	cui           *gocui.Gui
	dockerClient  *docker.Client
	views         map[string]widgets.Widget
	selectedImage string
	activeCancel  context.CancelFunc
}

func New() (*Controller, error) {
	cui, err := gocui.NewGui(gocui.OutputNormal, true)
	if err != nil {
		return nil, fmt.Errorf("failed to create gui: %w", err)
	}
	return &Controller{
		cui:          cui,
		dockerClient: docker.NewClient(),
		views:        make(map[string]widgets.Widget),
	}, nil
}

func (g *Controller) DockerClient() *docker.Client {
	return g.dockerClient
}

func (g *Controller) CreateWidgets() error {
	maxX, maxY := g.cui.Size()

	g.views[widgets.Images] = widgets.NewImagesWidget(widgets.Images, g)
	g.views[widgets.Results] = widgets.NewInfoWidget(widgets.Results, g)
	g.views[widgets.Menu] = widgets.NewMenuWidget(widgets.Menu, 0, maxY-3, maxX-1, maxY-1, g)
	g.views[widgets.Status] = widgets.NewStatusWidget(widgets.Status, g)
	g.views[widgets.Host] = widgets.NewHostWidget(widgets.Host, g)

	g.SetManager()
	g.EnableMouse()

	return nil
}

func (g *Controller) SetManager() {
	views := make([]gocui.Manager, 0, len(g.views)+1)
	for _, v := range g.views {
		views = append(views, v)
	}

	views = append(views, gocui.ManagerFunc(flowLayout))
	g.cui.SetManager(views...)
}

func (g *Controller) Run() error {
	if err := g.cui.MainLoop(); err != nil && errors.Is(err, gocui.ErrQuit) {
		return fmt.Errorf("error occurred during the run main loop: %w", err)
	}
	return nil
}

func (g *Controller) Initialise() {
	g.cui.Update(func(gui *gocui.Gui) error {
		if err := g.configureGlobalKeys(); err != nil {
			return fmt.Errorf("failed to configure global keys: %w", err)
		}

		for _, v := range g.views {
			if err := v.ConfigureKeys(); err != nil {
				return fmt.Errorf("failed to configure view keys: %w", err)
			}
		}

		_, err := gui.SetCurrentView(widgets.Images)

		return err
	})
}

func (g *Controller) SetKeyBinding(viewName string, key interface{}, mod gocui.Modifier, handler func(*gocui.Gui, *gocui.View) error) error {
	if err := g.cui.SetKeybinding(viewName, key, mod, handler); err != nil {
		return fmt.Errorf("failed to set keybinding for %s: %w", key, err)
	}
	return nil
}

func (g *Controller) Close() {
	g.cui.Close()
}

func (g *Controller) ShowCursor() {
	g.cui.Cursor = true
}

func (g *Controller) HideCursor() {
	g.cui.Cursor = false
}

func (g *Controller) EnableMouse() {
	g.cui.Mouse = true
}

func (g *Controller) DisableMouse() {
	g.cui.Mouse = false
}

func (g *Controller) CancelCurrentScan(_ *gocui.Gui, _ *gocui.View) error {
	g.Lock()
	defer g.Unlock()
	if g.activeCancel != nil {
		g.UpdateStatus("Current scan cancelled.")
		g.activeCancel()
		g.activeCancel = nil
	}
	return nil
}

func (g *Controller) SetSelectedImage(selected string) {
	g.selectedImage = selected
}

func (g *Controller) ScanImage(ctx context.Context, imageName string) {
	g.cleanupResults()

	var cancellable context.Context
	g.Lock()
	defer g.Unlock()
	cancellable, g.activeCancel = context.WithCancel(ctx)
	go func() {
		report, err := g.dockerClient.ScanImage(cancellable, imageName, g)
		if err != nil {
			return
		}
		g.cui.Update(func(gocui *gocui.Gui) error {
			return g.renderResultsReport(imageName, report)
		})
	}()
}

func (g *Controller) ScanAllImages(ctx context.Context) {
	g.cleanupResults()

	var cancellable context.Context
	g.Lock()
	defer g.Unlock()
	cancellable, g.activeCancel = context.WithCancel(ctx)
	go func() {
		reports, err := g.dockerClient.ScanAllImages(cancellable, g)
		if err != nil {
			return
		}
		if err := g.renderResultsReportSummary(reports); err != nil {
			g.UpdateStatus(err.Error())
		}
		g.UpdateStatus("All images scanned.")
	}()
}

func flowLayout(g *gocui.Gui) error {
	imagesWidth := 0
	viewNames := []string{widgets.Images, widgets.Host, widgets.Results, widgets.Menu, widgets.Status}
	maxX, maxY := g.Size()
	x := 0
	for _, viewName := range viewNames {
		v, err := g.View(viewName)
		if err != nil {
			return fmt.Errorf("failed to get view %s: %w", viewName, err)
		}
		w, _ := v.Size()
		h := 1
		nextW := w
		nextH := maxY - 3
		nextX := x

		switch v.Name() {
		case widgets.Host:
			nextW = imagesWidth
			nextX = 0
			nextH = 3
		case widgets.Images:
			imagesWidth = w
			h = 4
		case widgets.Status:
			nextW = maxX - 1
			h = maxY - 5
		case widgets.Results:
			nextW = maxX - 1
			nextH = maxY - 6
		case widgets.Menu, widgets.Remote:
			continue
		}

		_, err = g.SetView(v.Name(), nextX, h, nextW, nextH, 0)
		if err != nil && errors.Is(err, gocui.ErrUnknownView) {
			return fmt.Errorf("%w", err)
		}
		if v.Name() == widgets.Images {
			x += nextW + 1
		}
	}
	return nil
}

func quit(_ *gocui.Gui, _ *gocui.View) error {
	return gocui.ErrQuit
}

func setView(g *gocui.Gui, v *gocui.View) error {
	_, err := g.SetCurrentView(v.Name())
	return err
}

func (g *Controller) scanRemote(gui *gocui.Gui, _ *gocui.View) error {
	maxX, maxY := gui.Size()

	g.ShowCursor()
	remote, err := widgets.NewInput(widgets.Remote, maxX, maxY, 150, g)
	if err != nil {
		return fmt.Errorf("failed to create remote input: %w", err)
	}
	gui.Update(func(g *gocui.Gui) error {
		gui.Mouse = false
		if err := remote.Layout(gui); err != nil {
			return fmt.Errorf("failed to layout remote input: %w", err)
		}
		_, err := gui.SetCurrentView(widgets.Remote)
		return fmt.Errorf("failed to set current view: %w", err)
	})
	return nil
}

func (g *Controller) renderResultsReport(imageName string, report *output.Report) error {
	if v, ok := g.views[widgets.Results].(*widgets.InfoWidget); ok {
		v.RenderReport(report, fmt.Sprintf(" %s ", imageName), "ALL")
		_, err := g.cui.SetCurrentView(widgets.Results)
		if err != nil {
			return fmt.Errorf("failed to set current view: %w", err)
		}
	}
	return nil
}

func (g *Controller) renderResultsReportSummary(reports []*output.Report) error {
	if v, ok := g.views[widgets.Results].(*widgets.InfoWidget); ok {
		v.UpdateResultsTable(reports)
		_, err := g.cui.SetCurrentView(widgets.Results)
		if err != nil {
			return fmt.Errorf("error setting current view: %w", err)
		}
	}
	return errors.New("failed to render results report summary") //nolint:goerr113
}

func (g *Controller) UpdateStatus(status string) {
	if v, ok := g.views[widgets.Status].(*widgets.StatusWidget); ok {
		v.UpdateStatus(status)
		v.RefreshView()
	}
}

func (g *Controller) ClearStatus() {
	g.UpdateStatus("")
}

func (g *Controller) RefreshImages() error {
	g.UpdateStatus("Refreshing images")
	images := g.dockerClient.ListImages()
	g.ClearStatus()
	if v, ok := g.views[widgets.Images].(*widgets.ImagesWidget); ok {
		return v.RefreshImages(images)
	}
	return errors.New("error getting the images view") //nolint:goerr113
}

func (g *Controller) RefreshView(viewName string) {
	g.cui.Update(func(_ *gocui.Gui) error {
		if v, ok := g.views[viewName]; ok {
			v.RefreshView()
		}
		return nil
	})
}

func (g *Controller) cleanupResults() {
	if v, err := g.cui.View(widgets.Results); err == nil {
		v.Clear()
		v.Subtitle = ""
	}
	_ = g.cui.DeleteView("filter")
}
