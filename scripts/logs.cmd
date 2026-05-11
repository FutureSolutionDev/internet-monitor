@echo off
cd /d "%~dp0.."

rem Check local logs folder first, then installed location
if exist logs (
    explorer logs
    exit /b
)

set INSTALL_LOGS=%LOCALAPPDATA%\Programs\InternetMonitor\logs
if exist "%INSTALL_LOGS%" (
    explorer "%INSTALL_LOGS%"
    exit /b
)

echo [INFO] No logs folder found yet.
echo       Run the app first to generate logs.
