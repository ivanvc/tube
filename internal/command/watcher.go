package command

import (
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/ivanvc/tube/internal/config"
	"github.com/ivanvc/tube/internal/log"
)

// Watcher holds a wrapper for a fsnotify.Watcher. It debounces the events.
type Watcher struct {
	*fsnotify.Watcher
	logger   log.Logger
	activity chan *fsnotify.Event
}

const timeout = 250 * time.Millisecond

// Returns a NewWatcher.
func NewWatcher(cfg *config.Config, logger log.Logger) *Watcher {
	var w *fsnotify.Watcher
	if cfg.WatchForChanges {
		var err error
		w, err = fsnotify.NewWatcher()
		if err != nil {
			logger.Log().Fatal("Error starting watcher", "error", err)
		}
	}
	return &Watcher{
		Watcher:  w,
		logger:   logger,
		activity: make(chan *fsnotify.Event),
	}
}

// Runs the watching loop.
func (w *Watcher) Run() {
	if w.Watcher == nil {
		return
	}

	if err := w.Add("."); err != nil {
		w.logger.Log().Error("Error adding watch for current directory", "error", err)
	}

	stream := make(chan *fsnotify.Event)
	go debounce(timeout, stream, func(event *fsnotify.Event) {
		w.logger.Log().Infof("%s modified, reloading command", event.Name)
		w.activity <- event
	})

	for {
		select {
		case event, ok := <-w.Events:
			if !ok {
				return
			}
			stream <- &event
		case err, ok := <-w.Errors:
			if !ok {
				return
			}
			w.logger.Log().Error("Error watching for changes", "error", err, "ok", ok)
		}
	}
}

// Closes the underlying fsnotify.Watcher, if initialized.
func (w *Watcher) Close() error {
	if w.Watcher != nil {
		return w.Watcher.Close()
	}
	return nil
}

// Returns the Activity channel.
func (w *Watcher) Activity() <-chan *fsnotify.Event {
	return w.activity
}

func debounce(d time.Duration, in chan *fsnotify.Event, cb func(*fsnotify.Event)) {
	var ev *fsnotify.Event
	for {
		select {
		case ev = <-in:
		case <-time.After(d):
			if ev != nil {
				cb(ev)
				ev = nil
			}
		}
	}
}
