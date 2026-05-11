@echo off
taskkill /F /IM internet-monitor.exe >nul 2>&1
if %errorlevel% equ 0 (
    echo [OK] Internet Monitor stopped.
) else (
    echo [INFO] Internet Monitor is not running.
)
