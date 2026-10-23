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
