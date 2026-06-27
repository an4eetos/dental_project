package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
	"github.com/anuarkuanysh/dental_project/internal/port"
)

type ContentMediaRepository struct {
	pool *pgxpool.Pool
}

func NewContentMediaRepository(pool *pgxpool.Pool) *ContentMediaRepository {
	return &ContentMediaRepository{pool: pool}
}

func (r *ContentMediaRepository) Create(ctx context.Context, params port.CreateContentMediaParams) (int64, error) {
	const q = `
INSERT INTO content_media (content_item_id, mime_type, data)
VALUES ($1, $2, $3)
RETURNING id`

	var id int64
	if err := r.pool.QueryRow(ctx, q, params.ContentItemID, params.MIMEType, params.Data).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *ContentMediaRepository) GetByID(ctx context.Context, id int64) (port.ContentMediaRecord, error) {
	const q = `
SELECT id, content_item_id, mime_type, data
FROM content_media
WHERE id = $1`

	var rec port.ContentMediaRecord
	err := r.pool.QueryRow(ctx, q, id).Scan(&rec.ID, &rec.ContentItemID, &rec.MIMEType, &rec.Data)
	if errors.Is(err, pgx.ErrNoRows) {
		return port.ContentMediaRecord{}, domainerrors.ErrContentMediaNotFound
	}
	return rec, err
}

func (r *ContentMediaRepository) LinkToContentItem(ctx context.Context, contentItemID int64, mediaIDs []int64) error {
	if len(mediaIDs) == 0 {
		return nil
	}
	const q = `UPDATE content_media SET content_item_id = $1 WHERE id = ANY($2)`
	_, err := r.pool.Exec(ctx, q, contentItemID, mediaIDs)
	return err
}

func (r *ContentMediaRepository) DeleteUnlinkedForItem(ctx context.Context, contentItemID int64, keepMediaIDs []int64) error {
	if len(keepMediaIDs) == 0 {
		const q = `DELETE FROM content_media WHERE content_item_id = $1`
		_, err := r.pool.Exec(ctx, q, contentItemID)
		return err
	}
	const q = `DELETE FROM content_media WHERE content_item_id = $1 AND NOT (id = ANY($2))`
	_, err := r.pool.Exec(ctx, q, contentItemID, keepMediaIDs)
	return err
}

func (r *ContentMediaRepository) DeleteOrphaned(ctx context.Context) error {
	const q = `DELETE FROM content_media WHERE content_item_id IS NULL`
	_, err := r.pool.Exec(ctx, q)
	return err
}

func (r *ContentMediaRepository) MediaIDsExist(ctx context.Context, ids []int64) (bool, error) {
	if len(ids) == 0 {
		return true, nil
	}
	const q = `SELECT COUNT(*) = $1 FROM content_media WHERE id = ANY($2)`
	var ok bool
	if err := r.pool.QueryRow(ctx, q, len(ids), ids).Scan(&ok); err != nil {
		return false, err
	}
	return ok, nil
}

func (r *ContentMediaRepository) FindContentItemIDByMediaID(ctx context.Context, mediaID int64) (*int64, error) {
	const q = `SELECT content_item_id FROM content_media WHERE id = $1`
	var itemID *int64
	err := r.pool.QueryRow(ctx, q, mediaID).Scan(&itemID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domainerrors.ErrContentMediaNotFound
	}
	return itemID, err
}
