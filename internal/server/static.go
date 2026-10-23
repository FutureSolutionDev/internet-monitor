package server

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed assets
var uiFiles embed.FS

func staticHandler() http.Handler {
	return http.FileServer(http.FS(uiFiles))
}

func serveEmbedded(w http.ResponseWriter, name string) {
	data, err := fs.ReadFile(uiFiles, name)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	_, _ = w.Write(data)
}

func (s *Server) serveIndex(w http.ResponseWriter, _ *http.Request) {
	serveEmbedded(w, "assets/index.html")
}

func (s *Server) serveReportPage(w http.ResponseWriter, _ *http.Request) {
	serveEmbedded(w, "assets/report.html")
}
