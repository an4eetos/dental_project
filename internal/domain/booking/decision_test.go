package booking

import "testing"

func TestDoctorDecision_Valid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		decision DoctorDecision
		want     bool
	}{
		{DecisionInPerson, true},
		{DecisionVideo, true},
		{DecisionReject, true},
		{DoctorDecision("maybe"), false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(string(tt.decision), func(t *testing.T) {
			t.Parallel()
			if got := tt.decision.Valid(); got != tt.want {
				t.Fatalf("Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}
