package sound

import (
	"internet-monitor/internal/assets"
	"os"
	"path/filepath"
	"sync"
)

const CustomFileName = "notification.wav"

var (
	mu          sync.RWMutex
	appDir      string
	defaultPath string
	once        sync.Once
)

var Logf func(format string, args ...interface{})

func logf(format string, args ...interface{}) {
	if Logf != nil {
		Logf(format, args...)
	}
}

func SetDir(dir string) {
	mu.Lock()
	appDir = dir
	mu.Unlock()
}

func CustomPath() string {
	mu.RLock()
	defer mu.RUnlock()
	if appDir == "" {
		return ""
	}
	return filepath.Join(appDir, CustomFileName)
}

func RingtonePath() string {
	if custom := CustomPath(); custom != "" {
		if _, err := os.Stat(custom); err == nil {
			return custom
		}
	}
	once.Do(func() {
		data := assets.NotificationWAV()
		if len(data) == 0 {
			return
		}
		dir, err := os.MkdirTemp("", "internet-monitor-")
		if err != nil {
			return
		}
		path := filepath.Join(dir, "notification.wav")
		if os.WriteFile(path, data, 0o644) == nil {
			defaultPath = path
		}
	})
	return defaultPath
}
