@echo off
cd /d "%~dp0..\web"

where bun >nul 2>&1
if %errorlevel% neq 0 (
    echo [ERROR] bun is required to build the dashboard: https://bun.sh
    exit /b 1
)

echo [Internet Monitor] Building dashboard UI...
call bun install --frozen-lockfile || exit /b 1
call bun run build || exit /b 1
echo [OK] Dashboard built: web\dist
