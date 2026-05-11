package dashboard

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"internet-monitor/monitor"
	"net/http"
	"strings"
	"sync"
	"time"
)

//go:embed index.html
var indexHTML string

const maxHistory = 60
const maxEvents = 100

type EventEntry struct {
	Time      string  `json:"time"`
	EventType string  `json:"event_type"`
	Duration  float64 `json:"duration_seconds"`
	Reason    string  `json:"reason"`
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
	LatencyHistory []int64      `json:"latency_history"`
	Events         []EventEntry `json:"events"`
}

type Server struct {
	port    int
	clients map[chan string]struct{}
	mu      sync.Mutex

	stateMu        sync.RWMutex
	status         string
	latencyMs      int64
	packetLoss     float64
	tcpPingOK      bool
	httpOK         bool
	dnsOK          bool
	totalChecks    int
	disconnections int
	startTime      time.Time
	latencyHistory []int64
	events         []EventEntry
}

func NewServer(port int) *Server {
	return &Server{
		port:      port,
		clients:   make(map[chan string]struct{}),
		startTime: time.Now(),
		status:    "checking",
	}
}

func (s *Server) Start() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.serveIndex)
	mux.HandleFunc("/events", s.serveSSE)
	mux.HandleFunc("/api/state", s.serveState)
	go http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", s.port), mux)
}

func (s *Server) URL() string {
	return fmt.Sprintf("http://localhost:%d", s.port)
}

func (s *Server) UpdateTick(result monitor.CheckResult, status monitor.Status) {
	s.stateMu.Lock()
	s.status = status.String()
	s.latencyMs = result.LatencyMs
	s.packetLoss = result.PacketLoss
	s.tcpPingOK = result.TCPPingOK
	s.httpOK = result.HTTPOK
	s.dnsOK = result.DNSOK
	s.totalChecks++
	s.latencyHistory = append(s.latencyHistory, result.LatencyMs)
	if len(s.latencyHistory) > maxHistory {
		s.latencyHistory = s.latencyHistory[1:]
	}
	s.stateMu.Unlock()
	s.broadcast("tick")
}

func (s *Server) AddEvent(event monitor.Event) {
	parts := []string{}
	if event.Reason.TCPPingFailed {
		parts = append(parts, "TCP ping")
	}
	if event.Reason.HTTPFailed {
		parts = append(parts, "HTTP")
	}
	if event.Reason.DNSFailed {
		parts = append(parts, "DNS")
	}
	reason := "—"
	if len(parts) > 0 {
		reason = strings.Join(parts, ", ") + " failed"
	}
	if event.Reason.PacketLossPct > 20 {
		reason += fmt.Sprintf(" (loss %.0f%%)", event.Reason.PacketLossPct)
	}

	entry := EventEntry{
		Time:      event.Timestamp.Format("15:04:05"),
		EventType: event.EventType,
		Duration:  event.DurationSeconds,
		Reason:    reason,
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

func (s *Server) serveIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, indexHTML)
}

func (s *Server) serveState(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.snapshot("init"))
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

	// Send current state immediately on connect
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

func (s *Server) snapshot(msgType string) Snapshot {
	s.stateMu.RLock()
	defer s.stateMu.RUnlock()

	hist := make([]int64, len(s.latencyHistory))
	copy(hist, s.latencyHistory)
	evts := make([]EventEntry, len(s.events))
	copy(evts, s.events)

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
