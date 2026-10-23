package app

import (
	"context"
	"internet-monitor/internal/config"
	"internet-monitor/internal/core"
	"internet-monitor/internal/logger"
	"internet-monitor/internal/monitor"
	"internet-monitor/internal/notifytext"
	"internet-monitor/internal/server"
	"internet-monitor/internal/sound"
	"internet-monitor/internal/tray"
	"internet-monitor/internal/updater"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Options struct {
	Version          string
	TestNotification func(lang string)
}

type App struct {
	Dir     string
	Version string
	Store   *config.Store
	Log     *logger.Logger
	Server  *server.Server
	Engine  *core.Engine

	stop chan struct{}
	once sync.Once
}

func Dir() string {
	exe, err := os.Executable()
	if err != nil {
		wd, _ := os.Getwd()
		return wd
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	return filepath.Dir(exe)
}

func New(opts Options) (*App, error) {
	dir := Dir()
	if err := os.Chdir(dir); err != nil {
		log.Printf("warning: could not chdir to %s: %v", dir, err)
	}

	store, err := config.Open(filepath.Join(dir, "config.json"))
	if err != nil {
		log.Printf("config: %v", err)
	}
	cfg := store.Get()

	lgr, err := logger.New(cfg)
	if err != nil {
		return nil, err
	}
	if lerr := store.LoadError(); lerr != nil {
		lgr.AppLog("CONFIG %v", lerr)
	}
	lgr.AppLog("START version=%s dir=%s", opts.Version, dir)

	tray.Logf = lgr.AppLog
	sound.Logf = lgr.AppLog
	sound.SetDir(dir)
	notifytext.SetLang(cfg.Language)

	srv := server.New(server.Options{
		Store:               store,
		Logger:              lgr,
		Version:             opts.Version,
		AppDir:              dir,
		NativeNotifications: true,
		Hooks: server.Hooks{
			TestNotification: opts.TestNotification,
			PlaySound:        sound.Play,
			TestWebhook:      lgr.SendTestWebhook,
			CheckUpdate:      func() (*updater.Info, error) { return updater.Check(opts.Version) },
			ApplyUpdate:      updater.Apply,
			Restart:          updater.Restart,
		},
	})
	if err := srv.Start(cfg.DashboardPort); err != nil {
		lgr.AppLog("%v — falling back to a free port", err)
		if err := srv.Start(0); err != nil {
			return nil, err
		}
	}
	lgr.AppLog("dashboard listening on %s", srv.URL())

	engine := core.New(cfg, monitor.NewChecker(cfg), lgr, opts.Version)
	store.Subscribe(engine.ApplyConfig)

	return &App{
		Dir:     dir,
		Version: opts.Version,
		Store:   store,
		Log:     lgr,
		Server:  srv,
		Engine:  engine,
		stop:    make(chan struct{}),
	}, nil
}

func (a *App) Run(notifiers ...core.Notifier) {
	multi := core.MultiNotifier{server.NewNotifier(a.Server)}
	multi = append(multi, notifiers...)
	a.Engine.Notifier = multi
	a.Engine.Start()
	go a.updateLoop()
}

func (a *App) updateLoop() {
	select {
	case <-time.After(30 * time.Second):
	case <-a.stop:
		return
	}
	for {
		if info, err := updater.Check(a.Version); err == nil && info.HasUpdate {
			a.Log.AppLog("UPDATE available: %s (current: %s)", info.LatestVersion, info.CurrentVersion)
			a.Server.SetUpdateInfo(info)
		}
		select {
		case <-time.After(6 * time.Hour):
		case <-a.stop:
			return
		}
	}
}

func (a *App) Stop() {
	a.once.Do(func() {
		close(a.stop)
		a.Engine.Stop()
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := a.Server.Shutdown(ctx); err != nil {
			a.Log.AppLog("dashboard shutdown: %v", err)
		}
		a.Log.AppLog("STOP")
		_ = a.Log.Close()
		ReleaseSingleInstance()
	})
}
