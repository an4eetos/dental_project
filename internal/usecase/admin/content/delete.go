package content

import (
	"context"

	"github.com/anuarkuanysh/dental_project/internal/port"
	adminuc "github.com/anuarkuanysh/dental_project/internal/usecase/admin"
)

type Delete struct {
	Content port.ContentRepository
	Media   port.ContentMediaRepository
	Users   port.UserRepository
}

func (uc *Delete) Execute(ctx context.Context, adminUserID, contentID int64) error {
	if err := adminuc.RequireAdmin(ctx, uc.Users, adminUserID); err != nil {
		return err
	}
	if err := uc.Content.Delete(ctx, contentID); err != nil {
		return err
	}
	return uc.Media.DeleteOrphaned(ctx)
}
