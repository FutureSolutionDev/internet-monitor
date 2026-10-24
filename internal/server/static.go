package server

import (
	"bytes"
	"internet-monitor/web"
	"io/fs"
	"net/http"
	"path"
	"strings"
	"time"
)

const notBuiltPage = `<!doctype html><html><head><meta charset="utf-8"><title>Internet Monitor</title></head>
<body style="font-family:system-ui;background:#0c1015;color:#f3f4f6;display:grid;place-items:center;height:100vh;margin:0">
<p>The dashboard UI is not built. Run <code>bun run build</code> in <code>web/</code> and rebuild the app.</p></body></html>`

var uiFS = func() fs.FS {
	sub, err := fs.Sub(web.Dist, "dist")
	if err != nil {
		return web.Dist
	}
	return sub
}()

func (s *Server) serveUI(w http.ResponseWriter, r *http.Request) {
	if protectedPath(r.URL.Path) {
		writeError(w, http.StatusNotFound, "not_found", "no such endpoint")
		return
	}
	name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
	if name != "" && name != "index.html" {
		if data, err := fs.ReadFile(uiFS, name); err == nil {
			if strings.HasPrefix(name, "app/") {
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			} else {
				w.Header().Set("Cache-Control", "no-cache")
			}
			http.ServeContent(w, r, name, time.Time{}, bytes.NewReader(data))
			return
		}
		if path.Ext(name) != "" {
			http.NotFound(w, r)
			return
		}
	}
	index, err := fs.ReadFile(uiFS, "index.html")
	if err != nil {
		index = []byte(notBuiltPage)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	_, _ = w.Write(index)
}
