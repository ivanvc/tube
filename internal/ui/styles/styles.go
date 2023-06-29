package styles

import "github.com/charmbracelet/lipgloss"

var (
	FooterText = lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
	Logo       = lipgloss.NewStyle().Foreground(lipgloss.Color("141")).PaddingRight(1)
	Footer     = lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("4"))
	Viewport   = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("4")).Padding(0, 1)
	Help = lipgloss.NewStyle().Padding(0, 1)
)
