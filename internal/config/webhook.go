package config

import (
	"net/url"
	"strings"
)

var webhookHosts = map[string]string{
	"discord.com":        "/api/webhooks/",
	"discordapp.com":     "/api/webhooks/",
	"canary.discord.com": "/api/webhooks/",
	"ptb.discord.com":    "/api/webhooks/",
	"hooks.slack.com":    "/services/",
}

func parseWebhook(raw string) (*url.URL, bool) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Scheme != "https" || u.User != nil || u.Port() != "" {
		return nil, false
	}
	prefix, ok := webhookHosts[strings.ToLower(u.Hostname())]
	if !ok || !strings.HasPrefix(u.Path, prefix) || len(u.Path) <= len(prefix) {
		return nil, false
	}
	return u, true
}

func IsSupportedWebhook(raw string) bool {
	_, ok := parseWebhook(raw)
	return ok
}

func IsDiscordWebhook(raw string) bool {
	u, ok := parseWebhook(raw)
	return ok && strings.Contains(strings.ToLower(u.Hostname()), "discord")
}

func IsSlackWebhook(raw string) bool {
	u, ok := parseWebhook(raw)
	return ok && strings.EqualFold(u.Hostname(), "hooks.slack.com")
}
