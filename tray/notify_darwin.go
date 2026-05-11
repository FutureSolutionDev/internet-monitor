//go:build darwin

package tray

import (
	"fmt"
	"os/exec"
)

func Notify(title, message string) {
	script := fmt.Sprintf(`display notification "%s" with title "%s"`, message, title)
	exec.Command("osascript", "-e", script).Start()
}

func OpenURL(url string) {
	exec.Command("open", url).Start()
}

func OpenFolder(path string) {
	exec.Command("open", path).Start()
}
