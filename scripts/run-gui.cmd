@echo off
cd /d "%~dp0.."

rem Priority 1: TDM-GCC
if exist "C:\TDM-GCC-64\bin\gcc.exe" (
    set "PATH=C:\TDM-GCC-64\bin;%PATH%"
    set "CC=C:\TDM-GCC-64\bin\gcc.exe"
    goto :run
)

rem Priority 2: MSYS2 ucrt64
if exist "C:\msys64\ucrt64\bin\gcc.exe" (
    set "PATH=C:\msys64\ucrt64\bin;%PATH%"
    set "CC=C:\msys64\ucrt64\bin\gcc.exe"
    goto :run
)

rem Priority 3: MSYS2 mingw64
if exist "C:\msys64\mingw64\bin\gcc.exe" (
    set "PATH=C:\msys64\mingw64\bin;%PATH%"
    set "CC=C:\msys64\mingw64\bin\gcc.exe"
    goto :run
)

echo [ERROR] No compatible GCC found.
echo         Run scripts\build-gui.cmd to install TDM-GCC automatically.
pause & exit /b 1

:run
set CGO_ENABLED=1
go run ./cmd/gui/
