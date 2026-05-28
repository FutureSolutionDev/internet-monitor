// Package dashboard serves the web-based monitoring UI and REST API over HTTP.
package dashboard

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"internet-monitor/config"
	"internet-monitor/logger"
	"internet-monitor/report"
	"internet-monitor/speedtest"
	"internet-monitor/startup"
	"internet-monitor/types"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

//go:embed assets
var staticFiles embed.FS

// FaviconPNG returns the raw bytes of assets/favicon.png for use as the tray icon.
func FaviconPNG() []byte {
	data, _ := staticFiles.ReadFile("assets/favicon.png")
	return data
}

// UpdateInfo holds update availability data — exported so callers can populate it.
type UpdateInfo struct {
	HasUpdate      bool   `json:"has_update"`
	LatestVersion  string `json:"latest_version"`
	CurrentVersion string `json:"current_version"`
	DownloadURL    string `json:"download_url"`
	ReleaseNotes   string `json:"release_notes"`
}

// internal alias for the stored snapshot (same fields)
type updateSnapshot = UpdateInfo

// SetUpdateInfo is called by the updater goroutine when a new version is found.
func (s *Server) SetUpdateInfo(info *updateSnapshot) {
	s.updateMu.Lock()
	s.updateInfo = info
	s.updateMu.Unlock()
	// Broadcast so the dashboard banner appears without refresh
	s.broadcast("tick")
}

// serveUpdate handles GET (check status) and POST (apply update).
func (s *Server) serveUpdate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		s.updateMu.RLock()
		info := s.updateInfo
		s.updateMu.RUnlock()
		if info == nil {
			w.Write([]byte(`{"has_update":false}`))
			return
		}
		json.NewEncoder(w).Encode(info)

	case http.MethodPost:
		s.updateMu.RLock()
		info := s.updateInfo
		s.updateMu.RUnlock()

		if info == nil || !info.HasUpdate {
			http.Error(w, `{"error":"no update available"}`, http.StatusBadRequest)
			return
		}
		if s.OnApplyUpdate == nil {
			http.Error(w, `{"error":"updater not configured"}`, http.StatusInternalServerError)
			return
		}

		if err := s.OnApplyUpdate(info.DownloadURL); err != nil {
			resp, _ := json.Marshal(map[string]string{"error": err.Error()})
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(resp)
			return
		}

		w.Write([]byte(`{"ok":true,"restart":true}`))

		// Restart after a short delay so the HTTP response is flushed
		if s.OnRestartApp != nil {
			go func() {
				time.Sleep(500 * time.Millisecond)
				s.OnRestartApp()
			}()
		}

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// RingtoneMp3 returns the raw bytes of the embedded Ringtone.mp3.
func RingtoneMp3() []byte {
	data, _ := staticFiles.ReadFile("assets/Ringtone.mp3")
	return data
}

const maxHistory = 60
const maxEvents = 100
const maxTicks = 20

// ── Types ─────────────────────────────────────────────────────

// TickEntry is a single check-cycle result stored for the recent-checks table.
type TickEntry struct {
	Time       string  `json:"time"`
	Status     string  `json:"status"`
	LatencyMs  int64   `json:"latency_ms"`
	PacketLoss float64 `json:"packet_loss_pct"`
	TCPOk      bool    `json:"tcp_ok"`
	HTTPOk     bool    `json:"http_ok"`
	DNSOk      bool    `json:"dns_ok"`
}

// EventEntry is the in-memory / SSE representation of a connectivity event.
// Reason is sent as structured booleans so the client can translate it.
type EventEntry struct {
	Time       string  `json:"time"`
	EventType  string  `json:"event_type"`
	Duration   float64 `json:"duration_seconds"`
	TCPFailed  bool    `json:"tcp_failed"`
	HTTPFailed bool    `json:"http_failed"`
	DNSFailed  bool    `json:"dns_failed"`
	PacketLoss float64 `json:"packet_loss_pct"`
	LatencyMs  int64   `json:"latency_ms"`
}

type Snapshot struct {
	Type             string       `json:"type"`
	Status           string       `json:"status"`
	LatencyMs        int64        `json:"latency_ms"`
	PacketLoss       float64      `json:"packet_loss"`
	TCPPingOK        bool         `json:"tcp_ping_ok"`
	HTTPOK           bool         `json:"http_ok"`
	DNSOK            bool         `json:"dns_ok"`
	Diagnosis        string       `json:"diagnosis"`
	TotalChecks      int          `json:"total_checks"`
	Disconnections   int          `json:"disconnections"`
	UptimeSeconds    float64      `json:"uptime_seconds"`
	UptimePct        float64      `json:"uptime_pct"`
	JitterMs         int64        `json:"jitter_ms"`
	LatencyHistory   []int64      `json:"latency_history"`
	Events           []EventEntry `json:"events"`
	Ticks            []TickEntry  `json:"ticks"`
	UpdateInfo       *UpdateInfo  `json:"update_info,omitempty"`
	SystemNotifs     bool         `json:"system_notifs"`
	SpeedTestRunning bool         `json:"speed_test_running"`
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
	// Legacy single-value fields kept for backward compat with old JS
	HTTPTarget testTargetResult `json:"http_target"`
	DNSTarget  testTargetResult `json:"dns_target"`
}

