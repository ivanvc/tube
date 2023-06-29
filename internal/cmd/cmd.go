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
	logger *log.Logger
}

// Returns a new command manager.
func New(logger *log.Logger) *Manager {
	return &Manager{logger: logger}
}

// Runs the command.
func (m *Manager) Run(command []string, sub chan string) error {
	if len(command) == 0 {
		return fmt.Errorf("No program to run")
	}
	m.logger.Info("Starting new process", "command", command[0], "args", command[1:])

	m.Cmd = exec.Command(command[0], command[1:]...)
	m.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	stdout, err := m.StdoutPipe()
	if err != nil {
		m.logger.Error("Error piping stdout", "error", err)
		return err
	}
	go pipeToSub("stdout", stdout, m.logger, sub)

	stderr, err := m.StderrPipe()
	if err != nil {
		m.logger.Error("Error piping stderr", "error", err)
		return err
	}
	go pipeToSub("stderr", stderr, m.logger, sub)

	if err := m.Cmd.Run(); err != nil {
		m.logger.Error("Error executing process", "error", err)
		return err
	}
	m.logger.Info("Process exited", "command", command[0])

	return nil
}

// Stops the process, wait for it to stop.
func (m *Manager) Stop() error {
	if m.Cmd == nil {
		return nil
	}

	if err := syscall.Kill(-m.Process.Pid, syscall.SIGKILL); err != nil {
		m.logger.Error("Error trying to stop command", "command", m.Args[0], "error", err)
		return err
	}
	_, err := m.Process.Wait()
	if err != nil {
		m.logger.Error("Error trying to stop command", "command", m.Args[0], "error", err)
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
		m.logger.Error("Error killing command", "command", m.Args[0], "error", err)
		return err
	}
	return nil
}

func pipeToSub(t string, r io.ReadCloser, log *log.Logger, sub chan string) {
	reader := bufio.NewReader(r)
	for {
		line, err := reader.ReadSlice('\n')
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Error("Error reading "+t, "error", err)
			return
		}
		sub <- string(line)
	}
}
