package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"internet-monitor/config"
	"internet-monitor/monitor"
	"internet-monitor/types"
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

// ReadFile reads a log file under the same mutex that guards writes, so a
// reader never observes a torn line from a concurrent append.
func (l *Logger) ReadFile(path string) ([]byte, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return os.ReadFile(path)
}

// SetConfig swaps the logger's configuration under the same mutex that guards
// all reads of l.cfg (Log / LogSpeedTest), so a live config change is race-free.
// Ensures the new LogDir exists before swapping so future writes don't fail
// silently when the dashboard saves a non-existent path.
func (l *Logger) SetConfig(cfg *config.Config) {
	if cfg == nil {
		return
	}
	if cfg.LogDir != "" {
		if err := os.MkdirAll(cfg.LogDir, 0755); err != nil {
			l.AppLog("ERROR create log dir on config reload: %v", err)
			return
		}
	}
	l.mu.Lock()
	l.cfg = cfg
	l.mu.Unlock()
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
	if l.cfg.TelegramBotToken != "" && l.cfg.TelegramChatID != "" {
		go l.sendTelegram(l.cfg.TelegramBotToken, l.cfg.TelegramChatID, TelegramEventText(event))
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

// LogSample appends a one-minute metric aggregate to metrics_YYYY-MM-DD.jsonl.
func (l *Logger) LogSample(s types.MetricSample) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Use the sample's own timestamp (the minute bucket) so a 23:59 flush after
	// midnight lands in the correct day's file rather than today's.
	date := s.Timestamp
	if date.IsZero() {
		date = time.Now()
	}
	filename := filepath.Join(
		l.cfg.LogDir,
		fmt.Sprintf("metrics_%s.jsonl", date.Format("2006-01-02")),
	)
	data, err := json.Marshal(s)
	if err != nil {
		l.AppLog("ERROR marshal metric sample: %v", err)
		return
	}
	data = append(data, '\n')

	f, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err == nil {
		f.Write(data)
		f.Close()
	} else {
		l.AppLog("ERROR open metrics log: %v", err)
	}
}

// CleanupOldLogs deletes connectivity/speedtest/metrics log files older than
// maxAge (by file modification time). Best-effort; errors are logged only.
func (l *Logger) CleanupOldLogs(maxAge time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	entries, err := os.ReadDir(l.cfg.LogDir)
	if err != nil {
		return
	}
	cutoff := time.Now().Add(-maxAge)
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".jsonl") {
			continue
		}
		if !strings.HasPrefix(name, "connectivity_") &&
			!strings.HasPrefix(name, "speedtest_") &&
			!strings.HasPrefix(name, "metrics_") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			if err := os.Remove(filepath.Join(l.cfg.LogDir, name)); err != nil {
				l.AppLog("ERROR removing old log %s: %v", name, err)
			}
		}
	}
}

// SendSpeedTestResult sends the speed test result to the webhook and/or Telegram.
// If belowThreshold is true, the payload is styled as a warning alert.
func (l *Logger) SendSpeedTestResult(webhookURL string, event SpeedTestEvent, thresholdMbps float64, belowThreshold bool) {
	if webhookURL != "" && IsSupportedWebhook(webhookURL) {
		go l.sendWebhook(webhookURL, BuildSpeedResultPayload(event, thresholdMbps, belowThreshold, webhookURL))
	}
	l.mu.Lock()
	tgToken, tgChat := l.cfg.TelegramBotToken, l.cfg.TelegramChatID
	l.mu.Unlock()
	if tgToken != "" && tgChat != "" {
		go l.sendTelegram(tgToken, tgChat, TelegramSpeedText(event, thresholdMbps, belowThreshold))
	}
}

// sendTelegram posts a Markdown message to a chat via the Telegram Bot API.
func (l *Logger) sendTelegram(token, chatID, text string) {
	defer func() {
		if r := recover(); r != nil {
			l.AppLog("PANIC recovered in telegram send: %v", r)
		}
	}()
	if token == "" || chatID == "" {
		return
	}
	body, _ := json.Marshal(map[string]string{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "Markdown",
	})
	url := "https://api.telegram.org/bot" + token + "/sendMessage"
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		l.AppLog("TELEGRAM send error: %v", err)
		return
	}
	resp.Body.Close()
}

// buildEventPayload creates a Discord+generic compatible payload for a connectivity event.
// sendWebhook sends a payload to the configured webhook URL.
// It logs success/failure to app.log (issue 9).
func (l *Logger) sendWebhook(url string, payload interface{}) {
	defer func() {
		if r := recover(); r != nil {
			l.AppLog("PANIC recovered in webhook send: %v", r)
		}
	}()

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
