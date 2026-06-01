# Changelog

All notable changes to this project are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

GitHub release notes for each tag are auto-generated from commits; this file is
the curated, human-readable summary.

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
