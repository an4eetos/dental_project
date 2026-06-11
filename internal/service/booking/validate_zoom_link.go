package booking

import (
	"net/url"
	"strings"

	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
)

// ParseZoomLink validates an http(s) meeting URL.
func ParseZoomLink(raw string) (string, error) {
	link := strings.TrimSpace(raw)
	if link == "" {
		return "", domainerrors.ErrInvalidZoomLink
	}

	u, err := url.Parse(link)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", domainerrors.ErrInvalidZoomLink
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", domainerrors.ErrInvalidZoomLink
	}
	return link, nil
}
