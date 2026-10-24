package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"internet-monitor/internal/config"
	"internet-monitor/internal/i18n"
	"internet-monitor/internal/logger"
	"internet-monitor/internal/startup"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

type rateLimiter struct {
	mu    sync.Mutex
	every time.Duration
	last  map[string]time.Time
}

func newRateLimiter(every time.Duration) *rateLimiter {
	return &rateLimiter{every: every, last: make(map[string]time.Time)}
}

func (r *rateLimiter) allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if t, ok := r.last[key]; ok && time.Since(t) < r.every {
		return false
	}
	r.last[key] = time.Now()
	return true
}

func (r *rateLimiter) forget(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.last, key)
}

type testTargetResult struct {
	Target    string `json:"target"`
	OK        bool   `json:"ok"`
	LatencyMs int64  `json:"latency_ms"`
	Error     string `json:"error,omitempty"`
}

type testTargetsResponse struct {
	PingTargets []testTargetResult `json:"ping_targets"`
	HTTPTargets []testTargetResult `json:"http_targets"`
	DNSTargets  []testTargetResult `json:"dns_targets"`
	HTTPTarget  testTargetResult   `json:"http_target"`
	DNSTarget   testTargetResult   `json:"dns_target"`
}

func cleanList(in []string) []string {
	c := config.Config{PingTargets: in}
	c.Sanitize()
	return c.PingTargets
}

func runProbes(targets []string, fn func(string) testTargetResult) []testTargetResult {
	out := make([]testTargetResult, len(targets))
	var wg sync.WaitGroup
	for i, t := range targets {
		wg.Add(1)
		go func() {
			defer wg.Done()
			out[i] = fn(t)
		}()
	}
	wg.Wait()
	return out
}

func (s *Server) serveTestTargets(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PingTargets []string `json:"ping_targets"`
		HTTPTargets []string `json:"http_targets"`
		DNSTargets  []string `json:"dns_targets"`
		HTTPTarget  string   `json:"http_target"`
		DNSTarget   string   `json:"dns_target"`
		Notify      *bool    `json:"notify"`
	}
	if !decodeJSON(w, r, 16<<10, &req) {
		return
	}
	if req.HTTPTarget != "" && len(req.HTTPTargets) == 0 {
		req.HTTPTargets = []string{req.HTTPTarget}
	}
	if req.DNSTarget != "" && len(req.DNSTargets) == 0 {
		req.DNSTargets = []string{req.DNSTarget}
	}

	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()
	client := &http.Client{
		Timeout:       5 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		Transport:     &http.Transport{DisableKeepAlives: true},
	}

	var resp testTargetsResponse
	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		resp.PingTargets = runProbes(cleanList(req.PingTargets), func(t string) testTargetResult {
			dctx, c := context.WithTimeout(ctx, 3*time.Second)
			defer c()
			start := time.Now()
			conn, err := (&net.Dialer{}).DialContext(dctx, "tcp", t)
			if err != nil {
				return testTargetResult{Target: t, Error: simplifyError(err.Error())}
			}
			conn.Close()
			return testTargetResult{Target: t, OK: true, LatencyMs: time.Since(start).Milliseconds()}
		})
	}()
	go func() {
		defer wg.Done()
		resp.HTTPTargets = runProbes(cleanList(req.HTTPTargets), func(t string) testTargetResult {
			u, err := url.Parse(t)
			if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
				return testTargetResult{Target: t, Error: "invalid_url"}
			}
			hreq, _ := http.NewRequestWithContext(ctx, http.MethodGet, t, nil)
			start := time.Now()
			res, err := client.Do(hreq)
			if err != nil {
				return testTargetResult{Target: t, Error: simplifyError(err.Error())}
			}
			res.Body.Close()
			if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusNoContent {
				return testTargetResult{Target: t, Error: fmt.Sprintf("http_%d", res.StatusCode)}
			}
			return testTargetResult{Target: t, OK: true, LatencyMs: time.Since(start).Milliseconds()}
		})
	}()
	go func() {
		defer wg.Done()
		resp.DNSTargets = runProbes(cleanList(req.DNSTargets), func(t string) testTargetResult {
			dctx, c := context.WithTimeout(ctx, 3*time.Second)
			defer c()
			start := time.Now()
			if _, err := net.DefaultResolver.LookupHost(dctx, t); err != nil {
				return testTargetResult{Target: t, Error: simplifyError(err.Error())}
			}
			return testTargetResult{Target: t, OK: true, LatencyMs: time.Since(start).Milliseconds()}
		})
	}()
	wg.Wait()
	if len(resp.HTTPTargets) > 0 {
		resp.HTTPTarget = resp.HTTPTargets[0]
	}
	if len(resp.DNSTargets) > 0 {
		resp.DNSTarget = resp.DNSTargets[0]
	}
	writeJSON(w, http.StatusOK, resp)

	if (req.Notify == nil || *req.Notify) && s.tester.allow("test-targets-webhook") {
		go s.sendTestWebhook(resp)
	}
}

