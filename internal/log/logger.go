package log

import (
	"bufio"
	stdlog "log"

	"github.com/charmbracelet/log"
)

// Logger is the interface for a logger that can output to a buffer or stdout.
type Logger interface {
	Reader() *bufio.Reader
	GetStandardLogWithErrorLevel() *stdlog.Logger
	GetStandardLog() *stdlog.Logger
	Log() *log.Logger
}
