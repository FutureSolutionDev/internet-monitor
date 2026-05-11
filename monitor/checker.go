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
	history    []bool // TCP ping results for packet loss calculation
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
				MaxIdleConns:        0,
				TLSHandshakeTimeout: 3 * time.Second,
			},
		},
	}
}

func (c *Checker) Check() CheckResult {
	result := CheckResult{Timestamp: time.Now()}

	// TCP ping: try each target, record success + latency of first success
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

	// Packet loss: track last N TCP ping outcomes
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

	// HTTP check
	resp, err := c.httpClient.Get(c.cfg.HTTPTarget)
	if err == nil {
		resp.Body.Close()
		result.HTTPOK = resp.StatusCode == 204 || resp.StatusCode == 200
	}

	// DNS check
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, dnsErr := net.DefaultResolver.LookupHost(ctx, c.cfg.DNSTarget)
	result.DNSOK = dnsErr == nil

	return result
}
