package appointment

import (
	"context"
	"time"

	"github.com/anuarkuanysh/dental_project/internal/domain/booking"
	"github.com/anuarkuanysh/dental_project/internal/port"
	bookingvalidate "github.com/anuarkuanysh/dental_project/internal/service/booking"
)

const (
	reminderOneDay  = 24 * time.Hour
	reminderOneHour = time.Hour
)

// SendReminders delivers one-day and one-hour Telegram reminders for confirmed online visits.
type SendReminders struct {
	Appointments port.AppointmentRepository
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

	items, err := uc.Appointments.ListConfirmedWithZoomLink(ctx)
	if err != nil {
		return err
	}

	for _, item := range items {
		if err := uc.processItem(ctx, item, now, loc); err != nil {
			return err
		}
	}
	return nil
}

func (uc *SendReminders) processItem(
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
	zoomLink := item.Appointment.ZoomLink

	if until <= reminderOneHour && item.Appointment.Reminder1hAt == nil {
		message := bookingvalidate.FormatReminderOneHourMessage(dateStr, timeStr, zoomLink)
		if err := uc.Sender.SendText(ctx, item.Patient.TelegramID, message); err != nil {
			return err
		}
		return uc.Appointments.MarkReminderSent(ctx, item.Appointment.ID, booking.ReminderOneHour, now.UTC())
	}

	if until <= reminderOneDay && item.Appointment.Reminder1dAt == nil {
		message := bookingvalidate.FormatReminderOneDayMessage(dateStr, timeStr, zoomLink)
		if err := uc.Sender.SendText(ctx, item.Patient.TelegramID, message); err != nil {
			return err
		}
		return uc.Appointments.MarkReminderSent(ctx, item.Appointment.ID, booking.ReminderOneDay, now.UTC())
	}

	return nil
}
