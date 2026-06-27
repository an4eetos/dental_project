package admin

import (
	"context"

	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
	"github.com/anuarkuanysh/dental_project/internal/domain/identity"
	"github.com/anuarkuanysh/dental_project/internal/port"
)

func RequireAdmin(ctx context.Context, users port.UserRepository, adminUserID int64) error {
	user, err := users.GetByID(ctx, adminUserID)
	if err != nil {
		return err
	}
	if user.Role != identity.RoleAdmin {
		return domainerrors.ErrForbidden
	}
	return nil
}
