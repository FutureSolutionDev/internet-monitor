package dashboard

import (
	"bytes"
	"context"
	"encoding/json"
	"embed"
	"fmt"
	"internet-monitor/config"
	"internet-monitor/logger"
	"internet-monitor/monitor"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
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

// ── Types ─────────────────────────────────────────────────────

// EventEntry is the in-memory / SSE representation of a connectivity event.
// Reason is sent as structured booleans so the client can translate it.
type EventEntry struct {
	Time          string  `json:"time"`
	EventType     string  `json:"event_type"`
	Duration      float64 `json:"duration_seconds"`
	TCPFailed     bool    `json:"tcp_failed"`
	HTTPFailed    bool    `json:"http_failed"`
	DNSFailed     bool    `json:"dns_failed"`
	PacketLoss    float64 `json:"packet_loss_pct"`
	LatencyMs     int64   `json:"latency_ms"`
}

type Snapshot struct {
	Type           string       `json:"type"`
	Status         string       `json:"status"`
	LatencyMs      int64        `json:"latency_ms"`
	PacketLoss     float64      `json:"packet_loss"`
	TCPPingOK      bool         `json:"tcp_ping_ok"`
	HTTPOK         bool         `json:"http_ok"`
	DNSOK          bool         `json:"dns_ok"`
	TotalChecks    int          `json:"total_checks"`
	Disconnections int          `json:"disconnections"`
	UptimeSeconds  float64      `json:"uptime_seconds"`
	UptimePct      float64      `json:"uptime_pct"`
	LatencyHistory []int64      `json:"latency_history"`
	Events         []EventEntry `json:"events"`
}

type testTargetResult struct {
	Target    string `json:"target"`
	OK        bool   `json:"ok"`
	LatencyMs int64  `json:"latency_ms"`
	Error     string `json:"error,omitempty"`
}

type testTargetsResponse struct {
	PingTargets []testTargetResult `json:"ping_targets"`
	HTTPTarget  testTargetResult   `json:"http_target"`
	DNSTarget   testTargetResult   `json:"dns_target"`
}

// ── Server ────────────────────────────────────────────────────

type Server struct {
	port       int
	configPath string
	logDir     string
	clients    map[chan string]struct{}
	mu         sync.Mutex

	version string

	OnConfigChange      func(*config.Config)
	OnTestNotification  func()
	OnTestWebhook       func(url string) string
	OnApplyUpdate       func(downloadURL string) error
	OnRestartApp        func()

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
	connectedTicks int
	startTime      time.Time
	latencyHistory []int64
	events         []EventEntry
}

func NewServer(port int, configPath, logDir, version string) *Server {
	return &Server{
		port:       port,
		configPath: configPath,
		logDir:     logDir,
		version:    version,
		clients:    make(map[chan string]struct{}),
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

	go http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", s.port), mux)
}

func (s *Server) URL() string {
	return fmt.Sprintf("http://localhost:%d", s.port)
}

// ── Public update methods ─────────────────────────────────────

func (s *Server) UpdateTick(result monitor.CheckResult, status monitor.Status) {
	s.stateMu.Lock()
	s.status = status.String()
	s.latencyMs = result.LatencyMs
	s.packetLoss = result.PacketLoss
	s.tcpPingOK = result.TCPPingOK
	s.httpOK = result.HTTPOK
	s.dnsOK = result.DNSOK
	s.totalChecks++
	if status == monitor.StatusConnected {
		s.connectedTicks++
	}
	s.latencyHistory = append(s.latencyHistory, result.LatencyMs)
	if len(s.latencyHistory) > maxHistory {
		s.latencyHistory = s.latencyHistory[1:]
	}
	s.stateMu.Unlock()
	s.broadcast("tick")
}

func (s *Server) AddEvent(event monitor.Event) {
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
		pretty, _ := json.MarshalIndent(cfg, "", "  ")
		if err := os.WriteFile(s.configPath, pretty, 0644); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if s.OnConfigChange != nil {
			s.OnConfigChange(&cfg)
		}
		w.Write([]byte(`{"ok":true}`))

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
	filename := filepath.Join(s.logDir, fmt.Sprintf("connectivity_%s.jsonl", date))
	data, err := os.ReadFile(filename)
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
	entries, err := os.ReadDir(s.logDir)
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
		HTTPTarget  string   `json:"http_target"`
		DNSTarget   string   `json:"dns_target"`
	}
	body, _ := io.ReadAll(r.Body)
	json.Unmarshal(body, &req)

	resp := testTargetsResponse{
		PingTargets: make([]testTargetResult, 0),
	}

	// Test TCP ping targets
	for _, target := range req.PingTargets {
		target = strings.TrimSpace(target)
		if target == "" {
			continue
		}
		start := time.Now()
		conn, err := net.DialTimeout("tcp", target, 3*time.Second)
		lat := time.Since(start).Milliseconds()
		r := testTargetResult{Target: target}
		if err == nil {
			conn.Close()
			r.OK = true
			r.LatencyMs = lat
		} else {
			r.Error = simplifyError(err.Error())
		}
		resp.PingTargets = append(resp.PingTargets, r)
	}

	// Test HTTP target
	if req.HTTPTarget != "" {
		start := time.Now()
		client := &http.Client{
			Timeout: 5 * time.Second,
			Transport: &http.Transport{DisableKeepAlives: true},
		}
		httpResp, err := client.Get(req.HTTPTarget)
		lat := time.Since(start).Milliseconds()
		resp.HTTPTarget = testTargetResult{Target: req.HTTPTarget}
		if err == nil {
			httpResp.Body.Close()
			resp.HTTPTarget.OK = true
			resp.HTTPTarget.LatencyMs = lat
		} else {
			resp.HTTPTarget.Error = simplifyError(err.Error())
		}
	}

	// Test DNS target
	if req.DNSTarget != "" {
		start := time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_, err := net.DefaultResolver.LookupHost(ctx, req.DNSTarget)
		lat := time.Since(start).Milliseconds()
		resp.DNSTarget = testTargetResult{Target: req.DNSTarget}
		if err == nil {
			resp.DNSTarget.OK = true
			resp.DNSTarget.LatencyMs = lat
		} else {
			resp.DNSTarget.Error = simplifyError(err.Error())
		}
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
	if results.HTTPTarget.Target != "" {
		tr.HTTPTarget = &logger.TestResult{
			Target:    results.HTTPTarget.Target,
			OK:        results.HTTPTarget.OK,
			LatencyMs: results.HTTPTarget.LatencyMs,
			Error:     results.HTTPTarget.Error,
		}
	}
	if results.DNSTarget.Target != "" {
		tr.DNSTarget = &logger.TestResult{
			Target:    results.DNSTarget.Target,
			OK:        results.DNSTarget.OK,
			LatencyMs: results.DNSTarget.LatencyMs,
			Error:     results.DNSTarget.Error,
		}
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
	body := make([]byte, 2048)
	n, _ := r.Body.Read(body)
	json.Unmarshal(body[:n], &req)

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
		}
	}
}

// ── Internal ──────────────────────────────────────────────────

func (s *Server) snapshot(msgType string) Snapshot {
	s.stateMu.RLock()
	defer s.stateMu.RUnlock()

	hist := make([]int64, len(s.latencyHistory))
	copy(hist, s.latencyHistory)
	evts := make([]EventEntry, len(s.events))
	copy(evts, s.events)

	uptimePct := 0.0
	if s.totalChecks > 0 {
		uptimePct = float64(s.connectedTicks) / float64(s.totalChecks) * 100
	}

	return Snapshot{
		Type:           msgType,
		Status:         s.status,
		LatencyMs:      s.latencyMs,
		PacketLoss:     s.packetLoss,
		TCPPingOK:      s.tcpPingOK,
		HTTPOK:         s.httpOK,
		DNSOK:          s.dnsOK,
		TotalChecks:    s.totalChecks,
		Disconnections: s.disconnections,
		UptimeSeconds:  time.Since(s.startTime).Seconds(),
		UptimePct:      uptimePct,
		LatencyHistory: hist,
		Events:         evts,
	}
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
