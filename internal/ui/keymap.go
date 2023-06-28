package ui

import "github.com/charmbracelet/bubbles/key"

type keymap struct {
	reload key.Binding
	quit   key.Binding
}

func newKeymap() keymap {
	return keymap{
		reload: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "reload the running program"),
		),
		quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
	}
}
