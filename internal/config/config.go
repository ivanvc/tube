package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const envVarPrefix = "TUBE"

// Config holds the application configuration options.
type Config struct {
	ListenHost   string
	ListenPort   string
	ListenScheme string

	ServerBaseURL string

	ExecCommand     []string
	WatchForChanges bool

	StandaloneMode bool
	ShowVersion    bool
}

// Loads the configuration.
func Load() *Config {
	c := new(Config)
	loadStringOption(
		&c.ListenHost,
		"host",
		"localhost",
		"The host where the traffic will be forwarded to.",
	)
	loadStringOption(
		&c.ListenScheme,
		"scheme",
		"http",
		"The scheme to use for the forwarding.",
	)
	loadStringOption(
		&c.ServerBaseURL,
		"server-base-url",
		"https://localtunnel.me",
		"The local tunner server URL.",
	)
	loadBoolOption(
		&c.StandaloneMode,
		"standalone",
		false,
		"Set this option if you don't want to use the Terminal UI.",
	)
	loadBoolOption(
		&c.WatchForChanges,
		"watch",
		false,
		"Watch for changes in the current directory, and restart command.",
	)
	loadBoolOption(
		&c.ShowVersion,
		"version",
		false,
		"Print the version and exit.",
	)
	flag.Usage = func() {
		fmt.Fprintf(
			flag.CommandLine.Output(),
			`Usage:	%s [options] [port] [command to execute...]:

Starts a localtunnel.me tunnel on the  specified port. You can specify the
options by argument  flags,  or  by  setting an environment variable, i.e.
TUBE_HOST or --host. Arguments take precedence over environment variables.

The port can be either specified  as the first argument  or  the TUBE_PORT
environment variable.
The  command  to execute, is optional, it can be  the  last  argument,  of
specied by setting TUBE_EXEC_COMMAND.

Options:
`,
			os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	loadArgumentOptions(&c.ListenPort, &c.ExecCommand)
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

func loadBoolOption(ptr *bool, option string, fallback bool, help string) {
	flag.BoolVar(
		ptr,
		option,
		parseBool(loadEnvVar(option, strconv.FormatBool(fallback))),
		help,
	)
}

func loadStringOption(ptr *string, option, fallback, help string) {
	flag.StringVar(ptr, option, loadEnvVar(option, fallback), help)
}

func loadEnvVar(option, fallback string) string {
	if v, ok := os.LookupEnv(formatEnvVar(option)); ok {
		return v
	}
	return fallback
}

func loadArgumentOptions(port *string, program *[]string) {
	if v, ok := os.LookupEnv(formatEnvVar("port")); ok {
		*port = v
	}
	if v, ok := os.LookupEnv(formatEnvVar("exec-command")); ok {
		*program = strings.Split(v, " ")
	}

	if len(flag.Args()) > 0 {
		if _, err := strconv.Atoi(flag.Arg(0)); err == nil {
			*port = flag.Arg(0)
			if len(flag.Args()) > 2 {
				*program = flag.Args()[1:]
			}
		} else {
			*program = flag.Args()
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

func parseBool(value string) bool {
	b, err := strconv.ParseBool(value)
	if err != nil {
		return false
	}
	return b
}
