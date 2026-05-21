package booking

import (
	"testing"
	"time"
)

func TestParsePreferredDate(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)

	ok, err := ParsePreferredDate("2026-05-25", now)
	if err != nil || ok.Format(DateLayout) != "2026-05-25" {
		t.Fatalf("expected ok, got %v err=%v", ok, err)
	}

	if _, err := ParsePreferredDate("2026-05-19", now); err == nil {
		t.Fatal("expected past date error")
	}
}

func TestParsePreferredTime(t *testing.T) {
	t.Parallel()

	if _, err := ParsePreferredTime("10:30"); err != nil {
		t.Fatalf("valid slot: %v", err)
	}
	if _, err := ParsePreferredTime("21:00"); err == nil {
		t.Fatal("expected after hours error")
	}
}
