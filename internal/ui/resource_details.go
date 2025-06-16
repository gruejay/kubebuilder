package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ResourceDetails struct {
	name         string
	resourceType string
	content      string
}

func NewResourceDetails(name string, resourceType string, content string) ResourceDetails {
	return ResourceDetails{
		name:         name,
		resourceType: resourceType,
		content:      content,
	}
}

func (r *ResourceDetails) CreateView() tview.Primitive {
	textView := tview.NewTextView().
		SetTextColor(tcell.ColorWhite)
	fmt.Fprintf(textView, "%s", r.content)
	textView.SetBorder(true).
		SetBorderColor(tcell.ColorLightBlue).
		SetTitle(fmt.Sprintf(" %s: %s (Press Esc to return) ", r.resourceType, r.name)).
		SetTitleColor(tcell.ColorWhite)

	return textView
}
