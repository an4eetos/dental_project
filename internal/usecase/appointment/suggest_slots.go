package appointment

import (
	"context"

	"github.com/anuarkuanysh/dental_project/internal/port"
	bookingvalidate "github.com/anuarkuanysh/dental_project/internal/service/booking"
)

// SuggestSlots generates available slot text for a doctor rejecting / rescheduling.
type SuggestSlots struct {
	Appointments port.AppointmentRepository
	Users        port.UserRepository
	Clock        port.Clock
}

type SuggestSlotsResult struct {
	SuggestedText string
}

func (uc *SuggestSlots) Execute(ctx context.Context, doctorUserID, appointmentID int64) (SuggestSlotsResult, error) {
	if err := requireDoctor(ctx, uc.Users, doctorUserID); err != nil {
		return SuggestSlotsResult{}, err
	}

	item, err := uc.Appointments.GetWithPatientByID(ctx, appointmentID)
	if err != nil {
		return SuggestSlotsResult{}, err
	}

	text := bookingvalidate.SuggestAvailableSlots(uc.Clock.Now(), item.Appointment.PreferredDate)
	return SuggestSlotsResult{SuggestedText: text}, nil
}
