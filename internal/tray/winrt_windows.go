//go:build windows

package tray

import (
	"fmt"
	"sync"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	vtblQueryInterface = 0
	vtblRelease        = 2
	vtblFirstMethod    = 6
	vtblMaxEntries     = 32
)

var (
	combase                    = windows.NewLazySystemDLL("combase.dll")
	procRoActivateInstance     = combase.NewProc("RoActivateInstance")
	procRoGetActivationFactory = combase.NewProc("RoGetActivationFactory")
	procWindowsCreateString    = combase.NewProc("WindowsCreateString")
	procWindowsDeleteString    = combase.NewProc("WindowsDeleteString")
)

type hstring uintptr

var (
	outMu   sync.Mutex
	outSlot unsafe.Pointer
)

func withOut(fn func(slot uintptr) uintptr) (unsafe.Pointer, uintptr) {
	outMu.Lock()
	defer outMu.Unlock()
	outSlot = nil
	hr := fn(uintptr(unsafe.Pointer(&outSlot)))
	return outSlot, hr
}

type comObject struct {
	ptr unsafe.Pointer
}

func mustGUID(s string) windows.GUID {
	g, err := windows.GUIDFromString(s)
	if err != nil {
		panic(err)
	}
	return g
}

func hresultError(op string, hr uintptr) error {
	return fmt.Errorf("%s failed: HRESULT 0x%08X", op, uint32(hr))
}

func newHString(s string) (hstring, error) {
	u16, err := windows.UTF16FromString(s)
	if err != nil {
		return 0, err
	}
	var h hstring
	var first *uint16
	if len(u16) > 1 {
		first = &u16[0]
	}
	hr, _, _ := procWindowsCreateString.Call(
		uintptr(unsafe.Pointer(first)),
		uintptr(len(u16)-1),
		uintptr(unsafe.Pointer(&h)),
	)
	if hr != 0 {
		return 0, hresultError("WindowsCreateString", hr)
	}
	return h, nil
}

func (h hstring) release() {
	if h != 0 {
		procWindowsDeleteString.Call(uintptr(h))
	}
}

func (o comObject) valid() bool {
	return o.ptr != nil
}

func (o comObject) method(index int) uintptr {
	vtbl := *(**[vtblMaxEntries]uintptr)(o.ptr)
	return vtbl[index]
}

func (o comObject) call(index int, args ...uintptr) uintptr {
	all := make([]uintptr, 0, len(args)+1)
	all = append(all, uintptr(o.ptr))
	all = append(all, args...)
	hr, _, _ := syscall.SyscallN(o.method(index), all...)
	return hr
}

func (o comObject) release() {
	if o.valid() {
		o.call(vtblRelease)
	}
}

func (o comObject) queryInterface(iid *windows.GUID) (comObject, error) {
	out, hr := withOut(func(slot uintptr) uintptr {
		return o.call(vtblQueryInterface, uintptr(unsafe.Pointer(iid)), slot)
	})
	if hr != 0 {
		return comObject{}, hresultError("QueryInterface", hr)
	}
	return comObject{ptr: out}, nil
}

func roActivateInstance(class string) (comObject, error) {
	h, err := newHString(class)
	if err != nil {
		return comObject{}, err
	}
	defer h.release()
	out, hr := withOut(func(slot uintptr) uintptr {
		r, _, _ := procRoActivateInstance.Call(uintptr(h), slot)
		return r
	})
	if hr != 0 {
		return comObject{}, hresultError("RoActivateInstance("+class+")", hr)
	}
	return comObject{ptr: out}, nil
}

func roGetActivationFactory(class string, iid *windows.GUID) (comObject, error) {
	h, err := newHString(class)
	if err != nil {
		return comObject{}, err
	}
	defer h.release()
	out, hr := withOut(func(slot uintptr) uintptr {
		r, _, _ := procRoGetActivationFactory.Call(uintptr(h), uintptr(unsafe.Pointer(iid)), slot)
		return r
	})
	if hr != 0 {
		return comObject{}, hresultError("RoGetActivationFactory("+class+")", hr)
	}
	return comObject{ptr: out}, nil
}
