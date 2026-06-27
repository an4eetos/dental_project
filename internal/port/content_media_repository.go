package port

import (
	"context"
)

type CreateContentMediaParams struct {
	ContentItemID *int64
	MIMEType      string
	Data          []byte
}

type ContentMediaRecord struct {
	ID            int64
	ContentItemID *int64
	MIMEType      string
	Data          []byte
}

// ContentMediaRepository stores uploaded images and videos for content blocks.
type ContentMediaRepository interface {
	Create(ctx context.Context, params CreateContentMediaParams) (int64, error)
	GetByID(ctx context.Context, id int64) (ContentMediaRecord, error)
	LinkToContentItem(ctx context.Context, contentItemID int64, mediaIDs []int64) error
	DeleteUnlinkedForItem(ctx context.Context, contentItemID int64, keepMediaIDs []int64) error
	DeleteOrphaned(ctx context.Context) error
	MediaIDsExist(ctx context.Context, ids []int64) (bool, error)
	FindContentItemIDByMediaID(ctx context.Context, mediaID int64) (*int64, error)
}
