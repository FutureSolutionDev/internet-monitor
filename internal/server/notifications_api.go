package server

import (
	"encoding/json"
	"internet-monitor/internal/alerts"
	"io"
	"net/http"
	"strconv"
	"time"
)

const defaultNotificationLimit = 50

func (s *Server) PublishNotification(r alerts.Record) {
	unread := 0
	if s.history != nil {
		unread = s.history.Unread()
	}
	s.broadcastJSON(map[string]interface{}{"type": "notification", "notification": r, "unread": unread})
}

func (s *Server) serveNotifications(w http.ResponseWriter, r *http.Request) {
	if s.history == nil {
		writeError(w, http.StatusServiceUnavailable, "unavailable", "notification history not ready")
		return
	}
	limit := defaultNotificationLimit
	if v := r.URL.Query().Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			writeError(w, http.StatusBadRequest, "invalid", "limit must be a positive number")
			return
		}
		limit = min(n, alerts.MaxListLimit)
	}
	var lastRead *time.Time
	if t := s.history.LastRead(); !t.IsZero() {
		lastRead = &t
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"items":        s.history.List(limit),
		"unread":       s.history.Unread(),
		"last_read_at": lastRead,
	})
}

func (s *Server) serveNotificationsRead(w http.ResponseWriter, r *http.Request) {
	if s.history == nil {
		writeError(w, http.StatusServiceUnavailable, "unavailable", "notification history not ready")
		return
	}
	now := time.Now()
	until := now
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1024))
	if err != nil {
		writeError(w, http.StatusRequestEntityTooLarge, "too_large", "request body too large")
		return
	}
	if len(body) > 0 {
		var req struct {
			Until *time.Time `json:"until"`
		}
		if err := json.Unmarshal(body, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid_json", "invalid JSON body")
			return
		}
		if req.Until != nil && req.Until.Before(now) {
			until = *req.Until
		}
	}
	unread := s.history.MarkRead(until)
	s.broadcastJSON(map[string]interface{}{"type": "notifications_read", "unread": unread})
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "unread": unread})
}
