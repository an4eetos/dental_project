package port

import (
	"context"

	"github.com/anuarkuanysh/dental_project/internal/domain/identity"
)

// UpsertUserParams is input for creating or updating a user from Telegram auth.
type UpsertUserParams struct {
	Profile identity.TelegramProfile
	Role    identity.Role
}

// UserRepository persists Telegram users.
type UserRepository interface {
	UpsertByTelegramID(ctx context.Context, params UpsertUserParams) (identity.User, error)
	GetByID(ctx context.Context, id int64) (identity.User, error)
}
