package appointment

import (
	"context"
	"time"

	"github.com/anuarkuanysh/dental_project/internal/domain/booking"
	"github.com/anuarkuanysh/dental_project/internal/port"
	bookingvalidate "github.com/anuarkuanysh/dental_project/internal/service/booking"
)

const (
	reminderOneDay        = 24 * time.Hour
	reminderOneHour       = time.Hour
	zoomMissingLeadWindow = 24 * time.Hour
)

// SendReminders delivers reminders to patients and doctors, and nudges doctors about missing Zoom links.
type SendReminders struct {
	Appointments port.AppointmentRepository
	Users        port.UserRepository
	Sender       port.MessageSender
	Clock        port.Clock
	Location     *time.Location
}

func (uc *SendReminders) Execute(ctx context.Context) error {
	now := uc.Clock.Now()
	loc := uc.Location
	if loc == nil {
		loc = time.UTC
	}

	if err := uc.sendAppointmentReminders(ctx, now, loc); err != nil {
		return err
	}
	return uc.notifyMissingZoomLinks(ctx, now, loc)
}

func (uc *SendReminders) sendAppointmentReminders(
	ctx context.Context,
	now time.Time,
	loc *time.Location,
) error {
	items, err := uc.Appointments.ListConfirmedForReminders(ctx)
	if err != nil {
		return err
	}

	for _, item := range items {
		if err := uc.processReminderItem(ctx, item, now, loc); err != nil {
			return err
		}
	}
	return nil
}

func (uc *SendReminders) processReminderItem(
	ctx context.Context,
	item booking.AppointmentWithPatient,
	now time.Time,
	loc *time.Location,
) error {
	apptAt := bookingvalidate.AppointmentAt(item.Appointment, loc)
	if !apptAt.After(now) {
		return nil
	}

	until := apptAt.Sub(now)
	dateStr := item.Appointment.PreferredDate.Format(bookingvalidate.DateLayout)
	timeStr := item.Appointment.PreferredTime.Format(bookingvalidate.TimeLayout)
	visitType := item.Appointment.VisitType.String()
	zoomLink := item.Appointment.ZoomLink
	patientName := patientDisplayName(item)

	if until <= reminderOneHour && item.Appointment.Reminder1hAt == nil {
		message := bookingvalidate.FormatPatientReminderOneHour(dateStr, timeStr, visitType, zoomLink)
		if err := uc.Sender.SendText(ctx, item.Patient.TelegramID, message); err != nil {
			return err
		}
		if err := uc.Appointments.MarkReminderSent(ctx, item.Appointment.ID, booking.ReminderOneHour, now.UTC()); err != nil {
			return err
		}
	}

	if until <= reminderOneDay && item.Appointment.Reminder1dAt == nil {
		message := bookingvalidate.FormatPatientReminderOneDay(dateStr, timeStr, visitType, zoomLink)
		if err := uc.Sender.SendText(ctx, item.Patient.TelegramID, message); err != nil {
			return err
		}
		if err := uc.Appointments.MarkReminderSent(ctx, item.Appointment.ID, booking.ReminderOneDay, now.UTC()); err != nil {
			return err
		}
	}

	doctorTelegramID, ok, err := uc.doctorTelegramID(ctx, item.Appointment.RespondedBy)
	if err != nil || !ok {
		return err
	}

	if until <= reminderOneHour && item.Appointment.DoctorReminder1hAt == nil {
		message := bookingvalidate.FormatDoctorReminderOneHour(patientName, dateStr, timeStr, visitType, zoomLink)
		if err := uc.Sender.SendText(ctx, doctorTelegramID, message); err != nil {
			return err
		}
		if err := uc.Appointments.MarkReminderSent(ctx, item.Appointment.ID, booking.DoctorReminderOneHour, now.UTC()); err != nil {
			return err
		}
	}

	if until <= reminderOneDay && item.Appointment.DoctorReminder1dAt == nil {
		message := bookingvalidate.FormatDoctorReminderOneDay(patientName, dateStr, timeStr, visitType, zoomLink)
		if err := uc.Sender.SendText(ctx, doctorTelegramID, message); err != nil {
			return err
		}
		if err := uc.Appointments.MarkReminderSent(ctx, item.Appointment.ID, booking.DoctorReminderOneDay, now.UTC()); err != nil {
			return err
		}
	}

	return nil
}

func (uc *SendReminders) notifyMissingZoomLinks(ctx context.Context, now time.Time, loc *time.Location) error {
	items, err := uc.Appointments.ListVideoMissingZoom(ctx)
	if err != nil {
		return err
	}

	for _, item := range items {
		apptAt := bookingvalidate.AppointmentAt(item.Appointment, loc)
		if !apptAt.After(now) {
			continue
		}
		if apptAt.Sub(now) > zoomMissingLeadWindow {
			continue
		}

		doctorTelegramID, ok, err := uc.doctorTelegramID(ctx, item.Appointment.RespondedBy)
		if err != nil || !ok {
			continue
		}

		dateStr := item.Appointment.PreferredDate.Format(bookingvalidate.DateLayout)
		timeStr := item.Appointment.PreferredTime.Format(bookingvalidate.TimeLayout)
		message := bookingvalidate.FormatDoctorZoomMissingMessage(patientDisplayName(item), dateStr, timeStr)
		if err := uc.Sender.SendText(ctx, doctorTelegramID, message); err != nil {
			return err
		}
		if err := uc.Appointments.MarkReminderSent(ctx, item.Appointment.ID, booking.ZoomMissingNotified, now.UTC()); err != nil {
			return err
		}
	}
	return nil
}

func (uc *SendReminders) doctorTelegramID(ctx context.Context, doctorUserID *int64) (int64, bool, error) {
	if doctorUserID == nil || *doctorUserID <= 0 {
		return 0, false, nil
	}
	doctor, err := uc.Users.GetByID(ctx, *doctorUserID)
	if err != nil {
		return 0, false, err
	}
	return doctor.TelegramID, true, nil
}
