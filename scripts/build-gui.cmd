@echo off
setlocal EnableDelayedExpansion
cd /d "%~dp0.."

echo ============================================
echo  Internet Monitor GUI - Build
echo ============================================
echo.

rem ── TDM-GCC مطلوب لـ Go CGO — POSIX threads لا تعمل ────────
rem    TDM-GCC يستخدم win32 threads وهو الوحيد المضمون مع Go

:find_toolchain

rem أولوية 1: TDM-GCC (الأفضل لـ Go CGO)
if exist "C:\TDM-GCC-64\bin\g++.exe" (
    set "PATH=C:\TDM-GCC-64\bin;%PATH%"
    set "CC=C:\TDM-GCC-64\bin\gcc.exe"
    set "CXX=C:\TDM-GCC-64\bin\g++.exe"
    goto :verify
)

rem أولوية 2: MSYS2 ucrt64
if exist "C:\msys64\ucrt64\bin\g++.exe" (
    set "PATH=C:\msys64\ucrt64\bin;%PATH%"
    set "CC=C:\msys64\ucrt64\bin\gcc.exe"
    set "CXX=C:\msys64\ucrt64\bin\g++.exe"
    goto :verify
)

rem أولوية 3: MSYS2 mingw64
if exist "C:\msys64\mingw64\bin\g++.exe" (
    set "PATH=C:\msys64\mingw64\bin;%PATH%"
    set "CC=C:\msys64\mingw64\bin\gcc.exe"
    set "CXX=C:\msys64\mingw64\bin\g++.exe"
    goto :verify
)

rem فحص الـ PATH الحالي — لكن تجاهل POSIX threads
where g++ >nul 2>&1
if %errorlevel% equ 0 (
    for /f "tokens=*" %%G in ('g++ -v 2^>^&1') do (
        echo %%G | findstr /i "posix" >nul
        if !errorlevel! equ 0 (
            echo [WARN] الـ GCC الموجود يستخدم POSIX threads وهو غير متوافق مع Go CGO.
            echo        سيتم تحميل TDM-GCC...
            goto :install
        )
    )
    rem POSIX مش موجود — جرّب البناء بالـ GCC الحالي
    goto :verify
)

:install
echo [!] تثبيت TDM-GCC تلقائياً...
echo.

rem --- MSYS2 via winget ---
echo [1/3] Trying winget MSYS2...
winget install --id MSYS2.MSYS2 --silent --accept-package-agreements --accept-source-agreements >nul 2>&1

if exist "C:\msys64\usr\bin\bash.exe" (
    echo [2/3] Installing gcc via pacman...
    C:\msys64\usr\bin\bash -lc "pacman -S --noconfirm mingw-w64-ucrt-x86_64-gcc" >nul 2>&1
    if exist "C:\msys64\ucrt64\bin\g++.exe" (
        set "PATH=C:\msys64\ucrt64\bin;%PATH%"
        set "CC=C:\msys64\ucrt64\bin\gcc.exe"
        set "CXX=C:\msys64\ucrt64\bin\g++.exe"
        echo [OK] MSYS2 ucrt64 gcc installed.
        goto :verify
    )
)

rem --- TDM-GCC مباشر ---
echo [2/3] Downloading TDM-GCC...
set TDM_URL=https://github.com/jmeubank/tdm-gcc/releases/download/v10.3.0-tdm64-2/tdm64-gcc-10.3.0-2.exe
set TDM_EXE=%TEMP%\tdm-gcc-setup.exe
set TDM_DIR=C:\TDM-GCC-64

powershell -NoProfile -Command ^
  "try { Invoke-WebRequest '%TDM_URL%' -OutFile '%TDM_EXE%' -UseBasicParsing; Write-Host '[OK] Downloaded.' } catch { Write-Host '[ERR]' $_.Exception.Message; exit 1 }"

if %errorlevel% neq 0 (
    echo [ERROR] Download failed. Install manually: https://jmeubank.github.io/tdm-gcc/download/
    pause & exit /b 1
)

echo [3/3] Installing TDM-GCC (~1 min)...
start /wait "" "%TDM_EXE%" /S /D=%TDM_DIR%
del /f "%TDM_EXE%" >nul 2>&1

if exist "%TDM_DIR%\bin\g++.exe" (
    set "PATH=%TDM_DIR%\bin;%PATH%"
    set "CC=%TDM_DIR%\bin\gcc.exe"
    set "CXX=%TDM_DIR%\bin\g++.exe"
    echo [OK] TDM-GCC installed.
    goto :verify
)

echo [ERROR] Installation failed.
echo         Download manually: https://jmeubank.github.io/tdm-gcc/download/
pause & exit /b 1

rem ── التحقق ───────────────────────────────────────────────────
:verify
echo.
gcc --version 2>nul | findstr /v "^$"
g++ --version 2>nul | findstr /v "^$"

rem تحذير إذا كان POSIX threads
g++ -v 2>&1 | findstr /i "posix" >nul
if %errorlevel% equ 0 (
    echo.
    echo [WARN] هذا GCC يستخدم POSIX threads ← قد يفشل مع Go CGO.
    echo        الحل المضمون: https://jmeubank.github.io/tdm-gcc/download/
    echo.
)

rem ── البناء ───────────────────────────────────────────────────
:build
echo.
echo [Building] internet-monitor-gui.exe ...
echo.

set CGO_ENABLED=1
powershell -NoProfile -Command ^
  "$env:CGO_ENABLED='1'; $env:CC='%CC%'; $env:CXX='%CXX%'; go build -ldflags='-H=windowsgui -s -w' -o internet-monitor-gui.exe ./cmd/gui/ 2>&1" ^
  > build_output.txt 2>&1

rem تحقق إذا الملف اتعمل
if exist internet-monitor-gui.exe (
    for %%A in (internet-monitor-gui.exe) do echo [OK] Done: %%~zA bytes — %%~fA
    del /f build_output.txt >nul 2>&1
) else (
    echo [ERROR] Build failed. Details:
    echo --------------------------------------------------------
    type build_output.txt
    echo --------------------------------------------------------
    echo.
    echo إذا كان الخطأ "posix threads" أو "runtime/cgo":
    echo   ثبّت TDM-GCC: https://jmeubank.github.io/tdm-gcc/download/
    echo   ثم شغّل السكريبت مرة ثانية.
    del /f build_output.txt >nul 2>&1
    pause & exit /b 1
)

endlocal
