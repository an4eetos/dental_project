package subscription

import (
	"context"
	"errors"
	"time"

	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
	subdomain "github.com/anuarkuanysh/dental_project/internal/domain/subscription"
	"github.com/anuarkuanysh/dental_project/internal/port"
)

// ConfirmPayment records a successful Stars payment and extends subscription access.
type ConfirmPayment struct {
	Subscriptions port.SubscriptionRepository
	Signer        port.InvoicePayloadSigner
	Clock         port.Clock
	Plan          PlanConfig
}

type ConfirmPaymentInput struct {
	Payload                 string
	TelegramPaymentChargeID string
	StarsAmount             int
}

func (uc *ConfirmPayment) Execute(ctx context.Context, input ConfirmPaymentInput) error {
	userID, err := uc.Signer.Verify(input.Payload)
	if err != nil {
		return err
	}

	exists, err := uc.Subscriptions.PaymentExists(ctx, input.TelegramPaymentChargeID)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	_, err = uc.Subscriptions.RecordPayment(ctx, port.RecordPaymentParams{
		UserID:                  userID,
		TelegramPaymentChargeID: input.TelegramPaymentChargeID,
		StarsAmount:             input.StarsAmount,
	})
	if errors.Is(err, domainerrors.ErrPaymentAlreadyRecorded) {
		return nil
	}
	if err != nil {
		return err
	}

	now := uc.Clock.Now()
	var currentExpiry *time.Time
	sub, subErr := uc.Subscriptions.GetByUserID(ctx, userID)
	if subErr == nil {
		currentExpiry = &sub.ExpiresAt
	} else if !errors.Is(subErr, domainerrors.ErrSubscriptionNotFound) {
		return subErr
	}

	expiresAt := computeExpiresAt(currentExpiry, now, uc.Plan.Duration)
	_, err = uc.Subscriptions.ExtendSubscription(ctx, port.ExtendSubscriptionParams{
		UserID:    userID,
		ExpiresAt: expiresAt,
	})
	return err
}

func computeExpiresAt(current *time.Time, now time.Time, duration time.Duration) time.Time {
	base := now
	if current != nil && current.After(now) {
		base = *current
	}
	return base.Add(duration)
}

// IsActiveAt checks subscription validity at a point in time.
func IsActiveAt(sub subdomain.Subscription, now time.Time) bool {
	return subdomain.IsActive(sub, now)
}
