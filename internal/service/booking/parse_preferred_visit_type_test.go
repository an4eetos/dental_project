package booking

import (
	"errors"
	"testing"

	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
)

func TestParsePreferredVisitType_Valid(t *testing.T) {
	t.Parallel()

	for _, raw := range []string{"in_person", "video", " in_person "} {
		got, err := ParsePreferredVisitType(raw)
		if err != nil {
			t.Fatalf("ParsePreferredVisitType(%q): %v", raw, err)
		}
		if got != "in_person" && got != "video" {
			t.Fatalf("unexpected type: %q", got)
		}
	}
}

func TestParsePreferredVisitType_Invalid(t *testing.T) {
	t.Parallel()

	_, err := ParsePreferredVisitType("phone")
	if !errors.Is(err, domainerrors.ErrInvalidPreferredVisitType) {
		t.Fatalf("err = %v", err)
	}
}
