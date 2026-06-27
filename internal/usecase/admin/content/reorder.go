package content

import (
	"context"

	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
	"github.com/anuarkuanysh/dental_project/internal/port"
	adminuc "github.com/anuarkuanysh/dental_project/internal/usecase/admin"
)

type ReorderInput struct {
	AdminUserID int64
	IDs         []int64
}

type Reorder struct {
	Content port.ContentRepository
	Users   port.UserRepository
}

func (uc *Reorder) Execute(ctx context.Context, in ReorderInput) error {
	if err := adminuc.RequireAdmin(ctx, uc.Users, in.AdminUserID); err != nil {
		return err
	}
	if len(in.IDs) == 0 {
		return domainerrors.ErrInvalidContentBlocks
	}
	return uc.Content.UpdateSortOrder(ctx, in.IDs)
}
