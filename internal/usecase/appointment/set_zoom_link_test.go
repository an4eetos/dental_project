package appointment

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/anuarkuanysh/dental_project/internal/domain/booking"
	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
	"github.com/anuarkuanysh/dental_project/internal/domain/identity"
)

func confirmedVideoAppointment(id int64) booking.AppointmentWithPatient {
	item := pendingAppointment(id)
	item.Appointment.Status = booking.StatusConfirmed
	item.Appointment.VisitType = booking.VisitTypeVideo
	return item
}

func TestSetZoomLink_SendsLinkToPatient(t *testing.T) {
	t.Parallel()

	sender := &recordingSender{}
	repo := &stubAppointmentRepo{
		items: map[int64]booking.AppointmentWithPatient{1: confirmedVideoAppointment(1)},
	}
	users := &stubUserRepo{users: map[int64]identity.User{5: doctorUser(5)}}

	uc := &SetZoomLink{
		Appointments: repo,
		Users:        users,
		Sender:       sender,
	}

	item, err := uc.Execute(context.Background(), SetZoomLinkInput{
		DoctorUserID:  5,
		AppointmentID: 1,
		ZoomLink:      "https://zoom.us/j/abc",
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if item.Appointment.ZoomLink != "https://zoom.us/j/abc" {
		t.Fatalf("zoom link = %q", item.Appointment.ZoomLink)
	}
	if len(sender.messages) != 1 || sender.messages[0].chatID != 1000 {
		t.Fatalf("messages: %+v", sender.messages)
	}
	if !strings.Contains(sender.messages[0].text, "https://zoom.us/j/abc") {
		t.Fatalf("message: %s", sender.messages[0].text)
	}
	if len(repo.zoomUpdates) != 1 {
		t.Fatalf("zoom updates: %+v", repo.zoomUpdates)
	}
}

func TestSetZoomLink_NotVideo_ReturnsError(t *testing.T) {
	t.Parallel()

	item := pendingAppointment(1)
	item.Appointment.Status = booking.StatusConfirmed
	item.Appointment.VisitType = booking.VisitTypeInPerson

	uc := &SetZoomLink{
		Appointments: &stubAppointmentRepo{items: map[int64]booking.AppointmentWithPatient{1: item}},
		Users:        &stubUserRepo{users: map[int64]identity.User{5: doctorUser(5)}},
		Sender:       &recordingSender{},
	}

	_, err := uc.Execute(context.Background(), SetZoomLinkInput{
		DoctorUserID:  5,
		AppointmentID: 1,
		ZoomLink:      "https://zoom.us/j/abc",
	})
	if !errors.Is(err, domainerrors.ErrAppointmentNotVideo) {
		t.Fatalf("err = %v", err)
	}
}

func TestSetZoomLink_InvalidLink(t *testing.T) {
	t.Parallel()

	uc := &SetZoomLink{
		Appointments: &stubAppointmentRepo{
			items: map[int64]booking.AppointmentWithPatient{1: confirmedVideoAppointment(1)},
		},
		Users:  &stubUserRepo{users: map[int64]identity.User{5: doctorUser(5)}},
		Sender: &recordingSender{},
	}

	_, err := uc.Execute(context.Background(), SetZoomLinkInput{
		DoctorUserID:  5,
		AppointmentID: 1,
		ZoomLink:      "bad",
	})
	if !errors.Is(err, domainerrors.ErrInvalidZoomLink) {
		t.Fatalf("err = %v", err)
	}
}

func TestSetZoomLink_NotConfirmed(t *testing.T) {
	t.Parallel()

	uc := &SetZoomLink{
		Appointments: &stubAppointmentRepo{
			items: map[int64]booking.AppointmentWithPatient{1: pendingAppointment(1)},
		},
		Users:  &stubUserRepo{users: map[int64]identity.User{5: doctorUser(5)}},
		Sender: &recordingSender{},
	}

	_, err := uc.Execute(context.Background(), SetZoomLinkInput{
		DoctorUserID:  5,
		AppointmentID: 1,
		ZoomLink:      "https://zoom.us/j/abc",
	})
	if !errors.Is(err, domainerrors.ErrAppointmentNotFound) {
		t.Fatalf("err = %v", err)
	}
}

func TestSendReminders_OneHour_SendsToPatientAndDoctor(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 6, 10, 13, 30, 0, 0, time.UTC)
	item := confirmedVideoAppointment(1)
	item.Appointment.PreferredDate = time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)
	item.Appointment.PreferredTime = time.Date(0, 1, 1, 14, 0, 0, 0, time.UTC)
	item.Appointment.ZoomLink = "https://zoom.us/j/1"
	doctorID := int64(5)
	item.Appointment.RespondedBy = &doctorID

	repo := &stubAppointmentRepo{
		items:         map[int64]booking.AppointmentWithPatient{1: item},
		listConfirmed: []booking.AppointmentWithPatient{item},
	}
	sender := &recordingSender{}
	users := &stubUserRepo{users: map[int64]identity.User{5: doctorUser(5)}}

	uc := &SendReminders{
		Appointments: repo,
		Users:        users,
		Sender:       sender,
		Clock:        fixedClock{now: now},
		Location:     time.UTC,
	}

	if err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if len(sender.messages) < 2 {
		t.Fatalf("want at least patient+doctor messages, got %d: %+v", len(sender.messages), sender.messages)
	}

	var patientMsg, doctorMsg bool
	for _, m := range sender.messages {
		if m.chatID == 1000 {
			patientMsg = true
		}
		if m.chatID == doctorUser(5).TelegramID {
			doctorMsg = true
		}
	}
	if !patientMsg || !doctorMsg {
		t.Fatalf("messages: %+v", sender.messages)
	}
}

