# sysinformer

A simple, portable system information CLI tool for macOS and Linux, written in Go.

## Features

- System information (OS, kernel, uptime, users, etc.)
- CPU, memory, disk, network, latency, services, and container info
- Easy-to-use command-line flags
- No config file required

## Installation

Clone the repo and build with Go:

```sh
git clone https://github.com/timmyb824/sysinformer.git
cd sysinformer
go build -o sysinfo
```

## Usage

Run the CLI with the desired flags:

```sh
./sysinfo --system     # Show system info
./sysinfo --cpu        # Show CPU info
./sysinfo --memory     # Show memory info
./sysinfo --disks      # Show disk info
./sysinfo --network    # Show network info
./sysinfo --latency    # Show latency info
./sysinfo --services   # Show services info
./sysinfo --containers # Show container info
./sysinfo --all        # Show all info
```

Short flags (e.g., `-s`, `-c`, etc.) are also supported.

## License

See [LICENSE](LICENSE) for details.
