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
	CreatedAt     time.Time
}

// CreateInput carries validated booking parameters.
type CreateInput struct {
	User          identity.User
	PreferredDate time.Time
	PreferredTime time.Time
}
