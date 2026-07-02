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

func TestCreate_NotifiesDoctors(t *testing.T) {
	t.Parallel()

	sender := &recordingSender{}
	repo := &stubAppointmentRepo{
		createFn: func(appt booking.Appointment) booking.Appointment {
			appt.ID = 42
			return appt
		},
	}
	users := &stubUserRepo{users: map[int64]identity.User{10: patientUser(10)}}

	uc := &Create{
		Appointments: repo,
		Users:        users,
		Doctors:      stubDoctorRegistry{ids: []int64{111, 222}},
		Sender:       sender,
		Clock:        fixedClock{now: testNow()},
	}

	appt, err := uc.Execute(context.Background(), CreateInput{
		UserID:             10,
		PreferredDate:      "2026-06-10",
		PreferredTime:      "14:00",
		PreferredVisitType: "video",
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if appt.ID != 42 || appt.Status != booking.StatusPending {
		t.Fatalf("appointment: %+v", appt)
	}
	if appt.PreferredVisitType != booking.VisitTypeVideo {
		t.Fatalf("preferred visit type: %s", appt.PreferredVisitType)
	}
	if len(sender.messages) != 2 {
		t.Fatalf("want 2 doctor notifications, got %d", len(sender.messages))
	}
	for i, wantID := range []int64{111, 222} {
		if sender.messages[i].chatID != wantID {
			t.Fatalf("message %d chatID = %d", i, sender.messages[i].chatID)
		}
		if !strings.Contains(sender.messages[i].text, "Новая заявка") {
			t.Fatalf("message %d: %s", i, sender.messages[i].text)
		}
		if !strings.Contains(sender.messages[i].text, "Предпочтение пациента: Видеоконсультация") {
			t.Fatalf("message %d missing preference: %s", i, sender.messages[i].text)
		}
	}
}

func TestCreate_DoctorCannotBook(t *testing.T) {
	t.Parallel()

	uc := &Create{
		Appointments: &stubAppointmentRepo{},
		Users: &stubUserRepo{users: map[int64]identity.User{
			5: doctorUser(5),
		}},
		Doctors: stubDoctorRegistry{},
		Sender:  &recordingSender{},
		Clock:   fixedClock{now: testNow()},
	}

	_, err := uc.Execute(context.Background(), CreateInput{
		UserID:             5,
		PreferredDate:      "2026-06-10",
		PreferredTime:      "14:00",
		PreferredVisitType: "in_person",
	})
	if !errors.Is(err, domainerrors.ErrForbidden) {
		t.Fatalf("err = %v", err)
	}
}

func TestCreate_InvalidDate(t *testing.T) {
	t.Parallel()

	uc := &Create{
		Appointments: &stubAppointmentRepo{},
		Users:        &stubUserRepo{users: map[int64]identity.User{10: patientUser(10)}},
		Doctors:      stubDoctorRegistry{},
		Sender:       &recordingSender{},
		Clock:        fixedClock{now: testNow()},
	}

	_, err := uc.Execute(context.Background(), CreateInput{
		UserID:             10,
		PreferredDate:      "2020-01-01",
		PreferredTime:      "14:00",
		PreferredVisitType: "in_person",
	})
	if !errors.Is(err, domainerrors.ErrInvalidPreferredDate) {
		t.Fatalf("err = %v", err)
	}
}
