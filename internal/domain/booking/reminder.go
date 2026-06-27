package booking

// ReminderKind identifies which automated reminder was delivered.
type ReminderKind string

const (
	ReminderOneDay        ReminderKind = "1d"
	ReminderOneHour       ReminderKind = "1h"
	DoctorReminderOneDay  ReminderKind = "doctor_1d"
	DoctorReminderOneHour ReminderKind = "doctor_1h"
	ZoomMissingNotified   ReminderKind = "zoom_missing"
)
