@echo off
cd /d "%~dp0.."

echo [Internet Monitor] Building DEBUG (console visible)...
go build -o internet-monitor-debug.exe .
if %errorlevel% neq 0 (
    echo [ERROR] Build failed!
    pause
    exit /b 1
)

echo [OK] Built: internet-monitor-debug.exe
echo.
echo To run: internet-monitor-debug.exe
