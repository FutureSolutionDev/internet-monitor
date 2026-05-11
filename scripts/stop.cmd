@echo off
setlocal
set FOUND=0

taskkill /F /IM internet-monitor.exe >nul 2>&1
if %errorlevel% equ 0 set FOUND=1

taskkill /F /IM internet-monitor-gui.exe >nul 2>&1
if %errorlevel% equ 0 set FOUND=1

taskkill /F /IM internet-monitor-windows.exe >nul 2>&1
if %errorlevel% equ 0 set FOUND=1

taskkill /F /IM internet-monitor-gui-windows.exe >nul 2>&1
if %errorlevel% equ 0 set FOUND=1

if %FOUND% equ 1 (
    echo [OK] All Internet Monitor processes stopped.
) else (
    echo [INFO] No Internet Monitor processes were running.
)
