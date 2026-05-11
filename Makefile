.PHONY: build build-gui build-debug run stop install uninstall logs clean

BINARY       = internet-monitor.exe
GUI_BINARY   = internet-monitor-gui.exe
DEBUG_BINARY = internet-monitor-debug.exe

build:
	go build -ldflags="-H=windowsgui -s -w" -o $(BINARY) .
	@echo Built: $(BINARY)

build-debug:
	go build -o $(DEBUG_BINARY) .
	@echo Built: $(DEBUG_BINARY)

run: build
	start "" $(BINARY)

run-debug: build-debug
	.\$(DEBUG_BINARY)

stop:
	@taskkill /F /IM $(BINARY) 2>nul || echo Not running.

install:
	@scripts\install.cmd

uninstall:
	@scripts\uninstall.cmd

logs:
	@scripts\logs.cmd

build-gui:
	set CGO_ENABLED=1 && go build -ldflags="-H=windowsgui -s -w" -o $(GUI_BINARY) ./cmd/gui/
	@echo Built: $(GUI_BINARY)

clean:
	@if exist $(BINARY) del $(BINARY) && echo Deleted $(BINARY)
	@if exist $(GUI_BINARY) del $(GUI_BINARY) && echo Deleted $(GUI_BINARY)
	@if exist $(DEBUG_BINARY) del $(DEBUG_BINARY) && echo Deleted $(DEBUG_BINARY)
