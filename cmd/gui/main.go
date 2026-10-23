//go:generate go-winres simply --arch amd64 --icon ../../internal/assets/favicon.png

package main

import (
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

	a, err := app.New(app.Options{Version: Version, TestNotification: TestNotification})
	if err != nil {
		log.Fatalf("startup failed: %v", err)
	}

	w := webview.New(false)
	defer w.Destroy()
	w.SetTitle("Internet Monitor")
	w.SetSize(1180, 780, webview.HintNone)

	hwnd := uintptr(w.Window())
	if err := w.Bind("nativeMinimizeToTray", func() { hideWindow(hwnd) }); err != nil {
		a.Log.AppLog("webview bind failed: %v", err)
	}

	tray.SetCustomIcon(assets.FaviconPNG())
	setWindowIcon(hwnd)

	stopTray := initTray(w, hwnd)
	a.Run(&guiNotifier{})

	w.Navigate(a.Server.URL())
	w.Run()

	a.Stop()
	stopTray()
	os.Exit(0)
}
