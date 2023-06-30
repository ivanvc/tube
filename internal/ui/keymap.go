package ui

import "github.com/charmbracelet/bubbles/key"

type keymap struct {
	reload      key.Binding
	quit        key.Binding
	editCommand key.Binding

	editingSave   key.Binding
	editingCancel key.Binding
}

func newKeymap() keymap {
	return keymap{
		reload: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "reload the running program"),
		),
		quit: key.NewBinding(
			key.WithKeys("ctrl+c", "esc", "q"),
			key.WithHelp("q/esc", "quit"),
		),
		editCommand: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit command"),
		),
		editingCancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel editing command"),
		),
		editingSave: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "save command, stop current and start a new process"),
		),
	}
}
