@echo off
cd /d "%~dp0.."

if not exist internet-monitor.exe (
    echo [Internet Monitor] Binary not found, building first...
    call scripts\build.cmd
    if %errorlevel% neq 0 exit /b 1
)

echo [Internet Monitor] Starting...
start "" internet-monitor.exe
echo [OK] Running — check your system tray.
