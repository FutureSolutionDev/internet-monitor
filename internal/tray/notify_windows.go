//go:build windows

package tray

import (
	"log"
	"os"

	"golang.org/x/sys/windows"
)

func notifyLogf(format string, args ...interface{}) {
	if Logf != nil {
		Logf(format, args...)
	} else {
		log.Printf(format, args...)
	}
}

func Notify(title, message string) {
	playTraySound()
	ShowNotification(title, message)
}

func ShowNotification(title, message string) {
	if err := showToast(title, message); err != nil {
		notifyLogf("[notify] native toast failed: %v", err)
		return
	}
	notifyLogf("[notify] native toast shown: title=%q", title)
}

func shellOpen(target string) error {
	verb, err := windows.UTF16PtrFromString("open")
	if err != nil {
		return err
	}
	file, err := windows.UTF16PtrFromString(target)
	if err != nil {
		return err
	}
	return windows.ShellExecute(0, verb, file, nil, nil, windows.SW_SHOWNORMAL)
}

func OpenURL(url string) {
	valid, err := ValidateWebURL(url)
	if err != nil {
		notifyLogf("[notify] OpenURL rejected %q: %v", url, err)
		return
	}
	if err := shellOpen(valid); err != nil {
		notifyLogf("[notify] OpenURL failed: %v", err)
	}
}

func OpenFolder(path string) {
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		notifyLogf("[notify] OpenFolder rejected %q: not an existing directory", path)
		return
	}
	if err := shellOpen(path); err != nil {
		notifyLogf("[notify] OpenFolder failed: %v", err)
	}
}
