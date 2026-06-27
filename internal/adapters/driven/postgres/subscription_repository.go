package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
	"github.com/anuarkuanysh/dental_project/internal/domain/subscription"
	"github.com/anuarkuanysh/dental_project/internal/port"
)

// SubscriptionRepository implements port.SubscriptionRepository.
type SubscriptionRepository struct {
	pool *pgxpool.Pool
}

func NewSubscriptionRepository(pool *pgxpool.Pool) *SubscriptionRepository {
	return &SubscriptionRepository{pool: pool}
}

func (r *SubscriptionRepository) GetByUserID(ctx context.Context, userID int64) (subscription.Subscription, error) {
	const q = `
SELECT user_id, expires_at, created_at, updated_at
FROM subscriptions
WHERE user_id = $1`

	row := r.pool.QueryRow(ctx, q, userID)
	var sub subscription.Subscription
	err := row.Scan(&sub.UserID, &sub.ExpiresAt, &sub.CreatedAt, &sub.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return subscription.Subscription{}, domainerrors.ErrSubscriptionNotFound
	}
	if err != nil {
		return subscription.Subscription{}, err
	}
	return sub, nil
}

func (r *SubscriptionRepository) ExtendSubscription(ctx context.Context, params port.ExtendSubscriptionParams) (subscription.Subscription, error) {
	const q = `
INSERT INTO subscriptions (user_id, expires_at)
VALUES ($1, $2)
ON CONFLICT (user_id) DO UPDATE SET
    expires_at = EXCLUDED.expires_at,
    updated_at = NOW()
RETURNING user_id, expires_at, created_at, updated_at`

	row := r.pool.QueryRow(ctx, q, params.UserID, params.ExpiresAt)
	var sub subscription.Subscription
	if err := row.Scan(&sub.UserID, &sub.ExpiresAt, &sub.CreatedAt, &sub.UpdatedAt); err != nil {
		return subscription.Subscription{}, err
	}
	return sub, nil
}

func (r *SubscriptionRepository) RecordPayment(ctx context.Context, params port.RecordPaymentParams) (subscription.Payment, error) {
	const q = `
INSERT INTO subscription_payments (user_id, telegram_payment_charge_id, stars_amount)
VALUES ($1, $2, $3)
ON CONFLICT (telegram_payment_charge_id) DO NOTHING
RETURNING id, user_id, telegram_payment_charge_id, stars_amount, created_at`

	row := r.pool.QueryRow(ctx, q, params.UserID, params.TelegramPaymentChargeID, params.StarsAmount)
	var p subscription.Payment
	err := row.Scan(&p.ID, &p.UserID, &p.TelegramPaymentChargeID, &p.StarsAmount, &p.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return subscription.Payment{}, domainerrors.ErrPaymentAlreadyRecorded
	}
	if err != nil {
		return subscription.Payment{}, err
	}
	return p, nil
}

func (r *SubscriptionRepository) PaymentExists(ctx context.Context, chargeID string) (bool, error) {
	const q = `SELECT EXISTS(SELECT 1 FROM subscription_payments WHERE telegram_payment_charge_id = $1)`
	var exists bool
	if err := r.pool.QueryRow(ctx, q, chargeID).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}