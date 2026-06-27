package appointment

import (
	"context"

	"github.com/anuarkuanysh/dental_project/internal/domain/booking"
	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
	"github.com/anuarkuanysh/dental_project/internal/port"
	bookingvalidate "github.com/anuarkuanysh/dental_project/internal/service/booking"
)

// SetZoomLink lets a doctor attach a Zoom URL to a confirmed video consultation.
type SetZoomLink struct {
	Appointments port.AppointmentRepository
	Users        port.UserRepository
	Sender       port.MessageSender
}

// SetZoomLinkInput is the usecase input.
type SetZoomLinkInput struct {
	DoctorUserID  int64
	AppointmentID int64
	ZoomLink      string
}

func (uc *SetZoomLink) Execute(ctx context.Context, in SetZoomLinkInput) (booking.AppointmentWithPatient, error) {
	if err := requireDoctor(ctx, uc.Users, in.DoctorUserID); err != nil {
		return booking.AppointmentWithPatient{}, err
	}

	zoomLink, err := bookingvalidate.ParseZoomLink(in.ZoomLink)
	if err != nil {
		return booking.AppointmentWithPatient{}, err
	}

	item, err := uc.Appointments.GetWithPatientByID(ctx, in.AppointmentID)
	if err != nil {
		return booking.AppointmentWithPatient{}, err
	}
	if item.Appointment.Status != booking.StatusConfirmed {
		return booking.AppointmentWithPatient{}, domainerrors.ErrAppointmentNotFound
	}
	if item.Appointment.VisitType != booking.VisitTypeVideo {
		return booking.AppointmentWithPatient{}, domainerrors.ErrAppointmentNotVideo
	}

	if err := uc.Appointments.UpdateZoomLink(ctx, booking.ZoomLinkUpdate{
		ID:       in.AppointmentID,
		ZoomLink: zoomLink,
	}); err != nil {
		return booking.AppointmentWithPatient{}, err
	}

	dateStr := item.Appointment.PreferredDate.Format(bookingvalidate.DateLayout)
	timeStr := item.Appointment.PreferredTime.Format(bookingvalidate.TimeLayout)
	message := bookingvalidate.FormatZoomLinkAddedMessage(dateStr, timeStr, zoomLink)
	if err := uc.Sender.SendText(ctx, item.Patient.TelegramID, message); err != nil {
		return booking.AppointmentWithPatient{}, err
	}

	item.Appointment.ZoomLink = zoomLink
	item.Appointment.Reminder1dAt = nil
	item.Appointment.Reminder1hAt = nil
	item.Appointment.DoctorReminder1dAt = nil
	item.Appointment.DoctorReminder1hAt = nil
	item.Appointment.ZoomMissingNotifiedAt = nil
	return item, nil
}
