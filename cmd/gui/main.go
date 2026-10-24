//go:generate go-winres simply --arch amd64 --icon ../../internal/assets/favicon.png

package main

import (
	"internet-monitor/internal/alerts"
	"internet-monitor/internal/app"
	"internet-monitor/internal/assets"
	"internet-monitor/internal/tray"
	"log"
	"os"

	webview "github.com/webview/webview_go"
)

var Version = "dev"

func main() {
	if !app.AcquireSingleInstance() {
		return
	}

	a, err := app.New(app.Options{
		Version: Version,
		Desktop: func(func() string) alerts.Desktop { return tray.Desktop{} },
	})
	if err != nil {
		log.Fatalf("startup failed: %v", err)
	}

	w := webview.New(false)
	if w == nil {
		a.Log.AppLog("webview unavailable; opening the dashboard in the browser instead")
		tray.OpenURL(a.Server.LaunchURL())
		a.Run()
		select {}
	}
	w.SetTitle("Internet Monitor")
	w.SetSize(1180, 780, webview.HintNone)

	hwnd := uintptr(w.Window())
	if err := w.Bind("nativeMinimizeToTray", func() { hideWindow(hwnd) }); err != nil {
		a.Log.AppLog("webview bind failed: %v", err)
	}
	if err := w.Bind("nativeOpenURL", func(raw string) {
		u, err := tray.ValidateExternalURL(raw)
		if err != nil {
			a.Log.AppLog("open external link rejected %q: %v", raw, err)
			return
		}
		tray.OpenURL(u)
	}); err != nil {
		a.Log.AppLog("webview bind failed: %v", err)
	}

	tray.SetCustomIcon(assets.FaviconPNG())
	setWindowIcon(hwnd)

	stopTray := initTray(w, hwnd)
	a.Run(&guiNotifier{})

	w.Navigate(a.Server.LaunchURL())
	w.Run()

	a.Stop()
	stopTray()
	w.Destroy()
	os.Exit(0)
}
