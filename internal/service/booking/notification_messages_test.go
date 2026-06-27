package booking

import (
	"strings"
	"testing"
)

func TestFormatVideoConfirmedMessage_WithAndWithoutZoom(t *testing.T) {
	t.Parallel()

	withZoom := FormatVideoConfirmedMessage("2026-06-10", "14:00", "https://zoom.us/j/1")
	if !strings.Contains(withZoom, "https://zoom.us/j/1") {
		t.Fatal("expected zoom link in message")
	}
	if strings.Contains(withZoom, "будет добавлена позже") {
		t.Fatal("should not mention pending link when zoom is present")
	}

	withoutZoom := FormatVideoConfirmedMessage("2026-06-10", "14:00", "")
	if !strings.Contains(withoutZoom, "будет добавлена позже") {
		t.Fatal("expected pending zoom notice")
	}
}

func TestFormatInPersonConfirmedMessage(t *testing.T) {
	t.Parallel()

	msg := FormatInPersonConfirmedMessage("2026-06-10", "10:00", "Вход со двора")
	if !strings.Contains(msg, "очный приём") || !strings.Contains(msg, "Вход со двора") {
		t.Fatalf("unexpected message: %s", msg)
	}
}

func TestFormatRejectedMessage(t *testing.T) {
	t.Parallel()

	msg := FormatRejectedMessage("2026-06-10", "10:00", "- 2026-06-12 в 14:00")
	if !strings.Contains(msg, "недоступно") || !strings.Contains(msg, "2026-06-12") {
		t.Fatalf("unexpected message: %s", msg)
	}
}

func TestFormatDoctorReminderOneDay_VideoMissingZoom(t *testing.T) {
	t.Parallel()

	msg := FormatDoctorReminderOneDay("Иван", "2026-06-10", "14:00", "video", "")
	if !strings.Contains(msg, "Zoom ещё не добавлена") {
		t.Fatalf("expected zoom warning: %s", msg)
	}

	msg = FormatDoctorReminderOneDay("Иван", "2026-06-10", "14:00", "in_person", "")
	if strings.Contains(msg, "Zoom") {
		t.Fatalf("in-person reminder should not mention zoom: %s", msg)
	}
}

func TestFormatPatientReminderOneHour_InPerson(t *testing.T) {
	t.Parallel()

	msg := FormatPatientReminderOneHour("2026-06-10", "14:00", "in_person", "")
	if !strings.Contains(msg, "очный приём") || strings.Contains(msg, "Zoom") {
		t.Fatalf("unexpected message: %s", msg)
	}
}
