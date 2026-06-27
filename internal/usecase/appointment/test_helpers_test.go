package appointment

import (
	"context"
	"time"

	"github.com/anuarkuanysh/dental_project/internal/domain/admin"
	"github.com/anuarkuanysh/dental_project/internal/domain/booking"
	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
	"github.com/anuarkuanysh/dental_project/internal/domain/identity"
	"github.com/anuarkuanysh/dental_project/internal/port"
)

type fixedClock struct{ now time.Time }

func (c fixedClock) Now() time.Time { return c.now }

type recordingSender struct {
	messages []sentMessage
	err      error
}

type sentMessage struct {
	chatID int64
	text   string
}

func (s *recordingSender) SendText(_ context.Context, chatID int64, text string) error {
	if s.err != nil {
		return s.err
	}
	s.messages = append(s.messages, sentMessage{chatID: chatID, text: text})
	return nil
}

type stubDoctorRegistry struct {
	ids []int64
}

func (r stubDoctorRegistry) IsDoctor(telegramID int64) bool {
	for _, id := range r.ids {
		if id == telegramID {
			return true
		}
	}
	return false
}

func (r stubDoctorRegistry) TelegramIDs() []int64 { return append([]int64(nil), r.ids...) }

type stubUserRepo struct {
	users map[int64]identity.User
}

func (r *stubUserRepo) UpsertByTelegramID(context.Context, port.UpsertUserParams) (identity.User, error) {
	panic("not implemented")
}

func (r *stubUserRepo) GetByID(_ context.Context, id int64) (identity.User, error) {
	u, ok := r.users[id]
	if !ok {
		return identity.User{}, domainerrors.ErrUserNotFound
	}
	return u, nil
}

func (r *stubUserRepo) GetByTelegramID(context.Context, int64) (identity.User, error) {
	panic("not implemented")
}

func (r *stubUserRepo) List(context.Context, port.ListUsersParams) ([]identity.User, error) {
	panic("not implemented")
}

func (r *stubUserRepo) GetOverviewByID(context.Context, int64) (admin.UserOverview, error) {
	panic("not implemented")
}

func (r *stubUserRepo) UpdateByAdmin(context.Context, int64, port.AdminUpdateUserParams) (identity.User, error) {
	panic("not implemented")
}

func (r *stubUserRepo) SetBlocked(context.Context, int64, bool) (identity.User, error) {
	panic("not implemented")
}

type stubAppointmentRepo struct {
	items              map[int64]booking.AppointmentWithPatient
	respondUpdates     []booking.RespondUpdate
	zoomUpdates        []booking.ZoomLinkUpdate
	remindersMarked    []booking.ReminderKind
	getErr             error
	updateRespondErr   error
	updateZoomErr      error
	listConfirmed      []booking.AppointmentWithPatient
	listMissingZoom    []booking.AppointmentWithPatient
	createFn           func(booking.Appointment) booking.Appointment
}

func (r *stubAppointmentRepo) Create(_ context.Context, appt booking.Appointment) (booking.Appointment, error) {
	if r.createFn != nil {
		return r.createFn(appt), nil
	}
	appt.ID = 1
	return appt, nil
}

func (r *stubAppointmentRepo) ListByUserID(context.Context, int64) ([]booking.Appointment, error) {
	panic("not implemented")
}

func (r *stubAppointmentRepo) ListAllWithPatients(context.Context) ([]booking.AppointmentWithPatient, error) {
	panic("not implemented")
}

func (r *stubAppointmentRepo) GetWithPatientByID(_ context.Context, id int64) (booking.AppointmentWithPatient, error) {
	if r.getErr != nil {
		return booking.AppointmentWithPatient{}, r.getErr
	}
	item, ok := r.items[id]
	if !ok {
		return booking.AppointmentWithPatient{}, domainerrors.ErrAppointmentNotFound
	}
	return item, nil
}

func (r *stubAppointmentRepo) UpdateRespond(_ context.Context, update booking.RespondUpdate) error {
	if r.updateRespondErr != nil {
		return r.updateRespondErr
	}
	r.respondUpdates = append(r.respondUpdates, update)
	item := r.items[update.ID]
	item.Appointment.PreferredDate = update.PreferredDate
	item.Appointment.PreferredTime = update.PreferredTime
	item.Appointment.VisitType = update.VisitType
	item.Appointment.ZoomLink = update.ZoomLink
	item.Appointment.DoctorNotes = update.DoctorNotes
	item.Appointment.RespondedBy = &update.RespondedBy
	if update.Decision == booking.DecisionReject {
		item.Appointment.Status = booking.StatusRejected
	} else {
		item.Appointment.Status = booking.StatusConfirmed
	}
	r.items[update.ID] = item
	return nil
}

func (r *stubAppointmentRepo) UpdateZoomLink(_ context.Context, update booking.ZoomLinkUpdate) error {
	if r.updateZoomErr != nil {
		return r.updateZoomErr
	}
	r.zoomUpdates = append(r.zoomUpdates, update)
	item := r.items[update.ID]
	item.Appointment.ZoomLink = update.ZoomLink
	r.items[update.ID] = item
	return nil
}

func (r *stubAppointmentRepo) ListConfirmedForReminders(context.Context) ([]booking.AppointmentWithPatient, error) {
	return append([]booking.AppointmentWithPatient(nil), r.listConfirmed...), nil
}

func (r *stubAppointmentRepo) ListVideoMissingZoom(context.Context) ([]booking.AppointmentWithPatient, error) {
	return append([]booking.AppointmentWithPatient(nil), r.listMissingZoom...), nil
}

func (r *stubAppointmentRepo) MarkReminderSent(_ context.Context, id int64, kind booking.ReminderKind, at time.Time) error {
	r.remindersMarked = append(r.remindersMarked, kind)
	item := r.items[id]
	switch kind {
	case booking.ReminderOneDay:
		item.Appointment.Reminder1dAt = &at
	case booking.ReminderOneHour:
		item.Appointment.Reminder1hAt = &at
	case booking.DoctorReminderOneDay:
		item.Appointment.DoctorReminder1dAt = &at
	case booking.DoctorReminderOneHour:
		item.Appointment.DoctorReminder1hAt = &at
	case booking.ZoomMissingNotified:
		item.Appointment.ZoomMissingNotifiedAt = &at
	}
	r.items[id] = item
	return nil
}

func (r *stubAppointmentRepo) DeleteScheduledBefore(context.Context, time.Time, *time.Location) (int64, error) {
	panic("not implemented")
}

func testNow() time.Time {
	return time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
}

func pendingAppointment(id int64) booking.AppointmentWithPatient {
	return booking.AppointmentWithPatient{
		Appointment: booking.Appointment{
			ID:            id,
			UserID:        10,
			PreferredDate: time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC),
			PreferredTime: time.Date(0, 1, 1, 14, 0, 0, 0, time.UTC),
			Status:        booking.StatusPending,
		},
		Patient: identity.PatientSummary{
			ID:         10,
			TelegramID: 1000,
			FirstName:  "Анна",
			LastName:   "Иванова",
			Username:   "anna",
		},
	}
}

func doctorUser(id int64) identity.User {
	return identity.User{
		ID:         id,
		TelegramID: 9000 + id,
		FirstName:  "Доктор",
		Role:       identity.RoleDoctor,
	}
}

func patientUser(id int64) identity.User {
	return identity.User{
		ID:         id,
		TelegramID: 1000,
		FirstName:  "Анна",
		LastName:   "Иванова",
		Username:   "anna",
		Role:       identity.RolePatient,
	}
}
