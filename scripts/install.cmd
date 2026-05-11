@echo off
setlocal EnableDelayedExpansion
cd /d "%~dp0.."

set APP_NAME=InternetMonitor
set INSTALL_DIR=%LOCALAPPDATA%\Programs\InternetMonitor
set EXE=internet-monitor.exe
set REG_PATH=HKCU\Software\Microsoft\Windows\CurrentVersion\Run

echo ============================================
echo  Internet Monitor — Installer
echo ============================================
echo.

rem Build if needed
if not exist %EXE% (
    echo [1/4] Building binary...
    call scripts\build.cmd
    if %errorlevel% neq 0 exit /b 1
) else (
    echo [1/4] Binary found: %EXE%
)

rem Create install directory
echo [2/4] Installing to: %INSTALL_DIR%
if not exist "%INSTALL_DIR%" mkdir "%INSTALL_DIR%"
if not exist "%INSTALL_DIR%\logs" mkdir "%INSTALL_DIR%\logs"

rem Copy exe
copy /Y %EXE% "%INSTALL_DIR%\%EXE%" >nul
echo       Copied: %EXE%

rem Copy config only if it exists — app creates default on first run if missing
if exist config.json (
    copy /Y config.json "%INSTALL_DIR%\config.json" >nul
    echo       Copied: config.json
) else (
    echo       No config.json found — app will create default on first run.
)

rem Register startup (HKCU — no admin needed)
echo [3/4] Adding to Windows startup...
reg add "%REG_PATH%" /v "%APP_NAME%" /t REG_SZ /d "\"%INSTALL_DIR%\%EXE%\"" /f >nul
if %errorlevel% equ 0 (
    echo       Registered in: %REG_PATH%
) else (
    echo [WARN] Could not write to registry.
)

echo [4/4] Done.
echo.
echo  Installed at : %INSTALL_DIR%
echo  Startup key  : %REG_PATH%\%APP_NAME%
echo  Logs folder  : %INSTALL_DIR%\logs
echo.

set /p START=Start now? (y/n):
if /i "!START!"=="y" (
    start "" "%INSTALL_DIR%\%EXE%"
    echo [OK] Started. Check your system tray.
)

endlocal
