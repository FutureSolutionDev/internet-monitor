package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const (
	clientBuffer   = 32
	maxSSEClients  = 16
	heartbeatEvery = 15 * time.Second
	writeTimeout   = 10 * time.Second
)

type sseClient struct {
	ch      chan []byte
	evicted chan struct{}
	once    sync.Once
}

func (c *sseClient) evict() { c.once.Do(func() { close(c.evicted) }) }

type hub struct {
	mu      sync.Mutex
	clients map[*sseClient]struct{}
}

func newHub() *hub { return &hub{clients: make(map[*sseClient]struct{})} }

func (h *hub) subscribe() (*sseClient, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if len(h.clients) >= maxSSEClients {
		return nil, false
	}
	c := &sseClient{ch: make(chan []byte, clientBuffer), evicted: make(chan struct{})}
	h.clients[c] = struct{}{}
	return c, true
}

func (h *hub) unsubscribe(c *sseClient) {
	h.mu.Lock()
	delete(h.clients, c)
	h.mu.Unlock()
}

func (h *hub) count() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.clients)
}

func (h *hub) publish(msg []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for c := range h.clients {
		select {
		case c.ch <- msg:
		default:
			c.evict()
			delete(h.clients, c)
		}
	}
}

func (s *Server) broadcast(msgType string) {
	if s.hub.count() == 0 {
		return
	}
	data, err := json.Marshal(s.snapshot(msgType))
	if err != nil {
		return
	}
	s.hub.publish(data)
}

func (s *Server) broadcastJSON(v interface{}) {
	if s.hub.count() == 0 {
		return
	}
	data, err := json.Marshal(v)
	if err != nil {
		return
	}
	s.hub.publish(data)
}

func (s *Server) serveSSE(w http.ResponseWriter, r *http.Request) {
	c, ok := s.hub.subscribe()
	if !ok {
		http.Error(w, "too many live connections", http.StatusServiceUnavailable)
		return
	}
	defer s.hub.unsubscribe(c)

	h := w.Header()
	h.Set("Content-Type", "text/event-stream")
	h.Set("Cache-Control", "no-store")
	h.Set("Connection", "keep-alive")
	h.Set("X-Accel-Buffering", "no")

	rc := http.NewResponseController(w)
	write := func(format string, args ...interface{}) bool {
		_ = rc.SetWriteDeadline(time.Now().Add(writeTimeout))
		if _, err := fmt.Fprintf(w, format, args...); err != nil {
			return false
		}
		return rc.Flush() == nil
	}

	init, err := json.Marshal(s.snapshot("init"))
	if err != nil || !write("retry: 2000\ndata: %s\n\n", init) {
		return
	}

	heartbeat := time.NewTicker(heartbeatEvery)
	defer heartbeat.Stop()
	for {
		select {
		case msg := <-c.ch:
			if !write("data: %s\n\n", msg) {
				return
			}
		case <-heartbeat.C:
			if !write("data: {\"type\":\"ping\"}\n\n") {
				return
			}
		case <-c.evicted:
			return
		case <-r.Context().Done():
			return
		case <-s.shutdown:
			return
		}
	}
}