func TestSendReminders_NotifyMissingZoom_Within24Hours(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 6, 10, 10, 0, 0, 0, time.UTC)
	item := confirmedVideoAppointment(1)
	item.Appointment.PreferredDate = time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)
	item.Appointment.PreferredTime = time.Date(0, 1, 1, 20, 0, 0, 0, time.UTC)
	doctorID := int64(5)
	item.Appointment.RespondedBy = &doctorID

	repo := &stubAppointmentRepo{
		items:           map[int64]booking.AppointmentWithPatient{1: item},
		listMissingZoom: []booking.AppointmentWithPatient{item},
	}
	sender := &recordingSender{}
	users := &stubUserRepo{users: map[int64]identity.User{5: doctorUser(5)}}

	uc := &SendReminders{
		Appointments: repo,
		Users:        users,
		Sender:       sender,
		Clock:        fixedClock{now: now},
		Location:     time.UTC,
	}

	if err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	found := false
	for _, m := range sender.messages {
		if strings.Contains(m.text, "Нужна ссылка на Zoom") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected zoom missing message, got: %+v", sender.messages)
	}
	if len(repo.remindersMarked) == 0 || repo.remindersMarked[len(repo.remindersMarked)-1] != booking.ZoomMissingNotified {
		t.Fatalf("reminders marked: %v", repo.remindersMarked)
	}
}

func TestSendReminders_OneDay_SendsPatientReminder(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 6, 9, 14, 0, 0, 0, time.UTC)
	item := confirmedVideoAppointment(1)
	item.Appointment.PreferredDate = time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)
	item.Appointment.PreferredTime = time.Date(0, 1, 1, 14, 0, 0, 0, time.UTC)
	item.Appointment.VisitType = booking.VisitTypeInPerson

	repo := &stubAppointmentRepo{
		items:         map[int64]booking.AppointmentWithPatient{1: item},
		listConfirmed: []booking.AppointmentWithPatient{item},
	}
	sender := &recordingSender{}

	uc := &SendReminders{
		Appointments: repo,
		Users:        &stubUserRepo{},
		Sender:       sender,
		Clock:        fixedClock{now: now},
		Location:     time.UTC,
	}

	if err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if len(sender.messages) != 1 || sender.messages[0].chatID != 1000 {
		t.Fatalf("messages: %+v", sender.messages)
	}
	if !strings.Contains(sender.messages[0].text, "очный приём") {
		t.Fatalf("message: %s", sender.messages[0].text)
	}
}

func TestSendReminders_NotifyMissingZoom_SkipsBeyond24Hours(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 6, 8, 10, 0, 0, 0, time.UTC)
	item := confirmedVideoAppointment(1)
	item.Appointment.PreferredDate = time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)
	item.Appointment.PreferredTime = time.Date(0, 1, 1, 20, 0, 0, 0, time.UTC)
	doctorID := int64(5)
	item.Appointment.RespondedBy = &doctorID

	repo := &stubAppointmentRepo{listMissingZoom: []booking.AppointmentWithPatient{item}}
	sender := &recordingSender{}
	users := &stubUserRepo{users: map[int64]identity.User{5: doctorUser(5)}}

	uc := &SendReminders{
		Appointments: repo,
		Users:        users,
		Sender:       sender,
		Clock:        fixedClock{now: now},
		Location:     time.UTC,
	}

	if err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if len(sender.messages) != 0 {
		t.Fatalf("expected no messages beyond 24h window, got %+v", sender.messages)
	}
}

func TestSuggestSlots_ReturnsGeneratedText(t *testing.T) {
	t.Parallel()

	repo := &stubAppointmentRepo{
		items: map[int64]booking.AppointmentWithPatient{1: pendingAppointment(1)},
	}
	users := &stubUserRepo{users: map[int64]identity.User{5: doctorUser(5)}}

	uc := &SuggestSlots{
		Appointments: repo,
		Users:        users,
		Clock:        fixedClock{now: testNow()},
	}

	result, err := uc.Execute(context.Background(), 5, 1)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.HasPrefix(result.SuggestedText, "Доступные варианты") {
		t.Fatalf("text: %s", result.SuggestedText)
	}
}

func TestSuggestSlots_NonDoctor_Forbidden(t *testing.T) {
	t.Parallel()

	uc := &SuggestSlots{
		Appointments: &stubAppointmentRepo{},
		Users: &stubUserRepo{users: map[int64]identity.User{
			10: patientUser(10),
		}},
		Clock: fixedClock{now: testNow()},
	}

	_, err := uc.Execute(context.Background(), 10, 1)
	if !errors.Is(err, domainerrors.ErrForbidden) {
		t.Fatalf("err = %v", err)
	}
}
