package main

import (
	"context"
	"internet-monitor/config"
	"internet-monitor/core"
	"internet-monitor/dashboard"
	"internet-monitor/logger"
	"internet-monitor/monitor"
	"internet-monitor/notifytext"
	"internet-monitor/sound"
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
		if err := os.Chdir(filepath.Dir(exePath)); err != nil {
			log.Printf("warning: could not chdir to exe dir: %v — config/logs will use the current working directory", err)
		}
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
	// Route notification/sound diagnostics to logs/app.log (the standard log
	// package is invisible in a -H=windowsgui build).
	tray.Logf = lgr.AppLog
	sound.Logf = lgr.AppLog

	dash := dashboard.NewServer(cfg.DashboardPort, "config.json", cfg.LogDir, Version, lgr)
	dash.OnApplyUpdate = updater.Apply
	dash.OnRestartApp = updater.Restart
	dash.OnTestWebhook = lgr.SendTestWebhook
	dash.SetNativeNotifications(true) // tray build shows OS toasts
	dash.Start()

	checker := monitor.NewChecker(cfg)
	t := tray.New(cfg.LogDir, dash.URL())

	engine := core.New(cfg, checker, lgr, Version)
	dash.OnConfigChange = engine.ApplyConfig
	engine.Notifier = core.MultiNotifier{
		tray.NewNotifier(t),
		dashboard.NewNotifier(dash),
	}
	engine.OnUpdateAvailable = func(info *updater.Info) {
		dash.SetUpdateInfo(&dashboard.UpdateInfo{
			HasUpdate:      info.HasUpdate,
			LatestVersion:  info.LatestVersion,
			CurrentVersion: info.CurrentVersion,
			DownloadURL:    info.DownloadURL,
			ReleaseNotes:   info.ReleaseNotes,
		})
	}

	if png := dashboard.FaviconPNG(); len(png) > 0 {
		tray.SetCustomIcon(png)
	}
	dash.OnTestNotification = func(lang string) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[notify] tray test notification panic recovered: %v", r)
			}
		}()
		title, body := notifytext.TestMessage(lang)
		tray.Notify(title, body)
	}

	engine.Start()
	systray.Run(t.OnReady, t.OnExit)
	engine.Stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := dash.Shutdown(shutdownCtx); err != nil {
		log.Printf("dashboard shutdown error: %v", err)
	}
	os.Exit(0)
}
