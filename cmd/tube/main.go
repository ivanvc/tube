package main

import (
	"os"
	"os/signal"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"

	"github.com/ivanvc/tube/internal/cmd"
	"github.com/ivanvc/tube/internal/config"
	intlog "github.com/ivanvc/tube/internal/log"
	"github.com/ivanvc/tube/internal/server"
	"github.com/ivanvc/tube/internal/ui"
)

func main() {
	cfg := config.Load()
	if cfg.StandaloneMode {
		startStandalone(cfg)
	} else {
		startTUI(cfg)
	}
}

func startTUI(cfg *config.Config) {
	logger := intlog.NewBuffered()
	server := server.New(cfg, logger)

	p := tea.NewProgram(ui.New(cfg, logger, server), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func startStandalone(cfg *config.Config) {
	logger := intlog.NewStdout()
	server := server.New(cfg, logger)
	mgr := cmd.New(logger)
	if _, err := server.StartListener(); err != nil {
		logger.Fatal("error initializing listener", "error", err)
	}
	go func() {
		if err := server.Serve(); err != nil {
			logger.Fatal("error initializing http server", "error", err)
		}
	}()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go mgr.Run(cfg.ExecCommand)
	<-done
	mgr.Stop()
}
