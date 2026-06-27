package port

import (
	"context"
	"time"

	"github.com/anuarkuanysh/dental_project/internal/domain/booking"
)

// AppointmentRepository stores patient appointments.
type AppointmentRepository interface {
	Create(ctx context.Context, appt booking.Appointment) (booking.Appointment, error)
	ListByUserID(ctx context.Context, userID int64) ([]booking.Appointment, error)
	ListAllWithPatients(ctx context.Context) ([]booking.AppointmentWithPatient, error)
	GetWithPatientByID(ctx context.Context, id int64) (booking.AppointmentWithPatient, error)
	UpdateRespond(ctx context.Context, update booking.RespondUpdate) error
	UpdateZoomLink(ctx context.Context, update booking.ZoomLinkUpdate) error
	ListConfirmedForReminders(ctx context.Context) ([]booking.AppointmentWithPatient, error)
	ListVideoMissingZoom(ctx context.Context) ([]booking.AppointmentWithPatient, error)
	MarkReminderSent(ctx context.Context, id int64, kind booking.ReminderKind, at time.Time) error
	DeleteScheduledBefore(ctx context.Context, before time.Time, loc *time.Location) (int64, error)
}
