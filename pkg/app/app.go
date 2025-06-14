package app

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"kubeguide/pkg/kubernetes"
	"kubeguide/pkg/ui"
)

type App struct {
	app                 *tview.Application
	kubeClient          *kubernetes.Client
	views               *ui.Views
	currentMode         string
	currentNamespace    string
	currentResourceType string
	pages               *tview.Pages
	namespaces          []string
	explorerList        *tview.List
}

func New() *App {
	app := tview.NewApplication()

	// Set VSCode dark theme colors
	tview.Styles.PrimitiveBackgroundColor = tcell.ColorBlack
	tview.Styles.ContrastBackgroundColor = tcell.ColorBlack
	tview.Styles.MoreContrastBackgroundColor = tcell.ColorBlack
	tview.Styles.BorderColor = tcell.ColorLightBlue
	tview.Styles.TitleColor = tcell.ColorWhite
	tview.Styles.GraphicsColor = tcell.ColorWhite
	tview.Styles.PrimaryTextColor = tcell.ColorWhite
	tview.Styles.SecondaryTextColor = tcell.ColorLightGray
	tview.Styles.TertiaryTextColor = tcell.ColorDarkGray
	tview.Styles.InverseTextColor = tcell.ColorBlack
	tview.Styles.ContrastSecondaryTextColor = tcell.ColorLightGray

	return &App{
		app:                 app,
		views:               ui.NewViews(app),
		currentMode:         "welcome",
		currentResourceType: "all",
		pages:               tview.NewPages(),
	}
}

func (a *App) Initialize() error {
	// Try to load Kubernetes config
	kubeClient, currentNamespace, err := kubernetes.NewClient()
	if err != nil {
		fmt.Printf("Warning: Unable to load kubeconfig: %v\n", err)
		a.currentNamespace = "default"
	} else {
		a.kubeClient = kubeClient
		a.currentNamespace = currentNamespace
	}

	// Load namespaces
	if a.kubeClient != nil {
		namespaces, err := a.kubeClient.GetNamespaces()
		if err == nil {
			a.namespaces = namespaces
		}
	}

	return nil
}

func (a *App) setupPages() {
	// Set pages background color
	a.pages.SetBackgroundColor(tcell.ColorBlack)

	// Create pages
	a.pages.AddPage("welcome", a.views.CreateWelcomeView(), true, true)
	a.explorerList = a.views.CreateExplorerView(a.currentNamespace, a.currentResourceType)
	a.pages.AddPage("explorer", a.explorerList, true, false)

	// Load initial resources if connected
	if a.kubeClient != nil {
		go a.loadResources()
	}

	// Set up explorer list selection handler
	a.explorerList.SetSelectedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		a.handleResourceSelection(mainText, secondaryText)
	})
}

func (a *App) setupKeyBindings() {
	a.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		focused := a.app.GetFocus()
		switch focused.(type) {
		case *tview.InputField:
			return event // Let inputs handle their own input
		}

		if event.Key() == tcell.KeyEsc {
			// Check if we're viewing resource details
			if a.pages.HasPage("resource-details") {
				a.pages.RemovePage("resource-details")
				a.pages.SwitchToPage("explorer")
				return nil
			}
			// Otherwise, return to welcome screen
			if a.currentMode != "welcome" {
				a.currentMode = "welcome"
				a.pages.SwitchToPage("welcome")
			}
		}
		switch event.Rune() {
		case 'q':
			a.app.Stop()
			return nil
		case 'e':
			if a.currentMode == "welcome" {
				a.currentMode = "explorer"
				a.pages.SwitchToPage("explorer")
			}
			return nil
		case 'n':
			if a.currentMode == "explorer" {
				a.showNamespaceSelector()
			}
			return nil
		case 'r':
			if a.currentMode == "explorer" {
				a.showResourceSelector()
			}
			return nil
		}
		return event
	})
}

func (a *App) showNamespaceSelector() {
	if len(a.namespaces) == 0 {
		return
	}

	a.views.CreateNamespaceSelector(a.namespaces, a.pages, func(selectedNs string) {
		a.currentNamespace = selectedNs
		a.views.UpdateExplorerTitle(a.explorerList, a.currentNamespace, a.currentResourceType)
		go a.loadResources()
	})
}

func (a *App) showResourceSelector() {
	a.views.CreateResourceSelector(a.pages, func(selectedResourceType string) {
		a.currentResourceType = selectedResourceType
		a.views.UpdateExplorerTitle(a.explorerList, a.currentNamespace, a.currentResourceType)
		go a.loadResources()
	})
}

func (a *App) loadResources() {
	if a.kubeClient == nil {
		a.explorerList.Clear()
		a.explorerList.AddItem("Error: Unable to connect to Kubernetes", "", 0, nil)
		a.app.Draw()
		return
	}

	// Clear the list
	a.explorerList.Clear()

	// Load resources based on current resource type filter
	switch a.currentResourceType {
	case "all":
		a.loadAllResources()
	case "pods":
		a.loadPods()
	case "services":
		a.loadServices()
	default:
		a.explorerList.AddItem(fmt.Sprintf("Resource type '%s' not yet implemented", a.currentResourceType), "", 0, nil)
	}

	a.app.Draw()
}

func (a *App) loadAllResources() {
	a.loadPods()
	a.loadServices()
}

func (a *App) loadPods() {
	pods, err := a.kubeClient.GetPodsInNamespace(a.currentNamespace)
	if err != nil {
		a.explorerList.AddItem(fmt.Sprintf("Error loading pods: %v", err), "", 0, nil)
	} else {
		if len(pods) == 0 && a.currentResourceType == "pods" {
			a.explorerList.AddItem("No pods found in this namespace", "", 0, nil)
		} else {
			for _, pod := range pods {
				a.explorerList.AddItem(fmt.Sprintf("%s: %s (%s)", pod.Type, pod.Name, pod.Status), pod.Name, 0, nil)
			}
		}
	}
}

func (a *App) loadServices() {
	services, err := a.kubeClient.GetServicesInNamespace(a.currentNamespace)
	if err != nil {
		a.explorerList.AddItem(fmt.Sprintf("Error loading services: %v", err), "", 0, nil)
	} else {
		if len(services) == 0 && a.currentResourceType == "services" {
			a.explorerList.AddItem("No services found in this namespace", "", 0, nil)
		} else {
			for _, svc := range services {
				a.explorerList.AddItem(fmt.Sprintf("%s: %s", svc.Type, svc.Name), svc.Name, 0, nil)
			}
		}
	}
}

func (a *App) Run() error {
	a.setupPages()
	a.setupKeyBindings()

	return a.app.SetRoot(a.pages, true).SetFocus(a.pages).Run()
}

func (a *App) handleResourceSelection(mainText string, resourceName string) {
	if a.kubeClient == nil || resourceName == "" {
		return
	}

	// Parse resource type from mainText (format: "ResourceType: ResourceName (Status)")
	parts := strings.Split(mainText, ":")
	if len(parts) < 2 {
		return
	}
	resourceType := strings.TrimSpace(parts[0])

	// Fetch resource details
	go func() {
		yamlContent, err := a.kubeClient.GetResourceDetails(resourceType, resourceName, a.currentNamespace)
		if err != nil {
			yamlContent = fmt.Sprintf("Error fetching resource details: %v", err)
		}

		// Create and show the details view
		a.app.QueueUpdateDraw(func() {
			detailsView := a.views.CreateResourceDetailsView(resourceName, resourceType, yamlContent)
			a.pages.AddPage("resource-details", detailsView, true, true)
			a.pages.SwitchToPage("resource-details")
		})
	}()
}
