package content

import (
	"net/url"
	"regexp"
	"strings"

	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
)

var youtubeIDPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{11}$`)

// ParseYouTubeID extracts a video ID from a raw ID or common YouTube URL forms.
func ParseYouTubeID(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", domainerrors.ErrInvalidYouTubeID
	}

	if youtubeIDPattern.MatchString(raw) {
		return raw, nil
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return "", domainerrors.ErrInvalidYouTubeID
	}

	host := strings.ToLower(parsed.Host)
	host = strings.TrimPrefix(host, "www.")
	host = strings.TrimPrefix(host, "m.")

	switch host {
	case "youtu.be":
		id := strings.Trim(parsed.Path, "/")
		if youtubeIDPattern.MatchString(id) {
			return id, nil
		}
	case "youtube.com", "youtube-nocookie.com":
		if parsed.Path == "/watch" {
			id := parsed.Query().Get("v")
			if youtubeIDPattern.MatchString(id) {
				return id, nil
			}
		}
		if strings.HasPrefix(parsed.Path, "/embed/") {
			id := strings.TrimPrefix(parsed.Path, "/embed/")
			id = strings.Split(id, "/")[0]
			if youtubeIDPattern.MatchString(id) {
				return id, nil
			}
		}
		if strings.HasPrefix(parsed.Path, "/shorts/") {
			id := strings.TrimPrefix(parsed.Path, "/shorts/")
			id = strings.Split(id, "/")[0]
			if youtubeIDPattern.MatchString(id) {
				return id, nil
			}
		}
	}

	return "", domainerrors.ErrInvalidYouTubeID
}
