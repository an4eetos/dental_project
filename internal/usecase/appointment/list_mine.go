package appointment

import (
	"context"

	"github.com/anuarkuanysh/dental_project/internal/domain/booking"
	"github.com/anuarkuanysh/dental_project/internal/port"
)

// ListMine returns appointments for the authenticated user.
type ListMine struct {
	Appointments port.AppointmentRepository
}

func (uc *ListMine) Execute(ctx context.Context, userID int64) ([]booking.Appointment, error) {
	return uc.Appointments.ListByUserID(ctx, userID)
}
