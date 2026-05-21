package appointment

import (
	"context"

	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
	"github.com/anuarkuanysh/dental_project/internal/domain/booking"
	"github.com/anuarkuanysh/dental_project/internal/domain/identity"
	"github.com/anuarkuanysh/dental_project/internal/port"
)

// ListForDoctor returns all patient appointments for clinic staff.
type ListForDoctor struct {
	Appointments port.AppointmentRepository
	Users        port.UserRepository
}

func (uc *ListForDoctor) Execute(ctx context.Context, doctorUserID int64) ([]booking.AppointmentWithPatient, error) {
	user, err := uc.Users.GetByID(ctx, doctorUserID)
	if err != nil {
		return nil, err
	}
	if user.Role != identity.RoleDoctor {
		return nil, domainerrors.ErrForbidden
	}
	return uc.Appointments.ListAllWithPatients(ctx)
}
