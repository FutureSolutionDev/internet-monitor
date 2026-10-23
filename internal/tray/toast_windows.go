//go:build windows

package tray

import (
	"fmt"
	"os"
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
	hrSOK              = 0
)

var (
	iidXMLDocumentIO         = mustGUID("{6CD0E74E-EE65-4489-9EBF-CA43E87BA637}")
	iidXMLDocument           = mustGUID("{F7F3A506-1E87-42D6-BCFB-B8C809FA5494}")
	iidToastNotificationFact = mustGUID("{04124B20-82C6-4229-B109-FD9ED4662B53}")
	iidToastManagerStatics   = mustGUID("{50AC103F-D235-4598-BBEF-98FE4D1A3AD4}")
	registerAppIDOnce        sync.Once
	registerAppIDErr         error
)

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
		if exe, err := os.Executable(); err == nil {
			if err := k.SetStringValue("IconUri", exe); err != nil {
				registerAppIDErr = fmt.Errorf("set IconUri: %w", err)
			}
		}
	})
	return registerAppIDErr
}

func showToast(title, body string) error {
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

	xmlText, err := newHString(BuildToastXML(title, body))
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
