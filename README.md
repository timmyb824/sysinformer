# sysinformer

A simple, portable system and web information CLI tool for macOS and Linux, written in Go.

## Features

- System information (OS, kernel, uptime, users, etc.)
- CPU, memory, disk, network, latency, services, and container info
- Website diagnostics (ping, HTTP, DNS, SSL, WHOIS, traceroute)
- Easy-to-use command-line flags
- No config file required

## Installation

Clone the repo and build with Go:

```sh
git clone https://github.com/timmyb824/sysinformer.git
cd sysinformer
go build -o sysinformer
```

Install with curl and one of the pre-built binaries e.g.,:

```sh
sudo sh -c 'curl -fSsL https://github.com/timmyb824/sysinformer/releases/download/v1.0.5/sysinformer-linux-amd64 -o /usr/local/bin/sysinformer && chmod +x /usr/local/bin/sysinformer'
```

Install with Go:

_Please note this is supported as more of a convenience and methods above are preferred._

```sh
go install github.com/timmyb824/sysinformer@latest
```

## Usage

Run the CLI with the desired flags:

System information:

```sh
sysinformer --system     # Show system info
sysinformer --cpu        # Show CPU info
sysinformer --memory     # Show memory info
sysinformer --disks      # Show disk info
sysinformer --network    # Show network info
sysinformer --latency    # Show latency info
sysinformer --services   # Show services info
sysinformer --containers # Show container info
sysinformer --all        # Show all info
```

Short flags (e.g., `-s`, `-c`, etc.) are also supported.

Website diagnostics:

```sh
sysinformer web example.com --full
sysinformer web https://example.com --http --ssl
sysinformer web example.com --dns
sysinformer web example.com --ping --count 5
sysinformer web example.com --timeout 15 --trace
```

The `web` command supports the following options:

- `--ping`
- `--latency`
- `--dns`
- `--http`
- `--ssl`
- `--whois`
- `--trace`
- `--full`
- `--timeout` (seconds)
- `--count` (ping count)

## License

See [LICENSE](LICENSE) for details.
