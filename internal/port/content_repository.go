package port

import (
	"context"

	"github.com/anuarkuanysh/dental_project/internal/domain/content"
)

type CreateContentParams struct {
	Title       string
	Description string
	Access      content.AccessLevel
	Published   bool
	SortOrder   int
	Blocks      []content.Block
}

type UpdateContentParams struct {
	Title       string
	Description string
	Access      content.AccessLevel
	Published   bool
	Blocks      []content.Block
}

// ContentRepository persists educational content items.
type ContentRepository interface {
	ListPublished(ctx context.Context) ([]content.ContentItem, error)
	ListAll(ctx context.Context) ([]content.ContentItem, error)
	GetByID(ctx context.Context, id int64) (content.ContentItem, error)
	Create(ctx context.Context, params CreateContentParams) (content.ContentItem, error)
	Update(ctx context.Context, id int64, params UpdateContentParams) (content.ContentItem, error)
	Delete(ctx context.Context, id int64) error
	UpdateSortOrder(ctx context.Context, ids []int64) error
	NextSortOrder(ctx context.Context) (int, error)
}
