package booking

import (
	"strings"

	"github.com/anuarkuanysh/dental_project/internal/domain/booking"
	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
)

// ParsePreferredVisitType validates the patient's requested visit format.
func ParsePreferredVisitType(raw string) (booking.VisitType, error) {
	v := booking.VisitType(strings.TrimSpace(raw))
	if !v.Valid() {
		return "", domainerrors.ErrInvalidPreferredVisitType
	}
	return v, nil
}
