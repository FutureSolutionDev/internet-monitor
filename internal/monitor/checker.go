package monitor

import (
	"context"
	"internet-monitor/internal/config"
	"internet-monitor/internal/types"
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

func (c *Checker) SetConfig(cfg *config.Config) {
	c.cfg = cfg
}

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

	useICMP := c.cfg.UseICMP
	tcp := probe("tcp", c.cfg.PingTargets, func(target string) (bool, int64) {
		host := target
		if h, _, err := net.SplitHostPort(target); err == nil {
			host = h
		}
		if useICMP {
			if lat, ok := icmpPing(host, 2*time.Second); ok {
				return true, lat
			}
		}
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

	if useICMP {
		if gw, ok := defaultGatewayIPv4(); ok {
			result.Gateway = gw
			_, result.GatewayOK = icmpPing(gw, time.Second)
		}
	}

	return result
}
