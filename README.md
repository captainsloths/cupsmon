# CUPS Monitor

A simple Go application that monitors CUPS print server health and alerts PagerDuty when issues occur.

## Features

- Monitors CUPS server via HTTP health checks
- Sends PagerDuty alerts on state changes (down/recovered)
- Automatic incident resolution when service recovers
- Different severity levels (critical for service down, error for queue issues)
- Minimal dependencies (Go standard library only)
- Configuration via `.env` file

## Quick Start

### Prerequisites

- Go 1.16 or later (for building)
- A PagerDuty account with Events API v2 integration
- Access to a CUPS server

### Installation

1. Clone or download this repository
2. Create a `.env` file in the project directory:
   ```bash
   CUPS_URL=http://localhost:631
   PAGERDUTY_ROUTING_KEY=your-routing-key-here
   ```

### Build

#### Local Build

```bash
go build -o cupsmon main.go
```

#### Cross-Compilation

**Linux x86_64:**
```bash
GOOS=linux GOARCH=amd64 go build -o cupsmon-linux-x86 main.go
```

**macOS x86_64:**
```bash
GOOS=darwin GOARCH=amd64 go build -o cupsmon-macos-x86 main.go
```

**Windows x86_64:**
```bash
GOOS=windows GOARCH=amd64 go build -o cupsmon-windows-x86.exe main.go
```

#### Build Options

You can also specify additional build flags:

```bash
# Strip symbols and reduce binary size
go build -ldflags="-s -w" -o cupsmon main.go

# Build with optimizations
go build -ldflags="-s -w" -trimpath -o cupsmon main.go
```

### Run

```bash
./cupsmon
```

The application will:
- Load configuration from `.env` file
- Check CUPS health every 30 seconds
- Send PagerDuty alerts only on state changes (healthy ↔ unhealthy)
- Automatically resolve incidents when CUPS recovers

## Configuration

Configuration is done via a `.env` file in the same directory as the binary:

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `CUPS_URL` | No | `http://localhost:631` | URL of the CUPS server to monitor |
| `PAGERDUTY_ROUTING_KEY` | Yes | - | PagerDuty Events API v2 routing key |

**Example `.env` file:**
```
CUPS_URL=http://your-cups-server:631
PAGERDUTY_ROUTING_KEY=R0ab1234567890cdef
```

### Getting a PagerDuty Routing Key

1. Log in to PagerDuty
2. Go to Services → Service Directory
3. Select or create a service
4. Go to Integrations tab
5. Add a new integration with "Events API v2"
6. Copy the Integration Key (this is your routing key)

## Health Check Logic

The monitor performs HTTP GET requests to the CUPS server:

**CUPS is considered healthy if:**
- HTTP GET request succeeds
- Response status code is 2xx or 3xx

**CUPS is considered unhealthy if:**
- Connection fails → **Critical severity** ("CUPS service is down")
- Response status code is 4xx or 5xx → **Error severity** ("CUPS queue not accepting jobs")

## Alert Behavior

- **Trigger**: Sent when CUPS transitions from healthy to unhealthy
  - `critical` severity for connection failures (service down)
  - `error` severity for HTTP 4xx/5xx responses (queue issues)
- **Resolve**: Sent when CUPS transitions from unhealthy to healthy
  - `info` severity with message "CUPS recovered"
- **State tracking**: Only sends alerts on state changes, preventing duplicate alerts
- **Automatic resolution**: PagerDuty incidents are automatically resolved when service recovers

## Logs

All output goes to stdout with timestamps:

```
2025/10/29 18:00:00 CUPS Monitor starting...
2025/10/29 18:00:00 Monitoring http://localhost:631
2025/10/29 18:00:15 CUPS down: CUPS service is down
2025/10/29 18:00:15 Alert sent
2025/10/29 18:01:45 CUPS recovered
2025/10/29 18:01:45 Alert sent
```

## Systemd Service (Optional)

To run as a systemd service:

1. Build the binary:
   ```bash
   GOOS=linux GOARCH=amd64 go build -o cupsmon main.go
   ```

2. Copy binary and config to system location:
   ```bash
   sudo cp cupsmon /usr/local/bin/
   sudo mkdir -p /usr/local/etc/cupsmon
   sudo cp .env /usr/local/etc/cupsmon/.env
   ```

3. Create service file `/etc/systemd/system/cupsmon.service`:
   ```ini
   [Unit]
   Description=CUPS Health Monitor
   After=network.target

   [Service]
   Type=simple
   User=nobody
   Group=nogroup
   WorkingDirectory=/usr/local/etc/cupsmon
   ExecStart=/usr/local/bin/cupsmon
   Restart=always
   RestartSec=10

   [Install]
   WantedBy=multi-user.target
   ```

4. Enable and start:
   ```bash
   sudo systemctl daemon-reload
   sudo systemctl enable cupsmon
   sudo systemctl start cupsmon
   sudo systemctl status cupsmon
   ```

5. View logs:
   ```bash
   sudo journalctl -u cupsmon -f
   ```

## Development

The entire application is in `main.go` for simplicity. Key functions:

- `loadEnv()`: Parses `.env` file and returns configuration map
- `checkCUPS()`: Tests CUPS connectivity and returns health status
- `checkAndAlert()`: Handles state tracking and alert logic
- `sendAlert()`: Sends PagerDuty Events API v2 alerts (trigger/resolve)

## Troubleshooting

### "PAGERDUTY_ROUTING_KEY required in .env"
Make sure your `.env` file exists in the same directory as the binary and contains a valid routing key.

### Health checks always failing
- Verify CUPS URL is correct and accessible
- Check firewall rules
- Test manually: `curl http://your-cups-server:631`

### Alerts not appearing in PagerDuty
- Verify routing key is correct
- Check logs for API errors (will show full PagerDuty error response)
- Ensure the PagerDuty service integration is enabled

### Binary not found after deployment
Make sure the `.env` file is in the same directory as the binary when running.

## License

MIT License - see LICENSE file for details.
