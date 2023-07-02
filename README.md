# Tube

Tube is a [Localtunnel] client. It uses the [go-localtunnel] client library to
establish a connection and expose a local port externally.

It can spawn a command which the tunnel will use, show its output, and reload it
on-demand, or watch for changes using [fsnotify], and automatically reload it.

<intro.gif>

## Installation

Download the binary from the [releases] page. Or run

```bash
go install github.com/ivanvc/tube@latest
```

## Usage

To start a tunnel, run:

```bash
tube [options] [local port] [command to execute]
```

The **command to execute** is optional, if specified will spawn the command, and
will show the output in the terminal output.

You can speficy any of the options by arguments to the program, or by an
environment variable, run `tube -h` to see all of the available options. The
port and command to execute can also be set from environment variables, by using
`TUBE_PORT` and `TUBE_EXEC_COMMAND`.

### Reload using watch

If you specify either `--watch` or the environment variable `TUBE_WATCH=1`, it
will watch for file changes in the current working directory using `fsnotify`.
Then, it will reload the **exec command** if specified.

### TUI mode

The default execution type has a terminal user interface (made with the
[Bubble Tea] framework). You can edit the command by typing `e`, and manually
reload with `r`.

<tui demo.gif>

### Standalone mode

It's also possible to run in standalone mode (using `--standalone` or by setting
`TUBE_STANDALONE=1`).

As the output of the executing program will be shown, if you want to see the
tunnel's URL, you can send either the `SIGUSR1` or `SIGUSR2` to `tube` (i.e.,
`pkill -USR1 tube`).

You can also manually reload the running command by sending `SIGHUP` to `tube`
(i.e., ` pkill -HUP tube`).

<standalone demo.gif>

## License

See [LICENSE](LICENSE) © [Ivan Valdes](https://github.com/ivanvc/)

[Localtunnel]: https://localtunnel.me
[go-localtunnel]: https://github.com/localtunnel/go-localtunnel
[releases]: https://github.com/ivanvc/tube/releases
[fsnotify]: https://github.com/fsnotify/fsnotify
[Bubble Tea]: https://github.com/charmbracelet/bubbletea