// ── Server ────────────────────────────────────────────────────

type Server struct {
	port       int
	configPath string
	lgr        *logger.Logger
	clients    map[chan string]struct{}
	mu         sync.Mutex

	srv          *http.Server
	shutdownCh   chan struct{}
	shutdownOnce sync.Once

	// cfgMu guards logDir, which can change at runtime when the config is saved.
	cfgMu  sync.RWMutex
	logDir string

	// nativeNotifs is set by each binary to indicate the host shows OS-native
	// notifications (so the local dashboard suppresses duplicate browser ones).
	nativeNotifs bool

	version string

	OnConfigChange     func(*config.Config)
	OnTestNotification func()
	OnTestWebhook      func(url string) string
	OnApplyUpdate      func(downloadURL string) error
	OnRestartApp       func()

	updateMu   sync.RWMutex
	updateInfo *updateSnapshot

	stateMu        sync.RWMutex
	status         string
	latencyMs      int64
	packetLoss     float64
	tcpPingOK      bool
	httpOK         bool
	dnsOK          bool
	totalChecks    int
	disconnections int
	upTicks        int
	startTime      time.Time
	latencyHistory []int64
	events         []EventEntry
	ticks          []TickEntry

	// Speed test state
	stRunning atomic.Bool
	stCancel  context.CancelFunc
	stLast    *logger.SpeedTestEvent
}

func NewServer(port int, configPath, logDir, version string, lgr *logger.Logger) *Server {
	return &Server{
		port:       port,
		configPath: configPath,
		logDir:     logDir,
		lgr:        lgr,
		version:    version,
		clients:    make(map[chan string]struct{}),
		shutdownCh: make(chan struct{}),
		startTime:  time.Now(),
		status:     "checking",
	}
}

