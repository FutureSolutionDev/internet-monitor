//go:build windows

package sound

import (
	"sync"
	"syscall"
	"time"
	"unsafe"
)

var (
	modWinmm = syscall.NewLazyDLL("winmm.dll")
	procMci  = modWinmm.NewProc("mciSendStringW")

	playMu  sync.Mutex
	playing bool
)

func mci(cmd string) {
	p, _ := syscall.UTF16PtrFromString(cmd)
	procMci.Call(uintptr(unsafe.Pointer(p)), 0, 0, 0)
}

// Play plays the ringtone (if any) via MCI for ~15s, stopping any prior play
// first. Synchronous: callers that don't want to block should use `go Play()`.
func Play() {
	path := RingtonePath()
	if path == "" {
		return
	}
	playMu.Lock()
	if playing {
		mci("stop im_ring")
		mci("close im_ring")
	}
	playing = true
	playMu.Unlock()

	mci(`open "` + path + `" type mpegvideo alias im_ring`)
	mci("play im_ring")
	time.Sleep(15 * time.Second)
	mci("stop im_ring")
	mci("close im_ring")

	playMu.Lock()
	playing = false
	playMu.Unlock()
}
