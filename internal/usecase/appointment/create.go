package appointment

import (
	"context"

	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
	"github.com/anuarkuanysh/dental_project/internal/domain/booking"
	bookingvalidate "github.com/anuarkuanysh/dental_project/internal/service/booking"
	"github.com/anuarkuanysh/dental_project/internal/port"
)

// Create books a new appointment for an authenticated user.
type Create struct {
	Appointments port.AppointmentRepository
	Users        port.UserRepository
	Clock        port.Clock
}

// CreateInput is the usecase input (raw strings validated in service layer).
type CreateInput struct {
	UserID          int64
	PreferredDate   string
	PreferredTime   string
}

func (uc *Create) Execute(ctx context.Context, in CreateInput) (booking.Appointment, error) {
	user, err := uc.Users.GetByID(ctx, in.UserID)
	if err != nil {
		return booking.Appointment{}, err
	}
	if user.Role.IsDoctor() {
		return booking.Appointment{}, domainerrors.ErrForbidden
	}

	now := uc.Clock.Now()
	date, err := bookingvalidate.ParsePreferredDate(in.PreferredDate, now)
	if err != nil {
		return booking.Appointment{}, err
	}
	slot, err := bookingvalidate.ParsePreferredTime(in.PreferredTime)
	if err != nil {
		return booking.Appointment{}, err
	}

	appt := booking.Appointment{
		UserID:        user.ID,
		PreferredDate: date,
		PreferredTime: slot,
		Status:        booking.StatusPending,
		CreatedAt:     now.UTC(),
	}

	return uc.Appointments.Create(ctx, appt)
}
