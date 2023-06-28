package log

import (
	"bufio"
	"io"
	stdlog "log"

	"github.com/charmbracelet/log"
)

// Logger holds a common logger writing to a reader to consume it from the UI.
type Logger struct {
	*log.Logger
	Reader *bufio.Reader
}

// Returns a new Logger.
func New() *Logger {
	r, w := io.Pipe()
	return &Logger{
		Logger: log.NewWithOptions(w, log.Options{ReportTimestamp: true}),
		Reader: bufio.NewReader(r),
	}
}

// Returns a standard log with the lever forced to Error.
func (l *Logger) GetStandardLogWithErrorLevel() *stdlog.Logger {
	return l.StandardLog(log.StandardLogOptions{
		ForceLevel: log.ErrorLevel,
	})
}

// Returns a standard log with standard options applied.
func (l *Logger) GetStandardLog() *stdlog.Logger {
	return l.StandardLog(log.StandardLogOptions{})
}
