package server

import (
	"internet-monitor/internal/alerts"
	"net/http"
	"testing"
	"time"
)

type notificationsBody struct {
	Items      []alerts.Record `json:"items"`
	Unread     int             `json:"unread"`
	LastReadAt *time.Time      `json:"last_read_at"`
}

func TestNotificationsAPI(t *testing.T) {
	e := newTestEnv(t)
	resp := e.do(t, http.MethodGet, "/api/notifications", "", nil)
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("without history: %d", resp.StatusCode)
	}
	resp.Body.Close()

	dir := t.TempDir()
	h := alerts.OpenHistory(func() string { return dir }, nil)
	e.s.history = h
	base := time.Now().Add(-time.Hour).UTC()
	for i, k := range []alerts.Kind{alerts.KindDown, alerts.KindRestored, alerts.KindTest} {
		h.Add(alerts.Record{ID: string(k), Time: base.Add(time.Duration(i) * time.Minute), Kind: k, Sent: []alerts.Channel{}})
	}

	resp = e.do(t, http.MethodGet, "/api/notifications?limit=2", "", nil)
	got := decode[notificationsBody](t, resp.Body)
	resp.Body.Close()
	if len(got.Items) != 2 || got.Items[0].Kind != alerts.KindTest || got.Unread != 3 || got.LastReadAt != nil {
		t.Fatalf("list = %+v", got)
	}

	for _, bad := range []string{"0", "-1", "abc"} {
		r := e.do(t, http.MethodGet, "/api/notifications?limit="+bad, "", nil)
		if r.StatusCode != http.StatusBadRequest {
			t.Errorf("limit=%s: %d", bad, r.StatusCode)
		}
		r.Body.Close()
	}

	until := base.Add(90 * time.Second).Format(time.RFC3339Nano)
	resp = e.do(t, http.MethodPost, "/api/notifications/read", `{"until":"`+until+`"}`, nil)
	if resp.StatusCode != http.StatusOK || decode[struct{ Unread int }](t, resp.Body).Unread != 1 {
		t.Fatalf("partial read failed: %d", resp.StatusCode)
	}
	resp.Body.Close()

	future := time.Now().Add(24 * time.Hour).Format(time.RFC3339)
	resp = e.do(t, http.MethodPost, "/api/notifications/read", `{"until":"`+future+`"}`, nil)
	resp.Body.Close()
	if h.LastRead().After(time.Now()) {
		t.Fatalf("read cursor moved into the future: %v", h.LastRead())
	}
	if h.Unread() != 0 {
		t.Fatalf("unread = %d", h.Unread())
	}

	resp = e.do(t, http.MethodPost, "/api/notifications/read", `{"until":`, nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("bad json: %d", resp.StatusCode)
	}
	resp.Body.Close()
}
