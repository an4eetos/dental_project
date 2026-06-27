package booking

import "testing"

func TestStatus_Valid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		status Status
		want   bool
	}{
		{StatusPending, true},
		{StatusConfirmed, true},
		{StatusRejected, true},
		{StatusCancelled, true},
		{Status("unknown"), false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(string(tt.status), func(t *testing.T) {
			t.Parallel()
			if got := tt.status.Valid(); got != tt.want {
				t.Fatalf("Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStatus_String(t *testing.T) {
	t.Parallel()
	if StatusConfirmed.String() != "confirmed" {
		t.Fatalf("unexpected string: %q", StatusConfirmed.String())
	}
}
