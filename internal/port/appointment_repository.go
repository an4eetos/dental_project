package port

import (
	"context"

	"github.com/anuarkuanysh/dental_project/internal/domain/booking"
)

// AppointmentRepository stores patient appointments.
type AppointmentRepository interface {
	Create(ctx context.Context, appt booking.Appointment) (booking.Appointment, error)
	ListByUserID(ctx context.Context, userID int64) ([]booking.Appointment, error)
	ListAllWithPatients(ctx context.Context) ([]booking.AppointmentWithPatient, error)
}
