package booking

import (
	"time"

	"github.com/anuarkuanysh/dental_project/internal/domain/identity"
)

// Appointment is a patient's requested visit slot.
type Appointment struct {
	ID            int64
	UserID        int64
	PreferredDate time.Time
	PreferredTime time.Time
	Status        Status
	ZoomLink      string
	OfferSentAt   *time.Time
	Reminder1dAt  *time.Time
	Reminder1hAt  *time.Time
	CreatedAt     time.Time
}

// OfferUpdate carries doctor-confirmed slot details for a patient.
type OfferUpdate struct {
	ID            int64
	PreferredDate time.Time
	PreferredTime time.Time
	ZoomLink      string
	OfferSentAt   time.Time
}

// CreateInput carries validated booking parameters.
type CreateInput struct {
	User          identity.User
	PreferredDate time.Time
	PreferredTime time.Time
}
