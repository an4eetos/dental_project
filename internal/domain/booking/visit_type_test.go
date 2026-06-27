package booking

import "testing"

func TestVisitType_Valid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		visit VisitType
		want  bool
	}{
		{VisitTypeInPerson, true},
		{VisitTypeVideo, true},
		{VisitType(""), false},
		{VisitType("phone"), false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(string(tt.visit), func(t *testing.T) {
			t.Parallel()
			if got := tt.visit.Valid(); got != tt.want {
				t.Fatalf("Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}
