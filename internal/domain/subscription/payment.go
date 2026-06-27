package subscription

import "time"

// Payment records a completed Telegram Stars purchase.
type Payment struct {
	ID                       int64
	UserID                   int64
	TelegramPaymentChargeID  string
	StarsAmount              int
	CreatedAt                time.Time
}
