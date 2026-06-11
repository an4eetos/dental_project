package appointment

import (
	"context"

	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
	"github.com/anuarkuanysh/dental_project/internal/domain/identity"
	"github.com/anuarkuanysh/dental_project/internal/port"
)

func requireDoctor(ctx context.Context, users port.UserRepository, userID int64) error {
	user, err := users.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user.Role != identity.RoleDoctor {
		return domainerrors.ErrForbidden
	}
	return nil
}