func (s *Server) Start() {
	mux := http.NewServeMux()

	// Static assets (CSS, JS, chart lib, favicon…)
	mux.Handle("/assets/", http.FileServer(http.FS(staticFiles)))

	// Dashboard index
	mux.HandleFunc("/", s.serveIndex)

	// Real-time stream
	mux.HandleFunc("/events", s.serveSSE)

	// JSON APIs
	mux.HandleFunc("/api/state", s.serveState)
	mux.HandleFunc("/api/version", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"version":%q}`, s.version)
	})
	mux.HandleFunc("/api/config", s.serveConfig)
	mux.HandleFunc("/api/logs", s.serveLogs)
	mux.HandleFunc("/api/log-dates", s.serveLogDates)
	mux.HandleFunc("/api/test-targets", s.serveTestTargets)
	mux.HandleFunc("/api/test-notification", s.serveTestNotification)
	mux.HandleFunc("/api/test-webhook", s.serveTestWebhook)
	mux.HandleFunc("/api/update", s.serveUpdate)
	mux.HandleFunc("/api/startup", s.serveStartup)
	mux.HandleFunc("/notification-sound", s.serveNotificationSound)
	mux.HandleFunc("/api/speed-test/start", s.serveSpeedTestStart)
	mux.HandleFunc("/api/speed-test/cancel", s.serveSpeedTestCancel)
	mux.HandleFunc("/api/speed-test/history", s.serveSpeedTestHistory)
	mux.HandleFunc("/api/report", s.serveReport)
	mux.HandleFunc("/report", s.serveReportPage)
	mux.HandleFunc("/metrics", s.serveMetrics)

	s.srv = &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", s.port),
		Handler: s.csrfGuard(mux),
	}
	go func() {
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("dashboard: failed to listen on %s: %v", s.srv.Addr, err)
			if s.lgr != nil {
				s.lgr.AppLog("FATAL dashboard ListenAndServe on %s: %v (port in use?)", s.srv.Addr, err)
			}
		}
	}()
}

// Shutdown gracefully stops the HTTP server: it signals SSE handlers to exit
// (so long-lived streams don't block the drain) and then shuts down the server
// within the given context's deadline. Safe to call once.
func (s *Server) Shutdown(ctx context.Context) error {
	s.shutdownOnce.Do(func() { close(s.shutdownCh) })
	if s.srv != nil {
		return s.srv.Shutdown(ctx)
	}
	return nil
}

func (s *Server) URL() string {
	return fmt.Sprintf("http://localhost:%d", s.port)
}

// SetNativeNotifications declares whether the host emits OS-native notifications.
func (s *Server) SetNativeNotifications(v bool) {
	s.cfgMu.Lock()
	s.nativeNotifs = v
	s.cfgMu.Unlock()
}

func (s *Server) getLogDir() string {
	s.cfgMu.RLock()
	defer s.cfgMu.RUnlock()
	return s.logDir
}

func (s *Server) setLogDir(d string) {
	s.cfgMu.Lock()
	s.logDir = d
	s.cfgMu.Unlock()
}

// csrfGuard rejects state-changing requests that carry a cross-origin Origin
// header. The dashboard binds to loopback, so this blocks a malicious web page
// from driving the local API (config writes, updates, speed tests) via the
// browser. Requests with no Origin (non-browser clients) are allowed.
func (s *Server) csrfGuard(next http.Handler) http.Handler {
	allowed := map[string]bool{
		fmt.Sprintf("http://localhost:%d", s.port): true,
		fmt.Sprintf("http://127.0.0.1:%d", s.port): true,
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
		default:
			if origin := r.Header.Get("Origin"); origin != "" && !allowed[origin] {
				http.Error(w, "cross-origin request denied", http.StatusForbidden)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// ── Public update methods ─────────────────────────────────────

func (s *Server) UpdateTick(result types.CheckResult, status types.Status) {
	s.stateMu.Lock()
	s.status = status.String()
	s.latencyMs = result.LatencyMs
	s.packetLoss = result.PacketLoss
	s.tcpPingOK = result.TCPPingOK
	s.httpOK = result.HTTPOK
	s.dnsOK = result.DNSOK
	s.totalChecks++
	// "Up" = reachable, including degraded (slow but online). Only a full
	// disconnection counts against uptime.
	if status != types.StatusDisconnected {
		s.upTicks++
	}
	s.latencyHistory = append(s.latencyHistory, result.LatencyMs)
	if len(s.latencyHistory) > maxHistory {
		s.latencyHistory = s.latencyHistory[1:]
	}
	tick := TickEntry{
		Time:       result.Timestamp.Format("15:04:05"),
		Status:     status.String(),
		LatencyMs:  result.LatencyMs,
		PacketLoss: result.PacketLoss,
		TCPOk:      result.TCPPingOK,
		HTTPOk:     result.HTTPOK,
		DNSOk:      result.DNSOK,
	}
	s.ticks = append([]TickEntry{tick}, s.ticks...)
	if len(s.ticks) > maxTicks {
		s.ticks = s.ticks[:maxTicks]
	}
	s.stateMu.Unlock()
	s.broadcast("tick")
}

func (s *Server) AddEvent(event types.Event) {
	entry := EventEntry{
		Time:       event.Timestamp.Format("15:04:05"),
		EventType:  event.EventType,
		Duration:   event.DurationSeconds,
		TCPFailed:  event.Reason.TCPPingFailed,
		HTTPFailed: event.Reason.HTTPFailed,
		DNSFailed:  event.Reason.DNSFailed,
		PacketLoss: event.Reason.PacketLossPct,
		LatencyMs:  event.Reason.AvgLatencyMs,
	}

	s.stateMu.Lock()
	s.events = append([]EventEntry{entry}, s.events...)
	if len(s.events) > maxEvents {
		s.events = s.events[:maxEvents]
	}
	if event.EventType == "disconnected" {
		s.disconnections++
	}
	s.stateMu.Unlock()
	s.broadcast("event")
}

// ── HTTP handlers ─────────────────────────────────────────────

func (s *Server) serveIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	data, err := staticFiles.ReadFile("assets/index.html")
	if err != nil {
		http.Error(w, "index.html not found", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(data)
}

func (s *Server) serveState(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.snapshot("init"))
}

func (s *Server) serveConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		data, err := os.ReadFile(s.configPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(data)

	case http.MethodPost:
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		var cfg config.Config
		if err := json.Unmarshal(body, &cfg); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		cfg.Sanitize()
		if strings.Contains(cfg.LogDir, "..") {
			http.Error(w, `{"error":"log_dir must not contain '..'"}`, http.StatusBadRequest)
			return
		}
		pretty, _ := json.MarshalIndent(cfg, "", "  ")
		if err := os.WriteFile(s.configPath, pretty, 0644); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Apply live where possible: log dir is read by this server's handlers,
		// so keep it in sync with what the engine's logger now writes.
		s.setLogDir(cfg.LogDir)
		if s.OnConfigChange != nil {
			s.OnConfigChange(&cfg)
		}
		// The HTTP listener is bound once at startup; a port change needs a restart.
		restartRequired := cfg.DashboardPort != s.port
		if restartRequired && s.lgr != nil {
			s.lgr.AppLog("CONFIG dashboard_port changed to %d — restart required to take effect", cfg.DashboardPort)
		}
		fmt.Fprintf(w, `{"ok":true,"restart_required":%t}`, restartRequired)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) serveLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	date := r.URL.Query().Get("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}
	filename := filepath.Join(s.getLogDir(), fmt.Sprintf("connectivity_%s.jsonl", date))
	data, err := s.readDataFile(filename)
	if os.IsNotExist(err) {
		w.Write([]byte("[]"))
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	lines := bytes.Split(bytes.TrimSpace(data), []byte{'\n'})
	entries := make([]json.RawMessage, 0, len(lines))
	for _, line := range lines {
		line = bytes.TrimSpace(line)
		if len(line) > 0 && json.Valid(line) {
			entries = append(entries, json.RawMessage(line))
		}
	}
	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}
	json.NewEncoder(w).Encode(entries)
}

func (s *Server) serveLogDates(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	entries, err := os.ReadDir(s.getLogDir())
	if err != nil {
		w.Write([]byte("[]"))
		return
	}
	dates := []string{}
	for _, e := range entries {
		name := e.Name()
		if !e.IsDir() && strings.HasPrefix(name, "connectivity_") && strings.HasSuffix(name, ".jsonl") {
			dates = append(dates, name[len("connectivity_"):len(name)-len(".jsonl")])
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(dates)))
	json.NewEncoder(w).Encode(dates)
}

// serveTestTargets tests ping/http/dns targets and returns per-target results.
// Used by the settings UI to validate before saving.
func (s *Server) serveTestTargets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		PingTargets []string `json:"ping_targets"`
		HTTPTargets []string `json:"http_targets"`
		DNSTargets  []string `json:"dns_targets"`
		// Legacy single-value (old JS clients)
		HTTPTarget string `json:"http_target"`
		DNSTarget  string `json:"dns_target"`
	}
	body, _ := io.ReadAll(r.Body)
	json.Unmarshal(body, &req)

	// Merge legacy single values into arrays
	if req.HTTPTarget != "" && len(req.HTTPTargets) == 0 {
		req.HTTPTargets = []string{req.HTTPTarget}
	}
	if req.DNSTarget != "" && len(req.DNSTargets) == 0 {
		req.DNSTargets = []string{req.DNSTarget}
	}

	httpClient := &http.Client{
		Timeout:       5 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		Transport:     &http.Transport{DisableKeepAlives: true},
	}
	resp := testTargetsResponse{PingTargets: make([]testTargetResult, 0)}

	for _, target := range req.PingTargets {
		target = strings.TrimSpace(target)
		if target == "" {
			continue
		}
		start := time.Now()
		conn, err := net.DialTimeout("tcp", target, 3*time.Second)
		lat := time.Since(start).Milliseconds()
		rt := testTargetResult{Target: target}
		if err == nil {
			conn.Close()
			rt.OK = true
			rt.LatencyMs = lat
		} else {
			rt.Error = simplifyError(err.Error())
		}
		resp.PingTargets = append(resp.PingTargets, rt)
	}

	for _, target := range req.HTTPTargets {
		target = strings.TrimSpace(target)
		if target == "" {
			continue
		}
		start := time.Now()
		httpResp, err := httpClient.Get(target)
		lat := time.Since(start).Milliseconds()
		rt := testTargetResult{Target: target}
		if err == nil {
			httpResp.Body.Close()
			// Mirror the live monitor: only 200/204 count as reachable, so the
			// test doesn't give a false green for a 3xx/4xx/5xx endpoint.
			if httpResp.StatusCode == 200 || httpResp.StatusCode == 204 {
				rt.OK = true
				rt.LatencyMs = lat
			} else {
				rt.Error = fmt.Sprintf("http_%d", httpResp.StatusCode)
			}
		} else {
			rt.Error = simplifyError(err.Error())
		}
		resp.HTTPTargets = append(resp.HTTPTargets, rt)
	}
	// Populate legacy field with first result for backward compat
	if len(resp.HTTPTargets) > 0 {
		resp.HTTPTarget = resp.HTTPTargets[0]
	}

	for _, target := range req.DNSTargets {
		target = strings.TrimSpace(target)
		if target == "" {
			continue
		}
		start := time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		_, err := net.DefaultResolver.LookupHost(ctx, target)
		cancel()
		lat := time.Since(start).Milliseconds()
		rt := testTargetResult{Target: target}
		if err == nil {
			rt.OK = true
			rt.LatencyMs = lat
		} else {
			rt.Error = simplifyError(err.Error())
		}
		resp.DNSTargets = append(resp.DNSTargets, rt)
	}
	if len(resp.DNSTargets) > 0 {
		resp.DNSTarget = resp.DNSTargets[0]
	}

	json.NewEncoder(w).Encode(resp)

	// Send results via webhook if configured (non-blocking)
	go s.sendTestWebhook(resp)
}

func (s *Server) sendTestWebhook(results testTargetsResponse) {
	data, err := os.ReadFile(s.configPath)
	if err != nil {
		return
	}
	var cfg config.Config
	if err := json.Unmarshal(data, &cfg); err != nil || cfg.WebhookURL == "" {
		return
	}
	if !logger.IsSupportedWebhook(cfg.WebhookURL) {
		return // Only Discord and Slack
	}

	// Convert to logger.TestResults for unified formatting
	tr := logger.TestResults{}
	for _, r := range results.PingTargets {
		tr.PingTargets = append(tr.PingTargets, logger.TestResult{
			Target:    r.Target,
			OK:        r.OK,
			LatencyMs: r.LatencyMs,
			Error:     r.Error,
		})
	}
	for _, r := range results.HTTPTargets {
		tr.HTTPTargets = append(tr.HTTPTargets, logger.TestResult{Target: r.Target, OK: r.OK, LatencyMs: r.LatencyMs, Error: r.Error})
	}
	for _, r := range results.DNSTargets {
		tr.DNSTargets = append(tr.DNSTargets, logger.TestResult{Target: r.Target, OK: r.OK, LatencyMs: r.LatencyMs, Error: r.Error})
	}

	payload := logger.BuildTestPayload(tr, cfg.WebhookURL)
	body, _ := json.Marshal(payload)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(cfg.WebhookURL, "application/json", bytes.NewReader(body))
	if err == nil {
		resp.Body.Close()
	}
}

func (s *Server) serveTestNotification(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if s.OnTestNotification != nil {
		go s.OnTestNotification()
	}
	w.Write([]byte(`{"ok":true}`))
}

func (s *Server) serveTestWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	// Read webhook URL from request body, or fall back to config file
	var req struct {
		URL string `json:"url"`
	}
	body, _ := io.ReadAll(http.MaxBytesReader(w, r.Body, 4096))
	json.Unmarshal(body, &req)

	url := req.URL
	if url == "" {
		// Fall back to config file
		data, err := os.ReadFile(s.configPath)
		if err == nil {
			var cfg config.Config
			if json.Unmarshal(data, &cfg) == nil {
				url = cfg.WebhookURL
			}
		}
	}

	if url == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"ok":false,"error":"webhook_url not configured"}`))
		return
	}

	if s.OnTestWebhook != nil {
		if errMsg := s.OnTestWebhook(url); errMsg != "" {
			resp, _ := json.Marshal(map[string]interface{}{"ok": false, "error": errMsg})
			w.Write(resp)
			return
		}
	}
	w.Write([]byte(`{"ok":true}`))
}

