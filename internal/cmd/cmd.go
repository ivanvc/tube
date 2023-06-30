package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"syscall"

	"github.com/ivanvc/tube/internal/log"
)

// Manager has the running program initialized by the tunnel.
type Manager struct {
	*exec.Cmd
	logger log.Logger
}

// Returns a new command manager.
func New(logger log.Logger) *Manager {
	return &Manager{logger: logger}
}

// Runs the command.
func (m *Manager) Run(command []string) error {
	if len(command) == 0 {
		return fmt.Errorf("No program to run")
	}
	m.logger.Log().Info("Starting new process", "command", command[0], "args", command[1:])

	m.Cmd = exec.Command(command[0], command[1:]...)
	m.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	stdout, err := m.StdoutPipe()
	if err != nil {
		m.logger.Log().Error("Error piping stdout", "error", err)
		return err
	}
	go pipeOutput("stdout", stdout, m.logger)

	stderr, err := m.StderrPipe()
	if err != nil {
		m.logger.Log().Error("Error piping stderr", "error", err)
		return err
	}
	go pipeOutput("stderr", stderr, m.logger)

	if err := m.Cmd.Run(); err != nil {
		m.logger.Log().Error("Error executing process", "error", err)
		return err
	}
	m.logger.Log().Info("Process exited", "command", command[0])

	return nil
}

// Stops the process, wait for it to stop.
func (m *Manager) Stop() error {
	if m.Cmd == nil {
		return nil
	}

	if err := syscall.Kill(-m.Process.Pid, syscall.SIGKILL); err != nil {
		m.logger.Log().Error("Error trying to stop command", "command", m.Args[0], "error", err)
		return err
	}
	_, err := m.Process.Wait()
	if err != nil {
		m.logger.Log().Error("Error trying to stop command", "command", m.Args[0], "error", err)
		return err
	}
	return nil
}

// Kills the process.
func (m *Manager) Kill() error {
	if m.Cmd == nil || !m.ProcessState.Exited() {
		return nil
	}

	if err := m.Process.Kill(); err != nil {
		m.logger.Log().Error("Error killing command", "command", m.Args[0], "error", err)
		return err
	}
	return nil
}

func pipeOutput(t string, r io.ReadCloser, logger log.Logger) {
	reader := bufio.NewReader(r)
	for {
		line, err := reader.ReadSlice('\n')
		if err == io.EOF {
			return
		}
		if err != nil {
			logger.Log().Error("Error reading "+t, "error", err)
			return
		}
		logger.Log().Print(string(line[:len(line)-1]))
	}
}