func (s *Server) sendTestWebhook(results testTargetsResponse) {
	cfg := s.config()
	if cfg.WebhookURL == "" || !logger.IsSupportedWebhook(cfg.WebhookURL) {
		return
	}
	conv := func(in []testTargetResult) []logger.TestResult {
		out := make([]logger.TestResult, 0, len(in))
		for _, r := range in {
			out = append(out, logger.TestResult{Target: r.Target, OK: r.OK, LatencyMs: r.LatencyMs, Error: r.Error})
		}
		return out
	}
	payload := logger.BuildTestPayload(logger.TestResults{
		PingTargets: conv(results.PingTargets),
		HTTPTargets: conv(results.HTTPTargets),
		DNSTargets:  conv(results.DNSTargets),
	}, cfg.WebhookURL)
	body, err := json.Marshal(payload)
	if err != nil {
		return
	}
	client := logger.NoRedirectClient(10 * time.Second)
	resp, err := client.Post(cfg.WebhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		s.logf("WEBHOOK test-targets error: %s", logger.RedactSecret(err.Error(), cfg.WebhookURL))
		return
	}
	resp.Body.Close()
}

func (s *Server) serveTestNotification(w http.ResponseWriter, r *http.Request) {
	if !s.toasts.allow("test") {
		writeError(w, http.StatusTooManyRequests, "rate_limited", "wait a moment before testing again")
		return
	}
	lang := i18n.Normalize(r.URL.Query().Get("lang"))
	volume, set, ok := volumeParam(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid", "volume must be between 0 and 100")
		return
	}
	var override *int
	if set {
		override = &volume
	}
	s.logf("NOTIFICATION test triggered by user (lang=%s)", lang)
	if s.hooks.TestNotification != nil {
		go s.hooks.TestNotification(lang, override)
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func volumeParam(r *http.Request) (volume int, set, ok bool) {
	v := r.URL.Query().Get("volume")
	if v == "" {
		return 0, false, true
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 0 || n > 100 {
		return 0, false, false
	}
	return n, true, true
}

func (s *Server) servePlaySound(w http.ResponseWriter, r *http.Request) {
	volume, set, ok := volumeParam(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid", "volume must be between 0 and 100")
		return
	}
	if !set {
		volume = 100
		if s.store != nil {
			volume = s.store.Get().Notifications.Sound.Volume
		}
	}
	if s.hooks.PlaySound != nil {
		go s.hooks.PlaySound(volume)
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) serveTestWebhook(w http.ResponseWriter, r *http.Request) {
	var req struct {
		URL string `json:"url"`
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 4096))
	if err != nil {
		writeError(w, http.StatusRequestEntityTooLarge, "too_large", "request body too large")
		return
	}
	if len(bytes.TrimSpace(body)) > 0 && json.Unmarshal(body, &req) != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid JSON body")
		return
	}
	target := strings.TrimSpace(req.URL)
	if target == "" {
		target = s.config().WebhookURL
	}
	if target == "" {
		writeError(w, http.StatusBadRequest, "not_configured", "webhook_url not configured")
		return
	}
	if u, err := url.Parse(target); err != nil || u.Scheme != "https" || !logger.IsSupportedWebhook(target) {
		writeError(w, http.StatusBadRequest, "unsupported", "Only Discord and Slack https webhooks are supported")
		return
	}
	if !s.tester.allow("test-webhook") {
		writeError(w, http.StatusTooManyRequests, "rate_limited", "please wait a minute before testing again")
		return
	}
	if s.hooks.TestWebhook != nil {
		if msg := s.hooks.TestWebhook(target); msg != "" {
			s.tester.forget("test-webhook")
			writeError(w, http.StatusBadGateway, "webhook_failed", msg)
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) serveStartupGet(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]bool{"supported": startup.Supported(), "enabled": startup.IsEnabled()})
}

func (s *Server) serveStartupPost(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if !decodeJSON(w, r, 1024, &req) {
		return
	}
	if err := startup.SetEnabled(req.Enabled); err != nil {
		s.logf("STARTUP toggle failed: %v", err)
		writeError(w, http.StatusInternalServerError, "startup_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true, "enabled": startup.IsEnabled()})
}

func simplifyError(e string) string {
	e = strings.ToLower(e)
	switch {
	case strings.Contains(e, "timeout"), strings.Contains(e, "timed out"), strings.Contains(e, "deadline exceeded"), strings.Contains(e, "did not properly respond"):
		return "timeout"
	case strings.Contains(e, "connection refused"), strings.Contains(e, "actively refused"):
		return "refused"
	case strings.Contains(e, "no such host"), strings.Contains(e, "name or service not known"), strings.Contains(e, "name resolution"):
		return "not_found"
	case strings.Contains(e, "no route"), strings.Contains(e, "network unreachable"), strings.Contains(e, "host unreachable"), strings.Contains(e, "network is down"):
		return "unreachable"
	case strings.Contains(e, "permission denied"), strings.Contains(e, "access is denied"):
		return "no_permission"
	case strings.Contains(e, "connectex"):
		return "refused"
	default:
		return "error"
	}
}
