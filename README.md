# 🌐 Internet Monitor

[![Build & Release](https://github.com/FutureSolutionDev/internet-monitor/actions/workflows/build.yml/badge.svg)](https://github.com/FutureSolutionDev/internet-monitor/actions/workflows/build.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev)
[![Platforms](https://img.shields.io/badge/platform-Windows%20%7C%20macOS%20%7C%20Linux-lightgrey)](#-build-from-source)

> [🇸🇦 اقرأ بالعربية](README.ar.md)

**An open-source tool for real-time internet connectivity monitoring.**

Runs silently in the background, logs every disconnection with its cause and duration, and displays a live visual dashboard in your browser with charts and instant notifications.

## 💡 Concept & Purpose

Users often complain about internet drops without any concrete evidence — no timestamps, no causes, no durations. **Internet Monitor** solves this practically:

- **End users** — Works automatically, sends instant notifications on every drop
- **Support teams** — Full data dashboard with exportable logs
- **Developers** — Detailed Discord/Slack webhook with complete per-check data

## ✨ Features

| Feature | Details |
| ------- | ------- |
| 🔍 Multi-layer checking | TCP Ping + HTTP + DNS simultaneously, all configurable as arrays |
| 📊 Live dashboard | Real-time latency chart + stats + event log with hover tooltips |
| 🔔 Instant notifications | Windows Toast + Discord/Slack Webhook + custom alert sound |
| 📋 Structured logs | Daily JSONL files, exportable as CSV from the dashboard |
| 🔄 Auto-update | Checks GitHub Releases, one-click update with automatic restart |
| 🌐 Bilingual UI | Arabic & English with full RTL support, toggle any time |
| 🖥️ Two modes | System Tray (background) + Standalone native window |
| 🔒 Single instance | Prevents running more than one copy at the same time |

## 🚀 Quick Start

Download a pre-built binary from [Releases](https://github.com/FutureSolutionDev/internet-monitor/releases/latest):

| File | OS | Mode |
| ---- | -- | ---- |
| `internet-monitor-windows.exe` | Windows 10/11 | System Tray |
| `internet-monitor-gui-windows.exe` | Windows 10/11 | Standalone window |
| `internet-monitor-macos-arm64` | macOS M1/M2/M3 | System Tray |
| `internet-monitor-macos-intel` | macOS Intel | System Tray |
| `internet-monitor-linux` | Ubuntu/Debian | System Tray |

**Windows — run once, then right-click tray icon → Open Dashboard:**

```bat
internet-monitor-windows.exe
```

**Windows — install to run automatically at startup:**

```bat
scripts\install.cmd
```

**macOS / Linux:**

```bash
chmod +x internet-monitor-*
./internet-monitor-macos-arm64
```

Open the dashboard at **<http://localhost:8765>**

## 🛠️ Build from Source

### Requirements

| Tool | Version | Note |
| ---- | ------- | ---- |
| [Go](https://go.dev/dl/) | 1.21+ | Required |
| GCC | any | Optional — native window version only |

### Tray version — no CGO needed

```bash
git clone https://github.com/FutureSolutionDev/internet-monitor.git
cd internet-monitor
go mod tidy
go build -ldflags="-H=windowsgui -s -w" -o internet-monitor.exe .
```

### Native window version — requires GCC

**Windows** — install GCC once (auto-installed by the script):

```bat
scripts\build-gui.cmd
```

**macOS / Linux** — GCC is pre-installed:

```bash
go build -ldflags="-s -w" -o internet-monitor-gui ./cmd/gui/
```

### Available scripts

```text
scripts\build.cmd        Build tray exe
scripts\build-gui.cmd    Build native window (auto-installs GCC if missing)
scripts\build-debug.cmd  Build with visible console (debugging)
scripts\run.cmd          Build and run
scripts\stop.cmd         Stop running instance
scripts\install.cmd      Install to Windows Startup
scripts\uninstall.cmd    Remove from Startup
scripts\logs.cmd         Open logs folder
```

### CI/CD — GitHub Actions

Every push to `master` triggers automatic builds for all platforms:

```text
Windows Tray  →  cross-compiled on Linux (CGO disabled)
Windows GUI   →  windows-latest runner
macOS         →  macos-latest (arm64 native + intel cross-compile)
Linux         →  ubuntu-22.04 (WebKitGTK 4.0)
```

## ⚙️ Configuration

`config.json` is created automatically on first run. Edit it directly or use the **Settings tab** in the dashboard.

```json
{
  "check_interval_sec": 5,
  "ping_targets":  ["8.8.8.8:53", "1.1.1.1:53"],
  "http_targets":  ["https://connectivitycheck.gstatic.com/generate_204"],
  "dns_targets":   ["www.google.com", "www.cloudflare.com"],
  "fail_threshold": 3,
  "packet_loss_threshold": 20.0,
  "latency_threshold_ms": 500,
  "log_dir": "logs",
  "webhook_url": "",
  "dashboard_port": 8765
}
```

| Field | Description |
| ----- | ----------- |
| `ping_targets` | TCP Ping addresses — tries all, OK if any succeeds |
| `http_targets` | HTTP URLs — tries in order until one returns 200/204 |
| `dns_targets` | DNS domains — tries in order until one resolves |
| `fail_threshold` | Consecutive failures before declaring disconnected |
| `webhook_url` | Discord or Slack webhook URL (leave empty to disable) |

## 📡 Webhook — Discord & Slack

Supports **Discord** and **Slack** only. Payloads are formatted per platform automatically.

**Discord embed example:**

```json
{
  "username": "Internet Monitor",
  "embeds": [{
    "title": "❌ Internet Disconnected",
    "color": 15681604,
    "fields": [
      {"name": "🔌 TCP Ping",    "value": "❌ Failed", "inline": true},
      {"name": "🌐 HTTP",        "value": "❌ Failed", "inline": true},
      {"name": "🔍 DNS",         "value": "✅ OK",     "inline": true},
      {"name": "📉 Packet Loss", "value": "85.0%",    "inline": true},
      {"name": "⏱️ Duration",    "value": "2m 15s",   "inline": true}
    ],
    "timestamp": "2026-05-11T14:30:00Z"
  }]
}
```

## 📂 Project Structure

```text
internet-monitor/
├── main.go                  Entry point — Tray version
├── singleton_*.go           Single-instance guard (per OS)
├── cmd/gui/                 Native window version
│   ├── main.go
│   ├── tray_windows.go      Systray in background goroutine (done channel prevents zombie)
│   ├── tray_stub.go         No-op for macOS/Linux
│   ├── notify_windows.go    Windows Toast + MCI sound playback
│   └── notify_unix.go       osascript / notify-send
├── config/                  Config struct, loading, auto-migration
├── monitor/                 Check engine — TCP / HTTP / DNS
├── dashboard/               HTTP server, SSE stream, all REST APIs
│   └── assets/              Embedded HTML / CSS / JS (Chart.js local)
├── logger/                  JSONL logging + Discord/Slack webhook formatting
├── tray/                    Icon generation (colored ring + favicon blend)
├── updater/                 GitHub Releases API + minio/selfupdate
├── .github/workflows/       build.yml — CI/CD for all platforms
└── scripts/                 Windows build / install / run helpers
```

## 🤝 Contributing

All contributions are welcome — bugs, features, translations, docs.

### 1. Fork and clone

```bash
git clone https://github.com/YOUR_USERNAME/internet-monitor.git
cd internet-monitor
go mod tidy
```

### 2. Create a branch

```bash
git checkout -b feature/your-feature-name
```

### 3. Commit convention

We use [Conventional Commits](https://www.conventionalcommits.org/) — commit type drives auto-versioning:

| Prefix | Version bump | Example |
| ------ | ------------ | ------- |
| `feat:` | minor (v1.1.0) | `feat: add dark mode toggle` |
| `fix:` | patch (v1.0.1) | `fix: resolve DNS timeout on macOS` |
| `BREAKING CHANGE` | major (v2.0.0) | in commit body |
| `docs:`, `refactor:` | patch | no functional change |

### 4. Open a Pull Request

- Describe the motivation clearly
- Attach a screenshot if there is any visual change
- Confirm `go build ./...` passes with no errors

### Areas looking for help

- Windows ARM64 build support
- macOS native menu bar integration
- Per-target latency history charts
- Mobile-responsive dashboard improvements
- Additional webhook providers (Telegram, Teams)

## 📋 Log Format

```text
logs/
  connectivity_2026-05-11.jsonl   One JSON event per line
  app.log                          App errors and webhook send status
```

Each JSONL event:

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

## 📄 License

MIT — Free for personal and commercial use.
Built with ❤️ by [FutureSolutionDev](https://github.com/FutureSolutionDev)
