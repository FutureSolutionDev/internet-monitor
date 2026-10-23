package server

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	sessionQuery  = "session"
	sessionHeader = "X-IM-Session"
	sessionFile   = ".session"
)

type SessionInfo struct {
	Port  int    `json:"port"`
	Token string `json:"token"`
	PID   int    `json:"pid"`
}

func newToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *Server) cookieName() string { return fmt.Sprintf("im_session_%d", s.port) }

func (s *Server) validToken(v string) bool {
	return v != "" && subtle.ConstantTimeCompare([]byte(v), []byte(s.token)) == 1
}

func (s *Server) authenticated(r *http.Request) bool {
	if s.validToken(r.Header.Get(sessionHeader)) {
		return true
	}
	c, err := r.Cookie(s.cookieName())
	return err == nil && s.validToken(c.Value)
}

func protectedPath(p string) bool {
	return strings.HasPrefix(p, "/api/") || p == "/events" || p == "/notification-sound"
}

func (s *Server) handleSessionLaunch(w http.ResponseWriter, r *http.Request) bool {
	q := r.URL.Query()
	tok := q.Get(sessionQuery)
	if tok == "" || r.Method != http.MethodGet || protectedPath(r.URL.Path) {
		return false
	}
	if s.validToken(tok) {
		http.SetCookie(w, &http.Cookie{
			Name:     s.cookieName(),
			Value:    s.token,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
		})
	}
	q.Del(sessionQuery)
	target := r.URL.Path
	if enc := q.Encode(); enc != "" {
		target += "?" + enc
	}
	http.Redirect(w, r, target, http.StatusSeeOther)
	return true
}

func (s *Server) LaunchURL() string {
	return fmt.Sprintf("http://127.0.0.1:%d/?%s=%s", s.port, sessionQuery, s.token)
}

func (s *Server) sessionPath() string { return filepath.Join(s.appDir, sessionFile) }

func (s *Server) writeSessionFile() {
	if s.appDir == "" {
		return
	}
	data, _ := json.Marshal(SessionInfo{Port: s.port, Token: s.token, PID: os.Getpid()})
	if err := os.WriteFile(s.sessionPath(), data, 0o600); err != nil {
		s.logf("session file: %v", err)
	}
}

func (s *Server) removeSessionFile() {
	if s.appDir != "" {
		_ = os.Remove(s.sessionPath())
	}
}

func ReadSession(appDir string) (SessionInfo, bool) {
	var info SessionInfo
	data, err := os.ReadFile(filepath.Join(appDir, sessionFile))
	if err != nil || json.Unmarshal(data, &info) != nil || info.Port == 0 || info.Token == "" {
		return SessionInfo{}, false
	}
	return info, true
}

func (i SessionInfo) LaunchURL() string {
	return fmt.Sprintf("http://127.0.0.1:%d/?%s=%s", i.Port, sessionQuery, i.Token)
}
