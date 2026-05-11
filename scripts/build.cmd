@echo off
cd /d "%~dp0.."

echo [Internet Monitor] Building...
go build -ldflags="-H=windowsgui -s -w" -o internet-monitor.exe .
if %errorlevel% neq 0 (
    echo [ERROR] Build failed!
    pause
    exit /b 1
)

for %%A in (internet-monitor.exe) do echo [OK] Built: %%~zA bytes ^| %%~fA