func (s *Server) serveStartup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		json.NewEncoder(w).Encode(map[string]interface{}{
			"supported": startup.Supported(),
			"enabled":   startup.IsEnabled(),
		})

	case http.MethodPost:
		var req struct {
			Enabled bool `json:"enabled"`
		}
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &req)

		if err := startup.SetEnabled(req.Enabled); err != nil {
			resp, _ := json.Marshal(map[string]interface{}{"ok": false, "error": err.Error()})
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(resp)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":      true,
			"enabled": startup.IsEnabled(),
		})

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) serveSSE(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	ch := make(chan string, 20)
	s.mu.Lock()
	s.clients[ch] = struct{}{}
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.clients, ch)
		s.mu.Unlock()
	}()

	if data, err := json.Marshal(s.snapshot("init")); err == nil {
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()
	}

	for {
		select {
		case msg := <-ch:
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		case <-r.Context().Done():
			return
		case <-s.shutdownCh:
			return
		}
	}
}

// readDataFile reads a log/history file under the logger's write lock when a
// logger is present, so reads don't interleave with concurrent appends.
func (s *Server) readDataFile(path string) ([]byte, error) {
	if s.lgr != nil {
		return s.lgr.ReadFile(path)
	}
	return os.ReadFile(path)
}

// ── Internal ──────────────────────────────────────────────────

