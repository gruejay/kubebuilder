package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Explorer struct {
	app *tview.Application
}

func NewExplorer(app *tview.Application) *Explorer {
	return &Explorer{app: app}
}

func (e *Explorer) CreateExplorerView(namespace string, resourceType string) *tview.List {
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

func (e *Explorer) UpdateExplorerTitle(list *tview.List, namespace string, resourceType string) {
	title := fmt.Sprintf(" Explorer Mode - Namespace: %s | Resource: %s (Press 'n'/'r' to change) ", namespace, resourceType)
	list.SetTitle(title)
}

func (e *Explorer) CreateResourceSelector(pages *tview.Pages, onSelect func(string)) {
	resourceTypes := []string{"all", "pods", "services", "deployments", "configmaps", "secrets", "ingresses", "daemonsets", "statefulsets", "jobs", "cronjobs"}
	title := " Resource Type Selector (Ctrl+J/K to navigate, Enter to select, Esc to cancel) "
	pageName := "resource-selector"
	fs := NewFuzzySelector(resourceTypes, title, pageName, pages, onSelect)

	inputField, matchList, err := fs.createSelector()
	if err != nil {
		return
	}
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

}
func (e *Explorer) CreateNamespaceSelector(namespaces []string, pages *tview.Pages, onSelect func(string)) {
	title := " Namespace Selector (Ctrl+J/K to navigate, Enter to select, Esc to cancel) "
	pageName := "namespace-selector"
	fs := NewFuzzySelector(namespaces, title, pageName, pages, onSelect)
	inputField, matchList, err := fs.createSelector()
	if err != nil {
		return
	}
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

}
