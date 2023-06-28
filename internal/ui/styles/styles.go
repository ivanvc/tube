package styles

import "github.com/charmbracelet/lipgloss"

var (
	Spinner  = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
	Header   = lipgloss.NewStyle().PaddingRight(1).Foreground(lipgloss.Color("3")).Padding(0, 1)
	Viewport = lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("4")).Padding(0, 1)
)
