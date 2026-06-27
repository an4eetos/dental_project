package appointment

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/anuarkuanysh/dental_project/internal/domain/booking"
	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
	"github.com/anuarkuanysh/dental_project/internal/domain/identity"
)

func TestRespond_VideoWithoutZoom_NotifiesPatientAndDoctor(t *testing.T) {
	t.Parallel()

	sender := &recordingSender{}
	repo := &stubAppointmentRepo{
		items: map[int64]booking.AppointmentWithPatient{1: pendingAppointment(1)},
	}
	users := &stubUserRepo{users: map[int64]identity.User{
		5: doctorUser(5),
	}}

	uc := &Respond{
		Appointments: repo,
		Users:        users,
		Sender:       sender,
		Clock:        fixedClock{now: testNow()},
	}

	item, err := uc.Execute(context.Background(), RespondInput{
		DoctorUserID:  5,
		AppointmentID: 1,
		Decision:      "video",
		PreferredDate: "2026-06-10",
		PreferredTime: "14:00",
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if item.Appointment.Status != booking.StatusConfirmed {
		t.Fatalf("status = %s", item.Appointment.Status)
	}
	if item.Appointment.VisitType != booking.VisitTypeVideo {
		t.Fatalf("visit type = %s", item.Appointment.VisitType)
	}
	if len(sender.messages) != 2 {
		t.Fatalf("want 2 messages, got %d", len(sender.messages))
	}
	if sender.messages[0].chatID != 1000 {
		t.Fatalf("patient chat id = %d", sender.messages[0].chatID)
	}
	if !strings.Contains(sender.messages[0].text, "будет добавлена позже") {
		t.Fatalf("patient message: %s", sender.messages[0].text)
	}
	if sender.messages[1].chatID != doctorUser(5).TelegramID {
		t.Fatalf("doctor chat id = %d", sender.messages[1].chatID)
	}
	if !strings.Contains(sender.messages[1].text, "Нужна ссылка на Zoom") {
		t.Fatalf("doctor message: %s", sender.messages[1].text)
	}
	if len(repo.remindersMarked) != 1 || repo.remindersMarked[0] != booking.ZoomMissingNotified {
		t.Fatalf("reminders marked: %v", repo.remindersMarked)
	}
}

func TestRespond_Reject_RequiresNotes(t *testing.T) {
	t.Parallel()

	repo := &stubAppointmentRepo{
		items: map[int64]booking.AppointmentWithPatient{1: pendingAppointment(1)},
	}
	users := &stubUserRepo{users: map[int64]identity.User{5: doctorUser(5)}}

	uc := &Respond{
		Appointments: repo,
		Users:        users,
		Sender:       &recordingSender{},
		Clock:        fixedClock{now: testNow()},
	}

	_, err := uc.Execute(context.Background(), RespondInput{
		DoctorUserID:  5,
		AppointmentID: 1,
		Decision:      "reject",
		DoctorNotes:   "   ",
	})
	if !errors.Is(err, domainerrors.ErrDoctorNotesRequired) {
		t.Fatalf("err = %v", err)
	}
}

func TestRespond_Reject_NotifiesPatient(t *testing.T) {
	t.Parallel()

	sender := &recordingSender{}
	repo := &stubAppointmentRepo{
		items: map[int64]booking.AppointmentWithPatient{1: pendingAppointment(1)},
	}
	users := &stubUserRepo{users: map[int64]identity.User{5: doctorUser(5)}}

	uc := &Respond{
		Appointments: repo,
		Users:        users,
		Sender:       sender,
		Clock:        fixedClock{now: testNow()},
	}

	item, err := uc.Execute(context.Background(), RespondInput{
		DoctorUserID:  5,
		AppointmentID: 1,
		Decision:      "reject",
		DoctorNotes:   "- 2026-06-12 в 10:00",
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if item.Appointment.Status != booking.StatusRejected {
		t.Fatalf("status = %s", item.Appointment.Status)
	}
	if len(sender.messages) != 1 || !strings.Contains(sender.messages[0].text, "недоступно") {
		t.Fatalf("messages: %+v", sender.messages)
	}
	if len(repo.respondUpdates) != 1 || repo.respondUpdates[0].Decision != booking.DecisionReject {
		t.Fatalf("respond updates: %+v", repo.respondUpdates)
	}
}

func TestRespond_InPerson_ConfirmsAndNotifies(t *testing.T) {
	t.Parallel()

	sender := &recordingSender{}
	repo := &stubAppointmentRepo{
		items: map[int64]booking.AppointmentWithPatient{1: pendingAppointment(1)},
	}
	users := &stubUserRepo{users: map[int64]identity.User{5: doctorUser(5)}}

	uc := &Respond{
		Appointments: repo,
		Users:        users,
		Sender:       sender,
		Clock:        fixedClock{now: testNow()},
	}

	_, err := uc.Execute(context.Background(), RespondInput{
		DoctorUserID:  5,
		AppointmentID: 1,
		Decision:      "in_person",
		PreferredDate: "2026-06-11",
		PreferredTime: "10:00",
		DoctorNotes:   "Кабинет 2",
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if len(sender.messages) != 1 || !strings.Contains(sender.messages[0].text, "очный приём") {
		t.Fatalf("message: %+v", sender.messages)
	}
	if repo.respondUpdates[0].VisitType != booking.VisitTypeInPerson {
		t.Fatalf("visit type = %s", repo.respondUpdates[0].VisitType)
	}
}

func TestRespond_NotPending_ReturnsError(t *testing.T) {
	t.Parallel()

	item := pendingAppointment(1)
	item.Appointment.Status = booking.StatusConfirmed

	uc := &Respond{
		Appointments: &stubAppointmentRepo{items: map[int64]booking.AppointmentWithPatient{1: item}},
		Users:        &stubUserRepo{users: map[int64]identity.User{5: doctorUser(5)}},
		Sender:       &recordingSender{},
		Clock:        fixedClock{now: testNow()},
	}

	_, err := uc.Execute(context.Background(), RespondInput{
		DoctorUserID:  5,
		AppointmentID: 1,
		Decision:      "video",
		PreferredDate: "2026-06-10",
		PreferredTime: "14:00",
	})
	if !errors.Is(err, domainerrors.ErrAppointmentNotPending) {
		t.Fatalf("err = %v", err)
	}
}

func TestRespond_VideoWithZoom_OnlyNotifiesPatient(t *testing.T) {
	t.Parallel()

	sender := &recordingSender{}
	repo := &stubAppointmentRepo{
		items: map[int64]booking.AppointmentWithPatient{1: pendingAppointment(1)},
	}
	users := &stubUserRepo{users: map[int64]identity.User{5: doctorUser(5)}}

	uc := &Respond{
		Appointments: repo,
		Users:        users,
		Sender:       sender,
		Clock:        fixedClock{now: testNow()},
	}

	_, err := uc.Execute(context.Background(), RespondInput{
		DoctorUserID:  5,
		AppointmentID: 1,
		Decision:      "video",
		PreferredDate: "2026-06-10",
		PreferredTime: "14:00",
		ZoomLink:      "https://zoom.us/j/ok",
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if len(sender.messages) != 1 {
		t.Fatalf("want 1 message, got %d", len(sender.messages))
	}
	if !strings.Contains(sender.messages[0].text, "https://zoom.us/j/ok") {
		t.Fatalf("message: %s", sender.messages[0].text)
	}
	if len(repo.remindersMarked) != 0 {
		t.Fatalf("unexpected reminders: %v", repo.remindersMarked)
	}
}

func TestPatientDisplayName(t *testing.T) {
	t.Parallel()

	name := patientDisplayName(pendingAppointment(1))
	if name != "Анна Иванова (@anna)" {
		t.Fatalf("name = %q", name)
	}

	item := pendingAppointment(1)
	item.Patient.Username = ""
	if patientDisplayName(item) != "Анна Иванова" {
		t.Fatalf("name without username = %q", patientDisplayName(item))
	}
}

func TestRespond_NonDoctor_Forbidden(t *testing.T) {
	t.Parallel()

	uc := &Respond{
		Appointments: &stubAppointmentRepo{},
		Users: &stubUserRepo{users: map[int64]identity.User{
			10: patientUser(10),
		}},
		Sender: &recordingSender{},
		Clock:  fixedClock{now: testNow()},
	}

	_, err := uc.Execute(context.Background(), RespondInput{
		DoctorUserID:  10,
		AppointmentID: 1,
		Decision:      "video",
		PreferredDate: "2026-06-10",
		PreferredTime: "14:00",
	})
	if !errors.Is(err, domainerrors.ErrForbidden) {
		t.Fatalf("err = %v", err)
	}
}
