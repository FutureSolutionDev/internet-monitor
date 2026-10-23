package server

import (
	"encoding/json"
	"errors"
	"internet-monitor/internal/jsonl"
	"internet-monitor/internal/report"
	"internet-monitor/internal/types"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	defaultLogLimit = 500
	maxLogLimit     = 5000
)

func parseDate(v string) (string, bool) {
	if v == "" {
		return time.Now().Format("2006-01-02"), true
	}
	if _, err := time.Parse("2006-01-02", v); err != nil {
		return "", false
	}
	return v, true
}

func parseLimit(v string, def, maxN int) int {
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return def
	}
	return min(n, maxN)
}

func (s *Server) tailFile(w http.ResponseWriter, prefix, date string, limit int) {
	path := filepath.Join(s.logDir(), prefix+date+".jsonl")
	entries, err := jsonl.Tail(path, limit)
	switch {
	case errors.Is(err, os.ErrNotExist):
		writeJSON(w, http.StatusOK, []json.RawMessage{})
	case err != nil:
		s.logf("read %s: %v", path, err)
		writeError(w, http.StatusInternalServerError, "read_failed", "could not read log file")
	default:
		writeJSON(w, http.StatusOK, entries)
	}
}

func (s *Server) serveLogs(w http.ResponseWriter, r *http.Request) {
	date, ok := parseDate(r.URL.Query().Get("date"))
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid_date", "date must be YYYY-MM-DD")
		return
	}
	s.tailFile(w, "connectivity_", date, parseLimit(r.URL.Query().Get("limit"), defaultLogLimit, maxLogLimit))
}

func (s *Server) serveSpeedTestHistory(w http.ResponseWriter, r *http.Request) {
	date, ok := parseDate(r.URL.Query().Get("date"))
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid_date", "date must be YYYY-MM-DD")
		return
	}
	s.tailFile(w, "speedtest_", date, parseLimit(r.URL.Query().Get("limit"), 20, 200))
}

func (s *Server) serveLogDates(w http.ResponseWriter, _ *http.Request) {
	entries, err := os.ReadDir(s.logDir())
	dates := []string{}
	if err == nil {
		for _, e := range entries {
			name := e.Name()
			if e.IsDir() || !strings.HasPrefix(name, "connectivity_") || !strings.HasSuffix(name, ".jsonl") {
				continue
			}
			d := strings.TrimSuffix(strings.TrimPrefix(name, "connectivity_"), ".jsonl")
			if _, ok := parseDate(d); ok {
				dates = append(dates, d)
			}
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(dates)))
	writeJSON(w, http.StatusOK, dates)
}

func validMonth(m string) bool {
	_, err := time.Parse("2006-01", m)
	return err == nil
}

func adjacentMonths(month string) (prev, next string) {
	t, err := time.ParseInLocation("2006-01", month, time.Local)
	if err != nil {
		return "", ""
	}
	return t.AddDate(0, -1, 0).Format("2006-01"), t.AddDate(0, 1, 0).Format("2006-01")
}

func readMonthly[T any](dir, prefix, month string) []T {
	if month == "" {
		return nil
	}
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
		_ = jsonl.Decode(filepath.Join(dir, name), func(v T) { out = append(out, v) })
	}
	return out
}

func (s *Server) serveReport(w http.ResponseWriter, r *http.Request) {
	month := r.URL.Query().Get("month")
	if !validMonth(month) {
		month = time.Now().Format("2006-01")
	}
	dir := s.logDir()
	prev, next := adjacentMonths(month)
	events := readMonthly[types.Event](dir, "connectivity_", month)
	events = append(events, readMonthly[types.Event](dir, "connectivity_", prev)...)
	events = append(events, readMonthly[types.Event](dir, "connectivity_", next)...)
	samples := readMonthly[types.MetricSample](dir, "metrics_", month)
	writeJSON(w, http.StatusOK, report.Summarize(events, samples, month, time.Now()))
}

type Availability struct {
	Today *float64 `json:"today"`
	Week  *float64 `json:"week"`
	Month *float64 `json:"month"`
}

type availabilityCache struct {
	s  *Server
	mu sync.RWMutex
	v  *Availability
	at time.Time
}

func pct(total, up int) *float64 {
	if total == 0 {
		return nil
	}
	v := float64(up) / float64(total) * 100
	return &v
}

func computeAvailability(dir string, now time.Time) Availability {
	var total, up, tToday, uToday, tWeek, uWeek int
	for i := 0; i < 30; i++ {
		date := now.AddDate(0, 0, -i).Format("2006-01-02")
		_ = jsonl.Decode(filepath.Join(dir, "metrics_"+date+".jsonl"), func(m types.MetricSample) {
			total += m.Samples
			up += m.UpSamples
		})
		switch i {
		case 0:
			tToday, uToday = total, up
		case 6:
			tWeek, uWeek = total, up
		}
	}
	return Availability{Today: pct(tToday, uToday), Week: pct(tWeek, uWeek), Month: pct(total, up)}
}

func (a *availabilityCache) refresh() *Availability {
	v := computeAvailability(a.s.logDir(), time.Now())
	a.mu.Lock()
	a.v, a.at = &v, time.Now()
	a.mu.Unlock()
	return &v
}

func (a *availabilityCache) get() *Availability {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.v
}

func (a *availabilityCache) loop(stop <-chan struct{}) {
	a.refresh()
	a.s.broadcast("availability")
	t := time.NewTicker(time.Minute)
	defer t.Stop()
	for {
		select {
		case <-stop:
			return
		case <-t.C:
			a.refresh()
			a.s.broadcast("availability")
		}
	}
}

func (s *Server) serveAvailability(w http.ResponseWriter, _ *http.Request) {
	v := s.avail.get()
	if v == nil {
		v = s.avail.refresh()
	}
	writeJSON(w, http.StatusOK, v)
}
