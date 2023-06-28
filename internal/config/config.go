package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/log"
)

const envVarPrefix = "LT"

// Config holds the application configuration options.
type Config struct {
	ListenHost   string
	ListenPort   string
	ListenScheme string

	ServerBaseURL string

	ExecProgram []string
}

// Loads the configuration.
func Load() *Config {
	c := new(Config)
	loadOption(
		&c.ListenHost,
		"host",
		"localhost",
		"The host where the traffic will be forwarded to.",
	)
	loadOption(
		&c.ListenScheme,
		"scheme",
		"http",
		"The scheme to use for the forwarding.",
	)
	loadOption(
		&c.ServerBaseURL,
		"server-base-url",
		"https://localtunnel.me",
		"The local tunner server URL.",
	)
	flag.Parse()

	loadArgumentOptions(&c.ListenPort, &c.ExecProgram)
	if len(c.ListenPort) == 0 {
		log.Fatalf("Port needs to be specified, either by the %s environment variable, or by the first argument to the program", formatEnvVar("port"))
	}
	return c
}

// Returns the host:port pair.
func (c *Config) ListenHostWithPort() string {
	return fmt.Sprintf("%s:%s", c.ListenHost, c.ListenPort)
}

// Returns the URL where to listen.
func (c *Config) ListenURL() string {
	return fmt.Sprintf("%s://%s", c.ListenScheme, c.ListenHostWithPort())
}

func loadOption(ptr *string, option, fallback, help string) {
	if v, ok := os.LookupEnv(formatEnvVar(option)); ok {
		fallback = v
	}
	flag.StringVar(ptr, option, fallback, help)
}

func loadArgumentOptions(port *string, program *[]string) {
	if v, ok := os.LookupEnv(formatEnvVar("port")); ok {
		*port = v
	}
	if len(os.Args) > 1 {
		if _, err := strconv.Atoi(os.Args[1]); err == nil {
			*port = os.Args[1]
			if len(os.Args) > 2 {
				*program = os.Args[2:]
			}
		} else {
			*program = os.Args[1:]
		}
	}
}

func formatEnvVar(envVar string) string {
	return fmt.Sprintf(
		"%s_%s",
		envVarPrefix,
		strings.ReplaceAll(strings.ToUpper(envVar), "-", "_"),
	)
}
