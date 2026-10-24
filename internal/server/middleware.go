package server

import (
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"runtime/debug"
	"strings"
)

const contentSecurityPolicy = "default-src 'self'; img-src 'self' data:; font-src 'self'; style-src 'self' 'unsafe-inline'; script-src 'self'; connect-src 'self'; object-src 'none'; base-uri 'none'; frame-ancestors 'none'; form-action 'self'"

func (s *Server) allowedHost(host string) bool {
	host = strings.ToLower(host)
	for _, h := range []string{"127.0.0.1", "localhost", "[::1]"} {
		if host == fmt.Sprintf("%s:%d", h, s.port) {
			return true
		}
	}
	return false
}

func (s *Server) allowedOrigin(origin string) bool {
	origin = strings.ToLower(origin)
	return origin == fmt.Sprintf("http://127.0.0.1:%d", s.port) ||
		origin == fmt.Sprintf("http://localhost:%d", s.port) ||
		origin == fmt.Sprintf("http://[::1]:%d", s.port)
}

func isStateChanging(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return false
	}
	return true
}

func acceptableBody(r *http.Request) bool {
	if r.ContentLength == 0 && len(r.TransferEncoding) == 0 {
		return true
	}
	mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		return false
	}
	return mt == "application/json" || mt == "multipart/form-data"
}

func (s *Server) guard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Cross-Origin-Resource-Policy", "same-origin")
		h.Set("Content-Security-Policy", contentSecurityPolicy)

		if !s.allowedHost(r.Host) {
			http.Error(w, "forbidden host", http.StatusForbidden)
			return
		}
		if origin := r.Header.Get("Origin"); origin != "" && !s.allowedOrigin(origin) {
			http.Error(w, "cross-origin request denied", http.StatusForbidden)
			return
		}
		if r.Header.Get("Sec-Fetch-Site") == "cross-site" {
			http.Error(w, "cross-site request denied", http.StatusForbidden)
			return
		}
		if s.handleSessionLaunch(w, r) {
			return
		}
		if protectedPath(r.URL.Path) && !s.authenticated(r) {
			writeError(w, http.StatusUnauthorized, "unauthorized", "open the dashboard from the Internet Monitor tray icon or window")
			return
		}
		if isStateChanging(r.Method) && !acceptableBody(r) {
			http.Error(w, "unsupported content type", http.StatusUnsupportedMediaType)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/api/") {
			h.Set("Cache-Control", "no-store")
		}

		defer func() {
			if rec := recover(); rec != nil {
				if rec == http.ErrAbortHandler {
					panic(rec)
				}
				s.logf("PANIC in %s %s: %v\n%s", r.Method, r.URL.Path, rec, debug.Stack())
				http.Error(w, "internal error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, msg string) {
	writeJSON(w, status, map[string]interface{}{"ok": false, "code": code, "error": msg})
}

func decodeJSON(w http.ResponseWriter, r *http.Request, limit int64, v interface{}) bool {
	r.Body = http.MaxBytesReader(w, r.Body, limit)
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(v); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "invalid JSON body")
		return false
	}
	return true
}
