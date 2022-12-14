package aws

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/awesome-gocui/gocui"
	"github.com/owenrumney/lazytrivy/pkg/config"
	"github.com/owenrumney/lazytrivy/pkg/controllers/base"
	"github.com/owenrumney/lazytrivy/pkg/dockerClient"
	"github.com/owenrumney/lazytrivy/pkg/logger"
	"github.com/owenrumney/lazytrivy/pkg/output"
	"github.com/owenrumney/lazytrivy/pkg/widgets"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type Controller struct {
	*base.Controller
	*state
}

func NewAWSController(cui *gocui.Gui, dockerClient *dockerClient.Client, cfg *config.Config) *Controller {

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

func (c *Controller) CreateWidgets(manager base.Manager) error {
	logger.Debugf("Creating AWS view widgets")

	maxX, maxY := c.Cui.Size()
	c.Views[widgets.Services] = widgets.NewServicesWidget(widgets.Services, c)
	c.Views[widgets.Results] = widgets.NewAWSResultWidget(widgets.Results, c)
	c.Views[widgets.Menu] = widgets.NewMenuWidget(widgets.Menu, 0, maxY-3, maxX-1, maxY-1)
	c.Views[widgets.Status] = widgets.NewStatusWidget(widgets.Status)
	c.Views[widgets.Account] = widgets.NewAccountWidget(widgets.Account, c.Config.AWS.AccountNo, c.Config.AWS.Region)

	for _, v := range c.Views {
		_ = v.Layout(c.Cui)
		manager.AddViews(v)
	}
	manager.AddViews(gocui.ManagerFunc(c.LayoutFunc))
	c.SetManager()

	return nil
}

func (c *Controller) Initialise() error {
	logger.Debugf("Initialising AWS controller")
	var outerErr error

	c.Cui.Update(func(gui *gocui.Gui) error {

		err := c.refreshServices()
		if err != nil {
			return err
		}

		logger.Debugf("Configuring keyboard shortcuts")
		if err := c.configureKeyBindings(); err != nil {
			return fmt.Errorf("failed to configure global keys: %w", err)
		}

		for _, v := range c.Views {
			if err := v.ConfigureKeys(gui); err != nil {
				return fmt.Errorf("failed to configure view keys: %w", err)
			}
		}

		if c.Config.AWS.AccountNo == "" {
			c.UpdateStatus("No AWS specified, press 's' to scan...")
		}

		_, err = gui.SetCurrentView(widgets.Services)
		if err != nil {
			outerErr = fmt.Errorf("failed to set current view: %w", err)
		}

		return err
	})

	return outerErr
}

func (c *Controller) update() error {
	if v, ok := c.Views[widgets.Account]; ok {
		if a, ok := v.(*widgets.AccountWidget); ok {
			logger.Debugf("Updating the AWS account details in the UI")
			a.UpdateAccount(c.Config.AWS.AccountNo, c.Config.AWS.Region)
			if err := c.refreshServices(); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Controller) Tab() widgets.Tab {
	return widgets.AWSTab
}

func (c *Controller) moveViewLeft(*gocui.Gui, *gocui.View) error {
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

func (c *Controller) moveViewRight(*gocui.Gui, *gocui.View) error {
	if c.Cui.CurrentView().Name() == widgets.Services {
		_, err := c.Cui.SetCurrentView(widgets.Results)
		if err != nil {
			return fmt.Errorf("error getting the results view: %w", err)
		}
	}
	return nil
}

func (c *Controller) switchAccount(gui *gocui.Gui, _ *gocui.View) error {

	logger.Debugf("Switching AWS account")
	accounts, err := c.listAccountNumbers()
	if err != nil {
		logger.Errorf("Failed to list AWS accounts. %s", err)
		return err
	}

	x, y := gui.Size()

	accountChoices := widgets.NewChoiceWidget("choice", x, y, " Choose or ESC ", accounts, c.UpdateAccount, c)
	if err := accountChoices.Layout(gui); err != nil {
		logger.Errorf("Failed to create account choice widget. %s", err)
		return fmt.Errorf("error when rendering account choices: %w", err)
	}
	gui.Update(func(gui *gocui.Gui) error {
		_, err := gui.SetCurrentView("choice")
		return err
	})

	return nil
}

func (c *Controller) switchRegion(gui *gocui.Gui, _ *gocui.View) error {
	logger.Debugf("Switching AWS region")
	regions, err := c.listRegions(c.Config.AWS.AccountNo)
	if err != nil {
		logger.Errorf("Failed to list AWS regions. %s", err)
		return err
	}

	x, y := gui.Size()
	regionChoices := widgets.NewChoiceWidget("choice", x, y, " Choose or ESC ", regions, c.UpdateRegion, c)

	if err := regionChoices.Layout(gui); err != nil {
		logger.Errorf("Failed to create region choice widget. %s", err)
		return fmt.Errorf("error when rendering region choices: %w", err)
	}
	gui.Update(func(gui *gocui.Gui) error {
		_, err := gui.SetCurrentView("choice")
		return err
	})

	return nil
}

func (c *Controller) discoverAccount(region string) (string, string, error) {
	ctx := context.Background()
	logger.Debugf("Loading credentials from default config")
	cfg, err := awsConfig.LoadDefaultConfig(ctx)
	if err != nil {
		return "", "", err
	}

	if cfg.Region == "" && region != "" {
		// set a default region just to get going
		cfg.Region = region
	}

	if regionEnv, ok := os.LookupEnv("AWS_REGION"); ok {
		logger.Infof("Using AWS_REGION environment variable")
		cfg.Region = regionEnv
	}

	svc := sts.NewFromConfig(cfg)
	result, err := svc.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		logger.Errorf("Error getting caller identity")
		return "", "", fmt.Errorf("failed to discover AWS caller identity: %w", err)
	}
	if result.Account == nil {
		return "", "", fmt.Errorf("missing account id for aws account")
	}

	logger.Debugf("Discovered AWS account %s", *result.Account)
	return *result.Account, cfg.Region, nil
}

func (c *Controller) scanAccount(gui *gocui.Gui, _ *gocui.View) error {
	c.UpdateStatus("Looking for credentials...")
	account, _, err := c.discoverAccount(c.Config.AWS.Region)
	if err != nil {
		if strings.HasPrefix(err.Error(), "failed to discover AWS caller identity") {
			c.UpdateStatus("Failed to discover AWS credentials.")
			return NewErrNoValidCredentials()
		}
		return err
	}

	c.UpdateStatus("Checking credentials for account...")
	if account != c.Config.AWS.AccountNo && c.Config.AWS.AccountNo != "" {
		c.UpdateStatus("Account number does not match credentials.")
		return fmt.Errorf("account number mismatch: %s != %s", account, c.Config.AWS.AccountNo)
	}

	_, err = gui.SetCurrentView(widgets.Status)
	if err != nil {
		return nil
	}
	go func() {
		report, err := c.UpdateCache(context.Background())
		if err != nil {
			c.UpdateStatus(fmt.Sprintf("Error scanning account: %s", err))
			return
		}
		gui.Update(func(g *gocui.Gui) error {
			c.UpdateStatus(fmt.Sprintf("Scan complete. Found %d results.", report.GetTotalMisconfigurations()))
			return nil
		})

		_, _ = gui.SetCurrentView(widgets.Results)
		if err := c.refreshServices(); err != nil {
			logger.Errorf("Error refreshing services: %v", err)
		}
		c.UpdateStatus("Account scan complete.")
	}()
	return nil
}

func (c *Controller) RenderAWSResultsReport(report *output.Report) error {
	if v, ok := c.Views[widgets.Results].(*widgets.AWSResultWidget); ok {
		v.RenderReport(report, "ALL")
		_, err := c.Cui.SetCurrentView(widgets.Results)
		if err != nil {
			return fmt.Errorf("failed to set current view: %w", err)
		}
	}
	return nil
}

func (c *Controller) RenderAWSResultsReportSummary(report *output.Report) error {
	if v, ok := c.Views[widgets.Results].(*widgets.AWSResultWidget); ok {
		v.UpdateResultsTable([]*output.Report{report}, c.Cui)

	}
	return fmt.Errorf("failed to render results report summary") //nolint:goerr113
}
