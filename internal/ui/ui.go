package ui

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ivanvc/lt/internal/cmd"
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

	spinner   spinner.Model
	keymap    keymap
	help      help.Model
	viewport  viewport.Model
	textInput textinput.Model
	ready     bool
	addr      string

	linesChan       chan string
	logger          *log.Logger
	viewportContent []string
	scrollLock      bool
	editingCommand  bool

	manager *cmd.Manager
}

const maxLines = 1000

// Returns a new UI
func New(cfg *config.Config, logger *log.Logger, server *server.Server) *ui {
	s := spinner.New()
	s.Spinner = spinner.MiniDot
	s.Style = styles.FooterText

	return &ui{
		cfg:             cfg,
		server:          server,
		keymap:          newKeymap(),
		spinner:         s,
		help:            help.New(),
		linesChan:       make(chan string),
		logger:          logger,
		viewportContent: make([]string, maxLines),
		manager:         cmd.New(logger),
		textInput:       textinput.New(),
	}
}

// Init implements tea.Model.
func (ui ui) Init() tea.Cmd {
	return tea.Batch(
		ui.spinner.Tick,
		textinput.Blink,
		listenForLogs(ui.linesChan, ui.logger.Reader),
		waitForLines(ui.linesChan),
		startListener(ui.server, ui.logger),
	)
}

// Update implements tea.Model.
func (ui ui) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmds []tea.Cmd
		cmd  tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if ui.editingCommand {
			switch {
			case key.Matches(msg, ui.keymap.editingSave):
				ui.editingCommand = false
				ui.cfg.ExecProgram = strings.Split(ui.textInput.Value(), " ")
				cmds = append(cmds, tea.Sequence(
					stopCommand(ui.manager),
					startCommand(ui.cfg, ui.manager, ui.linesChan),
				))
			case key.Matches(msg, ui.keymap.editingCancel):
				ui.editingCommand = false
			case msg.String() == "ctrl+c":
				return ui, tea.Sequence(
					tea.Batch(stopCommand(ui.manager), stopServer(ui.server)),
					tea.Quit,
				)
			}
		} else {
			switch {
			case key.Matches(msg, ui.keymap.quit):
				return ui, tea.Sequence(
					tea.Batch(stopCommand(ui.manager), stopServer(ui.server)),
					tea.Quit,
				)
			case key.Matches(msg, ui.keymap.reload):
				ui.logger.Print("Reloading")
				cmds = append(cmds, tea.Sequence(
					stopCommand(ui.manager),
					startCommand(ui.cfg, ui.manager, ui.linesChan),
				))
			case key.Matches(msg, ui.keymap.editCommand):
				ui.editingCommand = true
				ui.textInput.SetValue(strings.Join(ui.cfg.ExecProgram, " "))
				ui.textInput.Focus()
				return ui, tea.Batch(cmds...)
			case key.Matches(msg, ui.keymap.scrollLock):
				ui.scrollLock = !ui.scrollLock
			}
		}
	case tea.WindowSizeMsg:
		verticalMarginHeight := lipgloss.Height(ui.footerView()) + lipgloss.Height(ui.helpView())
		if !ui.ready {
			ui.ready = true
			ui.viewport.HighPerformanceRendering = true
			ui.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			ui.viewport.YPosition = int(lipgloss.Top)
			ui.viewport.Style = styles.Viewport
		} else {
			ui.viewport.Width = msg.Width
			ui.viewport.Height = msg.Height - verticalMarginHeight
		}
		ui.textInput.Width = msg.Width - lipgloss.Width(logo) - 2
		cmds = append(cmds, viewport.Sync(ui.viewport))
	case spinner.TickMsg:
		ui.spinner, cmd = ui.spinner.Update(msg)
		cmds = append(cmds, cmd)
	case newLineMsg:
		ui.viewportContent = append(ui.viewportContent, string(msg))
		if len(ui.viewportContent) > maxLines {
			ui.viewportContent = ui.viewportContent[len(ui.viewportContent)-maxLines:]
		}

		ui.viewport.SetContent(strings.Join(ui.viewportContent, ""))
		if !ui.scrollLock {
			ui.viewport.GotoBottom()
		}

		cmds = append(cmds, waitForLines(ui.linesChan))
	case listenerReadyMsg:
		ui.addr = string(msg)
		cmds = append(cmds, tea.Batch(
			startServer(ui.server, ui.logger),
			startCommand(ui.cfg, ui.manager, ui.linesChan),
		))
	}

	ui.viewport, cmd = ui.viewport.Update(msg)
	ui.textInput, cmd = ui.textInput.Update(msg)

	return ui, tea.Batch(append(cmds, cmd)...)
}

// View implements tea.Model.
func (ui ui) View() string {
	if !ui.ready {
		return "Loading..."
	}

	return fmt.Sprintf("%s\n%s\n%s", ui.viewport.View(), ui.footerView(), ui.helpView())
}

func (ui ui) footerView() string {
	var s string
	if len(ui.addr) == 0 {
		s = fmt.Sprintf("%s Establishing connection...", ui.spinner.View())
	} else if !ui.editingCommand {
		s = fmt.Sprintf("🌐 %s", ui.addr)
	} else {
		s = ui.textInput.View()
	}
	return styles.Footer.Render(lipgloss.JoinHorizontal(
		lipgloss.Center, styles.Logo.Render(logo), styles.FooterText.Render(s),
	))
}

func (ui ui) helpView() string {
	if ui.editingCommand {
		return styles.Help.Render(ui.help.ShortHelpView([]key.Binding{
			ui.keymap.editingCancel,
			ui.keymap.editingSave,
		}))
	} else {
		return styles.Help.Render(ui.help.ShortHelpView([]key.Binding{
			ui.keymap.reload,
			ui.keymap.editCommand,
			ui.keymap.scrollLock,
			ui.keymap.quit,
		}))
	}
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

func startCommand(cfg *config.Config, mgr *cmd.Manager, sub chan string) tea.Cmd {
	return func() tea.Msg {
		mgr.Run(cfg.ExecProgram, sub)
		return nil
	}
}

func stopCommand(mgr *cmd.Manager) tea.Cmd {
	return func() tea.Msg {
		mgr.Stop()
		return nil
	}
}

func stopServer(s *server.Server) tea.Cmd {
	return func() tea.Msg {
		s.Close()
		return nil
	}
}