func (s *Server) snapshot(msgType string) Snapshot {
	s.updateMu.RLock()
	updateInfo := s.updateInfo
	s.updateMu.RUnlock()

	s.cfgMu.RLock()
	nativeNotifs := s.nativeNotifs
	s.cfgMu.RUnlock()

	s.stateMu.RLock()
	defer s.stateMu.RUnlock()

	hist := make([]int64, len(s.latencyHistory))
	copy(hist, s.latencyHistory)
	evts := make([]EventEntry, len(s.events))
	copy(evts, s.events)
	ticks := make([]TickEntry, len(s.ticks))
	copy(ticks, s.ticks)

	uptimePct := 0.0
	if s.totalChecks > 0 {
		uptimePct = float64(s.upTicks) / float64(s.totalChecks) * 100
	}

	return Snapshot{
		Type:             msgType,
		Status:           s.status,
		LatencyMs:        s.latencyMs,
		PacketLoss:       s.packetLoss,
		TCPPingOK:        s.tcpPingOK,
		HTTPOK:           s.httpOK,
		DNSOK:            s.dnsOK,
		Diagnosis:        types.Diagnose(types.CheckResult{TCPPingOK: s.tcpPingOK, HTTPOK: s.httpOK, DNSOK: s.dnsOK}),
		TotalChecks:      s.totalChecks,
		Disconnections:   s.disconnections,
		UptimeSeconds:    time.Since(s.startTime).Seconds(),
		UptimePct:        uptimePct,
		JitterMs:         jitterOf(hist),
		LatencyHistory:   hist,
		Events:           evts,
		Ticks:            ticks,
		UpdateInfo:       updateInfo,
		SystemNotifs:     nativeNotifs,
		SpeedTestRunning: s.stRunning.Load(),
	}
}

