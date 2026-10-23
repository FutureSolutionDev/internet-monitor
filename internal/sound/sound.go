package sound

import (
	"internet-monitor/internal/dashboard"
	"os"
	"path/filepath"
	"sync"
)

var (
	defaultPath string
	once        sync.Once
)

var Logf func(format string, args ...interface{})

func RingtonePath() string {
	if wd, err := os.Getwd(); err == nil {
		custom := filepath.Join(wd, "notification.wav")
		if _, err := os.Stat(custom); err == nil {
			return custom
		}
	}
	once.Do(func() {
		data := dashboard.NotificationWav()
		if len(data) == 0 {
			return
		}
		dir, err := os.MkdirTemp("", "internet-monitor-")
		if err != nil {
			return
		}
		path := filepath.Join(dir, "notification.wav")
		if os.WriteFile(path, data, 0644) == nil {
			defaultPath = path
		}
	})
	return defaultPath
}
