@echo off
cd /d "%~dp0.."

echo [Internet Monitor GUI] Checking CGO...

where gcc >nul 2>&1
if %errorlevel% neq 0 (
    echo.
    echo [ERROR] gcc not found — CGO required for the GUI version.
    echo.
    echo Install TDM-GCC with one command:
    echo   winget install TDMGcc.TDMGcc
    echo.
    echo Then close and reopen this terminal, and run this script again.
    echo.
    pause
    exit /b 1
)

echo [OK] gcc found:
where gcc
echo.
echo [Internet Monitor GUI] Building native window version...

set CGO_ENABLED=1
go build -ldflags="-H=windowsgui -s -w" -o internet-monitor-gui.exe ./cmd/gui/

if %errorlevel% neq 0 (
    echo [ERROR] Build failed!
    pause
    exit /b 1
)

for %%A in (internet-monitor-gui.exe) do echo [OK] Built: %%~zA bytes ^| %%~fA
echo.
echo Run: internet-monitor-gui.exe
