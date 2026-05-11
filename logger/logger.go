package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"internet-monitor/config"
	"internet-monitor/monitor"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Logger struct {
	cfg    *config.Config
	mu     sync.Mutex
	appLog *log.Logger
}

func New(cfg *config.Config) (*Logger, error) {
	if err := os.MkdirAll(cfg.LogDir, 0755); err != nil {
		return nil, err
	}

	// App-level error log (issue 10)
	appLogFile := filepath.Join(cfg.LogDir, "app.log")
	f, err := os.OpenFile(appLogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	var appLog *log.Logger
	if err == nil {
		appLog = log.New(f, "", log.LstdFlags)
	}

	return &Logger{cfg: cfg, appLog: appLog}, nil
}

// AppLog writes an application-level error/info to logs/app.log
func (l *Logger) AppLog(format string, args ...interface{}) {
	if l.appLog != nil {
		l.appLog.Printf(format, args...)
	}
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
		l.AppLog("ERROR marshal event: %v", err)
		return
	}
	data = append(data, '\n')

	f, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err == nil {
		f.Write(data)
		f.Close()
	} else {
		l.AppLog("ERROR open log file: %v", err)
	}

	if l.cfg.WebhookURL != "" {
		go l.sendWebhook(l.cfg.WebhookURL, l.buildEventPayload(event))
	}
}

// buildEventPayload creates a Discord+generic compatible payload for a connectivity event.
func (l *Logger) buildEventPayload(event monitor.Event) map[string]interface{} {
	emoji := map[string]string{
		"connected":    "✅",
		"disconnected": "❌",
		"degraded":     "⚠️",
	}
	em := emoji[event.EventType]
	summary := fmt.Sprintf("%s Internet %s", em, event.EventType)
	if event.DurationSeconds > 0 {
		summary += fmt.Sprintf(" (%.0fs)", event.DurationSeconds)
	}

	return map[string]interface{}{
		"content":   summary, // Discord shows this
		"event":     event.EventType,
		"timestamp": event.Timestamp.UTC().Format(time.RFC3339),
		"duration_seconds": event.DurationSeconds,
		"reason":    event.Reason,
	}
}

// sendWebhook sends a payload to the configured webhook URL.
// It logs success/failure to app.log (issue 9).
func (l *Logger) sendWebhook(url string, payload interface{}) {
	body, err := json.Marshal(payload)
	if err != nil {
		l.AppLog("WEBHOOK marshal error: %v", err)
		return
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		l.AppLog("WEBHOOK send error: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		l.AppLog("WEBHOOK HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	} else {
		l.AppLog("WEBHOOK OK HTTP %d", resp.StatusCode)
	}
}

// SendTestWebhook sends a test payload to the given URL.
// Returns an error string if it fails, or "" on success.
func (l *Logger) SendTestWebhook(url string) string {
	payload := map[string]interface{}{
		"content":   "🔔 Internet Monitor — Webhook Test",
		"type":      "webhook_test",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"status":    "ok",
	}

	body, _ := json.Marshal(payload)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		l.AppLog("WEBHOOK TEST error: %v", err)
		return err.Error()
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		msg := fmt.Sprintf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
		l.AppLog("WEBHOOK TEST %s", msg)
		return msg
	}
	l.AppLog("WEBHOOK TEST OK HTTP %d", resp.StatusCode)
	return ""
}
