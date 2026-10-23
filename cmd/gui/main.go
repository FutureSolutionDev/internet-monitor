//go:generate go-winres simply --arch amd64 --icon ../../internal/dashboard/assets/favicon.png

package main

import (
	"context"
	"internet-monitor/internal/config"
	"internet-monitor/internal/core"
	"internet-monitor/internal/dashboard"
	"internet-monitor/internal/logger"
	"internet-monitor/internal/monitor"
	"internet-monitor/internal/notifytext"
	"internet-monitor/internal/sound"
	"internet-monitor/internal/tray"
	"internet-monitor/internal/updater"
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

	tray.Logf = lgr.AppLog
	sound.Logf = lgr.AppLog
	notifytext.SetLang(cfg.Language)

	dash := dashboard.NewServer(cfg.DashboardPort, "config.json", cfg.LogDir, Version, lgr)
	dash.OnTestNotification = TestNotification
	dash.OnTestWebhook = lgr.SendTestWebhook
	dash.OnApplyUpdate = updater.Apply
	dash.OnRestartApp = updater.Restart
	dash.OnCheckUpdate = func() (*dashboard.UpdateInfo, error) {
		info, err := updater.Check(Version)
		if err != nil {
			return nil, err
		}
		return &dashboard.UpdateInfo{
			HasUpdate:      info.HasUpdate,
			LatestVersion:  info.LatestVersion,
			CurrentVersion: info.CurrentVersion,
			DownloadURL:    info.DownloadURL,
			ReleaseNotes:   info.ReleaseNotes,
		}, nil
	}
	dash.OnPlaySound = sound.Play
	dash.SetNativeNotifications(true)
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

	w.SetSize(1100, 750, webview.HintNone)

	hwnd := uintptr(w.Window())
	w.Bind("nativeMinimizeToTray", func() { hideWindow(hwnd) })

	if png := dashboard.FaviconPNG(); len(png) > 0 {
		tray.SetCustomIcon(png)
	}

	setWindowIcon(hwnd)

	stopTray := initTray(w, hwnd)
	engine.Start()

	time.Sleep(150 * time.Millisecond)
	w.Navigate(dash.URL())
	w.Run()

	engine.Stop()
	stopTray()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	if err := dash.Shutdown(shutdownCtx); err != nil {
		log.Printf("dashboard shutdown error: %v", err)
	}
	cancel()
	w.Destroy()
	os.Exit(0)
}
