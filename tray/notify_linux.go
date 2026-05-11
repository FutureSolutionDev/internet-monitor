//go:build linux

package tray

import "os/exec"

func Notify(title, message string) {
	exec.Command("notify-send", "-a", "Internet Monitor", title, message).Start()
}

func OpenURL(url string) {
	exec.Command("xdg-open", url).Start()
}

func OpenFolder(path string) {
	exec.Command("xdg-open", path).Start()
}
