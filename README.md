# CUPS Monitor

A simple Go application that monitors CUPS print server health and alerts PagerDuty when issues occur.

## Features

- Monitors CUPS server via HTTP health checks
- Sends PagerDuty alerts on state changes (down/recovered)
- Configurable check intervals
- Graceful shutdown handling
- Minimal dependencies (Go standard library only)

## Quick Start

### Prerequisites

- Go 1.16 or later
- A PagerDuty account with Events API v2 integration
- Access to a CUPS server

### Installation

1. Clone or download this repository
2. Copy the example environment file:
   ```bash
   cp .env.example .env
   ```
3. Edit `.env` and set your configuration:
   ```bash
   CUPS_URL=http://your-cups-server:631
   PAGERDUTY_ROUTING_KEY=your-routing-key-here
   CHECK_INTERVAL=30s
   DEDUP_KEY=cups-monitor
   ```

### Build

```bash
go build -o cups-monitor main.go
```

### Run

```bash
./cups-monitor
```

The application will:
- Load configuration from `.env`
- Perform an initial health check
- Check CUPS health every `CHECK_INTERVAL`
- Send PagerDuty alerts when state changes occur

## Configuration

All configuration is done via environment variables in the `.env` file:

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `CUPS_URL` | No | `http://localhost:631` | URL of the CUPS server to monitor |
| `PAGERDUTY_ROUTING_KEY` | Yes | - | PagerDuty Events API v2 routing key |
| `CHECK_INTERVAL` | No | `30s` | How often to check CUPS health (Go duration format) |
| `DEDUP_KEY` | No | `cups-monitor` | Deduplication key for PagerDuty alerts |

### Getting a PagerDuty Routing Key

1. Log in to PagerDuty
2. Go to Services → Service Directory
3. Select or create a service
4. Go to Integrations tab
5. Add a new integration with "Events API v2"
6. Copy the Integration Key (this is your routing key)

## Health Check Logic

The monitor considers CUPS healthy if:
- HTTP GET request succeeds
- Response status code is 2xx or 3xx

CUPS is considered unhealthy if:
- Connection fails
- Response status code is 4xx or 5xx

## Alert Behavior

- **Trigger**: Sent when CUPS transitions from healthy to unhealthy
- **Resolve**: Sent when CUPS transitions from unhealthy to healthy
- **Deduplication**: Uses `DEDUP_KEY` to prevent duplicate alerts

The monitor only sends alerts on state changes, not on every check.

## Logs

All output goes to stdout with timestamps:

```
2025/10/28 10:00:00 Starting CUPS Monitor...
2025/10/28 10:00:00 Configuration loaded:
2025/10/28 10:00:00   CUPS URL: http://localhost:631
2025/10/28 10:00:00   PagerDuty Key: R0ab****c123
2025/10/28 10:00:00   Check Interval: 30s
2025/10/28 10:00:00   Dedup Key: cups-monitor
2025/10/28 10:00:00 ✓ Health check passed: http://localhost:631
```

## Systemd Service (Optional)

To run as a systemd service:

1. Build the binary:
   ```bash
   go build -o cups-monitor main.go
   ```

2. Copy binary to system location:
   ```bash
   sudo cp cups-monitor /usr/local/bin/
   sudo cp .env /usr/local/etc/cups-monitor.env
   ```

3. Create service file `/etc/systemd/system/cups-monitor.service`:
   ```ini
   [Unit]
   Description=CUPS Health Monitor
   After=network.target

   [Service]
   Type=simple
   User=nobody
   Group=nogroup
   WorkingDirectory=/usr/local/etc
   Environment="ENV_FILE=/usr/local/etc/cups-monitor.env"
   ExecStart=/usr/local/bin/cups-monitor
   Restart=always
   RestartSec=10

   [Install]
   WantedBy=multi-user.target
   ```

4. Enable and start:
   ```bash
   sudo systemctl enable cups-monitor
   sudo systemctl start cups-monitor
   sudo systemctl status cups-monitor
   ```

5. View logs:
   ```bash
   sudo journalctl -u cups-monitor -f
   ```

## Development

The entire application is in `main.go` for simplicity. Key functions:

- `loadEnv()`: Parses `.env` file
- `checkHealth()`: Tests CUPS connectivity
- `sendAlert()`: Sends PagerDuty Events API v2 alerts
- `performHealthCheck()`: Combines health checking with state tracking
- `main()`: Orchestrates the monitoring loop and graceful shutdown

## Troubleshooting

### "PAGERDUTY_ROUTING_KEY is required"
Make sure your `.env` file exists and contains a valid routing key.

### Health checks always failing
- Verify CUPS URL is correct and accessible
- Check firewall rules
- Test manually: `curl http://your-cups-server:631`

### Alerts not appearing in PagerDuty
- Verify routing key is correct
- Check logs for API errors
- Ensure the PagerDuty service integration is enabled

## License

This is a simple utility script. Use and modify as needed.