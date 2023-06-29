package server

import (
	"net/http"

	"github.com/ivanvc/lt/internal/config"
	"github.com/ivanvc/lt/internal/log"
	"github.com/localtunnel/go-localtunnel"
)

type Server struct {
	cfg      *config.Config
	logger   *log.Logger
	server   *http.Server
	listener *localtunnel.Listener
}

// Returns a new Server with the reverse proxy
func New(cfg *config.Config, logger *log.Logger) *Server {
	server := &http.Server{
		Handler:  newProxy(cfg, logger),
		ErrorLog: logger.GetStandardLogWithErrorLevel(),
	}

	return &Server{cfg: cfg, logger: logger, server: server}
}

// Starts localtunnel listener.
func (s *Server) StartListener() (string, error) {
	s.logger.Infof("forwarding traffic to %s", s.cfg.ListenURL())
	var err error
	s.listener, err = localtunnel.Listen(localtunnel.Options{
		Log:     s.logger.GetStandardLog(),
		BaseURL: s.cfg.ServerBaseURL,
	})
	if err != nil {
		return "", err
	}
	return s.listener.Addr().String(), nil
}

// Serve the Proxy for the listener.
func (s *Server) Serve() error {
	return s.server.Serve(s.listener)
}

// Terminates the HTTP server.
func (s *Server) Close() error {
	return s.server.Close()
}
