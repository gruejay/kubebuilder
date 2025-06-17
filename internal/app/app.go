package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/yaml"

	"kubeguide/internal/kubernetes"
	"kubeguide/internal/modes"
	"kubeguide/internal/navigation"
	"kubeguide/internal/ui"
)

type Resource struct {
	Type   string
	Name   string
	Status string
}

type App struct {
	app                 *tview.Application
	kubeClient          *kubernetes.UnifiedClient
	explorer            *ui.Explorer
	welcome             *ui.Welcome
	resourceDetails     *ui.ResourceDetails
	currentMode         modes.Mode
	currentNamespace    string
	currentResourceType string
	pages               *tview.Pages
	namespaces          []string
	explorerList        *tview.List
	keyBindings         *navigation.KeyBindings
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
		explorer:            ui.NewExplorer(app),
		welcome:             ui.NewWelcome("Welcome", ""),
		currentMode:         modes.Welcome,
		currentResourceType: "all",
		pages:               tview.NewPages(),
		keyBindings:         navigation.GetDefaultKeyBindings(),
	}
}

func (a *App) Initialize() error {
	// Try to load Kubernetes config
	kubeClient, err := kubernetes.NewUnifiedClient()
	if err != nil {
		fmt.Printf("Warning: Unable to load kubeconfig: %v\n", err)
		a.currentNamespace = "default"
	} else {
		a.kubeClient = kubeClient
		a.currentNamespace = "default" // Default namespace
	}

	// Load namespaces
	if a.kubeClient != nil {
		namespaces, err := a.getNamespaces()
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
	a.pages.AddPage("welcome", a.welcome.CreateWelcomeView(), true, true)
	a.explorerList = a.explorer.CreateExplorerView(a.currentNamespace, a.currentResourceType)
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

		// Check if help page is open first - if so, only handle help-related keys
		if a.pages.HasPage("help") {
			if event.Key() == tcell.KeyEsc || event.Rune() == 'q' {
				a.pages.RemovePage("help")
			}
			// Consume all other events when help is open
			return nil
		}

		if event.Key() == tcell.KeyEsc {
			// Check if we're viewing resource details
			if a.pages.HasPage("resource-details") {
				a.pages.RemovePage("resource-details")
				a.pages.SwitchToPage("explorer")
				return nil
			}
			// Otherwise, return to welcome screen
			if a.currentMode != modes.Welcome {
				a.currentMode = modes.Welcome
				a.pages.SwitchToPage("welcome")
			}
		}
		switch event.Rune() {
		case 'q':
			a.app.Stop()
			return nil
		case 'e':
			if a.currentMode == modes.Welcome {
				a.currentMode = modes.Explorer
				a.pages.SwitchToPage("explorer")
			}
			return nil
		case 'n':
			if a.currentMode == modes.Explorer {
				a.showNamespaceSelector()
			}
			return nil
		case 'r':
			if a.currentMode == modes.Explorer {
				a.showResourceSelector()
			}
			return nil
		case '?':
			a.showHelpView()
			return nil
		}
		return event
	})
}

func (a *App) showNamespaceSelector() {
	if len(a.namespaces) == 0 {
		return
	}

	a.explorer.CreateNamespaceSelector(a.namespaces, a.pages, func(selectedNs string) {
		a.currentNamespace = selectedNs
		a.explorer.UpdateExplorerTitle(a.explorerList, a.currentNamespace, a.currentResourceType)
		go a.loadResources()
	})
}

func (a *App) showResourceSelector() {
	a.explorer.CreateResourceSelector(a.pages, func(selectedResourceType string) {
		a.currentResourceType = selectedResourceType
		a.explorer.UpdateExplorerTitle(a.explorerList, a.currentNamespace, a.currentResourceType)
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
	case "pods", "services", "deployments", "configmaps", "secrets":
		a.loadResourcesByType(a.currentResourceType)
	default:
		a.explorerList.AddItem(fmt.Sprintf("Resource type '%s' not yet implemented", a.currentResourceType), "", 0, nil)
	}

	a.app.Draw()
}

func (a *App) loadAllResources() {
	resourceTypes := []string{"pods", "services", "deployments", "configmaps", "secrets"}
	for _, resourceType := range resourceTypes {
		a.loadResourcesByType(resourceType)
	}
}

func (a *App) loadResourcesByType(resourceType string) {
	resources, err := a.getResourcesInNamespace(resourceType, a.currentNamespace)
	if err != nil {
		a.explorerList.AddItem(fmt.Sprintf("Error loading %s: %v", resourceType, err), "", 0, nil)
	} else {
		if len(resources) == 0 && a.currentResourceType == resourceType {
			a.explorerList.AddItem(fmt.Sprintf("No %s found in this namespace", resourceType), "", 0, nil)
		} else {
			for _, resource := range resources {
				displayText := fmt.Sprintf("%s: %s (%s)", resource.Type, resource.Name, resource.Status)
				a.explorerList.AddItem(displayText, resource.Name, 0, nil)
			}
		}
	}
}

func (a *App) loadPods() {
	a.loadResourcesByType("pods")
}

func (a *App) loadServices() {
	a.loadResourcesByType("services")
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
		yamlContent, err := a.getResourceDetails(resourceType, resourceName, a.currentNamespace)
		if err != nil {
			yamlContent = fmt.Sprintf("Error fetching resource details: %v", err)
		}

		// Create and show the details view
		a.app.QueueUpdateDraw(func() {
			rd := ui.NewResourceDetails(resourceName, resourceType, yamlContent)
			detailsView := rd.CreateView()
			a.pages.AddPage("resource-details", detailsView, true, true)
			a.pages.SwitchToPage("resource-details")
		})
	}()
}

