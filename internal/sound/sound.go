package sound

import (
	"fmt"
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

var (
	cacheMu   sync.Mutex
	cacheKey  string
	cacheData []byte
)

func scaled(path string, volume int) ([]byte, string, bool) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, "", false
	}
	key := fmt.Sprintf("%s|%d|%d|%d", path, info.ModTime().UnixNano(), info.Size(), volume)
	cacheMu.Lock()
	defer cacheMu.Unlock()
	if key == cacheKey {
		return cacheData, key, cacheData != nil
	}
	var out []byte
	if info.Size() <= MaxScaleBytes {
		if data, err := os.ReadFile(path); err == nil {
			out, _ = ScaleWAV(data, volume)
		}
	}
	if out == nil {
		logf("[sound] cannot change the volume of %q; playing it unchanged", path)
	}
	cacheKey, cacheData = key, out
	return out, key, out != nil
}

func Play() { PlayAt(100) }
