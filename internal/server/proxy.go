package server

import (
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/ivanvc/lt/internal/config"
	"github.com/ivanvc/lt/internal/log"
)

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
