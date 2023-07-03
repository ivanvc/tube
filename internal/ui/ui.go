package ui

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	cmd "github.com/ivanvc/tube/internal/command"
	"github.com/ivanvc/tube/internal/config"
	"github.com/ivanvc/tube/internal/log"
	"github.com/ivanvc/tube/internal/server"
	"github.com/ivanvc/tube/internal/ui/styles"
)

type newCommandLogLineMsg string
type newLogLineMsg string
type listenerReadyMsg string
type serverTerminatedMsg struct{}
type watcherGotChangesMsg struct{}

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

	logLinesChan    chan string
	commandLogsChan chan string
	changesChan     chan watcherGotChangesMsg
	commandReader   *bufio.Reader
	logger          log.Logger
	viewportContent []string
	editingCommand  bool

	manager *cmd.Manager
	watcher *cmd.Watcher
}

const maxLines = 1000

// Returns a new UI
func New(cfg *config.Config) *ui {
	s := spinner.New()
	s.Spinner = spinner.MiniDot
	s.Style = styles.FooterText
	ti := textinput.New()
	ti.Placeholder = "Command to execute"
	logger := log.NewBuffered()
	r, w := io.Pipe()

	return &ui{
		cfg:             cfg,
		server:          server.New(cfg, logger),
		keymap:          newKeymap(),
		spinner:         s,
		help:            help.New(),
		logLinesChan:    make(chan string),
		commandLogsChan: make(chan string),
		changesChan:     make(chan watcherGotChangesMsg),
		logger:          logger,
		viewportContent: make([]string, 0, maxLines),
		manager:         cmd.NewManager(logger, w),
		commandReader:   bufio.NewReader(r),
		textInput:       ti,
		watcher:         cmd.NewWatcher(cfg, logger),
	}
}

// Init implements tea.Model.
func (ui ui) Init() tea.Cmd {
	return tea.Batch(
		ui.spinner.Tick,
		textinput.Blink,
		listenForLogs(ui.logLinesChan, ui.logger.Reader()),
		waitForLogLines(ui.logLinesChan),
		listenForLogs(ui.commandLogsChan, ui.commandReader),
		waitForCommandLogs(ui.commandLogsChan),
		startListener(ui.server, ui.logger),
		listenForChanges(ui.watcher),
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
	case newCommandLogLineMsg:
		processLine(&ui.viewportContent, styles.CommandLogLine, string(msg))
		cmds = append(cmds, waitForCommandLogs(ui.commandLogsChan))
	case newLogLineMsg:
		processLine(&ui.viewportContent, styles.LogLine, string(msg))
		cmds = append(cmds, waitForLogLines(ui.logLinesChan))
	case watcherGotChangesMsg:
		ui.logger.Log().Info("Restarting")
		cmds = append(cmds,
			tea.Batch(
				tea.Sequence(
					stopCommand(ui.manager),
					startCommand(ui.cfg, ui.manager),
				),
				listenForChanges(ui.watcher),
			),
		)
	case listenerReadyMsg:
		ui.addr = string(msg)
		cmds = append(cmds, tea.Batch(
			startServer(ui.server, ui.logger),
			startCommand(ui.cfg, ui.manager),
			watchForChanges(ui.watcher),
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

func waitForCommandLogs(sub chan string) tea.Cmd {
	return func() tea.Msg {
		return newCommandLogLineMsg(<-sub)
	}
}

func waitForLogLines(sub chan string) tea.Cmd {
	return func() tea.Msg {
		return newLogLineMsg(<-sub)
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

func closeWatcher(w *cmd.Watcher) tea.Cmd {
	return func() tea.Msg {
		w.Close()
		return nil
	}
}

func listenForChanges(w *cmd.Watcher) tea.Cmd {
	return func() tea.Msg {
		<-w.Activity()
		return watcherGotChangesMsg{}
	}
}

func watchForChanges(w *cmd.Watcher) tea.Cmd {
	return func() tea.Msg {
		w.Run()
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

func processLine(lines *[]string, style lipgloss.Style, line string) {
	*lines = append(
		*lines,
		fmt.Sprintf("%s\n", style.Inline(true).Render(string(line))),
	)
	if len(*lines) > maxLines {
		*lines = (*lines)[len(*lines)-maxLines:]
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
