package appointment

import (
	"context"

	"github.com/anuarkuanysh/dental_project/internal/domain/booking"
	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
	"github.com/anuarkuanysh/dental_project/internal/port"
	bookingvalidate "github.com/anuarkuanysh/dental_project/internal/service/booking"
)

// Offer lets a doctor confirm a slot, attach a Zoom link, and notify the patient.
type Offer struct {
	Appointments port.AppointmentRepository
	Users        port.UserRepository
	Sender       port.MessageSender
	Clock        port.Clock
}

// OfferInput is the usecase input (raw strings validated in service layer).
type OfferInput struct {
	DoctorUserID  int64
	AppointmentID int64
	PreferredDate string
	PreferredTime string
	ZoomLink      string
}

func (uc *Offer) Execute(ctx context.Context, in OfferInput) (booking.AppointmentWithPatient, error) {
	if err := requireDoctor(ctx, uc.Users, in.DoctorUserID); err != nil {
		return booking.AppointmentWithPatient{}, err
	}

	item, err := uc.Appointments.GetWithPatientByID(ctx, in.AppointmentID)
	if err != nil {
		return booking.AppointmentWithPatient{}, err
	}
	if item.Appointment.Status == booking.StatusCancelled {
		return booking.AppointmentWithPatient{}, domainerrors.ErrAppointmentCancelled
	}

	now := uc.Clock.Now()
	date, err := bookingvalidate.ParsePreferredDate(in.PreferredDate, now)
	if err != nil {
		return booking.AppointmentWithPatient{}, err
	}
	slot, err := bookingvalidate.ParsePreferredTime(in.PreferredTime)
	if err != nil {
		return booking.AppointmentWithPatient{}, err
	}
	zoomLink, err := bookingvalidate.ParseZoomLink(in.ZoomLink)
	if err != nil {
		return booking.AppointmentWithPatient{}, err
	}

	offerSentAt := now.UTC()
	update := booking.OfferUpdate{
		ID:            in.AppointmentID,
		PreferredDate: date,
		PreferredTime: slot,
		ZoomLink:      zoomLink,
		OfferSentAt:   offerSentAt,
	}
	if err := uc.Appointments.UpdateOffer(ctx, update); err != nil {
		return booking.AppointmentWithPatient{}, err
	}

	dateStr := date.Format(bookingvalidate.DateLayout)
	timeStr := slot.Format(bookingvalidate.TimeLayout)
	message := bookingvalidate.FormatOfferMessage(dateStr, timeStr, zoomLink)
	if err := uc.Sender.SendText(ctx, item.Patient.TelegramID, message); err != nil {
		return booking.AppointmentWithPatient{}, err
	}

	item.Appointment.PreferredDate = date
	item.Appointment.PreferredTime = slot
	item.Appointment.Status = booking.StatusConfirmed
	item.Appointment.ZoomLink = zoomLink
	item.Appointment.OfferSentAt = &offerSentAt
	item.Appointment.Reminder1dAt = nil
	item.Appointment.Reminder1hAt = nil
	return item, nil
}
