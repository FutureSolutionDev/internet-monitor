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

func mci(cmd string) {
	p, _ := syscall.UTF16PtrFromString(cmd)
	procMci.Call(uintptr(unsafe.Pointer(p)), 0, 0, 0)
}

// Play stops any currently-playing ringtone and starts the latest one.
//
// MCI "play" returns immediately (asynchronous playback), so there is no
// blocking sleep: a rapid second call simply stops the first and plays the new
// one on the same alias. Every notification path funnels through here, so an
// in-progress sound is always replaced — no overlapping audio.
func Play() {
	path := RingtonePath()
	if path == "" {
		return
	}
	playMu.Lock()
	defer playMu.Unlock()

	// Kill whatever is currently playing, then start the latest.
	mci("stop im_ring")
	mci("close im_ring")
	mci(`open "` + path + `" type mpegvideo alias im_ring`)
	mci("play im_ring")
}
