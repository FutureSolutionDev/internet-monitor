package server

import (
	"fmt"
	"net/http"
	"strings"
)

func (s *Server) serveMetrics(w http.ResponseWriter, _ *http.Request) {
	snap := s.snapshot("metrics")
	stLast := s.speed.lastResult()

	b01 := func(ok bool) int {
		if ok {
			return 1
		}
		return 0
	}
	upVal := b01(snap.Status == "connected" || snap.Status == "degraded")

	var b strings.Builder
	metric := func(name, typ, help string, val interface{}) {
		fmt.Fprintf(&b, "# HELP %s %s\n# TYPE %s %s\n%s %v\n", name, help, name, typ, name, val)
	}
	fmt.Fprintf(&b, "# HELP internet_monitor_build_info Build information.\n# TYPE internet_monitor_build_info gauge\ninternet_monitor_build_info{version=%q} 1\n", s.version)
	metric("internet_monitor_up", "gauge", "1 if the internet is reachable (connected or degraded), else 0.", upVal)
	metric("internet_monitor_latency_ms", "gauge", "Most recent connectivity latency in milliseconds.", snap.LatencyMs)
	metric("internet_monitor_jitter_ms", "gauge", "Mean latency jitter over the recent history window.", snap.JitterMs)
	metric("internet_monitor_packet_loss_ratio", "gauge", "Most recent packet loss ratio (0-1).", snap.PacketLoss/100)
	metric("internet_monitor_tcp_ok", "gauge", "1 if the last TCP ping succeeded.", b01(snap.TCPPingOK))
	metric("internet_monitor_http_ok", "gauge", "1 if the last HTTP check succeeded.", b01(snap.HTTPOK))
	metric("internet_monitor_dns_ok", "gauge", "1 if the last DNS lookup succeeded.", b01(snap.DNSOK))
	metric("internet_monitor_checks_total", "counter", "Total checks performed since start.", snap.TotalChecks)
	metric("internet_monitor_up_checks_total", "counter", "Checks where the link was up since start.", snap.UpChecks)
	metric("internet_monitor_disconnections_total", "counter", "Disconnection events since start.", snap.Disconnections)
	metric("internet_monitor_uptime_ratio", "gauge", "Up checks divided by total checks (0-1).", snap.UptimePct/100)
	if stLast != nil {
		metric("internet_monitor_speedtest_download_mbps", "gauge", "Download speed (Mbps) from the last speed test.", stLast.DownloadMbps)
		metric("internet_monitor_speedtest_latency_ms", "gauge", "Latency (ms) from the last speed test.", stLast.LatencyMs)
		if stLast.UploadMbps != nil {
			metric("internet_monitor_speedtest_upload_mbps", "gauge", "Upload speed (Mbps) from the last speed test.", *stLast.UploadMbps)
		}
	}
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	_, _ = w.Write([]byte(b.String()))
}
