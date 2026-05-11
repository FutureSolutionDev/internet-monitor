// Package monitor performs TCP ping, HTTP, and DNS connectivity checks.
package monitor

import (
	"context"
	"internet-monitor/config"
	"net"
	"net/http"
	"time"
)

type Checker struct {
	cfg        *config.Config
	httpClient *http.Client
	history    []bool
	histSize   int
}

func NewChecker(cfg *config.Config) *Checker {
	return &Checker{
		cfg:      cfg,
		histSize: 10,
		history:  make([]bool, 0, 10),
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
			Transport: &http.Transport{
				DisableKeepAlives:   true,
				DisableCompression:  true,
				TLSHandshakeTimeout: 3 * time.Second,
			},
		},
	}
}

func (c *Checker) Check() CheckResult {
	result := CheckResult{Timestamp: time.Now()}

	// ── TCP Ping: try each target, succeed on first success ──────
	start := time.Now()
	for _, target := range c.cfg.PingTargets {
		conn, err := net.DialTimeout("tcp", target, 2*time.Second)
		if err == nil {
			conn.Close()
			result.TCPPingOK = true
			result.LatencyMs = time.Since(start).Milliseconds()
			break
		}
	}

	// Packet loss history (based on TCP ping results)
	c.history = append(c.history, result.TCPPingOK)
	if len(c.history) > c.histSize {
		c.history = c.history[1:]
	}
	failures := 0
	for _, ok := range c.history {
		if !ok {
			failures++
		}
	}
	if len(c.history) > 0 {
		result.PacketLoss = float64(failures) / float64(len(c.history)) * 100
	}

	// ── HTTP: try all targets, succeed on first 200/204 ──────────
	for _, target := range c.cfg.HTTPTargets {
		resp, err := c.httpClient.Get(target)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == 200 || resp.StatusCode == 204 {
				result.HTTPOK = true
				break
			}
		}
	}

	// ── DNS: try all targets, succeed on first successful resolve ─
	for _, target := range c.cfg.DNSTargets {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		_, err := net.DefaultResolver.LookupHost(ctx, target)
		cancel()
		if err == nil {
			result.DNSOK = true
			break
		}
	}

	return result
}
