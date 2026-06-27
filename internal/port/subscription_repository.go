package port

import (
	"context"
	"time"

	"github.com/anuarkuanysh/dental_project/internal/domain/subscription"
)

// ExtendSubscriptionParams extends or creates a subscription for a user.
type ExtendSubscriptionParams struct {
	UserID    int64
	ExpiresAt time.Time
}

// RecordPaymentParams stores a completed Stars payment.
type RecordPaymentParams struct {
	UserID                  int64
	TelegramPaymentChargeID string
	StarsAmount             int
}

// SubscriptionRepository persists subscription entitlements and payments.
type SubscriptionRepository interface {
	GetByUserID(ctx context.Context, userID int64) (subscription.Subscription, error)
	ExtendSubscription(ctx context.Context, params ExtendSubscriptionParams) (subscription.Subscription, error)
	RecordPayment(ctx context.Context, params RecordPaymentParams) (subscription.Payment, error)
	PaymentExists(ctx context.Context, chargeID string) (bool, error)
}