// jitterOf returns the mean absolute difference between consecutive latency
// samples — a simple jitter estimate over the recent history window.
func jitterOf(hist []int64) int64 {
	if len(hist) < 2 {
		return 0
	}
	var sum, n int64
	for i := 1; i < len(hist); i++ {
		d := hist[i] - hist[i-1]
		if d < 0 {
			d = -d
		}
		sum += d
		n++
	}
	return sum / n
}

func (s *Server) broadcast(msgType string) {
	data, err := json.Marshal(s.snapshot(msgType))
	if err != nil {
		return
	}
	msg := string(data)
	s.mu.Lock()
	for ch := range s.clients {
		select {
		case ch <- msg:
		default:
		}
	}
	s.mu.Unlock()
}

// simplifyError extracts a short, human-readable error from a net error string.
// simplifyError maps raw Go/Windows network errors to short codes.
// The frontend (app.js errCodes) translates these codes into user-friendly text.
func simplifyError(e string) string {
	e = strings.ToLower(e)
	switch {
	// Timeout — server didn't respond in time
	case strings.Contains(e, "timeout"),
		strings.Contains(e, "i/o timeout"),
		strings.Contains(e, "timed out"),
		strings.Contains(e, "did not properly respond"):
		return "timeout"

	// Connection refused — port closed or firewall blocked
	// "connectex" is Windows-specific for refused/timeout
	case strings.Contains(e, "connection refused"),
		strings.Contains(e, "actively refused"),
		strings.Contains(e, "connectex"):
		return "refused"

	// DNS / host not found
	case strings.Contains(e, "no such host"),
		strings.Contains(e, "no route"),
		strings.Contains(e, "name or service not known"),
		strings.Contains(e, "name resolution"):
		return "not_found"

	// Network-level unreachable
	case strings.Contains(e, "network unreachable"),
		strings.Contains(e, "host unreachable"),
		strings.Contains(e, "network is down"):
		return "unreachable"

	// Permission / firewall
	case strings.Contains(e, "permission denied"),
		strings.Contains(e, "access is denied"):
		return "no_permission"

	default:
		return "error"
	}
}

// ── Speed Test ────────────────────────────────────────────────

func (s *Server) serveSpeedTestStart(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Atomic check-and-set: two concurrent starts can't both pass.
	if !s.stRunning.CompareAndSwap(false, true) {
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(`{"error":"test_already_running"}`))
		return
	}

	data, _ := os.ReadFile(s.configPath)
	var cfg config.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		cfg = config.Default
	}
	cfg.Sanitize() // clamp parallel_connections etc. even for hand-edited files
	if len(cfg.SpeedTest.DownloadTargets) == 0 {
		cfg.SpeedTest = config.Default.SpeedTest
	}

	ctx, cancel := context.WithCancel(context.Background())
	// Store the cancel func before anyone can observe stRunning via a cancel
	// request, so a running test is always cancelable.
	s.stateMu.Lock()
	s.stCancel = cancel
	s.stateMu.Unlock()

	timeout := time.Duration(cfg.SpeedTest.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	stCfg := speedtest.Config{
		Endpoints:    cfg.SpeedTest.DownloadTargets,
		Parallel:     cfg.SpeedTest.ParallelConnections,
		Timeout:      timeout,
		UploadTarget: cfg.SpeedTest.UploadTarget,
	}

	go func() {
		defer func() {
			if rec := recover(); rec != nil && s.lgr != nil {
				s.lgr.AppLog("PANIC recovered in speed test goroutine: %v", rec)
			}
			cancel()
			s.stateMu.Lock()
			s.stCancel = nil
			s.stateMu.Unlock()
			// Release the run lock last, so a new test can only start once this
			// goroutine has fully torn down.
			s.stRunning.Store(false)
		}()

		totalSecs := stCfg.Timeout.Seconds()
		result, err := speedtest.Run(ctx, stCfg, func(mbps float64, elapsed time.Duration) {
			progress, _ := json.Marshal(map[string]interface{}{
				"type":            "speed_test_progress",
				"phase":           "download",
				"current_mbps":    mbps,
				"elapsed_seconds": elapsed.Seconds(),
				"total_seconds":   totalSecs,
				"done":            false,
			})
			s.broadcastRaw(string(progress))
		})

		if ctx.Err() != nil && (result == nil || result.DownloadMbps == 0) {
			cancelled, _ := json.Marshal(map[string]interface{}{
				"type":      "speed_test_progress",
				"done":      true,
				"cancelled": true,
			})
			s.broadcastRaw(string(cancelled))
			return
		}

		if err != nil || result == nil {
			return
		}

		// Upload phase (optional). Reuses the run's cancel context so a cancel
		// request stops it too.
		var uploadMbps *float64
		if stCfg.UploadTarget != "" {
			up, _ := json.Marshal(map[string]interface{}{
				"type": "speed_test_progress", "phase": "upload", "done": false,
			})
			s.broadcastRaw(string(up))
			if mbps, uerr := speedtest.MeasureUpload(ctx, stCfg); uerr == nil {
				uploadMbps = &mbps
			}
		}

		event := logger.SpeedTestEvent{
			Timestamp:       time.Now().UTC(),
			Event:           "speed_test",
			DownloadMbps:    result.DownloadMbps,
			UploadMbps:      uploadMbps,
			LatencyMs:       result.LatencyMs,
			DurationSeconds: result.DurationSeconds,
			Endpoints:       result.Endpoints,
			ParallelConns:   result.ParallelConns,
			TriggeredBy:     "user",
		}

		if s.lgr != nil {
			s.lgr.LogSpeedTest(event)
			// Always send webhook with speed test result if webhook is configured
			if cfg.WebhookURL != "" {
				belowThreshold := cfg.SpeedTest.AlertThresholdMbps > 0 &&
					result.DownloadMbps < cfg.SpeedTest.AlertThresholdMbps
				s.lgr.SendSpeedTestResult(cfg.WebhookURL, event, cfg.SpeedTest.AlertThresholdMbps, belowThreshold)
			}
		}

		s.stateMu.Lock()
		s.stLast = &event
		s.stateMu.Unlock()

		done, _ := json.Marshal(map[string]interface{}{
			"type":            "speed_test_progress",
			"phase":           "download",
			"current_mbps":    result.DownloadMbps,
			"elapsed_seconds": result.DurationSeconds,
			"done":            true,
			"result":          event,
		})
		s.broadcastRaw(string(done))
	}()

	w.Write([]byte(`{"ok":true}`))
}

