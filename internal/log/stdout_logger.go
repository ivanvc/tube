package log

import (
	"bufio"
	stdlog "log"

	"github.com/charmbracelet/log"
)

// StdoutLogger holds a common logger writing to a reader to consume it from the UI.
type StdoutLogger struct {
	*log.Logger
}

// Returns a new StdoutLogger that outputs to the stdout.
func NewStdout() *StdoutLogger {
	return &StdoutLogger{log.Default()}
}

// Returns a standard log with the lever forced to Error.
func (l *StdoutLogger) GetStandardLogWithErrorLevel() *stdlog.Logger {
	return l.StandardLog(log.StandardLogOptions{
		ForceLevel: log.ErrorLevel,
	})
}

// Returns a standard log with standard options applied.
func (l *StdoutLogger) GetStandardLog() *stdlog.Logger {
	return l.StandardLog(log.StandardLogOptions{})
}

// Returns the Buffered Reader
func (l *StdoutLogger) Reader() *bufio.Reader {
	return nil
}

// Returns the underneath logger.
func (l *StdoutLogger) Log() *log.Logger {
	return l.Logger
}
