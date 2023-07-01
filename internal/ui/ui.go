package ui

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ivanvc/tube/internal/cmd"
	"github.com/ivanvc/tube/internal/config"
	"github.com/ivanvc/tube/internal/log"
	"github.com/ivanvc/tube/internal/server"
	"github.com/ivanvc/tube/internal/ui/styles"
)

type newLineMsg string
type listenerReadyMsg string
type serverTerminatedMsg struct{}

type ui struct {
	cfg    *config.Config
	server *server.Server

	width          int
	viewportWidth  int
	viewportHeight int
	spinner        spinner.Model
	keymap         keymap
	help           help.Model
	textInput      textinput.Model
	ready          bool
	addr           string

	linesChan       chan string
	logger          log.Logger
	viewportContent []string
	editingCommand  bool

	manager *cmd.Manager
}

const maxLines = 1000

// Returns a new UI
func New(cfg *config.Config, logger log.Logger, server *server.Server) *ui {
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
		viewportContent: make([]string, 0, maxLines),
		manager:         cmd.New(logger),
		textInput:       textinput.New(),
	}
}

// Init implements tea.Model.
func (ui ui) Init() tea.Cmd {
	return tea.Batch(
		ui.spinner.Tick,
		textinput.Blink,
		listenForLogs(ui.linesChan, ui.logger.Reader()),
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
			case key.Matches(msg, ui.keymap.editing.save):
				ui.editingCommand = false
				ui.cfg.ExecCommand = strings.Split(ui.textInput.Value(), " ")
				cmds = append(cmds, tea.Sequence(
					stopCommand(ui.manager),
					startCommand(ui.cfg, ui.manager),
				))
			case key.Matches(msg, ui.keymap.editing.cancel):
				ui.editingCommand = false
			case key.Matches(msg, ui.keymap.editing.quit):
				return ui, quitSeq(ui)
			}
		} else {
			switch {
			case key.Matches(msg, ui.keymap.quit):
				return ui, quitSeq(ui)
			case key.Matches(msg, ui.keymap.reload):
				ui.logger.Log().Info("Reloading command")
				cmds = append(cmds, tea.Sequence(
					stopCommand(ui.manager),
					startCommand(ui.cfg, ui.manager),
				))
			case key.Matches(msg, ui.keymap.editCommand):
				ui.editingCommand = true
				ui.textInput.SetValue(strings.Join(ui.cfg.ExecCommand, " "))
				ui.textInput.Focus()
				return ui, tea.Batch(cmds...)
			}
		}
	case tea.WindowSizeMsg:
		verticalMarginHeight := lipgloss.Height(ui.footerView()) + 1
		const horizontalPadding = 2 * 2
		ui.width = msg.Width
		ui.viewportWidth = msg.Width - 2
		ui.viewportHeight = msg.Height - verticalMarginHeight
		if !ui.ready {
			ui.ready = true
		}
		ui.textInput.Width = msg.Width - lipgloss.Width(logo) - 2
	case spinner.TickMsg:
		ui.spinner, cmd = ui.spinner.Update(msg)
		cmds = append(cmds, cmd)
	case newLineMsg:
		ui.viewportContent = append(ui.viewportContent, string(msg))
		if len(ui.viewportContent) > maxLines {
			ui.viewportContent = ui.viewportContent[len(ui.viewportContent)-maxLines:]
		}

		cmds = append(cmds, waitForLines(ui.linesChan))
	case listenerReadyMsg:
		ui.addr = string(msg)
		cmds = append(cmds, tea.Batch(
			startServer(ui.server, ui.logger),
			startCommand(ui.cfg, ui.manager),
		))
	}

	ui.textInput, cmd = ui.textInput.Update(msg)

	return ui, tea.Batch(append(cmds, cmd)...)
}

// View implements tea.Model.
func (ui ui) View() string {
	if !ui.ready {
		return "Loading..."
	}

	logLines := strings.Split(styles.ViewportContent.MaxWidth(ui.viewportWidth-2).Render(strings.Join(ui.viewportContent, "")), "\n")
	logLines = logLines[max(0, len(logLines)-ui.viewportHeight):]

	return fmt.Sprintf(
		"%s\n%s",
		styles.Viewport.
			Width(ui.viewportWidth).
			Height(ui.viewportHeight).
			Render(
				styles.ViewportContent.
					MaxWidth(ui.viewportWidth-2).
					Height(ui.viewportHeight).
					MaxHeight(ui.viewportHeight).
					Render(strings.Join(logLines, "\n")),
			),
		ui.footerView(),
	)
}

func (ui ui) footerView() string {
	var s string
	if len(ui.addr) == 0 {
		s = fmt.Sprintf("%s Establishing connection...", ui.spinner.View())
	} else if !ui.editingCommand {
		s = fmt.Sprintf("🌐 %s", styles.Link.Render(ui.addr))
	} else {
		s = ui.textInput.View()
	}
	s = styles.FooterText.Render(s)
	hv := ui.helpView()
	l := styles.Logo.Render(logo)
	if lipgloss.Width(l+hv) > ui.width {
		return fmt.Sprintf(
			"%s\n%s",
			styles.Footer.Render(lipgloss.JoinHorizontal(lipgloss.Center, l, s)),
			styles.Help.Padding(0, 1).Render(hv),
		)
	}
	return styles.Footer.Render(
		lipgloss.JoinHorizontal(lipgloss.Center, l,
			lipgloss.JoinVertical(lipgloss.Top, s, hv),
		),
	)
}

func (ui ui) helpView() string {
	if ui.editingCommand {
		return ui.help.ShortHelpView([]key.Binding{
			ui.keymap.editing.save,
			ui.keymap.editing.cancel,
			ui.keymap.editing.quit,
		})
	} else {
		return ui.help.ShortHelpView([]key.Binding{
			ui.keymap.reload,
			ui.keymap.editCommand,
			ui.keymap.quit,
		})
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

func startListener(server *server.Server, logger log.Logger) tea.Cmd {
	return func() tea.Msg {
		address, err := server.StartListener()
		if err != nil {
			logger.Log().Fatal("error initializing listener", "error", err)
		}

		return listenerReadyMsg(address)
	}
}

func startServer(server *server.Server, logger log.Logger) tea.Cmd {
	return func() tea.Msg {
		if err := server.Serve(); err != nil {
			logger.Log().Fatal("error initializing http server", "error", err)
		}

		return serverTerminatedMsg{}
	}
}

func startCommand(cfg *config.Config, mgr *cmd.Manager) tea.Cmd {
	return func() tea.Msg {
		mgr.Run(cfg.ExecCommand)
		return nil
	}
}

func stopCommand(mgr *cmd.Manager) tea.Cmd {
	return func() tea.Msg {
		mgr.Stop()
		return nil
	}
}

func quitSeq(ui ui) tea.Cmd {
	return tea.Sequence(
		closeWatcher(ui.watcher),
		stopCommand(ui.manager),
		tea.Quit,
	)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
