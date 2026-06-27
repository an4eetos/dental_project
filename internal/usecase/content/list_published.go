package content

import (
	"context"

	contentdomain "github.com/anuarkuanysh/dental_project/internal/domain/content"
	contentsvc "github.com/anuarkuanysh/dental_project/internal/service/content"
	"github.com/anuarkuanysh/dental_project/internal/port"
)

type ItemView struct {
	ID          int64
	Title       string
	Description string
	Access      contentdomain.AccessLevel
	Locked      bool
	Blocks      []contentdomain.Block
	SortOrder   int
}

// ListPublished returns published content with masked blocks when locked.
type ListPublished struct {
	Content       port.ContentRepository
	Checker       port.SubscriptionChecker
	Users         port.UserRepository
}

func (uc *ListPublished) Execute(ctx context.Context, userID int64) ([]ItemView, error) {
	user, err := uc.Users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	entitled, err := userHasEntitlement(ctx, uc.Checker, user)
	if err != nil {
		return nil, err
	}

	items, err := uc.Content.ListPublished(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]ItemView, 0, len(items))
	for _, item := range items {
		view := toItemView(item, entitled)
		out = append(out, view)
	}
	return out, nil
}

func toItemView(item contentdomain.ContentItem, entitled bool) ItemView {
	locked := !contentdomain.IsAccessible(item.Access, entitled)
	blocks := item.Blocks
	if locked {
		blocks = contentsvc.MaskBlocks(blocks)
	}
	return ItemView{
		ID:          item.ID,
		Title:       item.Title,
		Description: item.Description,
		Access:      item.Access,
		Locked:      locked,
		Blocks:      blocks,
		SortOrder:   item.SortOrder,
	}
}
