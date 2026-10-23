//go:build linux

package tray

import (
	"internet-monitor/internal/sound"
	"os/exec"
)

func Notify(title, message string) {
	sound.Play()
	exec.Command("notify-send", "-a", "Internet Monitor", title, message).Start()
}

func OpenURL(url string) {
	exec.Command("xdg-open", url).Start()
}

func OpenFolder(path string) {
	exec.Command("xdg-open", path).Start()
}
