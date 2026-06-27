package port

import (
	"context"

	"github.com/anuarkuanysh/dental_project/internal/domain/identity"
)

// SubscriptionStatus is the entitlement state for a user.
type SubscriptionStatus struct {
	Active       bool
	ExpiresAt    *string
	StarsPrice   int
	DurationDays int
}

// SubscriptionChecker resolves whether a user has an active subscription.
type SubscriptionChecker interface {
	Check(ctx context.Context, user identity.User) (SubscriptionStatus, error)
}
