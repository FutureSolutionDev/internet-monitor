package server

import (
	"internet-monitor/internal/updater"
	"net/http"
	"sync"
	"time"
)

const (
	updateIdle        = "idle"
	updateDownloading = "downloading"
	updateRestarting  = "restarting"
	updateFailed      = "error"
)

type updateState struct {
	mu     sync.RWMutex
	info   *updater.Info
	status string
	err    string
}

func (u *updateState) get() (*updater.Info, string, string) {
	u.mu.RLock()
	defer u.mu.RUnlock()
	status := u.status
	if status == "" {
		status = updateIdle
	}
	if u.info == nil {
		return nil, status, u.err
	}
	cp := *u.info
	return &cp, status, u.err
}

func (u *updateState) begin() (*updater.Info, string) {
	u.mu.Lock()
	defer u.mu.Unlock()
	switch {
	case u.info == nil || !u.info.HasUpdate:
		return nil, "no_update"
	case u.status == updateDownloading || u.status == updateRestarting:
		return nil, "in_progress"
	}
	u.status, u.err = updateDownloading, ""
	cp := *u.info
	return &cp, ""
}

func (u *updateState) set(status, errMsg string) {
	u.mu.Lock()
	u.status, u.err = status, errMsg
	u.mu.Unlock()
}

func (s *Server) SetUpdateInfo(info *updater.Info) {
	s.update.mu.Lock()
	s.update.info = info
	s.update.mu.Unlock()
	s.broadcast("update")
}

func (s *Server) serveUpdateGet(w http.ResponseWriter, _ *http.Request) {
	info, status, errMsg := s.update.get()
	if info == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"has_update": false, "status": status})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"has_update":      info.HasUpdate,
		"latest_version":  info.LatestVersion,
		"current_version": info.CurrentVersion,
		"release_notes":   info.ReleaseNotes,
		"status":          status,
		"error":           errMsg,
	})
}

func (s *Server) serveUpdateCheck(w http.ResponseWriter, _ *http.Request) {
	if s.hooks.CheckUpdate == nil {
		writeError(w, http.StatusNotImplemented, "unavailable", "updater not configured")
		return
	}
	info, err := s.hooks.CheckUpdate()
	if err != nil {
		s.logf("UPDATE check failed: %v", err)
		writeError(w, http.StatusBadGateway, "check_failed", "could not reach the update server")
		return
	}
	s.SetUpdateInfo(info)
	writeJSON(w, http.StatusOK, info)
}

func (s *Server) serveUpdateApply(w http.ResponseWriter, _ *http.Request) {
	if s.hooks.ApplyUpdate == nil {
		writeError(w, http.StatusNotImplemented, "unavailable", "updater not configured")
		return
	}
	info, code := s.update.begin()
	switch code {
	case "no_update":
		writeError(w, http.StatusBadRequest, code, "no update available")
		return
	case "in_progress":
		writeError(w, http.StatusConflict, code, "update already in progress")
		return
	}
	s.broadcast("update")
	go func() {
		if err := s.hooks.ApplyUpdate(info.DownloadURL); err != nil {
			s.logf("UPDATE apply failed: %v", err)
			s.update.set(updateFailed, err.Error())
			s.broadcast("update")
			return
		}
		s.update.set(updateRestarting, "")
		s.broadcast("update")
		if s.hooks.Restart != nil {
			time.Sleep(750 * time.Millisecond)
			s.hooks.Restart()
		}
	}()
	writeJSON(w, http.StatusAccepted, map[string]interface{}{"ok": true, "status": updateDownloading})
}
