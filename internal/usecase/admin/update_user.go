package admin

import (
	"context"
	"strings"

	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
	"github.com/anuarkuanysh/dental_project/internal/domain/identity"
	"github.com/anuarkuanysh/dental_project/internal/port"
)

// UpdateUser applies admin edits to a patient or doctor profile.
type UpdateUser struct {
	Users port.UserRepository
}

type UpdateUserInput struct {
	AdminUserID  int64
	TargetUserID int64
	FirstName    *string
	LastName     *string
	Username     *string
	Role         *identity.Role
}

func (uc *UpdateUser) Execute(ctx context.Context, in UpdateUserInput) (identity.User, error) {
	if err := requireAdmin(ctx, uc.Users, in.AdminUserID); err != nil {
		return identity.User{}, err
	}
	if in.AdminUserID == in.TargetUserID {
		return identity.User{}, domainerrors.ErrForbidden
	}

	target, err := uc.Users.GetByID(ctx, in.TargetUserID)
	if err != nil {
		return identity.User{}, err
	}
	if target.Role == identity.RoleAdmin {
		return identity.User{}, domainerrors.ErrForbidden
	}

	params := port.AdminUpdateUserParams{}
	if in.FirstName != nil {
		name := strings.TrimSpace(*in.FirstName)
		if name == "" {
			return identity.User{}, domainerrors.ErrInvalidProfile
		}
		params.FirstName = &name
	}
	if in.LastName != nil {
		last := strings.TrimSpace(*in.LastName)
		params.LastName = &last
	}
	if in.Username != nil {
		username := strings.TrimSpace(*in.Username)
		params.Username = &username
	}
	if in.Role != nil {
		role := *in.Role
		if role != identity.RolePatient && role != identity.RoleDoctor {
			return identity.User{}, domainerrors.ErrInvalidRole
		}
		params.Role = &role
	}

	return uc.Users.UpdateByAdmin(ctx, in.TargetUserID, params)
}
