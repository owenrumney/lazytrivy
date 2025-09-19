package k8s

import (
	"context"
	"fmt"
	"strings"

	"github.com/awesome-gocui/gocui"
	"github.com/owenrumney/lazytrivy/pkg/config"
	"github.com/owenrumney/lazytrivy/pkg/controllers/base"
	"github.com/owenrumney/lazytrivy/pkg/engine"
	"github.com/owenrumney/lazytrivy/pkg/logger"
	"github.com/owenrumney/lazytrivy/pkg/output"
	"github.com/owenrumney/lazytrivy/pkg/widgets"
)

type Controller struct {
	*base.Controller
	*state
}

type Progress interface {
	UpdateStatus(status string)
}

func NewK8sController(cui *gocui.Gui, Engine *engine.Client, cfg *config.Config) *Controller {

	return &Controller{
		&base.Controller{
			Cui:        cui,
			Engine:     Engine,
			Views:      make(map[string]widgets.Widget),
			LayoutFunc: layout,
			HelpFunc:   help,
			Config:     cfg,
		},
		&state{
			namespaces: make(map[string]map[string][]string),
		},
	}

}

func (c *Controller) CreateWidgets(manager base.Manager) error {
	logger.Debugf("Creating K8s view widgets")

	maxX, maxY := c.Cui.Size()

	// K8s Host widget shows current context (top bar) - start with empty context
	c.Views[widgets.Host] = widgets.NewK8sContextWidget(widgets.Host, "", c)

	// Status widget
	c.Views[widgets.Status] = widgets.NewStatusWidget(widgets.Status)

	// Menu widget
	c.Views[widgets.Menu] = widgets.NewMenuWidget(widgets.Menu, 0, maxY-3, maxX-1, maxY-1)

	// K8s Tree widget for namespace/resource navigation
	c.Views[widgets.K8sTree] = widgets.NewK8sTreeWidget(widgets.K8sTree, c)

	// K8s Results widget for resource details and filtering
	c.Views[widgets.Results] = widgets.NewK8sResultWidget(widgets.Results, c)

	// Layout and configure widgets like other controllers
	for _, v := range c.Views {
		_ = v.Layout(c.Cui)
		manager.AddViews(v)
	}
	// Add the layout manager back with minimal layout
	manager.AddViews(gocui.ManagerFunc(c.LayoutFunc))
	c.SetManager()

	return nil
}

func (c *Controller) Initialise() error {
	logger.Debugf("Initialising K8s controller")
	var outerErr error

	c.Cui.Update(func(gui *gocui.Gui) error {
		logger.Debugf("Configuring keyboard shortcuts")
		if err := c.configureKeyBindings(); err != nil {
			return fmt.Errorf("failed to configure global keys: %w", err)
		}

		for _, v := range c.Views {
			if err := v.ConfigureKeys(gui); err != nil {
				return fmt.Errorf("failed to configure view keys: %w", err)
			}
		}

		_, err := gui.SetCurrentView(widgets.K8sTree)
		if err != nil {
			outerErr = fmt.Errorf("failed to set current view: %w", err)
		}

		return err
	})

	logger.Debugf("K8s controller initialised")
	return outerErr
}

func (c *Controller) Tab() widgets.Tab {
	return widgets.K8sTab
}

func (c *Controller) SetSelected(selected string) {
	if selected == "" {
		return
	}
	c.UpdateStatus(fmt.Sprintf("Selected: %s", selected))

	// Handle back navigation
	if strings.HasPrefix(selected, "⬅️ Back") {
		c.NavigateBack()
		return
	}

	// Parse the selection and update tree accordingly
	if c.state.selectedNamespace == "" {
		// Selecting a namespace - expand to show kinds
		namespaceName := strings.TrimPrefix(selected, "📁 ")
		c.state.setSelectedNamespace(namespaceName)
		c.expandNamespace(namespaceName)
	} else if c.state.selectedKind == "" {
		// Selecting a kind - expand to show resources
		kindName := strings.TrimPrefix(selected, "📦 ")
		c.state.setSelectedKind(kindName)
		c.expandKind(c.state.selectedNamespace, kindName)
	} else {
		// Selecting a specific resource
		resourceName := strings.TrimPrefix(selected, "📄 ")
		c.state.setSelectedResource(resourceName)
		c.showResourceDetails()
	}
}

func (c *Controller) ScanCluster(ctx context.Context) {
	c.Lock()
	defer c.Unlock()

	// Lazy load contexts on first use
	if len(c.state.availableContexts) == 0 {
		c.UpdateStatus("Loading K8s contexts...")
		if err := c.loadContexts(); err != nil {
			c.UpdateStatus(fmt.Sprintf("Error loading K8s contexts: %v", err))
			return
		}

		// Set current context if we don't have one
		if c.state.currentContext == "" {
			if currentContext, err := c.Engine.GetCurrentKubernetesContext(); err == nil {
				c.switchContext(currentContext)
			} else if len(c.state.availableContexts) > 0 {
				c.switchContext(c.state.availableContexts[0])
			} else {
				c.UpdateStatus("No K8s contexts available")
				return
			}
		}
	}

	var cancellable context.Context
	cancellable, c.ActiveCancel = context.WithCancel(ctx)

	go func() {
		c.UpdateStatus(fmt.Sprintf("Scanning K8s cluster with context: %s", c.state.currentContext))

		report, err := c.Engine.ScanKubernetes(cancellable, c.state.currentContext, c.Config, c)
		if err != nil {
			c.UpdateStatus(fmt.Sprintf("Error scanning cluster: %v", err))
			return
		}

		c.Cui.Update(func(gocui *gocui.Gui) error {
			return c.RenderK8sReport(report)
		})
	}()
}

