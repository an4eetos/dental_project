package subscription

import "time"

// Subscription grants a user access to premium content until ExpiresAt.
type Subscription struct {
	UserID    int64
	ExpiresAt time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

// IsActive reports whether the subscription is valid at the given time.
func IsActive(sub Subscription, now time.Time) bool {
	return sub.ExpiresAt.After(now)
}
