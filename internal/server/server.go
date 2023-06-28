package server

import (
	"fmt"
	"net/http"
	"net/http/httputil"

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

type proxy struct {
	httputil.ReverseProxy
	cfg    *config.Config
	logger *log.Logger
}

func newProxy(cfg *config.Config, logger *log.Logger) *proxy {
	p := &proxy{cfg: cfg, logger: logger}
	p.ReverseProxy.ErrorLog = logger.GetStandardLogWithErrorLevel()
	p.Director = p.getDirector
	return p
}

func (p *proxy) getDirector(req *http.Request) {
	defer func() {
		path := req.URL.Path
		if len(req.URL.RawQuery) > 0 {
			path += fmt.Sprintf("?%s", req.URL.RawQuery)
		}
		p.logger.Infof("%s %s", req.Method, path)
	}()

	req.URL.Scheme = p.cfg.ListenScheme
	req.URL.Host = p.cfg.ListenHostWithPort()
}