func (c *Controller) RenderK8sReport(report *output.Report) error {
	c.state.currentReport = report
	c.state.buildResourceTree(report)

	// Update the tree widget with scan results
	if treeWidget, ok := c.Views[widgets.K8sTree]; ok {
		// Build the tree view with namespaces
		var treeItems []string

		if len(c.state.namespaces) == 0 {
			treeItems = []string{" No resources with issues found "}
		} else {
			// Get sorted namespaces
			namespaces := c.state.getNamespaces()
			for _, namespace := range namespaces {
				treeItems = append(treeItems, fmt.Sprintf("📁 %s", namespace))
			}
		}

		// Check if the widget has UpdateTree method
		if updater, hasMethod := treeWidget.(interface{ UpdateTree([]string) }); hasMethod {
			updater.UpdateTree(treeItems)
		}

		// Set title to indicate we're viewing namespaces
		if titleSetter, hasMethod := treeWidget.(interface{ SetTitle(string) }); hasMethod {
			if len(c.state.namespaces) == 0 {
				titleSetter.SetTitle("K8s Resources")
			} else {
				titleSetter.SetTitle("Namespaces")
			}
		}
	}

	c.UpdateStatus("K8s report processed")
	return nil
}

func (c *Controller) moveViewLeft(*gocui.Gui, *gocui.View) error {
	if c.Cui.CurrentView().Name() == widgets.Results {
		_, err := c.Cui.SetCurrentView(widgets.K8sTree)
		if err != nil {
			return fmt.Errorf("error getting the k8s tree view: %w", err)
		}
	}
	return nil
}

func (c *Controller) moveViewRight(*gocui.Gui, *gocui.View) error {
	if c.Cui.CurrentView().Name() == widgets.K8sTree {
		_, err := c.Cui.SetCurrentView(widgets.Results)
		if err != nil {
			return fmt.Errorf("error getting the results view: %w", err)
		}
	}
	return nil
}

func (c *Controller) showResourceDetails() {
	// Find the specific resource and show its results
	if c.state.currentReport == nil {
		return
	}

	// Try multiple target formats to find the matching result
	possibleTargets := []string{
		fmt.Sprintf("%s/%s", c.state.selectedKind, c.state.selectedResource),
		fmt.Sprintf("%s/%s/%s", c.state.selectedNamespace, c.state.selectedKind, c.state.selectedResource),
		c.state.selectedResource,
		fmt.Sprintf("%s.%s", c.state.selectedKind, c.state.selectedResource),
	}

	var result *output.Result
	var err error
	var foundTarget string

	for _, target := range possibleTargets {
		result, err = c.state.currentReport.GetResultForTarget(target)
		if err == nil && result != nil {
			foundTarget = target
			break
		}
	}

	if result == nil {
		// If still not found, try to find by searching through all results
		for _, res := range c.state.currentReport.Results {
			if strings.Contains(res.Target, c.state.selectedResource) &&
				strings.Contains(res.Target, c.state.selectedKind) {
				result = res
				foundTarget = res.Target
				break
			}
		}
	}

	if result == nil {
		c.UpdateStatus(fmt.Sprintf("No results found for %s/%s (tried: %v)",
			c.state.selectedKind, c.state.selectedResource, possibleTargets))
		return
	}

	c.state.currentResult = result
	c.UpdateStatus(fmt.Sprintf("Showing details for %s", foundTarget))

	// Show results in the results widget
	if resultsWidget, ok := c.Views[widgets.Results].(*widgets.K8sResultWidget); ok {
		resultsWidget.RenderReport(result, c.state.currentReport, "ALL")
		_, _ = c.Cui.SetCurrentView(widgets.Results)
	}
}

func (c *Controller) loadContexts() error {
	contexts, err := c.Engine.GetKubernetesContexts()
	if err != nil {
		logger.Errorf("Failed to load K8s contexts from kubectl: %v", err)
		// Fallback to some default contexts
		c.state.availableContexts = []string{
			"docker-desktop",
			"minikube",
			"kind-kind",
		}
		return nil
	}

	c.state.availableContexts = contexts
	return nil
}

func (c *Controller) BackToParent(gui *gocui.Gui, _ *gocui.View) error {
	// Navigate back up the tree
	if c.state.selectedResource != "" {
		c.state.setSelectedResource("")
		c.UpdateStatus("Back to kinds")
	} else if c.state.selectedKind != "" {
		c.state.setSelectedKind("")
		c.UpdateStatus("Back to namespaces")
	} else if c.state.selectedNamespace != "" {
		c.state.setSelectedNamespace("")
		c.UpdateStatus("Back to cluster view")
	}
	return nil
}

