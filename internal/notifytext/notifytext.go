package notifytext

import (
	"fmt"
	"internet-monitor/internal/i18n"
	"internet-monitor/internal/types"
	"math"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

var (
	lang      atomic.Value
	listenMu  sync.Mutex
	listeners []func(string)
)

func SetLang(l string) {
	next := i18n.Normalize(l)
	if prev, ok := lang.Swap(next).(string); ok && prev == next {
		return
	}
	listenMu.Lock()
	fns := append([]func(string){}, listeners...)
	listenMu.Unlock()
	for _, fn := range fns {
		fn(next)
	}
}

func OnLangChange(fn func(string)) {
	listenMu.Lock()
	listeners = append(listeners, fn)
	listenMu.Unlock()
}

func StatusLabel(lang string, status types.Status) string {
	switch status {
	case types.StatusConnected:
		return i18n.T(lang, "tray.status.connected", nil)
	case types.StatusDegraded:
		return i18n.T(lang, "tray.status.degraded", nil)
	case types.StatusDisconnected:
		return i18n.T(lang, "tray.status.disconnected", nil)
	default:
		return i18n.T(lang, "tray.status.checking", nil)
	}
}

func StatusDetail(lang string, status types.Status, r types.CheckResult) string {
	label := StatusLabel(lang, status)
	if status == types.StatusDisconnected || (r.PacketLoss == 0 && r.LatencyMs == 0) {
		return label
	}
	return label + " · " + i18n.T(lang, "tray.detail", i18n.Vars{"loss": fmt.Sprintf("%.0f", r.PacketLoss), "ms": Latency(lang, r.LatencyMs)})
}

func Lang() string {
	if v, ok := lang.Load().(string); ok && v != "" {
		return v
	}
	return i18n.Default
}

func TestMessage(lang string) (title, body string) {
	return i18n.T(lang, "notify.test.title", nil), i18n.T(lang, "notify.test.body", nil)
}

func StatusFromEventType(eventType string) types.Status {
	switch eventType {
	case "connected":
		return types.StatusConnected
	case "degraded":
		return types.StatusDegraded
	default:
		return types.StatusDisconnected
	}
}

func Build(lang string, status types.Status, r types.CheckResult) (title, body string) {
	switch status {
	case types.StatusConnected:
		title = i18n.T(lang, "notify.restored.title", nil)
		if r.LatencyMs > 0 {
			return title, i18n.T(lang, "notify.restored.latency", i18n.Vars{"ms": Latency(lang, r.LatencyMs)})
		}
		return title, i18n.T(lang, "notify.restored.ok", nil)

	case types.StatusDegraded:
		return i18n.T(lang, "notify.degraded.title", nil),
			i18n.T(lang, "notify.degraded.body", i18n.Vars{"loss": fmt.Sprintf("%.0f", r.PacketLoss), "ms": Latency(lang, r.LatencyMs)})

	default:
		var parts []string
		if !r.TCPPingOK {
			parts = append(parts, "TCP")
		}
		if !r.HTTPOK {
			parts = append(parts, "HTTP")
		}
		if !r.DNSOK {
			parts = append(parts, "DNS")
		}
		title = i18n.T(lang, "notify.down.title", nil)
		if len(parts) > 0 {
			return title, i18n.T(lang, "notify.down.failed", i18n.Vars{"layers": strings.Join(parts, " + ")})
		}
		return title, i18n.T(lang, "notify.down.lost", nil)
	}
}

func Latency(lang string, ms int64) string {
	switch {
	case ms < 0:
		ms = 0
		fallthrough
	case ms < 1000:
		return i18n.T(lang, "duration.millis", i18n.Vars{"n": ms})
	case ms < 60000:
		s := float64(ms) / 1000
		digits := 2
		if s >= 10 {
			digits = 1
		}
		text := strings.TrimRight(strings.TrimRight(strconv.FormatFloat(s, 'f', digits, 64), "0"), ".")
		return i18n.T(lang, "duration.s", i18n.Vars{"s": text})
	default:
		return Duration(lang, float64(ms)/1000)
	}
}

func Duration(lang string, seconds float64) string {
	s := int(math.Round(seconds))
	if s < 0 {
		s = 0
	}
	switch {
	case s < 60:
		return i18n.T(lang, "duration.s", i18n.Vars{"s": s})
	case s < 3600:
		if s%60 == 0 {
			return i18n.T(lang, "duration.m", i18n.Vars{"m": s / 60})
		}
		return i18n.T(lang, "duration.ms", i18n.Vars{"m": s / 60, "s": s % 60})
	default:
		if (s%3600)/60 == 0 {
			return i18n.T(lang, "duration.h", i18n.Vars{"h": s / 3600})
		}
		return i18n.T(lang, "duration.hm", i18n.Vars{"h": s / 3600, "m": (s % 3600) / 60})
	}
}

func EscapeAppleScript(s string) string {
	return strings.NewReplacer("\\", "\\\\", "\"", "\\\"").Replace(s)
}
