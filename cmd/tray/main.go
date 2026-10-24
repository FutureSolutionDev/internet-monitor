package main

import (
	"internet-monitor/internal/alerts"
	"internet-monitor/internal/app"
	"internet-monitor/internal/assets"
	"internet-monitor/internal/server"
	"internet-monitor/internal/tray"
	"log"
	"os"

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
		Desktop: func(launchURL func() string) alerts.Desktop { return tray.Desktop{ActionURL: launchURL} },
	})
	if err != nil {
		log.Fatalf("startup failed: %v", err)
	}

	tray.SetCustomIcon(assets.FaviconPNG())
	t := tray.New(a.LogDir(), a.Server.LaunchURL())
	a.Run(tray.NewNotifier(t))
	systray.Run(t.OnReady, t.OnExit)
	a.Stop()
	os.Exit(0)
}

func openRunningDashboard() {
	if info, ok := server.ReadSession(app.Dir()); ok {
		tray.OpenURL(info.LaunchURL())
	}
}
