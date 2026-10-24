//go:build windows

package tray

import (
	"fmt"
	"internet-monitor/internal/assets"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

const (
	notifyAUMID        = "InternetMonitor"
	notifyDisplayName  = "Internet Monitor"
	aumidRegistryRoot  = `SOFTWARE\Classes\AppUserModelId\`
	classXMLDocument   = "Windows.Data.Xml.Dom.XmlDocument"
	classToast         = "Windows.UI.Notifications.ToastNotification"
	classToastManager  = "Windows.UI.Notifications.ToastNotificationManager"
	methodLoadXML      = vtblFirstMethod
	methodCreateToast  = vtblFirstMethod
	methodCreateNotifW = vtblFirstMethod + 1
	methodShow         = vtblFirstMethod
	methodPutTag       = vtblFirstMethod
	methodPutGroup     = vtblFirstMethod + 2
	hrSOK              = 0
)

var (
	iidXMLDocumentIO         = mustGUID("{6CD0E74E-EE65-4489-9EBF-CA43E87BA637}")
	iidXMLDocument           = mustGUID("{F7F3A506-1E87-42D6-BCFB-B8C809FA5494}")
	iidToastNotificationFact = mustGUID("{04124B20-82C6-4229-B109-FD9ED4662B53}")
	iidToastManagerStatics   = mustGUID("{50AC103F-D235-4598-BBEF-98FE4D1A3AD4}")
	iidToastNotification2    = mustGUID("{9DFB9FD1-143A-490E-90BF-B9FBA7132DE7}")
	registerAppIDOnce        sync.Once
	registerAppIDErr         error
)

func toastCacheDir() string {
	base, err := os.UserCacheDir()
	if err != nil {
		return ""
	}
	dir := filepath.Join(base, "InternetMonitor")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return ""
	}
	return dir
}

func toastIconPath() string {
	dir := toastCacheDir()
	if dir == "" {
		return ""
	}
	path := filepath.Join(dir, "toast-icon.png")
	if err := os.WriteFile(path, assets.FaviconPNG(), 0o644); err != nil {
		return ""
	}
	return path
}

var (
	statusIconMu    sync.Mutex
	statusIconPaths = map[string]string{}
)

func statusIconPath(name string, r, g, b uint8) string {
	statusIconMu.Lock()
	defer statusIconMu.Unlock()
	if p, ok := statusIconPaths[name]; ok {
		return p
	}
	path := ""
	if data, ok := badgePNG(96, 64, r, g, b); ok {
		if dir := toastCacheDir(); dir != "" {
			candidate := filepath.Join(dir, "status-"+name+".png")
			if os.WriteFile(candidate, data, 0o644) == nil {
				path = candidate
			}
		}
	}
	statusIconPaths[name] = path
	return path
}

func registerAppID() error {
	registerAppIDOnce.Do(func() {
		k, _, err := registry.CreateKey(registry.CURRENT_USER, aumidRegistryRoot+notifyAUMID, registry.SET_VALUE)
		if err != nil {
			registerAppIDErr = fmt.Errorf("create AUMID key: %w", err)
			return
		}
		defer k.Close()
		if err := k.SetStringValue("DisplayName", notifyDisplayName); err != nil {
			registerAppIDErr = fmt.Errorf("set DisplayName: %w", err)
			return
		}
		if icon := toastIconPath(); icon != "" {
			if err := k.SetStringValue("IconUri", icon); err != nil && Logf != nil {
				Logf("[notify] set IconUri: %v", err)
			}
		}
	})
	return registerAppIDErr
}

func showToast(c ToastContent, tag, group string) error {
	if err := registerAppID(); err != nil {
		return err
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if err := windows.CoInitializeEx(0, windows.COINIT_MULTITHREADED); err == nil || err == syscall.Errno(windows.S_FALSE) {
		defer windows.CoUninitialize()
	}

	doc, err := roActivateInstance(classXMLDocument)
	if err != nil {
		return err
	}
	defer doc.release()

	docIO, err := doc.queryInterface(&iidXMLDocumentIO)
	if err != nil {
		return err
	}
	defer docIO.release()

	xmlText, err := newHString(BuildToastXML(c))
	if err != nil {
		return err
	}
	defer xmlText.release()
	if hr := docIO.call(methodLoadXML, uintptr(xmlText)); hr != hrSOK {
		return hresultError("IXmlDocumentIO.LoadXml", hr)
	}

	xmlDoc, err := doc.queryInterface(&iidXMLDocument)
	if err != nil {
		return err
	}
	defer xmlDoc.release()

	toastFactory, err := roGetActivationFactory(classToast, &iidToastNotificationFact)
	if err != nil {
		return err
	}
	defer toastFactory.release()

	toastPtr, hr := withOut(func(slot uintptr) uintptr {
		return toastFactory.call(methodCreateToast, uintptr(xmlDoc.ptr), slot)
	})
	if hr != hrSOK {
		return hresultError("CreateToastNotification", hr)
	}
	toast := comObject{ptr: toastPtr}
	defer toast.release()
	if tag != "" {
		if err := tagToast(toast, tag, group); err != nil && Logf != nil {
			Logf("[notify] toast tag skipped: %v", err)
		}
	}

	manager, err := roGetActivationFactory(classToastManager, &iidToastManagerStatics)
	if err != nil {
		return err
	}
	defer manager.release()

	appID, err := newHString(notifyAUMID)
	if err != nil {
		return err
	}
	defer appID.release()

	notifierPtr, hr := withOut(func(slot uintptr) uintptr {
		return manager.call(methodCreateNotifW, uintptr(appID), slot)
	})
	if hr != hrSOK {
		return hresultError("CreateToastNotifierWithId", hr)
	}
	notifier := comObject{ptr: notifierPtr}
	defer notifier.release()

	if hr := notifier.call(methodShow, uintptr(toast.ptr)); hr != hrSOK {
		return hresultError("IToastNotifier.Show", hr)
	}
	return nil
}

func tagToast(toast comObject, tag, group string) error {
	t2, err := toast.queryInterface(&iidToastNotification2)
	if err != nil {
		return err
	}
	defer t2.release()
	h, err := newHString(tag)
	if err != nil {
		return err
	}
	defer h.release()
	if hr := t2.call(methodPutTag, uintptr(h)); hr != hrSOK {
		return hresultError("IToastNotification2.put_Tag", hr)
	}
	if group == "" {
		return nil
	}
	g, err := newHString(group)
	if err != nil {
		return err
	}
	defer g.release()
	if hr := t2.call(methodPutGroup, uintptr(g)); hr != hrSOK {
		return hresultError("IToastNotification2.put_Group", hr)
	}
	return nil
}
