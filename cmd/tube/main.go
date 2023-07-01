package main

import (
	"os"
	"os/signal"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"

	cmd "github.com/ivanvc/tube/internal/command"
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

func startStandalone(cfg *config.Config) {
	logger := intlog.NewStdout()
	if len(cfg.ExecCommand) > 0 {
		logger.SetPrefix("TUBE")
		log.TimestampStyle = log.TimestampStyle.Foreground(lipgloss.Color("3"))
		log.PrefixStyle = log.PrefixStyle.Foreground(lipgloss.Color("3"))
		log.SeparatorStyle = log.SeparatorStyle.Foreground(lipgloss.Color("11"))
	}
	server := server.New(cfg, logger)
	mgr := cmd.NewManager(logger, os.Stdout)
	watcher := cmd.NewWatcher(cfg, logger)
	defer mgr.Stop()
	defer watcher.Close()

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
	printTunnel := make(chan os.Signal)
	signal.Notify(printTunnel, syscall.SIGUSR1, syscall.SIGUSR2)

	go mgr.Run(cfg.ExecCommand)
	go watcher.Run()

	for {
		select {
		case <-watcher.Activity():
			mgr.Stop()
			go mgr.Run(cfg.ExecCommand)
		case <-printTunnel:
			log.Infof("Tunnel available at: %s", server.ListenerAddr())
		case <-done:
			return
		}
	}
}

func startTUI(cfg *config.Config) {
	p := tea.NewProgram(ui.New(cfg), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
