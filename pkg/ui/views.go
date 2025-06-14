package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/sahilm/fuzzy"
)

type Views struct {
	app *tview.Application
}

func NewViews(app *tview.Application) *Views {
	return &Views{app: app}
}

func (v *Views) CreateWelcomeView() tview.Primitive {
	welcomeMessage := "Welcome to kubeguide\nPress 'q' to quit\nPress 'e' for explorer"
	// Create a text view with large, centered text
	textView := tview.NewTextView().
		SetChangedFunc(func() { v.app.Draw() }).
		SetTextAlign(tview.AlignCenter).
		SetTextColor(tcell.ColorWhite)
	fmt.Fprintf(textView, "%s", welcomeMessage)

	// Add a border to make it more visible
	textView.SetBorder(true).
		SetBorderColor(tcell.ColorLightBlue).
		SetTitle(" Welcome ").
		SetTitleColor(tcell.ColorWhite)

	// Create a flex layout to center the text view
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().
			SetDirection(tview.FlexColumn).
			AddItem(nil, 0, 1, false).
			AddItem(textView, 40, 1, true).
			AddItem(nil, 0, 1, false),
			10, 1, true).
		AddItem(nil, 0, 1, false)

	return flex
}

func (v *Views) CreateExplorerView(namespace string, resourceType string) *tview.List {
	list := tview.NewList()
	list.SetMainTextColor(tcell.ColorWhite).
		SetSelectedTextColor(tcell.ColorBlack).
		SetSelectedBackgroundColor(tcell.ColorLightBlue).
		SetBackgroundColor(tcell.ColorBlack)

	list.SetBorder(true).
		SetBorderColor(tcell.ColorLightBlue).
		SetTitleColor(tcell.ColorWhite)

	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'j':
			return tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone)
		case 'k':
			return tcell.NewEventKey(tcell.KeyBacktab, 0, tcell.ModNone)
		}
		return event
	})

	title := fmt.Sprintf(" Explorer Mode - Namespace: %s | Resource: %s (Press 'n'/'r' to change) ", namespace, resourceType)
	list.SetTitle(title)
	return list
}

func (v *Views) UpdateExplorerTitle(list *tview.List, namespace string, resourceType string) {
	title := fmt.Sprintf(" Explorer Mode - Namespace: %s | Resource: %s (Press 'n'/'r' to change) ", namespace, resourceType)
	list.SetTitle(title)
}

func (v *Views) createGenericSelector(items []string, title string, pageName string, pages *tview.Pages, onSelect func(string)) {
	var filteredMatches []fuzzy.Match
	var selectedIndex int

	inputField := tview.NewInputField().
		SetLabel("Search: ").
		SetFieldWidth(50).
		SetFieldBackgroundColor(tcell.ColorBlack).
		SetFieldTextColor(tcell.ColorWhite).
		SetLabelColor(tcell.ColorLightBlue)

	// Initialize matchList
	matchList := tview.NewList()
	matchList.ShowSecondaryText(false)
	matchList.SetMainTextColor(tcell.ColorWhite)
	matchList.SetSelectedTextColor(tcell.ColorBlack)
	matchList.SetSelectedBackgroundColor(tcell.ColorLightBlue)
	matchList.SetBackgroundColor(tcell.ColorBlack)

	// Update match list based on search text
	updateMatches := func(text string) {
		matchList.Clear()
		selectedIndex = 0

		if text == "" {
			filteredMatches = make([]fuzzy.Match, len(items))
			for i, item := range items {
				filteredMatches[i] = fuzzy.Match{Str: item}
				matchList.AddItem(item, "", 0, nil)
			}
		} else {
			filteredMatches = fuzzy.Find(text, items)
			for _, match := range filteredMatches {
				matchList.AddItem(match.Str, "", 0, nil)
			}
		}

		if len(filteredMatches) > 0 {
			matchList.SetCurrentItem(0)
		}
	}

	// Initialize with all items
	updateMatches("")

	// Handle input changes for live filtering
	inputField.SetChangedFunc(func(text string) {
		updateMatches(text)
	})

	// Handle key events for navigation and selection
	inputField.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlJ, tcell.KeyTab: // Navigate down
			if len(filteredMatches) > 0 {
				selectedIndex = (selectedIndex + 1) % len(filteredMatches)
				matchList.SetCurrentItem(selectedIndex)
			}
			return nil
		case tcell.KeyCtrlK, tcell.KeyBacktab: // Navigate up
			if len(filteredMatches) > 0 {
				selectedIndex = (selectedIndex - 1 + len(filteredMatches)) % len(filteredMatches)
				matchList.SetCurrentItem(selectedIndex)
			}
			return nil
		case tcell.KeyEnter: // Select current match
			if len(filteredMatches) > 0 {
				selectedItem := filteredMatches[selectedIndex].Str
				onSelect(selectedItem)
				pages.RemovePage(pageName)
				pages.SwitchToPage("explorer")
			}
			return nil
		case tcell.KeyEscape: // Cancel
			pages.RemovePage(pageName)
			pages.SwitchToPage("explorer")
			return nil
		}
		return event
	})

	// Create flex layout with input field on top and match list below
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(inputField, 3, 0, true).
		AddItem(matchList, 0, 1, false)

	flex.SetBorder(true).
		SetTitle(title).
		SetBorderColor(tcell.ColorLightBlue).
		SetTitleColor(tcell.ColorWhite).
		SetBackgroundColor(tcell.ColorBlack)

	pages.AddPage(pageName, flex, true, false)
	pages.SwitchToPage(pageName)
	v.app.SetFocus(inputField)
}

func (v *Views) CreateResourceSelector(pages *tview.Pages, onSelect func(string)) {
	resourceTypes := []string{"all", "pods", "services", "deployments", "configmaps", "secrets", "ingresses", "daemonsets", "statefulsets", "jobs", "cronjobs"}
	v.createGenericSelector(resourceTypes, " Resource Type Selector (Ctrl+J/K to navigate, Enter to select, Esc to cancel) ", "resource-selector", pages, onSelect)
}
func (v *Views) CreateNamespaceSelector(namespaces []string, pages *tview.Pages, onSelect func(string)) {
	v.createGenericSelector(namespaces, " Namespace Selector (Ctrl+J/K to navigate, Enter to select, Esc to cancel) ", "namespace-selector", pages, onSelect)
}

func (v *Views) CreateResourceDetailsView(resourceName string, resourceType string, yamlContent string) tview.Primitive {
	textView := tview.NewTextView().
		SetChangedFunc(func() { v.app.Draw() }).
		SetTextColor(tcell.ColorWhite)
	fmt.Fprintf(textView, "%s", yamlContent)
	textView.SetBorder(true).
		SetBorderColor(tcell.ColorLightBlue).
		SetTitle(fmt.Sprintf(" %s: %s (Press Esc to return) ", resourceType, resourceName)).
		SetTitleColor(tcell.ColorWhite)

	return textView
}
