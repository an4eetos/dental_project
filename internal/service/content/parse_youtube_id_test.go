package content

import (
	"testing"

	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
)

func TestParseYouTubeID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		raw     string
		want    string
		wantErr bool
	}{
		{name: "raw id", raw: "zQZ3SGSwGBI", want: "zQZ3SGSwGBI"},
		{name: "watch url", raw: "https://www.youtube.com/watch?v=IFT7drSL35s", want: "IFT7drSL35s"},
		{name: "short url", raw: "https://youtu.be/FMU4zgGRbiE", want: "FMU4zgGRbiE"},
		{name: "embed url", raw: "https://www.youtube.com/embed/yKlH5tjZTxI", want: "yKlH5tjZTxI"},
		{name: "invalid", raw: "bad", wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ParseYouTubeID(tt.raw)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %q want %q", got, tt.want)
			}
		})
	}
}

func TestValidateBlocks_RejectsEmpty(t *testing.T) {
	t.Parallel()
	if err := ValidateBlocks(nil); err != domainerrors.ErrInvalidContentBlocks {
		t.Fatalf("got %v", err)
	}
}
