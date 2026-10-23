package server

import (
	"encoding/json"
	"errors"
	"internet-monitor/internal/config"
	"io"
	"net/http"
	"strconv"
	"strings"
)

const maxConfigBody = 64 << 10

func revisionFrom(r *http.Request) uint64 {
	for _, v := range []string{r.Header.Get("If-Match"), r.Header.Get("X-Config-Revision")} {
		v = strings.Trim(strings.TrimPrefix(strings.TrimSpace(v), "W/"), `"`)
		if n, err := strconv.ParseUint(v, 10, 64); err == nil {
			return n
		}
	}
	return 0
}

func setRevisionHeaders(w http.ResponseWriter, rev uint64) {
	w.Header().Set("ETag", `"`+strconv.FormatUint(rev, 10)+`"`)
	w.Header().Set("X-Config-Revision", strconv.FormatUint(rev, 10))
}

func (s *Server) serveConfigGet(w http.ResponseWriter, _ *http.Request) {
	if s.store == nil {
		writeError(w, http.StatusServiceUnavailable, "unavailable", "configuration store not ready")
		return
	}
	cfg, rev := s.store.Snapshot()
	setRevisionHeaders(w, rev)
	writeJSON(w, http.StatusOK, cfg)
}

func (s *Server) writeConfigError(w http.ResponseWriter, err error) {
	var fe *config.FieldError
	switch {
	case errors.As(err, &fe):
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{"ok": false, "code": "invalid", "field": fe.Field, "error": fe.Message})
	case errors.Is(err, config.ErrConflict):
		cfg, rev := s.store.Snapshot()
		setRevisionHeaders(w, rev)
		writeJSON(w, http.StatusConflict, map[string]interface{}{"ok": false, "code": "conflict", "error": err.Error(), "revision": rev, "config": cfg})
	default:
		s.logf("CONFIG save failed: %v", err)
		writeError(w, http.StatusInternalServerError, "save_failed", "could not save configuration")
	}
}

func (s *Server) serveConfigPost(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		writeError(w, http.StatusServiceUnavailable, "unavailable", "configuration store not ready")
		return
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxConfigBody))
	if err != nil || !json.Valid(body) {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid JSON body")
		return
	}
	prevPort := s.store.Get().DashboardPort
	cfg, rev, err := s.store.Update(revisionFrom(r), func(c *config.Config) error {
		return json.Unmarshal(body, c)
	})
	if err != nil {
		s.writeConfigError(w, err)
		return
	}
	restart := cfg.DashboardPort != s.port
	if cfg.DashboardPort != prevPort {
		s.logf("CONFIG dashboard_port changed to %d — restart required", cfg.DashboardPort)
	}
	setRevisionHeaders(w, rev)
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "revision": rev, "restart_required": restart, "config": cfg})
}

func (s *Server) serveLanguage(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		writeError(w, http.StatusServiceUnavailable, "unavailable", "configuration store not ready")
		return
	}
	lang := r.URL.Query().Get("lang")
	if lang != "ar" {
		lang = "en"
	}
	_, rev, err := s.store.Update(0, func(c *config.Config) error {
		c.Language = lang
		return nil
	})
	if err != nil {
		s.writeConfigError(w, err)
		return
	}
	setRevisionHeaders(w, rev)
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "revision": rev, "language": lang})
}
