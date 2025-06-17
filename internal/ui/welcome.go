package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Welcome struct {
	message string
	title   string
}

func NewWelcome(title string, message string) *Welcome {
	if message == "" {
		message = "Welcome to kubebuide!\n\nPress e to explore\n\nPress q to quit\n\nPress esc to go back\n\nPress ? for help"
	}
	w := Welcome{
		title:   title,
		message: message,
	}
	return &w
}

func (w *Welcome) CreateWelcomeView() tview.Primitive {
	// Create a text view with large, centered text
	textView := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetTextColor(tcell.ColorWhite)
	fmt.Fprintf(textView, "%s", w.message)

	// Add a border to make it more visible
	textView.SetBorder(true).
		SetBorderColor(tcell.ColorLightBlue).
		SetTitle(w.title).
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
