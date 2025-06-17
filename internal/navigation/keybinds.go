package navigation

import (
	"kubeguide/internal/modes"

	"github.com/gdamore/tcell/v2"
)

type KeyBind struct {
	Key         tcell.Key
	Rune        rune
	Description string
	Mode        modes.Mode
	Handler     func() bool // Returns true if event was handled
}

type KeyBindings struct {
	Bindings map[modes.Mode][]KeyBind
}

func NewKeyBindings() *KeyBindings {
	return &KeyBindings{
		Bindings: make(map[modes.Mode][]KeyBind),
	}
}

func (kb *KeyBindings) AddBinding(binding KeyBind) {
	kb.Bindings[binding.Mode] = append(kb.Bindings[binding.Mode], binding)
}

func (kb *KeyBindings) GetBindings(mode modes.Mode) []KeyBind {
	return kb.Bindings[mode]
}

func (kb *KeyBindings) HandleKey(mode modes.Mode, event *tcell.EventKey) bool {
	bindings := kb.GetBindings(mode)
	for _, binding := range bindings {
		if binding.Key != tcell.KeyNUL && event.Key() == binding.Key {
			if binding.Handler != nil {
				return binding.Handler()
			}
			return true
		}
		if binding.Rune != 0 && event.Rune() == binding.Rune {
			if binding.Handler != nil {
				return binding.Handler()
			}
			return true
		}
	}
	return false
}

// Default key bindings for all modes
func GetDefaultKeyBindings() *KeyBindings {
	kb := NewKeyBindings()
	
	// Global key bindings (apply to all modes)
	globalBindings := []KeyBind{
		{Key: tcell.KeyEsc, Description: "Go back/Exit", Mode: modes.Welcome},
		{Key: tcell.KeyEsc, Description: "Go back/Exit", Mode: modes.Explorer},
		{Key: tcell.KeyEsc, Description: "Go back/Exit", Mode: modes.ResourceDetails},
		{Rune: 'q', Description: "Quit application", Mode: modes.Welcome},
		{Rune: 'q', Description: "Quit application", Mode: modes.Explorer},
		{Rune: 'q', Description: "Quit application", Mode: modes.ResourceDetails},
		{Rune: '?', Description: "Show help", Mode: modes.Welcome},
		{Rune: '?', Description: "Show help", Mode: modes.Explorer},
		{Rune: '?', Description: "Show help", Mode: modes.ResourceDetails},
	}
	
	// Welcome mode specific bindings
	welcomeBindings := []KeyBind{
		{Rune: 'e', Description: "Enter Explorer mode", Mode: modes.Welcome},
	}
	
	// Explorer mode specific bindings
	explorerBindings := []KeyBind{
		{Rune: 'n', Description: "Switch namespace", Mode: modes.Explorer},
		{Rune: 'r', Description: "Switch resource type", Mode: modes.Explorer},
		{Key: tcell.KeyEnter, Description: "View resource details", Mode: modes.Explorer},
		{Rune: 'j', Description: "Move down", Mode: modes.Explorer},
		{Rune: 'k', Description: "Move up", Mode: modes.Explorer},
		{Rune: 'a', Description: "AI analysis (failed pods)", Mode: modes.Explorer},
	}
	
	// Add all bindings
	allBindings := append(globalBindings, welcomeBindings...)
	allBindings = append(allBindings, explorerBindings...)
	
	for _, binding := range allBindings {
		kb.AddBinding(binding)
	}
	
	return kb
}
