package content

import (
	"context"
	"strings"

	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
	contentdomain "github.com/anuarkuanysh/dental_project/internal/domain/content"
	contentsvc "github.com/anuarkuanysh/dental_project/internal/service/content"
	"github.com/anuarkuanysh/dental_project/internal/port"
	adminuc "github.com/anuarkuanysh/dental_project/internal/usecase/admin"
)

type UpdateInput struct {
	AdminUserID int64
	ContentID   int64
	Title       string
	Description string
	Access      string
	Published   bool
	Blocks      []contentdomain.Block
}

type Update struct {
	Content port.ContentRepository
	Media   port.ContentMediaRepository
	Users   port.UserRepository
}

func (uc *Update) Execute(ctx context.Context, in UpdateInput) (contentdomain.ContentItem, error) {
	if err := adminuc.RequireAdmin(ctx, uc.Users, in.AdminUserID); err != nil {
		return contentdomain.ContentItem{}, err
	}

	title := strings.TrimSpace(in.Title)
	if title == "" {
		return contentdomain.ContentItem{}, domainerrors.ErrInvalidContentTitle
	}

	access, ok := contentdomain.ParseAccessLevel(in.Access)
	if !ok {
		return contentdomain.ContentItem{}, domainerrors.ErrInvalidContentAccess
	}

	blocks, err := contentsvc.NormalizeBlocks(in.Blocks)
	if err != nil {
		return contentdomain.ContentItem{}, err
	}

	mediaIDs := contentsvc.ExtractMediaIDs(blocks)
	if len(mediaIDs) > 0 {
		exists, err := uc.Media.MediaIDsExist(ctx, mediaIDs)
		if err != nil {
			return contentdomain.ContentItem{}, err
		}
		if !exists {
			return contentdomain.ContentItem{}, domainerrors.ErrContentMediaNotFound
		}
	}

	item, err := uc.Content.Update(ctx, in.ContentID, port.UpdateContentParams{
		Title:       title,
		Description: strings.TrimSpace(in.Description),
		Access:      access,
		Published:   in.Published,
		Blocks:      blocks,
	})
	if err != nil {
		return contentdomain.ContentItem{}, err
	}

	if err := uc.Media.LinkToContentItem(ctx, item.ID, mediaIDs); err != nil {
		return contentdomain.ContentItem{}, err
	}
	if err := uc.Media.DeleteUnlinkedForItem(ctx, item.ID, mediaIDs); err != nil {
		return contentdomain.ContentItem{}, err
	}
	_ = uc.Media.DeleteOrphaned(ctx)

	return item, nil
}
