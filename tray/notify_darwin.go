//go:build darwin

package tray

import (
	"fmt"
	"internet-monitor/notifytext"
	"internet-monitor/sound"
	"os/exec"
)

func Notify(title, message string) {
	sound.Play() // match Windows/GUI: ring on every notification (no overlap)
	script := fmt.Sprintf(`display notification "%s" with title "%s"`,
		notifytext.EscapeAppleScript(message), notifytext.EscapeAppleScript(title))
	exec.Command("osascript", "-e", script).Start()
}

func OpenURL(url string) {
	exec.Command("open", url).Start()
}

func OpenFolder(path string) {
	exec.Command("open", path).Start()
}
