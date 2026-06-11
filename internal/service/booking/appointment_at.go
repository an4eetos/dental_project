package booking

import (
	"time"

	"github.com/anuarkuanysh/dental_project/internal/domain/booking"
)

// AppointmentAt combines stored date and time in the given location.
func AppointmentAt(appt booking.Appointment, loc *time.Location) time.Time {
	if loc == nil {
		loc = time.UTC
	}
	y, m, d := appt.PreferredDate.In(loc).Date()
	h, min, _ := appt.PreferredTime.Clock()
	return time.Date(y, m, d, h, min, 0, 0, loc)
}
