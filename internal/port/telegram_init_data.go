package port

import (
	"context"

	"github.com/anuarkuanysh/dental_project/internal/domain/identity"
)

// TelegramInitDataValidator verifies Telegram Mini App initData.
type TelegramInitDataValidator interface {
	Validate(ctx context.Context, initData string) (identity.TelegramProfile, error)
}