func (s *Server) serveSpeedTestCancel(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Only trigger cancellation; the run goroutine's defer clears stCancel and
	// releases stRunning once it has actually torn down. Flipping them here
	// would let a second test start while the first is still winding down.
	s.stateMu.Lock()
	cancel := s.stCancel
	s.stateMu.Unlock()
	if cancel != nil {
		cancel()
	}
	w.Write([]byte(`{"ok":true}`))
}

func (s *Server) serveSpeedTestHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	limit := 20
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			if n > 100 {
				n = 100
			}
			limit = n
		}
	}
	date := r.URL.Query().Get("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	filename := filepath.Join(s.getLogDir(), fmt.Sprintf("speedtest_%s.jsonl", date))
	data, err := s.readDataFile(filename)
	if os.IsNotExist(err) {
		w.Write([]byte("[]"))
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	lines := bytes.Split(bytes.TrimSpace(data), []byte{'\n'})
	entries := make([]json.RawMessage, 0, len(lines))
	for _, line := range lines {
		line = bytes.TrimSpace(line)
		if len(line) > 0 && json.Valid(line) {
			entries = append(entries, json.RawMessage(line))
		}
	}
	// Reverse to newest-first
	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}
	if len(entries) > limit {
		entries = entries[:limit]
	}
	json.NewEncoder(w).Encode(entries)
}

// ── Monthly report ────────────────────────────────────────────

func validMonth(m string) bool {
	if len(m) != 7 || m[4] != '-' {
		return false
	}
	for i, c := range m {
		if i == 4 {
			continue
		}
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// readMonthlyJSONL reads every "<prefix><month>-*.jsonl" file in the log dir
// (under the logger lock) and unmarshals each valid line into T.
func readMonthlyJSONL[T any](s *Server, prefix, month string) []T {
	dir := s.getLogDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	want := prefix + month
	var out []T
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasPrefix(name, want) || !strings.HasSuffix(name, ".jsonl") {
			continue
		}
		data, err := s.readDataFile(filepath.Join(dir, name))
		if err != nil {
			continue
		}
		for _, line := range bytes.Split(bytes.TrimSpace(data), []byte{'\n'}) {
			line = bytes.TrimSpace(line)
			if len(line) == 0 || !json.Valid(line) {
				continue
			}
			var v T
			if json.Unmarshal(line, &v) == nil {
				out = append(out, v)
			}
		}
	}
	return out
}

func (s *Server) serveReport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	month := r.URL.Query().Get("month")
	if !validMonth(month) {
		month = time.Now().Format("2006-01")
	}
	events := readMonthlyJSONL[types.Event](s, "connectivity_", month)
	samples := readMonthlyJSONL[types.MetricSample](s, "metrics_", month)
	json.NewEncoder(w).Encode(report.Summarize(events, samples, month, time.Now()))
}

