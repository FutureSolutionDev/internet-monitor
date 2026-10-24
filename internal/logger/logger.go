package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"internet-monitor/internal/config"
	"internet-monitor/internal/i18n"
	"internet-monitor/internal/monitor"
	"internet-monitor/internal/types"
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
			!strings.HasPrefix(name, "metrics_") &&
			!strings.HasPrefix(name, "notifications_") {
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

func (l *Logger) remoteTargets() (url, token, chat string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.cfg.WebhookURL, l.cfg.TelegramBotToken, l.cfg.TelegramChatID
}

func (l *Logger) SendEvent(lang string, event monitor.Event, kind string, webhook, telegram bool) {
	url, token, chat := l.remoteTargets()
	if webhook && url != "" && IsSupportedWebhook(url) {
		go l.sendWebhook(url, BuildEventPayload(lang, event, kind, url))
	}
	if telegram && token != "" && chat != "" {
		go l.sendTelegram(token, chat, TelegramEventText(lang, event, kind))
	}
}

func (l *Logger) SendSpeed(lang string, event SpeedTestEvent, thresholdMbps float64, below, webhook, telegram bool) {
	url, token, chat := l.remoteTargets()
	if webhook && url != "" && IsSupportedWebhook(url) {
		go l.sendWebhook(url, BuildSpeedResultPayload(lang, event, thresholdMbps, below, url))
	}
	if telegram && token != "" && chat != "" {
		go l.sendTelegram(token, chat, TelegramSpeedText(lang, event, thresholdMbps, below))
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
	client := NoRedirectClient(10 * time.Second)
	resp, err := client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		l.AppLog("TELEGRAM send error: %s", RedactSecret(err.Error(), token))
		return
	}
	resp.Body.Close()
	if resp.StatusCode >= 400 {
		l.AppLog("TELEGRAM HTTP %d", resp.StatusCode)
	}
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

	client := NoRedirectClient(10 * time.Second)
	resp, err := client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		l.AppLog("WEBHOOK send error: %s", RedactSecret(err.Error(), url))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		l.AppLog("WEBHOOK HTTP %d", resp.StatusCode)
	} else {
		l.AppLog("WEBHOOK OK HTTP %d", resp.StatusCode)
	}
}

func (l *Logger) SendTestWebhook(url string) string {
	if !IsSupportedWebhook(url) {
		return "Only Discord and Slack webhooks are supported"
	}

	l.mu.Lock()
	lang := l.cfg.Language
	l.mu.Unlock()
	title, text := i18n.T(lang, "webhook.test.title", nil), i18n.T(lang, "webhook.test.body", nil)

	var payload interface{}
	if IsDiscord(url) {
		payload = map[string]interface{}{
			"username": "Internet Monitor",
			"embeds": []map[string]interface{}{{
				"title":       title,
				"description": text,
				"color":       0x22C55E,
				"footer":      map[string]string{"text": "Internet Monitor"},
			}},
		}
	} else {
		payload = map[string]interface{}{
			"text": "*" + title + "* — " + text,
		}
	}

	body, _ := json.Marshal(payload)
	client := NoRedirectClient(10 * time.Second)
	resp, err := client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		msg := RedactSecret(err.Error(), url)
		l.AppLog("WEBHOOK TEST error: %s", msg)
		return msg
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		msg := fmt.Sprintf("the webhook service answered HTTP %d", resp.StatusCode)
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

func NoRedirectClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout:       timeout,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
}
