//go:build linux

package tray

import (
	"internet-monitor/sound"
	"os/exec"
)

func Notify(title, message string) {
	sound.Play() // match Windows/GUI: ring on every notification (no overlap)
	exec.Command("notify-send", "-a", "Internet Monitor", title, message).Start()
}

func OpenURL(url string) {
	exec.Command("xdg-open", url).Start()
}

func OpenFolder(path string) {
	exec.Command("xdg-open", path).Start()
}
