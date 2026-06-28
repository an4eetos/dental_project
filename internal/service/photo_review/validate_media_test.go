package photoreview_test

import (
	"testing"

	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
	photoreview "github.com/anuarkuanysh/dental_project/internal/domain/photo_review"
	photoreviewservice "github.com/anuarkuanysh/dental_project/internal/service/photo_review"
)

func TestValidateMediaSize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		media   photoreview.MediaType
		size    int
		max     int
		wantErr error
	}{
		{name: "photo ignores limit", media: photoreview.MediaTypePhoto, size: 30 << 20, max: 1 << 20},
		{name: "video within limit", media: photoreview.MediaTypeVideo, size: 1 << 20, max: 2 << 20},
		{name: "video exceeds limit", media: photoreview.MediaTypeVideo, size: 3 << 20, max: 2 << 20, wantErr: domainerrors.ErrSubmissionMediaTooLarge},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := photoreviewservice.ValidateMediaSize(tt.media, tt.size, tt.max)
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("expected %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestNormalizeVideoMIME(t *testing.T) {
	t.Parallel()

	if got := photoreviewservice.NormalizeVideoMIME("video/webm"); got != "video/webm" {
		t.Fatalf("got %q", got)
	}
	if got := photoreviewservice.NormalizeVideoMIME(""); got != "video/mp4" {
		t.Fatalf("got %q", got)
	}
}
