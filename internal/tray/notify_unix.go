//go:build linux || darwin

package tray

import (
	"fmt"
	"internet-monitor/internal/notifytext"
	"internet-monitor/internal/sound"
	"os"
	"os/exec"
	"runtime"
)

func logf(format string, args ...interface{}) {
	if Logf != nil {
		Logf(format, args...)
	}
}

func start(name string, args ...string) {
	cmd := exec.Command(name, args...)
	if err := cmd.Start(); err != nil {
		logf("[notify] %s failed: %v", name, err)
		return
	}
	go func() { _ = cmd.Wait() }()
}

func opener() string {
	if runtime.GOOS == "darwin" {
		return "open"
	}
	return "xdg-open"
}

func Notify(title, message string) {
	sound.Play()
	if runtime.GOOS == "darwin" {
		start("osascript", "-e", fmt.Sprintf(`display notification "%s" with title "%s"`,
			notifytext.EscapeAppleScript(message), notifytext.EscapeAppleScript(title)))
		return
	}
	start("notify-send", "-a", "Internet Monitor", title, message)
}

func OpenURL(url string) {
	if _, err := ValidateWebURL(url); err != nil {
		logf("[notify] OpenURL rejected: %v", err)
		return
	}
	start(opener(), url)
}

func OpenFolder(path string) {
	if info, err := os.Stat(path); err != nil || !info.IsDir() {
		logf("[notify] OpenFolder rejected: %q", path)
		return
	}
	start(opener(), path)
}
