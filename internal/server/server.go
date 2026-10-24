package server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"internet-monitor/internal/alerts"
	"internet-monitor/internal/config"
	"internet-monitor/internal/logger"
	"internet-monitor/internal/updater"
	"net"
	"net/http"
	"sync"
	"time"
)

type Hooks struct {
	TestNotification func(lang string, volume *int)
	PlaySound        func(volume int)
	TestWebhook      func(url string) string
	CheckUpdate      func() (*updater.Info, error)
	ApplyUpdate      func(downloadURL string) error
	Restart          func()
	SpeedResult      func(event logger.SpeedTestEvent, thresholdMbps float64, below bool)
}

type Options struct {
	Store               *config.Store
	Logger              *logger.Logger
	Version             string
	AppDir              string
	NativeNotifications bool
	Hooks               Hooks
	History             *alerts.History
}

type Server struct {
	store   *config.Store
	lgr     *logger.Logger
	version string
	appDir  string
	bootID  string
	token   string
	native  bool
	hooks   Hooks
	history *alerts.History

	port     int
	srv      *http.Server
	shutdown chan struct{}
	once     sync.Once

	hub    *hub
	live   *liveState
	update *updateState
	speed  *speedRunner
	avail  *availabilityCache
	tester *rateLimiter
	toasts *rateLimiter
}

func New(opts Options) *Server {
	s := &Server{
		store:    opts.Store,
		lgr:      opts.Logger,
		version:  opts.Version,
		appDir:   opts.AppDir,
		bootID:   newBootID(),
		token:    newToken(),
		native:   opts.NativeNotifications,
		hooks:    opts.Hooks,
		history:  opts.History,
		shutdown: make(chan struct{}),
		hub:      newHub(),
		live:     newLiveState(),
		update:   &updateState{},
		tester:   newRateLimiter(time.Minute),
		toasts:   newRateLimiter(2 * time.Second),
	}
	s.speed = &speedRunner{s: s}
	s.avail = &availabilityCache{s: s}
	if s.store != nil {
		s.store.Subscribe(func(config.Config) { s.broadcast("config") })
	}
	return s
}

func newBootID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *Server) logf(format string, args ...interface{}) {
	if s.lgr != nil {
		s.lgr.AppLog(format, args...)
	}
}

func (s *Server) config() config.Config {
	if s.store == nil {
		return config.Default()
	}
	return s.store.Get()
}

func (s *Server) logDir() string {
	if s.lgr != nil {
		return s.lgr.LogDir()
	}
	return s.config().LogDir
}

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", s.serveUI)
	mux.HandleFunc("GET /events", s.serveSSE)
	mux.HandleFunc("GET /metrics", s.serveMetrics)

	mux.HandleFunc("GET /api/state", s.serveState)
	mux.HandleFunc("GET /api/version", s.serveVersion)
	mux.HandleFunc("GET /api/config", s.serveConfigGet)
	mux.HandleFunc("POST /api/config", s.serveConfigPost)
	mux.HandleFunc("POST /api/language", s.serveLanguage)

	mux.HandleFunc("GET /api/logs", s.serveLogs)
	mux.HandleFunc("GET /api/log-dates", s.serveLogDates)
	mux.HandleFunc("GET /api/availability", s.serveAvailability)
	mux.HandleFunc("GET /api/report", s.serveReport)

	mux.HandleFunc("POST /api/test-targets", s.serveTestTargets)
	mux.HandleFunc("POST /api/test-notification", s.serveTestNotification)
	mux.HandleFunc("POST /api/test-webhook", s.serveTestWebhook)
	mux.HandleFunc("POST /api/play-sound", s.servePlaySound)
	mux.HandleFunc("GET /api/startup", s.serveStartupGet)
	mux.HandleFunc("POST /api/startup", s.serveStartupPost)

	mux.HandleFunc("GET /api/update", s.serveUpdateGet)
	mux.HandleFunc("POST /api/update", s.serveUpdateApply)
	mux.HandleFunc("POST /api/update/check", s.serveUpdateCheck)

	mux.HandleFunc("POST /api/speed-test/start", s.serveSpeedTestStart)
	mux.HandleFunc("POST /api/speed-test/cancel", s.serveSpeedTestCancel)
	mux.HandleFunc("GET /api/speed-test/history", s.serveSpeedTestHistory)

	mux.HandleFunc("GET /api/notifications", s.serveNotifications)
	mux.HandleFunc("POST /api/notifications/read", s.serveNotificationsRead)

	mux.HandleFunc("HEAD /notification-sound", s.serveSoundGet)
	mux.HandleFunc("GET /notification-sound", s.serveSoundGet)
	mux.HandleFunc("POST /notification-sound", s.serveSoundUpload)
	mux.HandleFunc("DELETE /notification-sound", s.serveSoundReset)
	return s.guard(mux)
}

func (s *Server) Start(port int) error {
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return fmt.Errorf("dashboard port %d is unavailable: %w", port, err)
	}
	s.port = ln.Addr().(*net.TCPAddr).Port
	s.srv = &http.Server{
		Handler:           s.routes(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       60 * time.Second,
		IdleTimeout:       2 * time.Minute,
		MaxHeaderBytes:    64 << 10,
	}
	go func() {
		if err := s.srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logf("dashboard server stopped: %v", err)
		}
	}()
	s.writeSessionFile()
	go s.speed.schedule(s.shutdown)
	go s.avail.loop(s.shutdown)
	return nil
}

func (s *Server) Port() int { return s.port }

func (s *Server) URL() string { return fmt.Sprintf("http://127.0.0.1:%d", s.port) }

func (s *Server) Shutdown(ctx context.Context) error {
	s.once.Do(func() {
		close(s.shutdown)
		s.removeSessionFile()
	})
	s.speed.cancel()
	if s.srv != nil {
		return s.srv.Shutdown(ctx)
	}
	return nil
}
