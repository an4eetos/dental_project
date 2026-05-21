package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
	"github.com/anuarkuanysh/dental_project/internal/domain/identity"
	"github.com/anuarkuanysh/dental_project/internal/port"
)

// UserRepository implements port.UserRepository.
type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) UpsertByTelegramID(ctx context.Context, params port.UpsertUserParams) (identity.User, error) {
	role := params.Role
	if !role.Valid() {
		role = identity.RolePatient
	}

	const q = `
INSERT INTO users (telegram_id, username, first_name, last_name, avatar_url, role)
VALUES ($1, $2, $3, $4, NULLIF($5, ''), $6)
ON CONFLICT (telegram_id) DO UPDATE SET
    username = EXCLUDED.username,
    first_name = EXCLUDED.first_name,
    last_name = EXCLUDED.last_name,
    avatar_url = EXCLUDED.avatar_url,
    role = EXCLUDED.role,
    updated_at = NOW()
RETURNING id, telegram_id, COALESCE(username, ''), first_name, COALESCE(last_name, ''),
          COALESCE(avatar_url, ''), role, created_at, updated_at`

	row := r.pool.QueryRow(ctx, q,
		params.Profile.TelegramID,
		nullIfEmpty(params.Profile.Username),
		params.Profile.FirstName,
		nullIfEmpty(params.Profile.LastName),
		params.Profile.AvatarURL,
		role.String(),
	)
	return scanUser(row)
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (identity.User, error) {
	const q = `
SELECT id, telegram_id, COALESCE(username, ''), first_name, COALESCE(last_name, ''),
       COALESCE(avatar_url, ''), role, created_at, updated_at
FROM users WHERE id = $1`

	row := r.pool.QueryRow(ctx, q, id)
	u, err := scanUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return identity.User{}, domainerrors.ErrUserNotFound
	}
	return u, err
}

func scanUser(row pgx.Row) (identity.User, error) {
	var u identity.User
	var role string
	err := row.Scan(
		&u.ID,
		&u.TelegramID,
		&u.Username,
		&u.FirstName,
		&u.LastName,
		&u.AvatarURL,
		&role,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		return identity.User{}, err
	}
	u.Role = identity.Role(role)
	if !u.Role.Valid() {
		u.Role = identity.RolePatient
	}
	return u, nil
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