// Helper methods using UnifiedClient GVR interface

func (a *App) getNamespaces() ([]string, error) {
	ctx := context.Background()
	nsGVR := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}

	var nsList v1.NamespaceList
	err := a.kubeClient.List(ctx, nsGVR, "", &nsList)
	if err != nil {
		return nil, err
	}

	var namespaces []string
	for _, ns := range nsList.Items {
		namespaces = append(namespaces, ns.Name)
	}

	return namespaces, nil
}

func (a *App) getResourcesInNamespace(resourceType, namespace string) ([]Resource, error) {
	ctx := context.Background()

	var gvr schema.GroupVersionResource
	switch strings.ToLower(resourceType) {
	case "pod", "pods":
		gvr = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
	case "service", "services":
		gvr = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "services"}
	case "deployment", "deployments":
		gvr = schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	case "configmap", "configmaps":
		gvr = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}
	case "secret", "secrets":
		gvr = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "secrets"}
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", resourceType)
	}

	var list unstructured.UnstructuredList
	err := a.kubeClient.List(ctx, gvr, namespace, &list)
	if err != nil {
		return nil, err
	}

	var resources []Resource
	for _, item := range list.Items {
		name := item.GetName()
		status := "Unknown"

		// Extract status based on resource type
		switch strings.ToLower(resourceType) {
		case "pod", "pods":
			if phase, found, _ := unstructured.NestedString(item.Object, "status", "phase"); found {
				status = phase
			}
		case "service", "services":
			if svcType, found, _ := unstructured.NestedString(item.Object, "spec", "type"); found {
				status = svcType
			}
		case "deployment", "deployments":
			if replicas, found, _ := unstructured.NestedInt64(item.Object, "status", "replicas"); found {
				if readyReplicas, readyFound, _ := unstructured.NestedInt64(item.Object, "status", "readyReplicas"); readyFound {
					status = fmt.Sprintf("%d/%d", readyReplicas, replicas)
				} else {
					status = fmt.Sprintf("0/%d", replicas)
				}
			}
		}

		resources = append(resources, Resource{
			Type:   item.GetKind(),
			Name:   name,
			Status: status,
		})
	}

	return resources, nil
}

func (a *App) getPodsInNamespace(namespace string) ([]Resource, error) {
	return a.getResourcesInNamespace("pods", namespace)
}

func (a *App) getServicesInNamespace(namespace string) ([]Resource, error) {
	return a.getResourcesInNamespace("services", namespace)
}

func (a *App) getResourceDetails(resourceType, resourceName, namespace string) (string, error) {
	ctx := context.Background()

	var gvr schema.GroupVersionResource
	switch strings.ToLower(resourceType) {
	case "pod":
		gvr = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
	case "service":
		gvr = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "services"}
	case "deployment":
		gvr = schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	case "configmap":
		gvr = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}
	case "secret":
		gvr = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "secrets"}
	default:
		return "", fmt.Errorf("unsupported resource type: %s", resourceType)
	}

	var obj unstructured.Unstructured
	err := a.kubeClient.Get(ctx, gvr, namespace, resourceName, &obj)
	if err != nil {
		return "", err
	}

	// Convert to YAML for display
	obj = kubernetes.CleanData(obj)
	yamlBytes, err := yaml.Marshal(obj.Object)
	if err != nil {
		return "", err
	}

	return string(yamlBytes), nil
}

func (a *App) showHelpView() {
	// Get key bindings for current mode
	bindings := a.keyBindings.GetBindings(a.currentMode)

	// Build help text from bindings
	helpText := fmt.Sprintf("Key Bindings - %s Mode:\n\n", a.currentMode)

	for _, binding := range bindings {
		var keyStr string
		if binding.Key != tcell.KeyNUL {
			switch binding.Key {
			case tcell.KeyEsc:
				keyStr = "Esc"
			case tcell.KeyEnter:
				keyStr = "Enter"
			default:
				keyStr = fmt.Sprintf("Key:%d", binding.Key)
			}
		} else if binding.Rune != 0 {
			keyStr = string(binding.Rune)
		}

		if keyStr != "" {
			helpText += fmt.Sprintf("  %-8s - %s\n", keyStr, binding.Description)
		}
	}

	textView := tview.NewTextView().
		SetText(helpText).
		SetDynamicColors(true).
		SetWrap(false)

	textView.SetBackgroundColor(tcell.ColorBlack)
	textView.SetTextColor(tcell.ColorWhite)
	textView.SetBorder(true).SetTitle(" Help - Press 'esc' or 'q' to close ")

	// Input capture is handled by the main app, so no need to set it here

	// Create a flex layout to position the help view in bottom right
	flex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(textView, 10, 1, true).
			AddItem(nil, 0, 1, false), 40, 1, true).
		AddItem(nil, 0, 1, false)

	flex.SetBackgroundColor(tcell.ColorBlack)

	a.pages.AddPage("help", flex, true, true)
}
