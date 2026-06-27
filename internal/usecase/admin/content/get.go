package content

import (
	"context"

	contentdomain "github.com/anuarkuanysh/dental_project/internal/domain/content"
	"github.com/anuarkuanysh/dental_project/internal/port"
	adminuc "github.com/anuarkuanysh/dental_project/internal/usecase/admin"
)

type Get struct {
	Content port.ContentRepository
	Users   port.UserRepository
}

func (uc *Get) Execute(ctx context.Context, adminUserID, contentID int64) (contentdomain.ContentItem, error) {
	if err := adminuc.RequireAdmin(ctx, uc.Users, adminUserID); err != nil {
		return contentdomain.ContentItem{}, err
	}
	return uc.Content.GetByID(ctx, contentID)
}
