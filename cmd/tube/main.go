package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"

	"github.com/ivanvc/tube/internal/config"
	intlog "github.com/ivanvc/tube/internal/log"
	"github.com/ivanvc/tube/internal/server"
	"github.com/ivanvc/tube/internal/ui"
)

func main() {
	cfg := config.Load()
	logger := intlog.New()
	server := server.New(cfg, logger)

	p := tea.NewProgram(ui.New(cfg, logger, server), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}