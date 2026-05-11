package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"internet-monitor/config"
	"internet-monitor/monitor"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Logger struct {
	cfg *config.Config
	mu  sync.Mutex
}

func New(cfg *config.Config) (*Logger, error) {
	if err := os.MkdirAll(cfg.LogDir, 0755); err != nil {
		return nil, err
	}
	return &Logger{cfg: cfg}, nil
}

func (l *Logger) Log(event monitor.Event) {
	l.mu.Lock()
	defer l.mu.Unlock()

	filename := filepath.Join(
		l.cfg.LogDir,
		fmt.Sprintf("connectivity_%s.jsonl", time.Now().Format("2006-01-02")),
	)

	data, err := json.Marshal(event)
	if err != nil {
		return
	}
	data = append(data, '\n')

	f, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err == nil {
		f.Write(data)
		f.Close()
	}

	if l.cfg.WebhookURL != "" {
		go l.sendWebhook(event)
	}
}

func (l *Logger) sendWebhook(event monitor.Event) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(l.cfg.WebhookURL, "application/json", bytes.NewReader(data))
	if err == nil {
		resp.Body.Close()
	}
}
