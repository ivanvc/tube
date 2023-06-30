package log

import (
	"bufio"
	"io"
	stdlog "log"

	"github.com/charmbracelet/log"
)

// BufferedLogger holds a common logger writing to a reader to consume it from the UI.
type BufferedLogger struct {
	*log.Logger
	reader *bufio.Reader
}

// Returns a new BufferedLogger that has the output to a buffered reader.
func NewBuffered() *BufferedLogger {
	r, w := io.Pipe()
	return &BufferedLogger{
		Logger: log.NewWithOptions(w, log.Options{ReportTimestamp: true}),
		reader: bufio.NewReader(r),
	}
}

// Returns a standard log with the lever forced to Error.
func (l *BufferedLogger) GetStandardLogWithErrorLevel() *stdlog.Logger {
	return l.StandardLog(log.StandardLogOptions{
		ForceLevel: log.ErrorLevel,
	})
}

// Returns a standard log with standard options applied.
func (l *BufferedLogger) GetStandardLog() *stdlog.Logger {
	return l.StandardLog(log.StandardLogOptions{})
}

// Returns the Buffered Reader
func (l *BufferedLogger) Reader() *bufio.Reader {
	return l.reader
}

// Returns the underneath logger.
func (l *BufferedLogger) Log() *log.Logger {
	return l.Logger
}
