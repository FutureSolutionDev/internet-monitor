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

func main() {
	if exePath, err := os.Executable(); err == nil {
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

	dash := dashboard.NewServer(cfg.DashboardPort, "config.json", cfg.LogDir)
	dash.Start()

	checker := monitor.NewChecker(cfg)
	t := tray.New(cfg, checker, lgr, dash.URL())
	t.OnTick  = dash.UpdateTick
	t.OnEvent = dash.AddEvent

	systray.Run(t.OnReady, t.OnExit)
}
