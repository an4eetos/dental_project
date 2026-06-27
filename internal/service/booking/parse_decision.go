package booking

import (
	"strings"

	domainbooking "github.com/anuarkuanysh/dental_project/internal/domain/booking"
	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
)

// ParseDecision validates a doctor response type.
func ParseDecision(raw string) (domainbooking.DoctorDecision, error) {
	decision := domainbooking.DoctorDecision(strings.TrimSpace(raw))
	if !decision.Valid() {
		return "", domainerrors.ErrInvalidDecision
	}
	return decision, nil
}

// ParseOptionalZoomLink validates a meeting URL when provided.
func ParseOptionalZoomLink(raw string) (string, error) {
	link := strings.TrimSpace(raw)
	if link == "" {
		return "", nil
	}
	return ParseZoomLink(link)
}

// ParseDoctorNotes trims rejection / clarification text.
func ParseDoctorNotes(raw string) (string, error) {
	notes := strings.TrimSpace(raw)
	if notes == "" {
		return "", domainerrors.ErrDoctorNotesRequired
	}
	return notes, nil
}
