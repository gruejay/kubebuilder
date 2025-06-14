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
	textView := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetText("Welcome to kubeguide\n\nPress 'e' for Explorer mode\nPress 'q' to quit").
		SetTextColor(tcell.ColorWhite).
		SetBackgroundColor(tcell.ColorBlack)

	return tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(textView, 0, 3, true).
			AddItem(nil, 0, 1, false), 0, 1, true).
		AddItem(nil, 0, 1, false)
}

func (v *Views) CreateExplorerView(namespace string) *tview.List {
	list := tview.NewList()
	list.SetMainTextColor(tcell.ColorWhite).
		SetSelectedTextColor(tcell.ColorBlack).
		SetSelectedBackgroundColor(tcell.ColorLightBlue).
		SetBackgroundColor(tcell.ColorBlack)
	
	list.SetBorder(true).
		SetBorderColor(tcell.ColorLightBlue).
		SetTitleColor(tcell.ColorWhite)
	
	title := fmt.Sprintf(" Explorer Mode - Namespace: %s (Press 'n' to change) ", namespace)
	list.SetTitle(title)
	return list
}

func (v *Views) UpdateExplorerTitle(list *tview.List, namespace string) {
	title := fmt.Sprintf(" Explorer Mode - Namespace: %s (Press 'n' to change) ", namespace)
	list.SetTitle(title)
}

func (v *Views) CreateNamespaceSelector(namespaces []string, pages *tview.Pages, onSelect func(string)) {
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
			filteredMatches = make([]fuzzy.Match, len(namespaces))
			for i, ns := range namespaces {
				filteredMatches[i] = fuzzy.Match{Str: ns}
				matchList.AddItem(ns, "", 0, nil)
			}
		} else {
			filteredMatches = fuzzy.Find(text, namespaces)
			for _, match := range filteredMatches {
				matchList.AddItem(match.Str, "", 0, nil)
			}
		}
		
		if len(filteredMatches) > 0 {
			matchList.SetCurrentItem(0)
		}
	}

	// Initialize with all namespaces
	updateMatches("")

	// Handle input changes for live filtering
	inputField.SetChangedFunc(func(text string) {
		updateMatches(text)
	})

	// Handle key events for navigation and selection
	inputField.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlJ: // Navigate down
			if len(filteredMatches) > 0 {
				selectedIndex = (selectedIndex + 1) % len(filteredMatches)
				matchList.SetCurrentItem(selectedIndex)
			}
			return nil
		case tcell.KeyCtrlK: // Navigate up
			if len(filteredMatches) > 0 {
				selectedIndex = (selectedIndex - 1 + len(filteredMatches)) % len(filteredMatches)
				matchList.SetCurrentItem(selectedIndex)
			}
			return nil
		case tcell.KeyEnter: // Select current match
			if len(filteredMatches) > 0 {
				selectedNs := filteredMatches[selectedIndex].Str
				onSelect(selectedNs)
				pages.RemovePage("namespace-selector")
				pages.SwitchToPage("explorer")
			}
			return nil
		case tcell.KeyEscape: // Cancel
			pages.RemovePage("namespace-selector")
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
		SetTitle(" Namespace Selector (Ctrl+J/K to navigate, Enter to select, Esc to cancel) ").
		SetBorderColor(tcell.ColorLightBlue).
		SetTitleColor(tcell.ColorWhite).
		SetBackgroundColor(tcell.ColorBlack)

	pages.AddPage("namespace-selector", flex, true, false)
	pages.SwitchToPage("namespace-selector")
	v.app.SetFocus(inputField)
}
