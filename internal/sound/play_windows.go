//go:build windows

package sound

import (
	"sync"
	"syscall"
	"unsafe"
)

var (
	modWinmm      = syscall.NewLazyDLL("winmm.dll")
	procPlaySound = modWinmm.NewProc("PlaySoundW")

	playMu   sync.Mutex
	wavUTF16 *uint16
	wavFor   string
)

const (
	sndAsync    = 0x0001
	sndFilename = 0x00020000
)

func logf(format string, args ...interface{}) {
	if Logf != nil {
		Logf(format, args...)
	}
}

func Play() {
	path := RingtonePath()
	if path == "" {
		logf("[sound] Play: no chime path (embedded extract failed?)")
		return
	}
	playMu.Lock()
	defer playMu.Unlock()

	if wavFor != path {
		p, err := syscall.UTF16PtrFromString(path)
		if err != nil {
			logf("[sound] bad path %q: %v", path, err)
			return
		}
		wavUTF16 = p
		wavFor = path
	}

	ret, _, _ := procPlaySound.Call(
		uintptr(unsafe.Pointer(wavUTF16)),
		0,
		uintptr(sndAsync|sndFilename),
	)
	if ret == 0 {
		logf("[sound] PlaySound failed for %q", path)
		return
	}
	logf("[sound] PlaySound ok %q", path)
}
