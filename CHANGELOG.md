# Changelog

All notable changes to this project are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

GitHub release notes for each tag are auto-generated from commits; this file is
the curated, human-readable summary.

## [Unreleased]

The v2 rebuild.

### Added
- **React dashboard** (React 19 + Vite + TypeScript in `web/`, embedded into the
  binaries via `web/embed.go`) with a live overview: status, plain-language
  diagnosis, latency chart, jitter, packet loss, availability for today / 7 / 30
  days, and per-target status.
- **History screen** with a per-day log, filters, and CSV export for a day, the
  last 7 days, or the last 30 days.
- **Monthly report page** at `/report?month=YYYY-MM` with "Print or save as PDF"
  (replaces `report.html`).
- Speed test screen with gauge, today's runs, scheduled tests and alert threshold.
- Custom alert sound upload (MP3/WAV, up to 30 s / 10 MB); Settings shows whether
  the custom sound is active and its length, with "Try the sound" and "Play like
  a real alert".
- **New default alert sound.**
- Dark and light themes, Arabic/English with full RTL, layout for phone-size windows.
- Unsaved-changes guard in Settings.
- CI `web` job (typecheck, lint, test, build); Go jobs reuse its `web-dist` artifact.

### Changed
- Repository layout: entry points in `cmd/tray` and `cmd/gui`, packages under
  `internal/`; the old vanilla-JS UI in `internal/server/assets` is removed.
- Building from source now requires [bun](https://bun.sh) (CI pins 1.3.14) in
  addition to Go. `make build`, `build-debug`, `build-gui` and `scripts\build*.cmd`
  build the web UI first.
- Settings saving is conflict-safe.
- Config handling moved to a hardened config store; the Go toolchain is
  upgraded to the version declared in `go.mod`.
- Dev mode: `go run ./cmd/tray` plus `bun run dev` in `web/` for the UI. The
  `.air.toml` config is no longer documented.
- CI refuses to publish a release unless every binary is present.

### Fixed
- **Uploaded custom alert sounds silently fell back to the default.** The
  uploaded file is now validated and converted, and is the sound that plays.
- The auto-updater never selects the GUI asset for a tray install.

### Security
- The local dashboard binds `127.0.0.1` and requires a per-launch session token;
  open it from the tray menu or app window.
- Webhook URL and Telegram token are never sent back to the browser (shown as
  "Saved — hidden").
- Webhook URLs must be `https` and on a Discord or Slack host.
- The updater is pinned to GitHub hosts.

## [0.10.0] - 2026-06-01

### Added
- **Manual "Check for updates" button** in the dashboard header — forces a fresh
  GitHub Releases check on demand instead of waiting for the background poll
  (`POST /api/update/check`).

### Changed
- **CI now mints a version tag and release only when every build passes.** The
  version job computes the next semver as a dry run; the tag and GitHub release
  are created by the release job, which is skipped if any platform build fails.
  A failing push no longer burns a version number.

### Fixed
- Workflow file failed to parse (an unquoted `:` in a step name) — every CI run
  failed at startup with no jobs.
- `gofmt`, cross-platform `tray.Logf` / `sound.Logf` declarations, and unused
  non-Windows tray stubs that broke the Linux `vet` / `golangci-lint` gates.

## [0.9.4] - 2026-06-01

The headline release that landed the Phase 0 + Phase 2 roadmap (PR #2).

### Added
- **Monthly outage report (PDF)** for ISP disputes — month picker, charts,
  chronological event log, per-layer cause breakdown. Neutral by default with an
  opt-in branding footer; handles outages that cross a month boundary. (Closes #3)
- **Upload speed test** with live progress, alongside download and ping.
- **speedtest.net-style circular gauge** — animated needle + arc across the
  ping → download → upload phases, with result cards.
- **Scheduled automatic speed tests** (`speed_test.schedule_minutes`).
- **Telegram notifications** (in addition to Discord / Slack).
- **Localized OS notifications** driven by `config.language` — no more mixed AR/EN.
- Per-target concurrent probing, DNS-vs-HTTP-vs-outage failure classification,
  LAN/ISP diagnosis, optional ICMP ping, latency jitter, and a Prometheus
  `/metrics` endpoint.
- Availability widget, copy-summary, and date-range CSV export.
- Self-update verifies the downloaded binary against the release `SHA256SUMS`.

### Changed
- Live config hot-reload from the dashboard (targets / thresholds / interval /
  webhook) without a restart.
- Unified ringtone playback through a single native backend — the previous sound
  is always stopped first, so test / preview / real-alert chimes never overlap.
- Config is sanitized on load and save; `check_interval_sec=0` can no longer
  panic the ticker.
- README leads with the ISP-dispute use case (EN + AR/RTL).

### Fixed
- Graceful HTTP shutdown; dashboard `ListenAndServe` errors are logged, not
  swallowed.
- CSRF guard, panic recovery, and request-body size limits on the local API.

## [0.8.1] - 2026-05-12

Last release before the 0.9.x → 0.10.0 feature wave. Earlier tags
(0.0.1 – 0.8.1) predate this changelog; see the GitHub Releases page for their
auto-generated notes.

[0.10.0]: https://github.com/FutureSolutionDev/internet-monitor/releases/tag/v0.10.0
[0.9.4]: https://github.com/FutureSolutionDev/internet-monitor/releases/tag/v0.9.4
[0.8.1]: https://github.com/FutureSolutionDev/internet-monitor/releases/tag/v0.8.1
