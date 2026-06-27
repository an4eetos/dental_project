package booking

import (
	"errors"
	"testing"

	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
)

func TestParseDecision(t *testing.T) {
	t.Parallel()

	tests := []struct {
		raw     string
		wantErr error
	}{
		{"in_person", nil},
		{"  video  ", nil},
		{"reject", nil},
		{"", domainerrors.ErrInvalidDecision},
		{"zoom", domainerrors.ErrInvalidDecision},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.raw, func(t *testing.T) {
			t.Parallel()
			got, err := ParseDecision(tt.raw)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("err = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if !got.Valid() {
				t.Fatalf("invalid decision: %q", got)
			}
		})
	}
}

func TestParseOptionalZoomLink(t *testing.T) {
	t.Parallel()

	link, err := ParseOptionalZoomLink("")
	if err != nil || link != "" {
		t.Fatalf("empty: link=%q err=%v", link, err)
	}

	link, err = ParseOptionalZoomLink("https://zoom.us/j/123")
	if err != nil || link != "https://zoom.us/j/123" {
		t.Fatalf("valid: link=%q err=%v", link, err)
	}

	_, err = ParseOptionalZoomLink("not-a-url")
	if !errors.Is(err, domainerrors.ErrInvalidZoomLink) {
		t.Fatalf("expected invalid zoom link, got %v", err)
	}
}

func TestParseDoctorNotes(t *testing.T) {
	t.Parallel()

	notes, err := ParseDoctorNotes("  Mon 10:00  ")
	if err != nil || notes != "Mon 10:00" {
		t.Fatalf("notes=%q err=%v", notes, err)
	}

	_, err = ParseDoctorNotes("   ")
	if !errors.Is(err, domainerrors.ErrDoctorNotesRequired) {
		t.Fatalf("expected notes required, got %v", err)
	}
}
