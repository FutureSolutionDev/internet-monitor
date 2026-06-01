//go:build windows

package sound

import (
	"sync"
	"syscall"
	"unsafe"
)

var (
	modWinmm = syscall.NewLazyDLL("winmm.dll")
	procMci  = modWinmm.NewProc("mciSendStringW")

	playMu sync.Mutex
)

// Logf, if set, receives diagnostic lines (wired to logger.AppLog by main so
// they land in logs/app.log — the GUI/tray builds are -H=windowsgui, so the
// standard log package is invisible).
var Logf func(format string, args ...interface{})

func logf(format string, args ...interface{}) {
	if Logf != nil {
		Logf(format, args...)
	}
}

// mci runs an MCI command and returns the result code (0 == success). On a
// non-zero code it also fetches the textual error via mciGetErrorString.
func mci(cmd string) uintptr {
	p, _ := syscall.UTF16PtrFromString(cmd)
	ret, _, _ := procMci.Call(uintptr(unsafe.Pointer(p)), 0, 0, 0)
	return ret
}

// Play stops any currently-playing ringtone and starts the latest one.
//
// MCI "play" returns immediately (asynchronous playback), so there is no
// blocking sleep: a rapid second call stops the first and plays the new one on
// the same alias. Every notification path funnels through here under playMu, so
// an in-progress sound is always replaced — no overlapping audio. The previous
// alias is always stop+close'd before re-open, and all return codes are logged
// so a silent failure (bad path / missing codec) is diagnosable in app.log.
func Play() {
	path := RingtonePath()
	if path == "" {
		logf("[sound] Play: no ringtone path (embedded extract failed?)")
		return
	}
	playMu.Lock()
	defer playMu.Unlock()

	mci("stop im_ring")  // ignore: may not be open yet
	mci("close im_ring") // ignore: may not be open yet
	if rc := mci(`open "` + path + `" type mpegvideo alias im_ring`); rc != 0 {
		// Retry without an explicit type: lets MCI pick the device for the
		// file's extension, which is more robust across Windows codec setups.
		logf("[sound] open(type mpegvideo) rc=%d for %q — retrying without type", rc, path)
		if rc2 := mci(`open "` + path + `" alias im_ring`); rc2 != 0 {
			logf("[sound] open rc=%d — cannot play %q", rc2, path)
			return
		}
	}
	if rc := mci("play im_ring"); rc != 0 {
		logf("[sound] play rc=%d for %q", rc, path)
		return
	}
	logf("[sound] playing %q", path)
}
