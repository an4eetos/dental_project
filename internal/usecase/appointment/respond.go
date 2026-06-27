package appointment

import (
	"context"
	"fmt"
	"strings"

	"github.com/anuarkuanysh/dental_project/internal/domain/booking"
	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
	"github.com/anuarkuanysh/dental_project/internal/port"
	bookingvalidate "github.com/anuarkuanysh/dental_project/internal/service/booking"
)

// Respond lets a doctor accept in person, accept video, or reject with reschedule notes.
type Respond struct {
	Appointments port.AppointmentRepository
	Users        port.UserRepository
	Sender       port.MessageSender
	Clock        port.Clock
}

// RespondInput is the usecase input (raw strings validated in service layer).
type RespondInput struct {
	DoctorUserID  int64
	AppointmentID int64
	Decision      string
	PreferredDate string
	PreferredTime string
	ZoomLink      string
	DoctorNotes   string
}

func (uc *Respond) Execute(ctx context.Context, in RespondInput) (booking.AppointmentWithPatient, error) {
	if err := requireDoctor(ctx, uc.Users, in.DoctorUserID); err != nil {
		return booking.AppointmentWithPatient{}, err
	}

	decision, err := bookingvalidate.ParseDecision(in.Decision)
	if err != nil {
		return booking.AppointmentWithPatient{}, err
	}

	item, err := uc.Appointments.GetWithPatientByID(ctx, in.AppointmentID)
	if err != nil {
		return booking.AppointmentWithPatient{}, err
	}
	if item.Appointment.Status != booking.StatusPending {
		return booking.AppointmentWithPatient{}, domainerrors.ErrAppointmentNotPending
	}

	now := uc.Clock.Now()
	var date = item.Appointment.PreferredDate
	var slot = item.Appointment.PreferredTime
	var zoomLink string
	var doctorNotes string
	var visitType booking.VisitType

	switch decision {
	case booking.DecisionReject:
		doctorNotes, err = bookingvalidate.ParseDoctorNotes(in.DoctorNotes)
		if err != nil {
			return booking.AppointmentWithPatient{}, err
		}
	case booking.DecisionInPerson:
		date, err = bookingvalidate.ParsePreferredDate(in.PreferredDate, now)
		if err != nil {
			return booking.AppointmentWithPatient{}, err
		}
		slot, err = bookingvalidate.ParsePreferredTime(in.PreferredTime)
		if err != nil {
			return booking.AppointmentWithPatient{}, err
		}
		doctorNotes = strings.TrimSpace(in.DoctorNotes)
		visitType = booking.VisitTypeInPerson
	case booking.DecisionVideo:
		date, err = bookingvalidate.ParsePreferredDate(in.PreferredDate, now)
		if err != nil {
			return booking.AppointmentWithPatient{}, err
		}
		slot, err = bookingvalidate.ParsePreferredTime(in.PreferredTime)
		if err != nil {
			return booking.AppointmentWithPatient{}, err
		}
		zoomLink, err = bookingvalidate.ParseOptionalZoomLink(in.ZoomLink)
		if err != nil {
			return booking.AppointmentWithPatient{}, err
		}
		doctorNotes = strings.TrimSpace(in.DoctorNotes)
		visitType = booking.VisitTypeVideo
	}

	respondedAt := now.UTC()
	update := booking.RespondUpdate{
		ID:            in.AppointmentID,
		Decision:      decision,
		PreferredDate: date,
		PreferredTime: slot,
		VisitType:     visitType,
		ZoomLink:      zoomLink,
		DoctorNotes:   doctorNotes,
		RespondedBy:   in.DoctorUserID,
		RespondedAt:   respondedAt,
	}
	if err := uc.Appointments.UpdateRespond(ctx, update); err != nil {
		return booking.AppointmentWithPatient{}, err
	}

	dateStr := date.Format(bookingvalidate.DateLayout)
	timeStr := slot.Format(bookingvalidate.TimeLayout)
	origDateStr := item.Appointment.PreferredDate.Format(bookingvalidate.DateLayout)
	origTimeStr := item.Appointment.PreferredTime.Format(bookingvalidate.TimeLayout)

	var message string
	switch decision {
	case booking.DecisionInPerson:
		message = bookingvalidate.FormatInPersonConfirmedMessage(dateStr, timeStr, doctorNotes)
	case booking.DecisionVideo:
		message = bookingvalidate.FormatVideoConfirmedMessage(dateStr, timeStr, zoomLink)
	case booking.DecisionReject:
		message = bookingvalidate.FormatRejectedMessage(origDateStr, origTimeStr, doctorNotes)
	}
	if err := uc.Sender.SendText(ctx, item.Patient.TelegramID, message); err != nil {
		return booking.AppointmentWithPatient{}, err
	}

	if decision == booking.DecisionVideo && zoomLink == "" {
		doctor, err := uc.Users.GetByID(ctx, in.DoctorUserID)
		if err != nil {
			return booking.AppointmentWithPatient{}, err
		}
		reminder := bookingvalidate.FormatDoctorZoomMissingMessage(
			patientDisplayName(item),
			dateStr,
			timeStr,
		)
		if err := uc.Sender.SendText(ctx, doctor.TelegramID, reminder); err != nil {
			return booking.AppointmentWithPatient{}, err
		}
		notifiedAt := now.UTC()
		if err := uc.Appointments.MarkReminderSent(ctx, in.AppointmentID, booking.ZoomMissingNotified, notifiedAt); err != nil {
			return booking.AppointmentWithPatient{}, err
		}
		item.Appointment.ZoomMissingNotifiedAt = &notifiedAt
	}

	item.Appointment.PreferredDate = date
	item.Appointment.PreferredTime = slot
	item.Appointment.ZoomLink = zoomLink
	item.Appointment.DoctorNotes = doctorNotes
	item.Appointment.VisitType = visitType
	item.Appointment.RespondedBy = &in.DoctorUserID
	item.Appointment.OfferSentAt = &respondedAt
	item.Appointment.Reminder1dAt = nil
	item.Appointment.Reminder1hAt = nil
	item.Appointment.DoctorReminder1dAt = nil
	item.Appointment.DoctorReminder1hAt = nil
	item.Appointment.ZoomMissingNotifiedAt = nil

	switch decision {
	case booking.DecisionReject:
		item.Appointment.Status = booking.StatusRejected
	default:
		item.Appointment.Status = booking.StatusConfirmed
	}

	return item, nil
}

func patientDisplayName(item booking.AppointmentWithPatient) string {
	parts := []string{item.Patient.FirstName, item.Patient.LastName}
	name := strings.TrimSpace(strings.Join(parts, " "))
	if item.Patient.Username != "" {
		return fmt.Sprintf("%s (@%s)", name, item.Patient.Username)
	}
	return name
}
