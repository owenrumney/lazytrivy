package aws

import (
	"context"
	"fmt"
	"os"

	"github.com/awesome-gocui/gocui"
	"github.com/owenrumney/lazytrivy/pkg/config"
	"github.com/owenrumney/lazytrivy/pkg/controllers/base"
	"github.com/owenrumney/lazytrivy/pkg/docker"
	"github.com/owenrumney/lazytrivy/pkg/widgets"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
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
		"<blue>[?]</blue> help", "switch <blue>[m]</blue>ode", "<red>[t]</red>erminate scan", "<red>[q]</red>uit",
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

func (c *Controller) UpdateAccount(account string) error {

	c.Config.AWS.AccountNo = account
	c.Config.AWS.Region = "us-east-1"
	return c.Config.Save()
}

func (c *Controller) UpdateRegion(region string) error {
	c.Config.AWS.Region = region
	return c.Config.Save()
}

func (c *Controller) Tab() widgets.Tab {
	return widgets.AWSTab
}

func (c *Controller) moveViewLeft(gui *gocui.Gui, view *gocui.View) error {
	if c.Cui.CurrentView().Name() == widgets.Results {
		_, err := c.Cui.SetCurrentView(widgets.Services)
		if err != nil {
			return fmt.Errorf("error getting the images view: %w", err)
		}
		if v, ok := c.Views[widgets.Images].(*widgets.ImagesWidget); ok {
			return v.SetSelectedImage(c.state.selectedService)
		}
	}
	return nil
}

func (c *Controller) moveViewRight(gui *gocui.Gui, view *gocui.View) error {
	if c.Cui.CurrentView().Name() == widgets.Services {
		_, err := c.Cui.SetCurrentView(widgets.Results)
		if err != nil {
			return fmt.Errorf("error getting the results view: %w", err)
		}
	}
	return nil
}

func (c *Controller) switchAccount(gui *gocui.Gui, view *gocui.View) error {

	accounts, err := c.listAccountNumbers()
	if err != nil {
		return err
	}

	x, y := gui.Size()

	accountChoices := widgets.NewChoiceWidget("choice", x/2-10, y/2-2, x/2+10, y/2+2, " Choose or ESC ", accounts, c.UpdateAccount, c)

	if err := accountChoices.Layout(gui); err != nil {
		return fmt.Errorf("error when rendering account choices: %w", err)
	}
	gui.Update(func(gui *gocui.Gui) error {
		_, err := gui.SetCurrentView("choice")
		return err
	})

	return nil
}

func (c *Controller) switchRegion(gui *gocui.Gui, view *gocui.View) error {

	regions, err := c.listRegions(c.Config.AWS.AccountNo)
	if err != nil {
		return err
	}

	x, y := gui.Size()

	regionChoices := widgets.NewChoiceWidget("choice", x/2-10, y/2-2, x/2+10, y/2+len(regions), " Choose or ESC ", regions, c.UpdateRegion, c)

	if err := regionChoices.Layout(gui); err != nil {
		return fmt.Errorf("error when rendering region choices: %w", err)
	}
	gui.Update(func(gui *gocui.Gui) error {
		_, err := gui.SetCurrentView("choice")
		return err
	})

	return nil
}

func (c *Controller) discoverAccount() (string, string, error) {
	ctx := context.Background()
	cfg, err := awsConfig.LoadDefaultConfig(ctx)
	if err != nil {
		return "", "", err
	}

	if regionEnv, ok := os.LookupEnv("AWS_REGION"); ok {
		cfg.Region = regionEnv
	}

	svc := sts.NewFromConfig(cfg)
	result, err := svc.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return "", "", fmt.Errorf("failed to discover AWS caller identity: %w", err)
	}
	if result.Account == nil {
		return "", "", fmt.Errorf("missing account id for aws account")
	}
	return *result.Account, cfg.Region, nil
}
