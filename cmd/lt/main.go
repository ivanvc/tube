package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/charmbracelet/log"
	localtunnel "github.com/localtunnel/go-localtunnel"
)

var (
	port   = flag.String("port", "", "The port to forward traffic to.")
	host   = flag.String("host", "localhost", "The host to forward traffic to.")
	scheme = flag.String("scheme", "http", "The host to forward traffic to.")
)

func main() {
	flag.Parse()

	if len(*port) == 0 {
		log.Fatal("port not given")
	}

	hostWithPort := fmt.Sprintf("%s:%s", *host, *port)

	stdlog := log.Default().StandardLog(log.StandardLogOptions{})

	// Setup a listener for localtunnel
	listener, err := localtunnel.Listen(localtunnel.Options{Log: stdlog})
	if err != nil {
		log.Fatal("error initializing listener", "error", err)
	}
	log.Infof("forwarding traffic to %s://%s", *scheme, hostWithPort)

	server := http.Server{Handler: newProxy(*scheme, hostWithPort)}
	server.Serve(listener)
}

type proxy struct {
	httputil.ReverseProxy
	host   string
	scheme string
}

func newProxy(scheme, host string) *proxy {
	p := &proxy{host: host}
	p.Director = p.getDirector
	return p
}

func (p *proxy) getDirector(req *http.Request) {
	defer func() {
		path := req.URL.Path
		if len(req.URL.RawQuery) > 0 {
			path += fmt.Sprintf("?%s", req.URL.RawQuery)
		}
		log.Infof("%s %s", req.Method, path)
	}()

	req.URL.Scheme = p.scheme
	req.URL.Host = p.host
}
