# ZBittorrent

ZBittorrent is a lightweight BitTorrent client implemented in Go. It aims to provide a simple and efficient way to download and share files using the BitTorrent protocol.

## Features

- Basic BitTorrent protocol implementation
- DHT (Distributed Hash Table) support
- PEX (Peer Exchange) support
- Configurable upload/download rate limits
- JSON/Text logging formats
- YAML-based configuration

## Requirements

- Go 1.24 or higher

## Installation

```bash
# Clone the repository
git clone https://github.com/al002/zbittorrent.git

# Change directory
cd zbittorrent

# Build the project
make build
```

## Configuration

Create a configuration file at `$HOME/.zbittorrent/config.yaml` or specify a custom path using the `--config` flag.

Example configuration:

```yaml
listen_port: 6881
enable_dht: true
enable_pex: true
max_uploads: 4
max_peers: 50
min_peers: 10
download_dir: "/path/to/downloads"
upload_rate_limit: 1024    # KB/s
download_rate_limit: 2048  # KB/s

log:
  dir: "/path/to/logs"
  level: "info"           # debug, info, warn, error
  format: "json"          # json or text
```

## Usage

```bash
# Run with default config location ($HOME/.zbittorrent/config.yaml)
./bin/zbittorrent

# Run with custom config file
./bin/zbittorrent --config /path/to/config.yaml
```

## Project Structure

```
.
├── cmd/                    # Command line interface
├── internal/              
│   ├── config/            # Configuration management
│   └── logger/            # Logging functionality
├── pkg/
│   └── bencode/          # Bencode encoding/decoding
├── go.mod                 # Go modules file
├── go.sum                 # Go modules checksums
└── main.go               # Application entry point
```

## Development Status

This project is currently under development. The following features are planned or in progress:

- [ ] Complete Bencode implementation
- [ ] Torrent file parsing
- [ ] Peer wire protocol
- [ ] DHT implementation
- [ ] PEX implementation
- [ ] Rate limiting
- [ ] Web UI
