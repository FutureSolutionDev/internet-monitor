//go:generate go-winres simply --arch amd64 --icon ../../dashboard/assets/favicon.png

package main

import (
	"internet-monitor/config"
	"internet-monitor/core"
	"internet-monitor/dashboard"
	"internet-monitor/logger"
	"internet-monitor/monitor"
	"internet-monitor/tray"
	"internet-monitor/updater"
	"log"
	"os"
	"path/filepath"
	"time"

	webview "github.com/webview/webview_go"
)

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

	dash := dashboard.NewServer(cfg.DashboardPort, "config.json", cfg.LogDir, Version, lgr)
	dash.OnTestNotification = TestNotification
	dash.OnTestWebhook = lgr.SendTestWebhook
	dash.OnApplyUpdate = updater.Apply
	dash.OnRestartApp = updater.Restart
	dash.SetNativeNotifications(true) // GUI build shows native notifications
	dash.Start()

	checker := monitor.NewChecker(cfg)
	engine := core.New(cfg, checker, lgr, Version)
	dash.OnConfigChange = engine.ApplyConfig
	engine.Notifier = core.MultiNotifier{
		dashboard.NewNotifier(dash),
		&guiNotifier{},
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

	w := webview.New(false)
	defer w.Destroy()
	w.SetTitle("مراقب الإنترنت — Internet Monitor")
	// HintNone: initial size only — no minimum or maximum constraints.
	w.SetSize(1100, 750, webview.HintNone)

	hwnd := uintptr(w.Window())
	w.Bind("nativeMinimizeToTray", func() { hideWindow(hwnd) })

	if png := dashboard.FaviconPNG(); len(png) > 0 {
		tray.SetCustomIcon(png)
	}
	// Set window icon from the favicon (no go-winres .syso in dev environment).
	setWindowIcon(hwnd)

	stopTray := initTray(w, hwnd)
	engine.Start()

	time.Sleep(150 * time.Millisecond)
	w.Navigate(dash.URL())
	w.Run()

	engine.Stop()
	stopTray()
	w.Destroy()
	os.Exit(0)
}
