package aws

import (
	"fmt"

	"github.com/awesome-gocui/gocui"
	"github.com/owenrumney/lazytrivy/pkg/config"
	"github.com/owenrumney/lazytrivy/pkg/controllers/base"
	"github.com/owenrumney/lazytrivy/pkg/docker"
	"github.com/owenrumney/lazytrivy/pkg/widgets"
)

type Controller struct {
	*base.Controller
	*state
}

func NewAWSController(cui *gocui.Gui, dockerClient *docker.Client, cfg *config.Config) *Controller {

	return &Controller{
		&base.Controller{
			Cui:          cui,
			DockerClient: dockerClient,
			Views:        make(map[string]widgets.Widget),
			LayoutFunc:   layout,
			HelpFunc:     help,
			Config:       cfg,
		},
		&state{
			cacheDirectory: cfg.AWS.CacheDirectory,
		},
	}
}

func (c *Controller) Initialise() error {
	var outerErr error

	c.Cui.Update(func(gui *gocui.Gui) error {

		services, err := c.accountRegionCacheServices(c.Config.AWS.AccountNo, c.Config.AWS.Region)
		if err != nil {
			return err
		}

		if v, ok := c.Views[widgets.Services].(*widgets.ServicesWidget); ok {
			if err := v.RefreshServices(services, 20); err != nil {
				return err
			}
		}

		if err := c.configureKeyBindings(); err != nil {
			return fmt.Errorf("failed to configure global keys: %w", err)
		}

		for _, v := range c.Views {
			if err := v.ConfigureKeys(); err != nil {
				return fmt.Errorf("failed to configure view keys: %w", err)
			}
		}

		_, err = gui.SetCurrentView(widgets.Services)
		if err != nil {
			outerErr = fmt.Errorf("failed to set current view: %w", err)
		}
		return err
	})

	return outerErr
}

func (c *Controller) CreateWidgets(manager base.Manager) error {
	menuItems := []string{
		"<blue>[?]</blue> help", "<red>[t]</red>erminate scan", "<red>[q]</red>uit",
		"\n\n<yellow>Navigation: Use arrow keys to navigate and ESC to exit screens</yellow>",
	}

	maxX, maxY := c.Cui.Size()
	c.Views[widgets.Services] = widgets.NewServicesWidget(widgets.Services, c)
	c.Views[widgets.Results] = widgets.NewAWSResultWidget(widgets.Results, c)
	c.Views[widgets.Menu] = widgets.NewMenuWidget(widgets.Menu, 0, maxY-3, maxX-1, maxY-1, menuItems)
	c.Views[widgets.Status] = widgets.NewStatusWidget(widgets.Status)
	c.Views[widgets.Account] = widgets.NewAccountWidget(widgets.Account, c.Config.AWS.AccountNo, c.Config.AWS.Region)

	for _, v := range c.Views {
		manager.AddViews(v)
	}
	manager.AddViews(gocui.ManagerFunc(c.LayoutFunc))

	c.SetManager()

	return nil
}

func (c *Controller) SetKeyBinding(viewName string, key interface{}, mod gocui.Modifier, handler func(*gocui.Gui, *gocui.View) error) error {
	if err := c.Cui.SetKeybinding(viewName, key, mod, handler); err != nil {
		return fmt.Errorf("failed to set keybinding for %s: %w", key, err)
	}
	return nil
}

func (c *Controller) Tab() widgets.Tab {
	return widgets.AWSTab
}