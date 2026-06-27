package subscription

import (
	"context"
	"errors"
	"time"

	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
	"github.com/anuarkuanysh/dental_project/internal/domain/identity"
	subdomain "github.com/anuarkuanysh/dental_project/internal/domain/subscription"
	"github.com/anuarkuanysh/dental_project/internal/port"
)

// GetStatus returns subscription entitlement for a user.
type GetStatus struct {
	Subscriptions port.SubscriptionRepository
	Clock         port.Clock
	Plan          PlanConfig
}

func (uc *GetStatus) Execute(ctx context.Context, user identity.User) (subdomain.Entitlement, error) {
	durationDays := int(uc.Plan.Duration.Hours() / 24)
	if durationDays < 1 {
		durationDays = 1
	}
	base := subdomain.Entitlement{
		StarsPrice:   uc.Plan.StarsPrice,
		DurationDays: durationDays,
	}

	if user.Role == identity.RoleDoctor || user.Role == identity.RoleAdmin {
		base.Active = true
		return base, nil
	}

	sub, err := uc.Subscriptions.GetByUserID(ctx, user.ID)
	if errors.Is(err, domainerrors.ErrSubscriptionNotFound) {
		return base, nil
	}
	if err != nil {
		return subdomain.Entitlement{}, err
	}

	now := uc.Clock.Now()
	if subdomain.IsActive(sub, now) {
		base.Active = true
		exp := sub.ExpiresAt
		base.ExpiresAt = &exp
	}
	return base, nil
}

// Checker adapts GetStatus to port.SubscriptionChecker.
type Checker struct {
	GetStatus *GetStatus
	Users     port.UserRepository
}

var _ port.SubscriptionChecker = (*Checker)(nil)

func (c *Checker) Check(ctx context.Context, user identity.User) (port.SubscriptionStatus, error) {
	entitlement, err := c.GetStatus.Execute(ctx, user)
	if err != nil {
		return port.SubscriptionStatus{}, err
	}
	return toPortSubscriptionStatus(entitlement), nil
}

func (c *Checker) CheckByUserID(ctx context.Context, userID int64) (port.SubscriptionStatus, error) {
	user, err := c.Users.GetByID(ctx, userID)
	if err != nil {
		return port.SubscriptionStatus{}, err
	}
	return c.Check(ctx, user)
}

func toPortSubscriptionStatus(entitlement subdomain.Entitlement) port.SubscriptionStatus {
	status := port.SubscriptionStatus{
		Active:       entitlement.Active,
		StarsPrice:   entitlement.StarsPrice,
		DurationDays: entitlement.DurationDays,
	}
	if entitlement.ExpiresAt != nil {
		formatted := entitlement.ExpiresAt.UTC().Format(time.RFC3339)
		status.ExpiresAt = &formatted
	}
	return status
}
