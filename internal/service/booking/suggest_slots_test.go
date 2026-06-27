package booking

import (
	"strings"
	"testing"
	"time"
)

func TestSuggestAvailableSlots_ProducesWeekdaySlots(t *testing.T) {
	t.Parallel()

	// Wednesday 2026-06-10 — next slots should start Thursday onward.
	now := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)
	patientDate := time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)

	text := SuggestAvailableSlots(now, patientDate)
	if !strings.HasPrefix(text, "Доступные варианты для записи:") {
		t.Fatalf("unexpected prefix: %s", text)
	}
	if !strings.Contains(text, "2026-06-11") {
		t.Fatalf("expected next weekday slot, got: %s", text)
	}
	if strings.Contains(text, "2026-06-13") {
		t.Fatal("Saturday should be skipped")
	}
}

func TestSuggestAvailableSlots_CountsSixEntries(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 6, 1, 9, 0, 0, 0, time.UTC) // Monday
	text := SuggestAvailableSlots(now, now)

	lines := strings.Split(text, "\n")
	slotLines := 0
	for _, line := range lines {
		if strings.HasPrefix(line, "- ") {
			slotLines++
		}
	}
	if slotLines != suggestSlotCount {
		t.Fatalf("want %d slot lines, got %d in:\n%s", suggestSlotCount, slotLines, text)
	}
}
