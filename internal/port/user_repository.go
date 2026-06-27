package port

import (
	"context"

	"github.com/anuarkuanysh/dental_project/internal/domain/admin"
	"github.com/anuarkuanysh/dental_project/internal/domain/identity"
)

// UpsertUserParams is input for creating or updating a user from Telegram auth.
type UpsertUserParams struct {
	Profile identity.TelegramProfile
	Role    identity.Role
}

// ListUsersParams filters users for admin listing.
type ListUsersParams struct {
	Role   identity.Role
	Search string
	Limit  int
	Offset int
}

// AdminUpdateUserParams is input for admin profile edits.
type AdminUpdateUserParams struct {
	FirstName *string
	LastName  *string
	Username  *string
	Role      *identity.Role
}

// UserRepository persists Telegram users.
type UserRepository interface {
	UpsertByTelegramID(ctx context.Context, params UpsertUserParams) (identity.User, error)
	GetByID(ctx context.Context, id int64) (identity.User, error)
	GetByTelegramID(ctx context.Context, telegramID int64) (identity.User, error)
	List(ctx context.Context, params ListUsersParams) ([]identity.User, error)
	GetOverviewByID(ctx context.Context, id int64) (admin.UserOverview, error)
	UpdateByAdmin(ctx context.Context, id int64, params AdminUpdateUserParams) (identity.User, error)
	SetBlocked(ctx context.Context, id int64, blocked bool) (identity.User, error)
}
