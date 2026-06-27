package admin

import (
	"time"

	"github.com/anuarkuanysh/dental_project/internal/domain/identity"
)

// UserOverview is an admin-facing view of a user with activity counts.
type UserOverview struct {
	ID                   int64
	TelegramID           int64
	Username             string
	FirstName            string
	LastName             string
	AvatarURL            string
	Role                 identity.Role
	Blocked              bool
	AppointmentCount     int64
	PhotoSubmissionCount int64
	CreatedAt            time.Time
	UpdatedAt            time.Time
}
