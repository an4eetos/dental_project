package content

import (
	"context"

	contentdomain "github.com/anuarkuanysh/dental_project/internal/domain/content"
	"github.com/anuarkuanysh/dental_project/internal/port"
	adminuc "github.com/anuarkuanysh/dental_project/internal/usecase/admin"
)

type List struct {
	Content port.ContentRepository
	Users   port.UserRepository
}

func (uc *List) Execute(ctx context.Context, adminUserID int64) ([]contentdomain.ContentItem, error) {
	if err := adminuc.RequireAdmin(ctx, uc.Users, adminUserID); err != nil {
		return nil, err
	}
	return uc.Content.ListAll(ctx)
}
