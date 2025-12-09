# Deployment Guide

## Prerequisites

- Go 1.22+
- Git
- `git-http-backend` (usually part of git-core or git package)

## Building

```bash
go build -o menshen ./src
```

## Configuration

Create a `config.yaml` file in the same directory as the binary:

```yaml
port: 8080
ttl: 48h
rate_limits:
  global: 0 # Unlimited
  repo: 0   # Unlimited
log_level: info
```

## Running

```bash
./menshen
```

## Systemd Service Example

Create `/etc/systemd/system/menshen.service`:

```ini
[Unit]
Description=Menshen Git Server
After=network.target

[Service]
Type=simple
User=menshen
WorkingDirectory=/opt/menshen
ExecStart=/opt/menshen/menshen
Restart=always
Environment=PORT=8080

[Install]
WantedBy=multi-user.target
```
