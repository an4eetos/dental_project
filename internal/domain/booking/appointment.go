package booking

import (
	"time"

	"github.com/anuarkuanysh/dental_project/internal/domain/identity"
)

// Appointment is a patient's requested visit slot.
type Appointment struct {
	ID                    int64
	UserID                int64
	PreferredDate         time.Time
	PreferredTime         time.Time
	Status                Status
	VisitType             VisitType
	ZoomLink              string
	DoctorNotes           string
	RespondedBy           *int64
	OfferSentAt           *time.Time
	Reminder1dAt          *time.Time
	Reminder1hAt          *time.Time
	DoctorReminder1dAt    *time.Time
	DoctorReminder1hAt    *time.Time
	ZoomMissingNotifiedAt *time.Time
	CreatedAt             time.Time
}

// RespondUpdate carries a doctor's decision on a pending appointment.
type RespondUpdate struct {
	ID            int64
	Decision      DoctorDecision
	PreferredDate time.Time
	PreferredTime time.Time
	VisitType     VisitType
	ZoomLink      string
	DoctorNotes   string
	RespondedBy   int64
	RespondedAt   time.Time
}

// ZoomLinkUpdate attaches or updates a Zoom URL for a video consultation.
type ZoomLinkUpdate struct {
	ID       int64
	ZoomLink string
}

// CreateInput carries validated booking parameters.
type CreateInput struct {
	User          identity.User
	PreferredDate time.Time
	PreferredTime time.Time
}