// serveMetrics exposes the live monitoring state in Prometheus text exposition
// format. Hand-written to avoid pulling in the prometheus client dependency.
func (s *Server) serveMetrics(w http.ResponseWriter, r *http.Request) {
	s.stateMu.RLock()
	status := s.status
	latency := s.latencyMs
	loss := s.packetLoss
	tcp, httpOK, dns := s.tcpPingOK, s.httpOK, s.dnsOK
	total, disc, up := s.totalChecks, s.disconnections, s.upTicks
	stLast := s.stLast
	hist := make([]int64, len(s.latencyHistory))
	copy(hist, s.latencyHistory)
	s.stateMu.RUnlock()

	b01 := func(ok bool) int {
		if ok {
			return 1
		}
		return 0
	}
	upVal := 0
	if status == "connected" || status == "degraded" {
		upVal = 1
	}
	uptimeRatio := 0.0
	if total > 0 {
		uptimeRatio = float64(up) / float64(total)
	}

	var b strings.Builder
	metric := func(name, typ, help string, val interface{}) {
		fmt.Fprintf(&b, "# HELP %s %s\n# TYPE %s %s\n%s %v\n", name, help, name, typ, name, val)
	}

	fmt.Fprintf(&b, "# HELP internet_monitor_build_info Build information.\n# TYPE internet_monitor_build_info gauge\ninternet_monitor_build_info{version=%q} 1\n", s.version)
	metric("internet_monitor_up", "gauge", "1 if the internet is reachable (connected or degraded), else 0.", upVal)
	metric("internet_monitor_latency_ms", "gauge", "Most recent connectivity latency in milliseconds.", latency)
	metric("internet_monitor_jitter_ms", "gauge", "Mean latency jitter over the recent history window.", jitterOf(hist))
	metric("internet_monitor_packet_loss_ratio", "gauge", "Most recent packet loss ratio (0-1).", loss/100)
	metric("internet_monitor_tcp_ok", "gauge", "1 if the last TCP ping succeeded.", b01(tcp))
	metric("internet_monitor_http_ok", "gauge", "1 if the last HTTP check succeeded.", b01(httpOK))
	metric("internet_monitor_dns_ok", "gauge", "1 if the last DNS lookup succeeded.", b01(dns))
	metric("internet_monitor_checks_total", "counter", "Total checks performed since start.", total)
	metric("internet_monitor_up_checks_total", "counter", "Checks where the link was up since start.", up)
	metric("internet_monitor_disconnections_total", "counter", "Disconnection events since start.", disc)
	metric("internet_monitor_uptime_ratio", "gauge", "Up checks divided by total checks (0-1).", uptimeRatio)
	if stLast != nil {
		metric("internet_monitor_speedtest_download_mbps", "gauge", "Download speed (Mbps) from the last speed test.", stLast.DownloadMbps)
		metric("internet_monitor_speedtest_latency_ms", "gauge", "Latency (ms) from the last speed test.", stLast.LatencyMs)
		if stLast.UploadMbps != nil {
			metric("internet_monitor_speedtest_upload_mbps", "gauge", "Upload speed (Mbps) from the last speed test.", *stLast.UploadMbps)
		}
	}

	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	w.Write([]byte(b.String()))
}

func (s *Server) serveReportPage(w http.ResponseWriter, r *http.Request) {
	data, err := staticFiles.ReadFile("assets/report.html")
	if err != nil {
		http.Error(w, "report.html not found", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(data)
}

func (s *Server) broadcastRaw(msg string) {
	s.mu.Lock()
	for ch := range s.clients {
		select {
		case ch <- msg:
		default:
		}
	}
	s.mu.Unlock()
}

// ── Custom notification sound ─────────────────────────────────

func (s *Server) customSoundPath() string {
	return filepath.Join(filepath.Dir(s.configPath), "notification.mp3")
}

// serveNotificationSound handles GET (serve sound) and POST (upload) and DELETE (reset).
func (s *Server) serveNotificationSound(w http.ResponseWriter, r *http.Request) {
	customPath := s.customSoundPath()

	customExists := false
	if _, err := os.Stat(customPath); err == nil {
		customExists = true
	}

	switch r.Method {
	case http.MethodHead:
		w.Header().Set("Content-Type", "audio/mpeg")
		if customExists {
			w.Header().Set("X-Custom-Sound", "1")
		}
		w.WriteHeader(http.StatusOK)

	case http.MethodGet:
		if customExists {
			w.Header().Set("Content-Type", "audio/mpeg")
			w.Header().Set("X-Custom-Sound", "1")
			http.ServeFile(w, r, customPath)
			return
		}
		// Fall back to embedded default
		data := RingtoneMp3()
		w.Header().Set("Content-Type", "audio/mpeg")
		w.Write(data)

	case http.MethodPost:
		r.Body = http.MaxBytesReader(w, r.Body, 10<<20) // 10 MB limit
		file, _, err := r.FormFile("sound")
		if err != nil {
			http.Error(w, `{"error":"invalid file"}`, http.StatusBadRequest)
			return
		}
		defer file.Close()
		data, err := io.ReadAll(file)
		if err != nil || len(data) == 0 {
			http.Error(w, `{"error":"read failed"}`, http.StatusBadRequest)
			return
		}
		if err := os.WriteFile(customPath, data, 0644); err != nil {
			http.Error(w, `{"error":"save failed"}`, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true}`))

	case http.MethodDelete:
		os.Remove(customPath)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true}`))

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
