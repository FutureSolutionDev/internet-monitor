.PHONY: build build-debug run stop install uninstall logs clean

BINARY       = internet-monitor.exe
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

clean:
	@if exist $(BINARY) del $(BINARY) && echo Deleted $(BINARY)
	@if exist $(DEBUG_BINARY) del $(DEBUG_BINARY) && echo Deleted $(DEBUG_BINARY)
