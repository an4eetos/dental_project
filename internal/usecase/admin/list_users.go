package admin

import (
	"context"
	"strings"

	admindomain "github.com/anuarkuanysh/dental_project/internal/domain/admin"
	"github.com/anuarkuanysh/dental_project/internal/domain/identity"
	"github.com/anuarkuanysh/dental_project/internal/port"
)

// ListUsers returns users matching admin filters.
type ListUsers struct {
	Users port.UserRepository
}

type ListUsersInput struct {
	AdminUserID int64
	Role        identity.Role
	Search      string
	Limit       int
	Offset      int
}

func (uc *ListUsers) Execute(ctx context.Context, in ListUsersInput) ([]identity.User, error) {
	if err := RequireAdmin(ctx, uc.Users, in.AdminUserID); err != nil {
		return nil, err
	}

	params := port.ListUsersParams{
		Search: strings.TrimSpace(in.Search),
		Limit:  in.Limit,
		Offset: in.Offset,
	}
	if in.Role.Valid() {
		params.Role = in.Role
	}
	return uc.Users.List(ctx, params)
}

// GetUser returns a user overview for admin inspection.
type GetUser struct {
	Users port.UserRepository
}

func (uc *GetUser) Execute(ctx context.Context, adminUserID, targetUserID int64) (admindomain.UserOverview, error) {
	if err := RequireAdmin(ctx, uc.Users, adminUserID); err != nil {
		return admindomain.UserOverview{}, err
	}
	return uc.Users.GetOverviewByID(ctx, targetUserID)
}
