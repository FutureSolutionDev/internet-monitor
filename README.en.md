# Internet Monitor

A free, open-source Windows tool that monitors your internet connection stability.
Runs silently in the background, logs every disconnection with its cause and duration,
and provides a real-time visual dashboard in your browser.

---

## What It Does

- Checks internet connectivity every 5 seconds (configurable)
- Detects 3 connection states:
  - **Disconnected** — no internet at all
  - **Degraded** — packet loss or high latency
  - **Connected** — everything working normally
- Logs every event to a daily JSONL file
- Shows a Windows notification on disconnect/reconnect
- Sends alerts to Discord / Slack / any webhook service
- Provides a **live dashboard** in your browser with real-time charts

---

## Installation & Usage

### Quick Start (Non-technical users)

1. Download `internet-monitor.exe`
2. Place it in a permanent folder (e.g. `C:\Tools\InternetMonitor\`)
3. Copy `config.json` to the same folder
4. Double-click `internet-monitor.exe`
5. An icon appears in your System Tray
6. Right-click the icon → **Open Dashboard** to view the dashboard

### Auto-start with Windows

```
scripts\install.cmd
```

Registers the app to start automatically when Windows boots.

---

## Tray Icon

| Icon | Meaning |
|------|---------|
| 🟢 Green | Connected and healthy |
| 🟡 Yellow | Connection degraded (slow or packet loss) |
| 🔴 Red | Internet disconnected |

---

## Dashboard

Open your browser to: **http://localhost:8765**

Or right-click the tray icon → **Open Dashboard**

### Dashboard Tabs:

**Dashboard tab**
- Large status card with current latency
- Connection quality grade (A / B / C / D / F)
- 5 stat cards: Uptime, Connection %, Drops, Avg Latency, Total Checks
- Live latency chart (last 60 checks) with hover tooltips
- TCP / HTTP / DNS check badges + packet loss %
- Recent events table

**Logs tab**
- Select any date to view full logs for that day
- Each row: time, event type, outage duration, reason, packet loss, latency
- Export to CSV button

**Settings tab**
- Edit all settings directly from the browser
- Instant save without restart (except port changes)

---

## Configuration (config.json)

```json
{
  "check_interval_sec": 5,
  "ping_targets": ["8.8.8.8:53", "1.1.1.1:53"],
  "http_target": "https://connectivitycheck.gstatic.com/generate_204",
  "dns_target": "www.google.com",
  "fail_threshold": 3,
  "packet_loss_threshold": 20.0,
  "latency_threshold_ms": 500,
  "log_dir": "logs",
  "webhook_url": "",
  "dashboard_port": 8765
}
```

| Field | Description | Default |
|-------|-------------|---------|
| `check_interval_sec` | How often to check (seconds) | 5 |
| `fail_threshold` | Consecutive failures before "disconnected" | 3 |
| `packet_loss_threshold` | Packet loss % threshold for "degraded" | 20% |
| `latency_threshold_ms` | Latency ms threshold for "degraded" | 500ms |
| `webhook_url` | Discord/Slack webhook URL for alerts | empty |
| `dashboard_port` | Dashboard HTTP port | 8765 |

---

## Webhook Payload

Add your webhook URL to `webhook_url` in settings.
Sent on every disconnect/reconnect/degraded event:

```json
{
  "timestamp": "2026-05-11T14:30:00Z",
  "event": "disconnected",
  "duration_seconds": 45.2,
  "reason": {
    "tcp_ping_failed": true,
    "http_failed": true,
    "dns_failed": false,
    "packet_loss_pct": 80.0,
    "avg_latency_ms": 0
  }
}
```

---

## Log Files

Stored in the `logs/` folder, one JSONL file per day:

```
logs/
  connectivity_2026-05-11.jsonl
  connectivity_2026-05-12.jsonl
  ...
```

---

## Scripts (Developers)

```bash
scripts\build.cmd        # Build release exe
scripts\build-debug.cmd  # Build with visible console (for debugging)
scripts\run.cmd          # Run directly
scripts\stop.cmd         # Stop running instance
scripts\install.cmd      # Install to Windows Startup
scripts\uninstall.cmd    # Remove from Startup
scripts\logs.cmd         # Open logs folder
```

Or with `make`:
```bash
make build
make install
make stop
make logs
```

---

## Build from Source

Requires: [Go 1.21+](https://go.dev/dl/)

```bash
git clone <repo>
cd internet-monitor
go mod tidy
go build -ldflags="-H=windowsgui -s -w" -o internet-monitor.exe .
```

---

## Connection Quality Grades

| Grade | Meaning |
|-------|---------|
| **A — Excellent** | >95% uptime, latency <200ms |
| **B — Good** | >95% uptime but occasionally slow |
| **C — Fair** | 80-95% uptime |
| **D — Poor** | 50-80% uptime |
| **F — Critical** | <50% uptime |

---

## License

MIT — Free for personal and commercial use.
Platform: Windows 10/11
