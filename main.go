package main

import (
	"internet-monitor/config"
	"internet-monitor/dashboard"
	"internet-monitor/logger"
	"internet-monitor/monitor"
	"internet-monitor/tray"
	"internet-monitor/updater"
	"log"
	"os"
	"path/filepath"
	"time"

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

	// Wire update callbacks
	dash.OnApplyUpdate = updater.Apply
	dash.OnRestartApp  = updater.Restart

	dash.Start()

	// Background update checker: first check after 30s, then every 6h
	go func() {
		time.Sleep(30 * time.Second)
		for {
			if info, err := updater.Check(Version); err == nil && info.HasUpdate {
				lgr.AppLog("UPDATE available: %s (current: %s)", info.LatestVersion, info.CurrentVersion)
				dash.SetUpdateInfo(&dashboard.UpdateInfo{
					HasUpdate:      info.HasUpdate,
					LatestVersion:  info.LatestVersion,
					CurrentVersion: info.CurrentVersion,
					DownloadURL:    info.DownloadURL,
					ReleaseNotes:   info.ReleaseNotes,
				})
			}
			time.Sleep(6 * time.Hour)
		}
	}()

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
