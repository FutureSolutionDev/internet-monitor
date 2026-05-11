package main

import (
	"internet-monitor/config"
	"internet-monitor/dashboard"
	"internet-monitor/logger"
	"internet-monitor/monitor"
	"internet-monitor/tray"
	"log"
	"os"
	"path/filepath"

	"github.com/getlantern/systray"
)

// Version is embedded at build time via: -ldflags "-X main.Version=v1.x.x"
var Version = "dev"

func main() {
	ensureSingleInstance()

	if exePath, err := os.Executable(); err == nil {
		if resolved, err2 := filepath.EvalSymlinks(exePath); err2 == nil {
			exePath = resolved
		}
		os.Chdir(filepath.Dir(exePath))
	}

	cfg, err := config.Load("config.json")
	if err != nil {
		log.Printf("config error: %v — using defaults", err)
		cfg = &config.Default
	}

	lgr, err := logger.New(cfg)
	if err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}

	dash := dashboard.NewServer(cfg.DashboardPort, "config.json", cfg.LogDir, Version)
	dash.Start()

	// Use favicon.png as the tray icon (embedded in the dashboard assets)
	if png := dashboard.FaviconPNG(); len(png) > 0 {
		tray.SetCustomIcon(png)
	}

	dash.OnTestNotification = func() {
		tray.Notify("اختبار الإشعار / Test Notification", "🔔 الإشعار يعمل بشكل صحيح")
	}
	dash.OnTestWebhook = lgr.SendTestWebhook

	checker := monitor.NewChecker(cfg)
	t := tray.New(cfg, checker, lgr, dash.URL())
	t.OnTick  = dash.UpdateTick
	t.OnEvent = dash.AddEvent

	systray.Run(t.OnReady, t.OnExit)
}
