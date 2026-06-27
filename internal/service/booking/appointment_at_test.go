package booking

import (
	"testing"
	"time"

	"github.com/anuarkuanysh/dental_project/internal/domain/booking"
)

func TestAppointmentAt(t *testing.T) {
	t.Parallel()

	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		t.Fatalf("load location: %v", err)
	}

	appt := booking.Appointment{
		PreferredDate: time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC),
		PreferredTime: time.Date(0, 1, 1, 14, 30, 0, 0, time.UTC),
	}

	got := AppointmentAt(appt, loc)
	if got.Year() != 2026 || got.Month() != time.June || got.Day() != 15 {
		t.Fatalf("unexpected date: %v", got)
	}
	if got.Hour() != 14 || got.Minute() != 30 {
		t.Fatalf("unexpected time: %v", got)
	}
	if got.Location() != loc {
		t.Fatalf("expected location %v, got %v", loc, got.Location())
	}
}
