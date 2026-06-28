package photoreview

import (
	"strings"

	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
	photoreview "github.com/anuarkuanysh/dental_project/internal/domain/photo_review"
)

const defaultMaxVideoBytes = 20 * 1024 * 1024

// ValidateMediaSize rejects submissions that exceed configured limits.
func ValidateMediaSize(mediaType photoreview.MediaType, size, maxVideoBytes int) error {
	if mediaType != photoreview.MediaTypeVideo {
		return nil
	}
	limit := maxVideoBytes
	if limit <= 0 {
		limit = defaultMaxVideoBytes
	}
	if size > limit {
		return domainerrors.ErrSubmissionMediaTooLarge
	}
	return nil
}

// NormalizeVideoMIME returns a supported video MIME type or a safe default.
func NormalizeVideoMIME(mimeHint string) string {
	mime := strings.ToLower(strings.TrimSpace(mimeHint))
	switch mime {
	case "video/mp4", "video/webm", "video/quicktime", "video/3gpp":
		return mime
	default:
		return "video/mp4"
	}
}
