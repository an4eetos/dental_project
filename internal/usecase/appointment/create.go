package appointment

import (
	"context"
	"strings"

	"github.com/anuarkuanysh/dental_project/internal/domain/booking"
	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
	"github.com/anuarkuanysh/dental_project/internal/domain/identity"
	"github.com/anuarkuanysh/dental_project/internal/port"
	bookingvalidate "github.com/anuarkuanysh/dental_project/internal/service/booking"
)

// Create books a new appointment for an authenticated user.
type Create struct {
	Appointments port.AppointmentRepository
	Users        port.UserRepository
	Doctors      port.DoctorRegistry
	Sender       port.MessageSender
	Clock        port.Clock
}

// CreateInput is the usecase input (raw strings validated in service layer).
type CreateInput struct {
	UserID              int64
	PreferredDate       string
	PreferredTime       string
	PreferredVisitType  string
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
	preferredVisitType, err := bookingvalidate.ParsePreferredVisitType(in.PreferredVisitType)
	if err != nil {
		return booking.Appointment{}, err
	}

	appt := booking.Appointment{
		UserID:             user.ID,
		PreferredDate:      date,
		PreferredTime:      slot,
		PreferredVisitType: preferredVisitType,
		Status:             booking.StatusPending,
		CreatedAt:          now.UTC(),
	}

	created, err := uc.Appointments.Create(ctx, appt)
	if err != nil {
		return booking.Appointment{}, err
	}

	dateStr := date.Format(bookingvalidate.DateLayout)
	timeStr := slot.Format(bookingvalidate.TimeLayout)
	patientName := formatPatientName(user)
	doctorMessage := bookingvalidate.FormatNewRequestDoctorMessage(
		patientName,
		dateStr,
		timeStr,
		preferredVisitType.String(),
	)
	for _, telegramID := range uc.Doctors.TelegramIDs() {
		_ = uc.Sender.SendText(ctx, telegramID, doctorMessage)
	}

	return created, nil
}

func formatPatientName(user identity.User) string {
	parts := []string{user.FirstName, user.LastName}
	name := strings.TrimSpace(strings.Join(parts, " "))
	if user.Username != "" {
		return name + " (@" + user.Username + ")"
	}
	return name
}
