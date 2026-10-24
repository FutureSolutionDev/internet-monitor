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

const (
	tcpTimeout  = 2 * time.Second
	httpTimeout = 5 * time.Second
	dnsTimeout  = 3 * time.Second
)

type Checker struct {
	mu         sync.Mutex
	cfg        config.Config
	httpClient *http.Client
	history    []bool
	histSize   int
}

func NewChecker(cfg config.Config) *Checker {
	return &Checker{
		cfg:      cfg.Clone(),
		histSize: 10,
		history:  make([]bool, 0, 10),
		httpClient: &http.Client{
			Timeout: httpTimeout,
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

func (c *Checker) SetConfig(cfg config.Config) {
	c.mu.Lock()
	c.cfg = cfg.Clone()
	c.mu.Unlock()
}

func probe(ctx context.Context, layer string, targets []string, fn func(context.Context, string) (bool, int64)) []types.TargetResult {
	results := make([]types.TargetResult, len(targets))
	var wg sync.WaitGroup
	for i, target := range targets {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ok, lat := fn(ctx, target)
			results[i] = types.TargetResult{Layer: layer, Target: target, OK: ok, LatencyMs: lat}
		}()
	}
	wg.Wait()
	return results
}

func (c *Checker) Check(ctx context.Context) CheckResult {
	c.mu.Lock()
	cfg := c.cfg
	c.mu.Unlock()

	result := CheckResult{Timestamp: time.Now()}
	var tcp, httpRes, dnsRes []types.TargetResult
	var gateway string
	var gatewayOK bool

	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		tcp = probe(ctx, "tcp", cfg.PingTargets, func(ctx context.Context, target string) (bool, int64) {
			return c.pingTCP(ctx, target, cfg.UseICMP)
		})
	}()
	go func() {
		defer wg.Done()
		httpRes = probe(ctx, "http", cfg.HTTPTargets, c.probeHTTP)
	}()
	go func() {
		defer wg.Done()
		dnsRes = probe(ctx, "dns", cfg.DNSTargets, probeDNS)
	}()
	if cfg.UseICMP {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if gw, ok := defaultGatewayIPv4(); ok {
				gateway = gw
				_, gatewayOK = icmpPing(gw, time.Second)
			}
		}()
	}
	wg.Wait()

	result.TCPPingOK = len(cfg.PingTargets) == 0
	for _, r := range tcp {
		if r.OK {
			result.TCPPingOK = true
			if result.LatencyMs == 0 || r.LatencyMs < result.LatencyMs {
				result.LatencyMs = r.LatencyMs
			}
		}
	}
	result.HTTPOK = len(cfg.HTTPTargets) == 0 || anyOK(httpRes)
	result.DNSOK = len(cfg.DNSTargets) == 0 || anyOK(dnsRes)
	result.Targets = append(append(append(result.Targets, tcp...), httpRes...), dnsRes...)
	result.Gateway, result.GatewayOK = gateway, gatewayOK
	result.PacketLoss = c.recordTCP(result.TCPPingOK)
	return result
}

func anyOK(rs []types.TargetResult) bool {
	for _, r := range rs {
		if r.OK {
			return true
		}
	}
	return false
}

func (c *Checker) recordTCP(ok bool) float64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.history = append(c.history, ok)
	if len(c.history) > c.histSize {
		c.history = c.history[1:]
	}
	failures := 0
	for _, v := range c.history {
		if !v {
			failures++
		}
	}
	return float64(failures) / float64(len(c.history)) * 100
}

func (c *Checker) pingTCP(ctx context.Context, target string, useICMP bool) (bool, int64) {
	if useICMP {
		host := target
		if h, _, err := net.SplitHostPort(target); err == nil {
			host = h
		}
		if lat, ok := icmpPing(host, tcpTimeout); ok {
			return true, lat
		}
	}
	dctx, cancel := context.WithTimeout(ctx, tcpTimeout)
	defer cancel()
	start := time.Now()
	conn, err := (&net.Dialer{}).DialContext(dctx, "tcp", target)
	if err != nil {
		return false, 0
	}
	conn.Close()
	return true, max(time.Since(start).Milliseconds(), 1)
}

func (c *Checker) probeHTTP(ctx context.Context, target string) (bool, int64) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return false, 0
	}
	start := time.Now()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, 0
	}
	resp.Body.Close()
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
		return true, max(time.Since(start).Milliseconds(), 1)
	}
	return false, 0
}

func probeDNS(ctx context.Context, target string) (bool, int64) {
	dctx, cancel := context.WithTimeout(ctx, dnsTimeout)
	defer cancel()
	start := time.Now()
	if _, err := net.DefaultResolver.LookupHost(dctx, target); err != nil {
		return false, 0
	}
	return true, max(time.Since(start).Milliseconds(), 1)
}
