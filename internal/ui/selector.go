package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/sahilm/fuzzy"
)

type FuzzySelector struct {
	title    string
	pageName string
	pages    *tview.Pages
	onSelect func(string)
	items    []string
}

func NewFuzzySelector(items []string, title string, pageName string, pages *tview.Pages, onSelect func(string)) *FuzzySelector {
	fs := FuzzySelector{
		title:    title,
		pageName: pageName,
		pages:    pages,
		onSelect: onSelect,
		items:    items,
	}
	return &fs
}

func (fs *FuzzySelector) createSelector() (*tview.InputField, *tview.List, error) {
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
			filteredMatches = make([]fuzzy.Match, len(fs.items))
			for i, item := range fs.items {
				filteredMatches[i] = fuzzy.Match{Str: item}
				matchList.AddItem(item, "", 0, nil)
			}
		} else {
			filteredMatches = fuzzy.Find(text, fs.items)
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
				fs.onSelect(selectedItem)
				fs.pages.RemovePage(fs.pageName)
				fs.pages.SwitchToPage("explorer")
			}
			return nil
		case tcell.KeyEscape: // Cancel
			fs.pages.RemovePage(fs.pageName)
			fs.pages.SwitchToPage("explorer")
			return nil
		}
		return event
	})

	return inputField, matchList, nil
}
