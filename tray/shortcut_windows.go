//go:build windows

package tray

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

// ensureStartMenuShortcut creates (once) a Start Menu shortcut for the app
// with the AppUserModelID property set. Windows 10/11 requires this for any
// non-packaged desktop app to show notification banner popups. Without the
// shortcut, balloon tips are silently redirected to Action Center only.
//
// The shortcut is written to:
//
//	%APPDATA%\Microsoft\Windows\Start Menu\Programs\Internet Monitor.lnk
//
// It is created using PowerShell + inline C# (Add-Type) to call the
// IShellLink + IPropertyStore COM interfaces. Runs once; subsequent calls
// return immediately when the file already exists.
// EnsureStartMenuShortcut is exported so cmd/gui can call it too.
func EnsureStartMenuShortcut(aumid string) {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		return
	}
	lnkPath := filepath.Join(appData,
		"Microsoft", "Windows", "Start Menu", "Programs",
		"Internet Monitor.lnk")

	if _, err := os.Stat(lnkPath); err == nil {
		return // shortcut already exists
	}

	exe, err := os.Executable()
	if err != nil {
		return
	}
	exe, _ = filepath.EvalSymlinks(exe)

	// Inline C# compiled at runtime via Add-Type.
	// IShellLinkW has exactly 18 vtable methods — all must be declared in order.
	// IPropertyStore methods: GetCount, GetAt, GetValue, SetValue, Commit.
	script := fmt.Sprintf(`
$lnk = '%s'; $exe = '%s'; $id = '%s'

Add-Type -TypeDefinition @"
using System;
using System.Runtime.InteropServices;
using System.Runtime.InteropServices.ComTypes;

[ComImport, Guid("000214F9-0000-0000-C000-000000000046"),
 InterfaceType(ComInterfaceType.InterfaceIsIUnknown)]
public interface IShellLinkW {
    void GetPath(System.Text.StringBuilder f, int c, IntPtr d, int g);
    void GetIDList(out IntPtr p);    void SetIDList(IntPtr p);
    void GetDescription(System.Text.StringBuilder n, int c);
    void SetDescription(string n);
    void GetWorkingDirectory(System.Text.StringBuilder d, int c);
    void SetWorkingDirectory(string d);
    void GetArguments(System.Text.StringBuilder a, int c);
    void SetArguments(string a);
    void GetHotkey(out short h);    void SetHotkey(short h);
    void GetShowCmd(out int s);     void SetShowCmd(int s);
    void GetIconLocation(System.Text.StringBuilder p, int c, out int i);
    void SetIconLocation(string p, int i);
    void SetRelativePath(string p, int r);
    void Resolve(IntPtr h, int f);
    void SetPath(string p);
}
[ComImport, Guid("00021401-0000-0000-C000-000000000046")]
public class ShellLink {}

[ComImport, Guid("886D8EEB-8CF2-4446-8D02-CDBA1DBDCF99"),
 InterfaceType(ComInterfaceType.InterfaceIsIUnknown)]
public interface IPropertyStore {
    int GetCount(out uint n);
    int GetAt(uint i, out PropKey k);
    int GetValue(ref PropKey k, out object v);
    [PreserveSig] int SetValue(ref PropKey k, ref PropVariant v);
    [PreserveSig] int Commit();
}
[StructLayout(LayoutKind.Sequential, Pack=4)]
public struct PropKey { public Guid fmtid; public uint pid; }
[StructLayout(LayoutKind.Explicit)]
public struct PropVariant {
    [FieldOffset(0)] public ushort vt;
    [FieldOffset(8)] public IntPtr ptr;
}
"@ -ErrorAction SilentlyContinue

try {
    $sl  = New-Object ShellLink
    $isl = [IShellLinkW]$sl
    $isl.SetPath($exe)
    $ps  = [IPropertyStore]$sl
    $pk  = New-Object PropKey
    $pk.fmtid = [Guid]"9F4C2855-9F79-4B39-A8D0-E1D42DE1D5F3"
    $pk.pid   = 5
    $pv       = New-Object PropVariant
    $pv.vt    = 31   # VT_LPWSTR
    $pv.ptr   = [Runtime.InteropServices.Marshal]::StringToCoTaskMemUni($id)
    $ps.SetValue([ref]$pk, [ref]$pv) | Out-Null
    $ps.Commit() | Out-Null
    [Runtime.InteropServices.Marshal]::FreeCoTaskMem($pv.ptr)
    ([IPersistFile]$sl).Save($lnk, $false)
    Write-Host "[notify] Start Menu shortcut created: $lnk"
} catch {
    Write-Host "[notify] shortcut creation failed: $_"
}
`, lnkPath, exe, aumid)

	cmd := exec.Command("powershell",
		"-WindowStyle", "Hidden",
		"-NonInteractive",
		"-Command", script)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	out, err := cmd.Output()
	if err != nil {
		log.Printf("[notify] ensureStartMenuShortcut: powershell error: %v", err)
	}
	if len(out) > 0 {
		log.Printf("[notify] %s", out)
	}
}
