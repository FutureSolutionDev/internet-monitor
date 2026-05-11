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
		if !IsSupportedWebhook(l.cfg.WebhookURL) {
			l.AppLog("WEBHOOK skipped: URL is not Discord or Slack")
		} else {
			go l.sendWebhook(l.cfg.WebhookURL, BuildEventPayload(event, l.cfg.WebhookURL))
		}
	}
}

// SpeedTestEvent is a completed speed test result written to speedtest_DATE.jsonl.
type SpeedTestEvent struct {
	Timestamp       time.Time `json:"timestamp"`
	Event           string    `json:"event"` // always "speed_test"
	DownloadMbps    float64   `json:"download_mbps"`
	UploadMbps      *float64  `json:"upload_mbps"`
	LatencyMs       int64     `json:"latency_ms"`
	DurationSeconds float64   `json:"duration_seconds"`
	Endpoints       []string  `json:"endpoints"`
	ParallelConns   int       `json:"parallel_connections"`
	TriggeredBy     string    `json:"triggered_by"`
}

// LogSpeedTest writes a speed test result to speedtest_YYYY-MM-DD.jsonl.
// Uses the same mutex as Log() to prevent concurrent writes to the same file.
func (l *Logger) LogSpeedTest(event SpeedTestEvent) {
	l.mu.Lock()
	defer l.mu.Unlock()

	filename := filepath.Join(
		l.cfg.LogDir,
		fmt.Sprintf("speedtest_%s.jsonl", time.Now().Format("2006-01-02")),
	)

	data, err := json.Marshal(event)
	if err != nil {
		l.AppLog("ERROR marshal speed test event: %v", err)
		return
	}
	data = append(data, '\n')

	f, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err == nil {
		f.Write(data)
		f.Close()
	} else {
		l.AppLog("ERROR open speedtest log: %v", err)
	}
}

// SendSpeedTestAlert sends a webhook notification when speed drops below threshold.
func (l *Logger) SendSpeedTestAlert(webhookURL string, event SpeedTestEvent, thresholdMbps float64) {
	if webhookURL == "" || !IsSupportedWebhook(webhookURL) {
		return
	}
	go l.sendWebhook(webhookURL, BuildSpeedAlertPayload(event, thresholdMbps, webhookURL))
}

// buildEventPayload creates a Discord+generic compatible payload for a connectivity event.
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
	if !IsSupportedWebhook(url) {
		return "Only Discord and Slack webhooks are supported"
	}

	// Build a proper test payload matching the platform format
	var payload interface{}
	if IsDiscord(url) {
		payload = map[string]interface{}{
			"username": "Internet Monitor",
			"embeds": []map[string]interface{}{{
				"title":       "🔔 Webhook Test",
				"description": "الاتصال بين Internet Monitor والـ Webhook يعمل بشكل صحيح ✅",
				"color":       0x22C55E,
				"footer":      map[string]string{"text": "Internet Monitor"},
			}},
		}
	} else {
		payload = map[string]interface{}{
			"text": ":bell: *Webhook Test* — Internet Monitor webhook is working correctly :white_check_mark:",
		}
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
