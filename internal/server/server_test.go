package server

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"internet-monitor/internal/config"
	"internet-monitor/internal/logger"
	"internet-monitor/internal/types"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type testEnv struct {
	s    *Server
	dir  string
	base string
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()
	dir := t.TempDir()
	store, err := config.Open(filepath.Join(dir, "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	cfg := store.Get()
	cfg.LogDir = filepath.Join(dir, "logs")
	lgr, err := logger.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	s := New(Options{Store: store, Logger: lgr, Version: "v9.9.9", AppDir: dir})
	if err := s.Start(0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = s.Shutdown(ctx)
		_ = lgr.Close()
	})
	return &testEnv{s: s, dir: dir, base: fmt.Sprintf("http://127.0.0.1:%d", s.Port())}
}

func (e *testEnv) do(t *testing.T, method, path, body string, headers map[string]string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(method, e.base+path, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		if k == "Host" {
			req.Host = v
			continue
		}
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { resp.Body.Close() })
	return resp
}

func decode[T any](t *testing.T, r io.Reader) T {
	t.Helper()
	var v T
	if err := json.NewDecoder(r).Decode(&v); err != nil {
		t.Fatal(err)
	}
	return v
}

func TestGuardRejectsForeignHost(t *testing.T) {
	e := newTestEnv(t)
	if r := e.do(t, "GET", "/api/config", "", map[string]string{"Host": fmt.Sprintf("evil.example:%d", e.s.Port())}); r.StatusCode != http.StatusForbidden {
		t.Errorf("rebinding host: status %d, want 403", r.StatusCode)
	}
	for _, h := range []string{"127.0.0.1", "localhost"} {
		if r := e.do(t, "GET", "/api/config", "", map[string]string{"Host": fmt.Sprintf("%s:%d", h, e.s.Port())}); r.StatusCode != http.StatusOK {
			t.Errorf("host %s: status %d, want 200", h, r.StatusCode)
		}
	}
}

func TestGuardRejectsCrossOriginAndBadContentType(t *testing.T) {
	e := newTestEnv(t)
	if r := e.do(t, "POST", "/api/play-sound", "", map[string]string{"Origin": "https://evil.example"}); r.StatusCode != http.StatusForbidden {
		t.Errorf("cross-origin POST: %d, want 403", r.StatusCode)
	}
	if r := e.do(t, "GET", "/api/config", "", map[string]string{"Origin": "https://evil.example"}); r.StatusCode != http.StatusForbidden {
		t.Errorf("cross-origin GET: %d, want 403", r.StatusCode)
	}
	if r := e.do(t, "GET", "/api/config", "", map[string]string{"Sec-Fetch-Site": "cross-site"}); r.StatusCode != http.StatusForbidden {
		t.Errorf("cross-site fetch: %d, want 403", r.StatusCode)
	}
	if r := e.do(t, "POST", "/api/config", "", map[string]string{"Content-Type": "text/plain"}); r.StatusCode != http.StatusBadRequest {
		t.Errorf("empty text/plain POST: %d, want 400 from handler", r.StatusCode)
	}
	req, _ := http.NewRequest("POST", e.base+"/api/config", strings.NewReader(`{"language":"ar"}`))
	req.Header.Set("Content-Type", "text/plain")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnsupportedMediaType {
		t.Errorf("text/plain body: %d, want 415", resp.StatusCode)
	}
	if r := e.do(t, "GET", "/api/config", "", nil); r.Header.Get("Content-Security-Policy") == "" || r.Header.Get("X-Frame-Options") != "DENY" {
		t.Error("security headers missing")
	}
}

func TestConfigRevisionFlow(t *testing.T) {
	e := newTestEnv(t)
	r := e.do(t, "GET", "/api/config", "", nil)
	if r.Header.Get("X-Config-Revision") != "1" {
		t.Fatalf("revision header = %q", r.Header.Get("X-Config-Revision"))
	}

	r = e.do(t, "POST", "/api/config", `{"check_interval_sec":7}`, map[string]string{"If-Match": `"1"`})
	if r.StatusCode != http.StatusOK {
		t.Fatalf("save: %d", r.StatusCode)
	}
	saved := decode[struct {
		Revision uint64        `json:"revision"`
		Config   config.Config `json:"config"`
	}](t, r.Body)
	if saved.Revision != 2 || saved.Config.CheckIntervalSec != 7 {
		t.Fatalf("saved = %+v", saved)
	}

	r = e.do(t, "POST", "/api/config", `{"check_interval_sec":8}`, map[string]string{"If-Match": `"1"`})
	if r.StatusCode != http.StatusConflict {
		t.Errorf("stale revision: %d, want 409", r.StatusCode)
	}

	r = e.do(t, "POST", "/api/config", `{"log_dir":"C:\\Windows"}`, nil)
	if r.StatusCode != http.StatusBadRequest {
		t.Fatalf("invalid log_dir: %d, want 400", r.StatusCode)
	}
	if body := decode[map[string]interface{}](t, r.Body); body["field"] != "log_dir" {
		t.Errorf("error body = %v", body)
	}

	data, _ := os.ReadFile(filepath.Join(e.dir, "config.json"))
	if !strings.Contains(string(data), `"check_interval_sec": 7`) {
		t.Errorf("config not persisted: %s", data)
	}
}

func TestSSEStreamsInitAndTicks(t *testing.T) {
	e := newTestEnv(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, "GET", e.base+"/events", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	rd := bufio.NewReader(resp.Body)
	next := func() Snapshot {
		for {
			line, err := rd.ReadString('\n')
			if err != nil {
				t.Fatal(err)
			}
			if strings.HasPrefix(line, "data: ") {
				var s Snapshot
				if err := json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &s); err != nil {
					t.Fatal(err)
				}
				return s
			}
		}
	}
	init := next()
	if init.Type != "init" || init.BootID == "" || init.Version != "v9.9.9" || init.ConfigRevision != 1 {
		t.Fatalf("init = %+v", init)
	}
	for e.s.hub.count() == 0 {
		time.Sleep(10 * time.Millisecond)
	}
	e.s.UpdateTick(types.CheckResult{TCPPingOK: true, HTTPOK: true, DNSOK: true, LatencyMs: 40, Timestamp: time.Now()}, types.StatusConnected)
	for {
		s := next()
		if s.Type == "tick" {
			if s.Status != "connected" || s.LatencyMs != 40 || s.AvgLatencyMs != 40 {
				t.Errorf("tick = %+v", s)
			}
			return
		}
	}
}

func TestHubEvictsSlowClient(t *testing.T) {
	h := newHub()
	c, ok := h.subscribe()
	if !ok {
		t.Fatal("subscribe failed")
	}
	for i := 0; i < clientBuffer+1; i++ {
		h.publish([]byte("x"))
	}
	select {
	case <-c.evicted:
	default:
		t.Fatal("slow client was not evicted")
	}
	if h.count() != 0 {
		t.Errorf("count = %d, want 0", h.count())
	}
}

func TestLogsValidateDateAndLimit(t *testing.T) {
	e := newTestEnv(t)
	if r := e.do(t, "GET", "/api/logs?date=../../etc", "", nil); r.StatusCode != http.StatusBadRequest {
		t.Errorf("bad date: %d, want 400", r.StatusCode)
	}
	logDir := e.s.logDir()
	date := time.Now().Format("2006-01-02")
	var lines []string
	for i := 0; i < 10; i++ {
		lines = append(lines, fmt.Sprintf(`{"event":"disconnected","n":%d}`, i))
	}
	if err := os.WriteFile(filepath.Join(logDir, "connectivity_"+date+".jsonl"), []byte(strings.Join(lines, "\n")), 0o600); err != nil {
		t.Fatal(err)
	}
	r := e.do(t, "GET", "/api/logs?limit=3", "", nil)
	got := decode[[]map[string]interface{}](t, r.Body)
	if len(got) != 3 || got[0]["n"].(float64) != 9 {
		t.Errorf("logs = %v", got)
	}
}

func TestUpdateApplyWithoutUpdate(t *testing.T) {
	e := newTestEnv(t)
	if r := e.do(t, "POST", "/api/update", "", nil); r.StatusCode != http.StatusBadRequest {
		t.Errorf("status %d, want 400", r.StatusCode)
	}
}

func TestServeMetrics(t *testing.T) {
	e := newTestEnv(t)
	e.s.UpdateTick(types.CheckResult{TCPPingOK: true, HTTPOK: true, DNSOK: true, LatencyMs: 42, Timestamp: time.Now()}, types.StatusConnected)
	body, _ := io.ReadAll(e.do(t, "GET", "/metrics", "", nil).Body)
	for _, want := range []string{
		`internet_monitor_build_info{version="v9.9.9"} 1`,
		"internet_monitor_up 1",
		"internet_monitor_latency_ms 42",
		"internet_monitor_checks_total 1",
		"internet_monitor_up_checks_total 1",
	} {
		if !strings.Contains(string(body), want) {
			t.Errorf("metrics missing %q\n%s", want, body)
		}
	}
}

func TestJitterOf(t *testing.T) {
	if got := jitterOf([]int64{10}); got != 0 {
		t.Errorf("single sample jitter = %d", got)
	}
	if got := jitterOf([]int64{10, 20, 10, 20}); got != 10 {
		t.Errorf("jitter = %d, want 10", got)
	}
}

func TestShutdownIsIdempotent(t *testing.T) {
	e := newTestEnv(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := e.s.Shutdown(ctx); err != nil {
		t.Fatal(err)
	}
	if err := e.s.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
}
