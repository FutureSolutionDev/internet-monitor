package main

import (
	"fmt"
	"internet-monitor/internal/app"
	"internet-monitor/internal/assets"
	"internet-monitor/internal/config"
	"internet-monitor/internal/notifytext"
	"internet-monitor/internal/tray"
	"log"
	"os"
	"path/filepath"

	"fyne.io/systray"
)

var Version = "dev"

func main() {
	if !app.AcquireSingleInstance() {
		openRunningDashboard()
		return
	}

	a, err := app.New(app.Options{
		Version: Version,
		TestNotification: func(lang string) {
			title, body := notifytext.TestMessage(lang)
			tray.Notify(title, body)
		},
	})
	if err != nil {
		log.Fatalf("startup failed: %v", err)
	}

	tray.SetCustomIcon(assets.FaviconPNG())
	t := tray.New(filepath.Join(a.Dir, a.Store.Get().LogDir), a.Server.URL())
	a.Run(tray.NewNotifier(t))
	systray.Run(t.OnReady, t.OnExit)
	a.Stop()
	os.Exit(0)
}

func openRunningDashboard() {
	cfg := config.Default()
	if data, err := os.ReadFile(filepath.Join(app.Dir(), "config.json")); err == nil {
		if parsed, perr := config.Parse(data); perr == nil {
			cfg = parsed
		}
	}
	tray.OpenURL(fmt.Sprintf("http://localhost:%d", cfg.DashboardPort))
}
