package subscription

import "time"

// Entitlement is a user's subscription access state and current plan terms for purchase UI.
type Entitlement struct {
	Active       bool
	ExpiresAt    *time.Time
	StarsPrice   int
	DurationDays int
}
