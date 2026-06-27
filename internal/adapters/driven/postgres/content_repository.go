package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
	"github.com/anuarkuanysh/dental_project/internal/domain/content"
	"github.com/anuarkuanysh/dental_project/internal/port"
)

type ContentRepository struct {
	pool *pgxpool.Pool
}

func NewContentRepository(pool *pgxpool.Pool) *ContentRepository {
	return &ContentRepository{pool: pool}
}

func (r *ContentRepository) ListPublished(ctx context.Context) ([]content.ContentItem, error) {
	const q = `
SELECT id, title, description, access, published, sort_order, blocks, created_at, updated_at
FROM content_items
WHERE published = TRUE
ORDER BY sort_order ASC, id ASC`

	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanContentItems(rows)
}

func (r *ContentRepository) ListAll(ctx context.Context) ([]content.ContentItem, error) {
	const q = `
SELECT id, title, description, access, published, sort_order, blocks, created_at, updated_at
FROM content_items
ORDER BY sort_order ASC, id ASC`

	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanContentItems(rows)
}

func (r *ContentRepository) GetByID(ctx context.Context, id int64) (content.ContentItem, error) {
	const q = `
SELECT id, title, description, access, published, sort_order, blocks, created_at, updated_at
FROM content_items
WHERE id = $1`

	row := r.pool.QueryRow(ctx, q, id)
	item, err := scanContentItem(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return content.ContentItem{}, domainerrors.ErrContentNotFound
	}
	return item, err
}

func (r *ContentRepository) Create(ctx context.Context, params port.CreateContentParams) (content.ContentItem, error) {
	raw, err := marshalBlocks(params.Blocks)
	if err != nil {
		return content.ContentItem{}, err
	}

	const q = `
INSERT INTO content_items (title, description, access, published, sort_order, blocks)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, title, description, access, published, sort_order, blocks, created_at, updated_at`

	row := r.pool.QueryRow(ctx, q,
		params.Title,
		params.Description,
		params.Access.String(),
		params.Published,
		params.SortOrder,
		raw,
	)
	return scanContentItem(row)
}

func (r *ContentRepository) Update(ctx context.Context, id int64, params port.UpdateContentParams) (content.ContentItem, error) {
	raw, err := marshalBlocks(params.Blocks)
	if err != nil {
		return content.ContentItem{}, err
	}

	const q = `
UPDATE content_items
SET title = $2,
    description = $3,
    access = $4,
    published = $5,
    blocks = $6,
    updated_at = NOW()
WHERE id = $1
RETURNING id, title, description, access, published, sort_order, blocks, created_at, updated_at`

	row := r.pool.QueryRow(ctx, q,
		id,
		params.Title,
		params.Description,
		params.Access.String(),
		params.Published,
		raw,
	)
	item, err := scanContentItem(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return content.ContentItem{}, domainerrors.ErrContentNotFound
	}
	return item, err
}

func (r *ContentRepository) Delete(ctx context.Context, id int64) error {
	const q = `DELETE FROM content_items WHERE id = $1`
	tag, err := r.pool.Exec(ctx, q, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domainerrors.ErrContentNotFound
	}
	return nil
}

func (r *ContentRepository) UpdateSortOrder(ctx context.Context, ids []int64) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	const q = `UPDATE content_items SET sort_order = $2, updated_at = NOW() WHERE id = $1`
	for i, id := range ids {
		tag, err := tx.Exec(ctx, q, id, i+1)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return domainerrors.ErrContentNotFound
		}
	}

	return tx.Commit(ctx)
}

func (r *ContentRepository) NextSortOrder(ctx context.Context) (int, error) {
	const q = `SELECT COALESCE(MAX(sort_order), 0) + 1 FROM content_items`
	var next int
	if err := r.pool.QueryRow(ctx, q).Scan(&next); err != nil {
		return 0, err
	}
	return next, nil
}

type contentScanner interface {
	Scan(dest ...any) error
}

func scanContentItem(row contentScanner) (content.ContentItem, error) {
	var item content.ContentItem
	var access string
	var raw []byte
	if err := row.Scan(
		&item.ID,
		&item.Title,
		&item.Description,
		&access,
		&item.Published,
		&item.SortOrder,
		&raw,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return content.ContentItem{}, err
	}
	item.Access = content.AccessLevel(access)
	blocks, err := unmarshalBlocks(raw)
	if err != nil {
		return content.ContentItem{}, err
	}
	item.Blocks = blocks
	return item, nil
}

func scanContentItems(rows pgx.Rows) ([]content.ContentItem, error) {
	var out []content.ContentItem
	for rows.Next() {
		item, err := scanContentItem(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func marshalBlocks(blocks []content.Block) ([]byte, error) {
	if blocks == nil {
		blocks = []content.Block{}
	}
	return json.Marshal(blocks)
}

func unmarshalBlocks(raw []byte) ([]content.Block, error) {
	if len(raw) == 0 {
		return []content.Block{}, nil
	}
	var blocks []content.Block
	if err := json.Unmarshal(raw, &blocks); err != nil {
		return nil, err
	}
	if blocks == nil {
		return []content.Block{}, nil
	}
	return blocks, nil
}
