package tray

import (
	"errors"
	"net/url"
	"strings"
)

var errUnsupportedURL = errors.New("only http and https URLs are allowed")

func ValidateWebURL(raw string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return "", err
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return "", errUnsupportedURL
	}
	if u.Hostname() == "" {
		return "", errUnsupportedURL
	}
	return u.String(), nil
}

var externalHosts = map[string]bool{
	"github.com":        true,
	"wa.me":             true,
	"chat.whatsapp.com": true,
	"discord.gg":        true,
	"discord.com":       true,
	"paypal.me":         true,
	"www.paypal.me":     true,
	"paypal.com":        true,
	"www.paypal.com":    true,
	"ipn.eg":            true,
}

func ValidateExternalURL(raw string) (string, error) {
	valid, err := ValidateWebURL(raw)
	if err != nil {
		return "", err
	}
	u, _ := url.Parse(valid)
	if !strings.EqualFold(u.Scheme, "https") || u.User != nil || !externalHosts[strings.ToLower(u.Hostname())] || u.Port() != "" {
		return "", errUnsupportedURL
	}
	return valid, nil
}
