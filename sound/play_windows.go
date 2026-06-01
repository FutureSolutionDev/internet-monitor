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
	wavUTF16 *uint16 // cached UTF-16 path of the chime
	wavFor   string  // which path wavUTF16 was built for
)

const (
	sndAsync    = 0x0001     // play asynchronously (return immediately)
	sndFilename = 0x00020000 // lpName is a filename
)

// Logf, if set, receives diagnostic lines (wired to logger.AppLog by main so
// they land in logs/app.log — GUI/tray builds are -H=windowsgui, so the
// standard log package is invisible).
var Logf func(format string, args ...interface{})

func logf(format string, args ...interface{}) {
	if Logf != nil {
		Logf(format, args...)
	}
}

// Play plays the notification chime, replacing any in-progress one.
//
// winmm PlaySound(SND_ASYNC|SND_FILENAME) inherently stops whatever it was
// previously playing and starts the new sound — there is a single playback
// channel per process, so a rapid second call cannot overlap the first. No
// device/alias bookkeeping (unlike MCI), so there's no desync to track:
// "latest wins, never overlaps" is built in.
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
