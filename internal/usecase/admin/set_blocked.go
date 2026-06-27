package admin

import (
	"context"

	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
	"github.com/anuarkuanysh/dental_project/internal/domain/identity"
	"github.com/anuarkuanysh/dental_project/internal/port"
)

// SetBlocked blocks or unblocks a user account.
type SetBlocked struct {
	Users port.UserRepository
}

type SetBlockedInput struct {
	AdminUserID  int64
	TargetUserID int64
	Blocked      bool
}

func (uc *SetBlocked) Execute(ctx context.Context, in SetBlockedInput) (identity.User, error) {
	if err := RequireAdmin(ctx, uc.Users, in.AdminUserID); err != nil {
		return identity.User{}, err
	}
	if in.AdminUserID == in.TargetUserID {
		return identity.User{}, domainerrors.ErrCannotBlockSelf
	}

	target, err := uc.Users.GetByID(ctx, in.TargetUserID)
	if err != nil {
		return identity.User{}, err
	}
	if target.Role == identity.RoleAdmin {
		return identity.User{}, domainerrors.ErrCannotBlockAdmin
	}

	return uc.Users.SetBlocked(ctx, in.TargetUserID, in.Blocked)
}
