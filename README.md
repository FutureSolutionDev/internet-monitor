# 🌐 Internet Monitor

![Logo](internal/assets/favicon.png)

[![Build & Release](https://github.com/FutureSolutionDev/internet-monitor/actions/workflows/build.yml/badge.svg)](https://github.com/FutureSolutionDev/internet-monitor/actions/workflows/build.yml)
[![License: Elastic 2.0](https://img.shields.io/badge/License-Elastic%202.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.27+-00ADD8?logo=go)](https://go.dev)
![Platforms](https://img.shields.io/badge/platform-Windows%20%7C%20macOS%20%7C%20Linux-lightgrey)

> [🇸🇦 اقرأ بالعربية](README.ar.md)

# Document your outages. Win the dispute with your ISP.

**Internet Monitor** is a free, source-available tool that captures **proof** of every internet drop — so you can stop arguing from memory and start attaching evidence to your support ticket.

It logs each outage with **per-layer detail (TCP / HTTP / DNS)** — so you can prove things like *"DNS failed while the link was up"* — with exact **timestamps and durations**, and exports a clean **monthly report (PDF)** or **CSV** you can hand to your provider. 100% local: no account, no cloud, no telemetry.

> It's also an excellent everyday connectivity monitor — live dashboard, latency/jitter trends, speed tests, and Discord/Slack/Telegram alerts.

---

![Dashboard](docs/screenshot-dashboard.png)

![Dashboard Arabic RTL](docs/screenshot-dashboard-ar.png)

![Speed Test Result](docs/screenshot-speed-result.png)

![History](docs/screenshot-logs.png)

![Settings](docs/screenshot-settings.png)

## 💡 Why Internet Monitor?

Most "my internet is down" complaints fail because you have **no proof** — no timestamps, no causes, no durations. Internet Monitor produces that evidence:

- **Per-layer logging (TCP / HTTP / DNS)** — prove *which* layer failed, not just "it was slow". (e.g. "DNS resolution failed while the link itself was reachable.")
- **Exact timestamps + durations** for every disconnection, plus monthly availability %.
- **Exportable evidence** — one-click **monthly outage report (PDF)** and **JSONL / CSV** you can attach to a ticket.
- **LAN-vs-ISP diagnosis** (optional ICMP) — distinguish "my router is down" from "my ISP is down".

## 🧾 How to use this against your ISP

1. **Run it** in the background (system tray or standalone window) — it checks every few seconds, 24/7.
2. **Let it capture the drops** — each outage is logged with its cause (TCP/HTTP/DNS), exact start time and duration.
3. **Generate the monthly report** — open the dashboard → **History**, open the monthly report (`/report?month=YYYY-MM`), pick the month, then click **Print or save as PDF**.
4. **Attach it to your ticket/complaint.** The report is **neutral by default** (reads as your own document); an opt-in "shareable" toggle can add a small footer if you want.

> Tip: enable **scheduled speed tests** and **ICMP/LAN-vs-ISP diagnosis** to back your case with latency, jitter, and throughput history.

## 👥 Who is it for?

| Who | What they get |
| --- | ------------- |
| 🧑 **End users** | Proof for ISP disputes + notifications on every drop + one-click speed test |
| 🛠️ **Support teams** | Full dashboard with exportable JSONL/CSV logs and monthly PDF reports |
| 👨‍💻 **Developers** | Detailed Discord / Slack / Telegram alerts with structured per-check data + a `/metrics` Prometheus endpoint |

---

## ✨ Features

| Feature | Details |
| ------- | ------- |
| 🔍 **Multi-layer checking** | TCP Ping + HTTP + DNS simultaneously — all targets configurable as arrays |
| 📊 **Live overview** | Connection status, a plain-language diagnosis (e.g. "router fine, outage at ISP"), latency chart, jitter, packet loss, availability for today / 7 days / 30 days, and per-target status |
| 🕘 **History** | Per-day event log with filters and CSV export (selected day, last 7 days, last 30 days) |
| ⚡ **Speed Test** | Download **and upload** measurement (Cloudflare endpoint) with a circular gauge, ping / download / upload results, today's runs, configurable parallel connections/timeout, **scheduled automatic tests** and a low-speed alert threshold |
| 🔔 **Instant notifications** | Windows Toast + Discord / Slack / **Telegram**, localized to your chosen language |
| 🎵 **Custom alert sound** | Upload an MP3 or WAV (up to 30 seconds / 10 MB). Settings shows whether the custom sound is active and its length; **Try the sound** and **Play like a real alert** let you test it. A new default sound ships built in |
| 📋 **Structured logs** | Daily JSONL files — connectivity events in `connectivity_DATE.jsonl`, speed tests in `speedtest_DATE.jsonl` |
| 📄 **Monthly report** | `/report?month=YYYY-MM` page with **Print or save as PDF** |
| 🔄 **Auto-update** | Checks GitHub Releases (background + a manual **"Check for updates"** button), one-click update with automatic restart |
| 🌐 **Bilingual UI** | Arabic & English with full RTL support, toggle anytime |
| 🎨 **Themes & small screens** | Dark and light themes; the layout adapts to phone-size windows |
| 🖥️ **Two modes** | System Tray (background) + Standalone native window |
| 🔒 **Single instance** | Prevents running more than one copy at the same time |
| ⚙️ **Safe settings** | Unsaved-changes guard and conflict-safe saving (a stale save never overwrites newer settings) |
| 🛡️ **Local-only dashboard** | Listens on `127.0.0.1` only, protected by a per-launch session token. Webhook URL and Telegram token are never sent back to the browser |

---

## 🚀 Quick Start

> **Windows SmartScreen warning?** Click **"More info" → "Run anyway"**. The binary is unsigned — the source code is fully auditable above.

Download the latest pre-built binary from [**Releases**](https://github.com/FutureSolutionDev/internet-monitor/releases/latest):

| File | OS | Mode |
| ---- | -- | ---- |
| `internet-monitor-windows.exe` | Windows 10/11 | System Tray |
| `internet-monitor-gui-windows.exe` | Windows 10/11 | Standalone window |
| `internet-monitor-macos-arm64` | macOS M1/M2/M3 | System Tray |
| `internet-monitor-macos-intel` | macOS Intel | System Tray |
| `internet-monitor-linux` | Ubuntu/Debian | System Tray |

**Windows — run once, then right-click the tray icon → Open Dashboard:**

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

Open the dashboard from the **tray menu** (or the app window in the GUI version).

> The dashboard only accepts requests that carry a per-launch session token. Opening <http://127.0.0.1:8765> directly in a browser shows an "open from the app" message — that is expected.

---

## 🧑‍💻 Dev Mode

You need [Go](https://go.dev/dl/) and [bun](https://bun.sh) (see [Build from Source](#️-build-from-source)). Build the web UI once so the Go embed has files to pick up:

```bash
git clone https://github.com/FutureSolutionDev/internet-monitor.git
cd internet-monitor
cd web && bun install --frozen-lockfile && bun run build && cd ..
go run ./cmd/tray
```

### Live UI work (Vite dev server)

With the backend running, start the Vite dev server in a second terminal. It proxies `/api`, `/events` and `/notification-sound` to `http://127.0.0.1:8765` (override with `IM_BACKEND`) and adds the session token for you. The token is the `token` value in the `.session` file next to the running executable.

```bash
cd web
IM_SESSION=<token> bun run dev
```

On Windows PowerShell: `$env:IM_SESSION="<token>"; bun run dev`.

> The native window (GUI) version requires CGO — see [Build from Source](#️-build-from-source) below.

---

## 🛠️ Build from Source

### Requirements

| Tool | Version | Note |
| ---- | ------- | ---- |
| [Go](https://go.dev/dl/) | see `go.mod` (currently 1.27.3) | Required |
| [bun](https://bun.sh) | 1.3.14 (the version CI uses) | Required — builds the React dashboard |
| GCC (TDM-GCC recommended) | any | Optional — native window version only (CGO) |

The dashboard in `web/` is embedded into the Go binaries (`web/embed.go`), so **build it first**. Without `web/dist`, `go build` fails.

### Web UI

```bash
cd web
bun install --frozen-lockfile
bun run build
cd ..
```

Other checks used by CI: `bun run typecheck`, `bun run lint`, `bun run test`.

### Tray version — no CGO needed on Windows

```bash
git clone https://github.com/FutureSolutionDev/internet-monitor.git
cd internet-monitor
cd web && bun install --frozen-lockfile && bun run build && cd ..
go build -ldflags="-H=windowsgui -s -w" -o internet-monitor.exe ./cmd/tray
```

### Native window version — requires CGO

Uses WebView2 on Windows and WebKitGTK/WebKit on Linux/macOS.

**Windows** — builds the web UI, then auto-installs TDM-GCC if missing:

```bat
scripts\build-gui.cmd
```

**macOS / Linux** — GCC is pre-installed (Linux also needs `libgtk-3-dev` and `libwebkit2gtk-4.0-dev`):

```bash
go build -ldflags="-s -w" -o internet-monitor-gui ./cmd/gui/
```

### Makefile

`make build`, `make build-debug` and `make build-gui` depend on `make web`, which runs `bun install --frozen-lockfile && bun run build` inside `web/`.

### Available scripts

The `build*.cmd` scripts call `scripts\build-web.cmd` first.

```text
scripts\build.cmd        Build tray exe
scripts\build-gui.cmd    Build native window (auto-installs GCC if missing)
scripts\build-debug.cmd  Build with visible console (debugging)
scripts\run.cmd          Build and run (tray)
scripts\run-gui.cmd      Build and run (native window — requires GCC)
scripts\stop.cmd         Stop running instance
scripts\install.cmd      Install to Windows Startup
scripts\uninstall.cmd    Remove from Startup
scripts\logs.cmd         Open logs folder
```

### CI/CD — GitHub Actions

Every push to `master` (and every pull request) runs the pipeline:

```text
Web UI        →  typecheck, lint, test, build (bun 1.3.14); uploads web/dist as an artifact
Quality       →  gofmt, go vet, go test, golangci-lint (needs Web UI)
Windows Tray  →  cross-compiled on Linux (CGO disabled)
Windows GUI   →  windows-latest runner
macOS         →  macos-latest (arm64 native + intel cross-compile; tray + GUI)
Linux         →  ubuntu-22.04 (WebKitGTK 4.0; tray + GUI)
```

The Go build jobs download the `web-dist` artifact instead of rebuilding the UI.

On push to `master`, a version tag and GitHub release are created **only if every build passes**. The release job checks that all 8 binaries exist and publishes a `SHA256SUMS` file with them.

---

## ⚙️ Configuration

`config.json` is created automatically on first run with sensible defaults.
Edit it directly or use the **Settings tab** in the dashboard.

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
  "telegram_bot_token": "",
  "telegram_chat_id": "",
  "dashboard_port": 8765,
  "use_icmp": false,
  "language": "en",
  "speed_test": {
    "download_targets": ["https://speed.cloudflare.com/__down"],
    "parallel_connections": 4,
    "timeout_seconds": 30,
    "upload_target": "https://speed.cloudflare.com/__up",
    "schedule_minutes": 0,
    "alert_threshold_mbps": 0
  }
}
```

| Field | Description |
| ----- | ----------- |
| `ping_targets` | TCP addresses — tries all in parallel, OK if any succeeds |
| `http_targets` | HTTP URLs — tries in order until one returns 200/204 |
| `dns_targets` | DNS domains — tries in order until one resolves |
| `check_interval_sec` | Seconds between checks (1–3600, default 5) |
| `fail_threshold` | Consecutive failures before declaring disconnected |
| `log_dir` | Log folder, relative to the app folder. Absolute paths and `..` are rejected |
| `webhook_url` | Discord or Slack webhook URL (leave empty to disable). Only `https` URLs on Discord (`discord.com`, `discordapp.com`) or Slack (`hooks.slack.com`) hosts are accepted |
| `telegram_bot_token`, `telegram_chat_id` | Telegram alerts. The dashboard shows a saved token as "Saved — hidden" and never receives it back |
| `dashboard_port` | Local dashboard port (1–65535, default 8765) |
| `use_icmp` | Also ping the router (ICMP) for LAN-vs-ISP diagnosis |
| `language` | `en` or `ar` |
| `speed_test.parallel_connections` | Parallel HTTP connections for download test (1–16, default 4) |
| `speed_test.timeout_seconds` | Download test duration in seconds (5–120, default 30) |
| `speed_test.upload_target` | POST endpoint for the upload test (default Cloudflare `__up`; empty disables upload) |
| `speed_test.schedule_minutes` | Run an automatic speed test every N minutes (0 = disabled) |
| `speed_test.alert_threshold_mbps` | Send webhook if speed drops below this value (0 = disabled) |

---

## 📡 Webhooks — Discord & Slack

Internet Monitor sends webhook notifications for three event types:

| Trigger | Payload |
| ------- | ------- |
| Connectivity event (connected / degraded / disconnected) | Per-layer status + packet loss + duration |
| Test Targets (from Settings) | Per-target latency and pass/fail |
| Speed Test result | Download speed + duration + latency |

### Discord example — disconnection event

```json
{
  "username": "Internet Monitor",
  "embeds": [{
    "title": "❌ Internet Disconnected",
    "color": 16007988,
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

### Discord example — speed test result

```json
{
  "username": "Internet Monitor",
  "embeds": [{
    "title": "🚀 Speed Test Completed",
    "color": 2278718,
    "fields": [
      {"name": "📥 Download", "value": "94.3 Mbps", "inline": true},
      {"name": "⏱️ Duration", "value": "10.0s",     "inline": true},
      {"name": "📡 Latency",  "value": "12ms",       "inline": true}
    ]
  }]
}
```

---

## ⚡ Speed Test

The built-in speed test uses an **adaptive HTTP download + upload** approach, shown live on a **speedtest.net-style circular gauge**:

- Measures ping, then **download**, then **upload** — the gauge animates through each phase
- Starts with small payloads and increases chunk size over the configured duration
- Opens multiple parallel connections (default: 4) to saturate modern high-speed links
- Uses **Cloudflare's speed endpoint** (`speed.cloudflare.com`) by default — no account required
- Optional **scheduled automatic tests** (`speed_test.schedule_minutes`) build a throughput history unattended
- Results are stored in `logs/speedtest_DATE.jsonl` (separate from connectivity logs)
- Today's runs are listed on the Speed test screen; the connectivity history can be exported as CSV from History

> **Privacy:** speed test payloads contain no device identifiers. Traffic goes only to the endpoint you configure in Settings → Speed Test.

---

## 📂 Project Structure

```text
internet-monitor/
├── cmd/
│   ├── tray/                 Entry point — system tray version
│   └── gui/                  Entry point — native window version (CGO + WebView2/WebKit)
├── internal/
│   ├── app/                  Shared bootstrap, single-instance guard
│   ├── assets/               Embedded icon
│   ├── audio/                Custom alert sound validation + conversion
│   ├── config/               Config struct, defaults, store, webhook allow-list
│   ├── core/                 Monitoring engine + Notifier interface
│   ├── jsonl/                Daily JSONL file reading/writing
│   ├── logger/               Logging + Discord/Slack webhook formatting
│   ├── monitor/              Check engine — TCP / HTTP / DNS / ICMP
│   ├── notifytext/           Localized notification text
│   ├── report/               Monthly report data
│   ├── server/               Local HTTP API, SSE stream, session token, static UI
│   ├── sound/                Alert sound playback
│   ├── speedtest/            Download/upload speed test engine
│   ├── startup/              OS startup registration
│   ├── tray/                 Tray icon, menu, OS notifications
│   ├── types/                Shared data types
│   └── updater/              GitHub Releases updater
├── web/                      React 19 + Vite + TypeScript dashboard
│   ├── src/                  App, screens (Overview, History, Speed, Report, Settings), i18n
│   ├── dist/                 Build output (generated)
│   └── embed.go              Embeds dist/ into the Go binaries
├── scripts/                  Windows build / install / run helpers
├── docs/                     Screenshots
└── .github/workflows/        build.yml — CI/CD for all platforms
```

---

## 📋 Log Format

```text
logs/
  connectivity_2026-05-11.jsonl   Connectivity events (one JSON per line)
  speedtest_2026-05-11.jsonl      Speed test results (one JSON per line)
  app.log                          App errors and webhook status
```

**Connectivity event:**

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

**Speed test result:**

```json
{
  "timestamp": "2026-05-11T14:35:00Z",
  "event": "speed_test",
  "download_mbps": 94.3,
  "upload_mbps": 18.7,
  "latency_ms": 12,
  "duration_seconds": 9.8,
  "endpoints": ["https://speed.cloudflare.com/__down"],
  "parallel_connections": 4,
  "triggered_by": "user"
}
```

---

## 🤝 Contributing

All contributions are welcome — bugs, features, translations, documentation.

### 1. Fork and clone

```bash
git clone https://github.com/YOUR_USERNAME/internet-monitor.git
cd internet-monitor
cd web && bun install --frozen-lockfile && cd ..
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
- Attach a screenshot for any visual change
- Confirm `go build ./...` and `go test ./...` pass (build `web/` first)
- For changes in `web/`, confirm `bun run typecheck`, `bun run lint` and `bun run test` pass

### Areas looking for help

- Windows ARM64 build support
- macOS native menu bar integration
- Per-target latency history charts
- Additional language support

### Adding a language

Each language has its own folder in `web/locales/`, with one JSON file per page
(`settings.json`, `faq.json`, `changelog.json`, `server.json` for desktop alerts and
Webhook/Telegram messages, and so on). Keys are nested: `{"faq": {"free": {"a": "…"}}}`
is the key `faq.free.a`. The dashboard, the monthly report and the app itself all
read from these same files, and a key defined twice is rejected.

1. Copy `web/locales/en/` to `web/locales/<code>/`, for example `fr/`.
2. Edit `meta.json`:
   - `name`: the language's own name, such as `Français`.
   - `dir`: `ltr` or `rtl`.
   - `intl`: the locale used for dates and numbers, such as `fr-FR`.
3. Translate every value in the other files. Keep each `{placeholder}` and each
   in-app link target such as `(settings#alerts)` exactly as it is in English.
4. Run `bun run test` in `web/`. It fails if a file or key is missing, a key isn't
   in English, or a placeholder doesn't match.

A new page file needs one extra line: import its English copy in
`web/src/lib/i18n.ts` (`MESSAGE_FILES`) so its keys are type-checked.

No code changes are needed. The new language appears in the language menu and
in Settings automatically.

---

## 🗺️ Roadmap

| Feature | Status |
| ------- | ------ |
| ⬆️ Upload speed measurement | ✅ Done |
| 📬 Telegram webhook support | ✅ Done |
| ⏰ Scheduled speed tests | ✅ Done |
| 📄 PDF report export (for ISP complaints) | ✅ Done |
| 🔄 Manual "Check for updates" button | ✅ Done |
| ⚛️ React dashboard (dark/light themes, phone-size layout) | ✅ Done |
| 🪟 Windows ARM64 build | 🔜 Planned |
| 📈 Per-ISP analytics & weekly reports | 💡 Proposed |
| 🔌 Plugin API for custom check types | 💡 Proposed |

---

## 📄 License

[Elastic License 2.0](LICENSE) — free for personal and commercial use, and the source is public.
You may not offer it to others as a hosted or managed service, and you must keep its license and
copyright notices. Versions before 2.0.0 were released under the MIT license.

Built with ❤️ by [FutureSolutionDev](https://github.com/FutureSolutionDev)
