package ui

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ivanvc/lt/internal/config"
	"github.com/ivanvc/lt/internal/log"
	"github.com/ivanvc/lt/internal/server"
	"github.com/ivanvc/lt/internal/ui/styles"
)

type newLineMsg string
type listenerReadyMsg string
type serverTerminatedMsg struct{}

type ui struct {
	cfg    *config.Config
	server *server.Server

	spinner  spinner.Model
	keymap   keymap
	help     help.Model
	viewport viewport.Model
	ready    bool
	addr     string

	linesChan       chan string
	logger          *log.Logger
	viewportContent []string
}

const maxLines = 1000

// Returns a new UI
func New(cfg *config.Config, logger *log.Logger, server *server.Server) *ui {

	s := spinner.New()
	s.Spinner = spinner.MiniDot
	s.Style = styles.Spinner

	return &ui{
		cfg:             cfg,
		server:          server,
		keymap:          newKeymap(),
		spinner:         s,
		help:            help.New(),
		linesChan:       make(chan string),
		logger:          logger,
		viewportContent: make([]string, maxLines),
	}
}

// Init implements tea.Model.
func (ui ui) Init() tea.Cmd {
	return tea.Batch(
		ui.spinner.Tick,
		listenForLogs(ui.linesChan, ui.logger.Reader),
		waitForLines(ui.linesChan),
		startListener(ui.server, ui.logger),
	)
}

// Update implements tea.Model.
func (ui ui) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, ui.keymap.quit):
			return ui, tea.Quit
		}
	case tea.WindowSizeMsg:
		verticalMarginHeight := 1 + lipgloss.Height(ui.helpView()) // header + helpView
		if !ui.ready {
			ui.ready = true
			ui.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			ui.viewport.YPosition = verticalMarginHeight
			ui.viewport.Style = styles.Viewport
		} else {
			ui.viewport.Width = msg.Width
			ui.viewport.Height = msg.Height - verticalMarginHeight
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		ui.spinner, cmd = ui.spinner.Update(msg)
		return ui, cmd
	case newLineMsg:
		ui.viewportContent = append(ui.viewportContent, string(msg))
		if len(ui.viewportContent) > maxLines {
			ui.viewportContent = ui.viewportContent[len(ui.viewportContent)-maxLines:]
		}
		scrollToBottom := ui.viewport.AtBottom()

		ui.viewport.SetContent(strings.Join(ui.viewportContent, ""))
		if scrollToBottom {
			ui.viewport.GotoBottom()
		} else {
			ui.viewport.SetYOffset(ui.viewport.YOffset + 1)
		}

		return ui, waitForLines(ui.linesChan)
	case listenerReadyMsg:
		ui.addr = string(msg)
		cmds := tea.Batch(
			startServer(ui.server, ui.logger),
			spawnCommand(ui.cfg, ui.logger, ui.linesChan),
		)
		return ui, cmds
	}
	return ui, nil
}

// View implements tea.Model.
func (ui ui) View() string {
	if !ui.ready {
		return "Loading..."
	}

	return fmt.Sprintf("%s\n%s\n%s", ui.headerView(), ui.viewport.View(), ui.helpView())
}

func (ui ui) headerView() string {
	var s string
	if len(ui.addr) == 0 {
		s = fmt.Sprintf("%s Establishing connection...", ui.spinner.View())
	} else {
		s = fmt.Sprintf("🌐 %s", ui.addr)
	}
	return styles.Header.Render(s)
}

func (ui ui) helpView() string {
	return ui.help.ShortHelpView([]key.Binding{
		ui.keymap.reload,
		ui.keymap.quit,
	})
}

func waitForLines(sub chan string) tea.Cmd {
	return func() tea.Msg {
		return newLineMsg(<-sub)
	}
}

func listenForLogs(sub chan string, reader *bufio.Reader) tea.Cmd {
	return func() tea.Msg {
		for {
			line, err := reader.ReadSlice('\n')
			if err != nil {
				return nil
			}
			sub <- string(line)
		}
	}
}

func startListener(server *server.Server, logger *log.Logger) tea.Cmd {
	return func() tea.Msg {
		address, err := server.StartListener()
		if err != nil {
			logger.Fatal("error initializing listener", "error", err)
		}

		return listenerReadyMsg(address)
	}
}

func startServer(server *server.Server, logger *log.Logger) tea.Cmd {
	return func() tea.Msg {
		if err := server.Serve(); err != nil {
			logger.Fatal("error initializing http server", "error", err)
		}

		return serverTerminatedMsg{}
	}
}

func spawnCommand(cfg *config.Config, logger *log.Logger, sub chan string) tea.Cmd {
	return func() tea.Msg {
		if len(cfg.ExecProgram) == 0 {
			return nil
		}
		logger.Info("Starting program", "program", cfg.ExecProgram[0], "args", cfg.ExecProgram[1:])

		cmd := exec.Command(cfg.ExecProgram[0], cfg.ExecProgram[1:]...)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			logger.Error("Error piping stdout", "error", err)
		}
		reader := bufio.NewReader(stdout)
		go func() {
			for {
				line, err := reader.ReadSlice('\n')
				if err != nil {
					logger.Error("Error reading stdout", "error", err)
					return
				}
				sub <- string(line)
			}
		}()

		stderr, err := cmd.StderrPipe()
		er := bufio.NewReader(stderr)
		if err != nil {
			logger.Error("Error piping stderr", "error", err)
		}

		go func() {
			for {
				line, err := er.ReadSlice('\n')
				if err != nil {
					logger.Error("Error reading sterr", "error", err)
					return
				}
				sub <- string(line)
			}
		}()
		if err := cmd.Run(); err != nil {
			logger.Error("Error executing program", "error", err)
		}
		logger.Info("Program exited", "error", cmd.Err)
		return nil
	}
}
