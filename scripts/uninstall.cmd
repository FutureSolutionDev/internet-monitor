@echo off
setlocal

set APP_NAME=InternetMonitor
set INSTALL_DIR=%LOCALAPPDATA%\Programs\InternetMonitor
set REG_PATH=HKCU\Software\Microsoft\Windows\CurrentVersion\Run

echo ============================================
echo  Internet Monitor — Uninstaller
echo ============================================
echo.

echo [1/3] Stopping process...
taskkill /F /IM internet-monitor.exe >nul 2>&1
if %errorlevel% equ 0 (echo       Stopped.) else (echo       Not running.)

echo [2/3] Removing from Windows startup...
reg delete "%REG_PATH%" /v "%APP_NAME%" /f >nul 2>&1
if %errorlevel% equ 0 (echo       Removed startup entry.) else (echo       Entry not found.)

echo [3/3] Done.
echo.
echo  Note: Files at "%INSTALL_DIR%" were NOT deleted.
echo  To delete them run:
echo    rmdir /s /q "%INSTALL_DIR%"
echo.

endlocal
