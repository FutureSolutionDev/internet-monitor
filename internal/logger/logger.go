package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"internet-monitor/internal/config"
	"internet-monitor/internal/monitor"
	"internet-monitor/internal/types"
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
	cfg     config.Config
	mu      sync.Mutex
	appMu   sync.Mutex
	appLog  *log.Logger
	appFile *os.File
}

const maxAppLogSize = 5 << 20

func New(cfg config.Config) (*Logger, error) {
	if err := os.MkdirAll(cfg.LogDir, 0o755); err != nil {
		return nil, err
	}

	appLogFile := filepath.Join(cfg.LogDir, "app.log")
	if info, err := os.Stat(appLogFile); err == nil && info.Size() > maxAppLogSize {
		_ = os.Rename(appLogFile, appLogFile+".1")
	}
	f, err := os.OpenFile(appLogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	var appLog *log.Logger
	if err == nil {
		appLog = log.New(f, "", log.LstdFlags)
	} else {
		f = nil
	}

	return &Logger{cfg: cfg.Clone(), appLog: appLog, appFile: f}, nil
}

func (l *Logger) Close() error {
	l.appMu.Lock()
	defer l.appMu.Unlock()
	if l.appFile == nil {
		return nil
	}
	err := l.appFile.Close()
	l.appFile = nil
	l.appLog = nil
	return err
}

func (l *Logger) LogDir() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.cfg.LogDir
}

func (l *Logger) SetConfig(cfg config.Config) {
	if cfg.LogDir != "" {
		if err := os.MkdirAll(cfg.LogDir, 0o755); err != nil {
			l.AppLog("ERROR create log dir on config reload: %v", err)
			return
		}
	}
	l.mu.Lock()
	l.cfg = cfg.Clone()
	l.mu.Unlock()
}

func (l *Logger) AppLog(format string, args ...interface{}) {
	l.appMu.Lock()
	defer l.appMu.Unlock()
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

type SpeedTestEvent struct {
	Timestamp       time.Time `json:"timestamp"`
	Event           string    `json:"event"`
	DownloadMbps    float64   `json:"download_mbps"`
	UploadMbps      *float64  `json:"upload_mbps"`
	LatencyMs       int64     `json:"latency_ms"`
	DurationSeconds float64   `json:"duration_seconds"`
	Endpoints       []string  `json:"endpoints"`
	ParallelConns   int       `json:"parallel_connections"`
	TriggeredBy     string    `json:"triggered_by"`
}

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

func (l *Logger) LogSample(s types.MetricSample) {
	l.mu.Lock()
	defer l.mu.Unlock()

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
		l.AppLog("TELEGRAM send error: %s", RedactSecret(err.Error(), token))
		return
	}
	resp.Body.Close()
}

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
		l.AppLog("WEBHOOK send error: %s", RedactSecret(err.Error(), url))
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

func (l *Logger) SendTestWebhook(url string) string {
	if !IsSupportedWebhook(url) {
		return "Only Discord and Slack webhooks are supported"
	}

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
		msg := RedactSecret(err.Error(), url)
		l.AppLog("WEBHOOK TEST error: %s", msg)
		return msg
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

func RedactSecret(msg, secret string) string {
	if secret == "" {
		return msg
	}
	return strings.ReplaceAll(msg, secret, "<redacted>")
}
