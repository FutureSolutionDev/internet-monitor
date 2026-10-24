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
	memWAV   []byte
)

const (
	sndAsync    = 0x0001
	sndMemory   = 0x0004
	sndFilename = 0x00020000
)

func PlayAt(volume int) {
	if volume <= 0 {
		return
	}
	path := RingtonePath()
	if path == "" {
		logf("[sound] Play: no chime path (embedded extract failed?)")
		return
	}
	if volume < 100 {
		if data, _, ok := scaled(path, volume); ok && len(data) > 0 {
			playMemory(data)
			return
		}
	}
	playFile(path)
}

func playMemory(data []byte) {
	if len(data) == 0 {
		return
	}
	playMu.Lock()
	defer playMu.Unlock()
	procPlaySound.Call(0, 0, 0)
	memWAV = data
	ret, _, _ := procPlaySound.Call(uintptr(unsafe.Pointer(&memWAV[0])), 0, uintptr(sndAsync|sndMemory))
	if ret == 0 {
		logf("[sound] PlaySound from memory failed")
	}
}

func playFile(path string) {
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
	}
}
