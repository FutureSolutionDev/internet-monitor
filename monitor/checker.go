// Package monitor performs TCP ping, HTTP, and DNS connectivity checks.
package monitor

import (
	"context"
	"internet-monitor/config"
	"internet-monitor/types"
	"net"
	"net/http"
	"sync"
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
			// Don't follow redirects: a captive portal that 302-redirects the
			// generate_204 probe to a login page must count as "not connected",
			// not as a success.
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
			Transport: &http.Transport{
				DisableKeepAlives:   true,
				DisableCompression:  true,
				TLSHandshakeTimeout: 3 * time.Second,
			},
		},
	}
}

// SetConfig swaps the checker's configuration. Only called from the engine's
// run goroutine (the same goroutine that calls Check), so no lock is needed.
func (c *Checker) SetConfig(cfg *config.Config) {
	c.cfg = cfg
}

// probe runs fn for each target concurrently and returns per-target results in
// input order. Bounding all targets of a layer to a single timeout keeps a
// check from taking targets×timeout when several are down.
func probe(layer string, targets []string, fn func(string) (bool, int64)) []types.TargetResult {
	results := make([]types.TargetResult, len(targets))
	var wg sync.WaitGroup
	for i, target := range targets {
		wg.Add(1)
		go func(i int, target string) {
			defer wg.Done()
			ok, lat := fn(target)
			results[i] = types.TargetResult{Layer: layer, Target: target, OK: ok, LatencyMs: lat}
		}(i, target)
	}
	wg.Wait()
	return results
}

func (c *Checker) Check() CheckResult {
	result := CheckResult{Timestamp: time.Now()}

	// ── TCP Ping ── (latency = fastest successful target)
	tcp := probe("tcp", c.cfg.PingTargets, func(target string) (bool, int64) {
		start := time.Now()
		conn, err := net.DialTimeout("tcp", target, 2*time.Second)
		if err != nil {
			return false, 0
		}
		conn.Close()
		return true, time.Since(start).Milliseconds()
	})
	for _, r := range tcp {
		if r.OK {
			result.TCPPingOK = true
			if result.LatencyMs == 0 || r.LatencyMs < result.LatencyMs {
				result.LatencyMs = r.LatencyMs
			}
		}
	}
	result.Targets = append(result.Targets, tcp...)

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

	// ── HTTP ── (ok = any target returns 200/204)
	httpRes := probe("http", c.cfg.HTTPTargets, func(target string) (bool, int64) {
		start := time.Now()
		resp, err := c.httpClient.Get(target)
		if err != nil {
			return false, 0
		}
		resp.Body.Close()
		if resp.StatusCode == 200 || resp.StatusCode == 204 {
			return true, time.Since(start).Milliseconds()
		}
		return false, 0
	})
	for _, r := range httpRes {
		if r.OK {
			result.HTTPOK = true
		}
	}
	result.Targets = append(result.Targets, httpRes...)

	// ── DNS ── (ok = any target resolves)
	dnsRes := probe("dns", c.cfg.DNSTargets, func(target string) (bool, int64) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		start := time.Now()
		if _, err := net.DefaultResolver.LookupHost(ctx, target); err != nil {
			return false, 0
		}
		return true, time.Since(start).Milliseconds()
	})
	for _, r := range dnsRes {
		if r.OK {
			result.DNSOK = true
		}
	}
	result.Targets = append(result.Targets, dnsRes...)

	return result
}