func (c *Controller) CancelCurrentScan(gui *gocui.Gui, _ *gocui.View) error {
	c.Lock()
	defer c.Unlock()
	if c.ActiveCancel != nil {
		logger.Debugf("Cancelling current K8s scan")
		c.UpdateStatus("Current K8s scan cancelled.")
		c.ActiveCancel()
		c.ActiveCancel = nil
		_, _ = gui.SetCurrentView(widgets.Results)
	}
	return nil
}

func (c *Controller) showContextChoice(gui *gocui.Gui, _ *gocui.View) error {
	// Lazy load contexts if not already loaded
	if len(c.state.availableContexts) == 0 {
		c.UpdateStatus("Loading K8s contexts...")
		if err := c.loadContexts(); err != nil {
			c.UpdateStatus(fmt.Sprintf("Error loading K8s contexts: %v", err))
			return err
		}
	}

	maxX, maxY := gui.Size()

	choiceWidget := widgets.NewChoiceWidget(
		"contexts",
		maxX, maxY,
		"Select K8s Context",
		c.state.availableContexts,
		func(selectedContext string) error {
			return c.switchContext(selectedContext)
		},
		c,
	)

	_ = choiceWidget.Layout(gui)
	_, err := gui.SetCurrentView("contexts")
	return err
}

func (c *Controller) switchContext(context string) error {
	c.state.setCurrentContext(context)

	// Update host widget display
	if hostWidget, ok := c.Views[widgets.Host].(*widgets.K8sContextWidget); ok {
		hostWidget.UpdateContext(context)
	}

	c.UpdateStatus(fmt.Sprintf("Switched to context: %s", context))
	return nil
}

// expandNamespace shows the kinds within a selected namespace
func (c *Controller) expandNamespace(namespace string) {
	if treeWidget, ok := c.Views[widgets.K8sTree]; ok {
		var treeItems []string

		// Add back button
		treeItems = append(treeItems, "⬅️ Back to namespaces")

		// Add kinds in this namespace (sorted)
		kinds := c.state.getKindsForNamespace(namespace)
		for _, kind := range kinds {
			treeItems = append(treeItems, fmt.Sprintf("📦 %s", kind))
		}

		if updater, hasMethod := treeWidget.(interface{ UpdateTree([]string) }); hasMethod {
			updater.UpdateTree(treeItems)
		}

		// Set title to show current namespace
		if titleSetter, hasMethod := treeWidget.(interface{ SetTitle(string) }); hasMethod {
			titleSetter.SetTitle(namespace)
		}
	}
}

// expandKind shows the resources within a selected kind
func (c *Controller) expandKind(namespace, kind string) {
	if treeWidget, ok := c.Views[widgets.K8sTree]; ok {
		var treeItems []string

		// Add back button
		treeItems = append(treeItems, "⬅️ Back to kinds")

		// Add resources of this kind (sorted)
		resources := c.state.getResourcesForKind(namespace, kind)
		for _, resource := range resources {
			treeItems = append(treeItems, fmt.Sprintf("📄 %s", resource))
		}

		if updater, hasMethod := treeWidget.(interface{ UpdateTree([]string) }); hasMethod {
			updater.UpdateTree(treeItems)
		}

		// Set title to show current namespace/kind
		if titleSetter, hasMethod := treeWidget.(interface{ SetTitle(string) }); hasMethod {
			titleSetter.SetTitle(fmt.Sprintf("%s/%s", namespace, kind))
		}
	}
}

// NavigateBack handles back navigation in the tree
func (c *Controller) NavigateBack() {
	if c.state.selectedResource != "" {
		// Go back from resource to kinds
		c.state.setSelectedResource("")
		c.expandKind(c.state.selectedNamespace, c.state.selectedKind)
	} else if c.state.selectedKind != "" {
		// Go back from kinds to namespace
		c.state.setSelectedKind("")
		c.expandNamespace(c.state.selectedNamespace)
	} else if c.state.selectedNamespace != "" {
		// Go back from namespace to root
		c.state.setSelectedNamespace("")
		c.showNamespaces()
	}
}

// showNamespaces displays the root level namespace view
func (c *Controller) showNamespaces() {
	if treeWidget, ok := c.Views[widgets.K8sTree]; ok {
		var treeItems []string

		if len(c.state.namespaces) == 0 {
			treeItems = []string{" Press 's' to scan cluster "}
		} else {
			// Get sorted namespaces
			namespaces := c.state.getNamespaces()
			for _, namespace := range namespaces {
				treeItems = append(treeItems, fmt.Sprintf("📁 %s", namespace))
			}
		}

		if updater, hasMethod := treeWidget.(interface{ UpdateTree([]string) }); hasMethod {
			updater.UpdateTree(treeItems)
		}

		// Set title to indicate we're viewing namespaces
		if titleSetter, hasMethod := treeWidget.(interface{ SetTitle(string) }); hasMethod {
			titleSetter.SetTitle("Namespaces")
		}
	}
}
