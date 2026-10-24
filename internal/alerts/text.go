package alerts

import (
	"fmt"
	"internet-monitor/internal/i18n"
	"internet-monitor/internal/notifytext"
	"internet-monitor/internal/types"
)

func resultOf(r types.EventReason) types.CheckResult {
	return types.CheckResult{
		TCPPingOK:  !r.TCPPingFailed,
		HTTPOK:     !r.HTTPFailed,
		DNSOK:      !r.DNSFailed,
		LatencyMs:  r.AvgLatencyMs,
		PacketLoss: r.PacketLossPct,
	}
}

func EventText(lang string, kind Kind, ev types.Event, duration float64) (title, body string) {
	r := resultOf(ev.Reason)
	switch kind {
	case KindRestored:
		title, body = notifytext.Build(lang, types.StatusConnected, r)
		if duration >= 1 {
			body = i18n.T(lang, "notify.restored.after", i18n.Vars{"duration": notifytext.Duration(lang, duration)}) + " · " + body
		}
		return title, body
	case KindReminder:
		return i18n.T(lang, "notify.reminder.title", nil),
			i18n.T(lang, "notify.reminder.body", i18n.Vars{"duration": notifytext.Duration(lang, duration)})
	case KindDegraded:
		return notifytext.Build(lang, types.StatusDegraded, r)
	default:
		return notifytext.Build(lang, types.StatusDisconnected, r)
	}
}

func SpeedText(lang string, s SpeedInfo) (title, body string) {
	title = i18n.T(lang, "notify.speed.title", nil)
	if s.Below && s.ThresholdMbps > 0 {
		title = i18n.T(lang, "notify.speed.slow_title", i18n.Vars{"threshold": mbps(s.ThresholdMbps)})
	}
	vars := i18n.Vars{"down": mbps(s.DownloadMbps), "ms": notifytext.Latency(lang, s.LatencyMs)}
	if s.UploadMbps != nil {
		vars["up"] = mbps(*s.UploadMbps)
		return title, i18n.T(lang, "notify.speed.body", vars)
	}
	return title, i18n.T(lang, "notify.speed.body_down", vars)
}

func TestText(lang string) (title, body string) { return notifytext.TestMessage(lang) }

func mbps(v float64) string { return fmt.Sprintf("%.1f", v) }
