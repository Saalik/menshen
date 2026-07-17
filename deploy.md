# Deployment Guide

## Prerequisites

- Go 1.22+
- Git
- `git-http-backend` (usually part of git-core or git package)

## Automated Installation (macOS & Linux)

An automated installation script is provided for both macOS and Linux. It builds the binary, sets up the configuration in `/etc/menshen`, and configures the service to run automatically.

```bash
chmod +x install.sh
./install.sh
```

## Manual Installation

### Building

```bash
go build -o menshen ./src
```

### Configuration

Menshen looks for its configuration file in the following order:
1. `config.yaml` in the current working directory.
2. `/etc/menshen/config.yaml`

Create a `config.yaml` with the following contents:

```yaml
port: 8080
ttl: 48h
rate_limits:
  global: 0 # Unlimited
  repo: 0   # Unlimited
log_level: info
```

### Running Manually

```bash
./menshen
```

## Systemd Service Example (Manual Setup)

If you are setting it up manually on Linux without `install.sh`, you can use the following unit file `/etc/systemd/system/menshen.service`:

```ini
[Unit]
Description=Menshen Git Server
After=network.target

[Service]
Type=simple
User=menshen
Group=menshen
WorkingDirectory=/opt/menshen
ExecStart=/usr/local/bin/menshen
Restart=always

[Install]
WantedBy=multi-user.target
```
