package booking

import (
	"errors"
	"testing"

	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
)

func TestParseZoomLink(t *testing.T) {
	t.Parallel()

	tests := []struct {
		raw     string
		want    string
		wantErr bool
	}{
		{"https://zoom.us/j/123", "https://zoom.us/j/123", false},
		{"  http://meet.example/room  ", "http://meet.example/room", false},
		{"", "", true},
		{"ftp://zoom.us/j/1", "", true},
		{"not-a-url", "", true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.raw, func(t *testing.T) {
			t.Parallel()
			got, err := ParseZoomLink(tt.raw)
			if tt.wantErr {
				if !errors.Is(err, domainerrors.ErrInvalidZoomLink) {
					t.Fatalf("err = %v", err)
				}
				return
			}
			if err != nil || got != tt.want {
				t.Fatalf("got %q err=%v", got, err)
			}
		})
	}
}
