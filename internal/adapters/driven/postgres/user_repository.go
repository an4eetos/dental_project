package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
	"github.com/anuarkuanysh/dental_project/internal/domain/admin"
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
          COALESCE(avatar_url, ''), role, is_blocked, created_at, updated_at`

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
       COALESCE(avatar_url, ''), role, is_blocked, created_at, updated_at
FROM users WHERE id = $1`

	row := r.pool.QueryRow(ctx, q, id)
	u, err := scanUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return identity.User{}, domainerrors.ErrUserNotFound
	}
	return u, err
}

func (r *UserRepository) GetByTelegramID(ctx context.Context, telegramID int64) (identity.User, error) {
	const q = `
SELECT id, telegram_id, COALESCE(username, ''), first_name, COALESCE(last_name, ''),
       COALESCE(avatar_url, ''), role, is_blocked, created_at, updated_at
FROM users WHERE telegram_id = $1`

	row := r.pool.QueryRow(ctx, q, telegramID)
	u, err := scanUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return identity.User{}, domainerrors.ErrUserNotFound
	}
	return u, err
}

func (r *UserRepository) List(ctx context.Context, params port.ListUsersParams) ([]identity.User, error) {
	limit := params.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset := params.Offset
	if offset < 0 {
		offset = 0
	}

	search := strings.TrimSpace(params.Search)
	roleFilter := ""
	if params.Role.Valid() {
		roleFilter = params.Role.String()
	}

	const q = `
SELECT id, telegram_id, COALESCE(username, ''), first_name, COALESCE(last_name, ''),
       COALESCE(avatar_url, ''), role, is_blocked, created_at, updated_at
FROM users
WHERE ($1 = '' OR role = $1)
  AND (
    $2 = '' OR username ILIKE '%' || $2 || '%'
    OR first_name ILIKE '%' || $2 || '%'
    OR last_name ILIKE '%' || $2 || '%'
    OR CAST(telegram_id AS TEXT) LIKE '%' || $2 || '%'
  )
ORDER BY created_at DESC
LIMIT $3 OFFSET $4`

	rows, err := r.pool.Query(ctx, q, roleFilter, search, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []identity.User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *UserRepository) GetOverviewByID(ctx context.Context, id int64) (admin.UserOverview, error) {
	const q = `
SELECT u.id, u.telegram_id, COALESCE(u.username, ''), u.first_name, COALESCE(u.last_name, ''),
       COALESCE(u.avatar_url, ''), u.role, u.is_blocked,
       (SELECT COUNT(*) FROM appointments a WHERE a.user_id = u.id),
       (SELECT COUNT(*) FROM photo_submissions ps WHERE ps.user_id = u.id),
       u.created_at, u.updated_at
FROM users u
WHERE u.id = $1`

	row := r.pool.QueryRow(ctx, q, id)
	overview, err := scanUserOverview(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return admin.UserOverview{}, domainerrors.ErrUserNotFound
	}
	return overview, err
}

func (r *UserRepository) UpdateByAdmin(ctx context.Context, id int64, params port.AdminUpdateUserParams) (identity.User, error) {
	setClauses := make([]string, 0, 4)
	args := make([]any, 0, 6)
	argN := 1

	if params.FirstName != nil {
		setClauses = append(setClauses, fmt.Sprintf("first_name = $%d", argN))
		args = append(args, strings.TrimSpace(*params.FirstName))
		argN++
	}
	if params.LastName != nil {
		setClauses = append(setClauses, fmt.Sprintf("last_name = $%d", argN))
		args = append(args, nullIfEmpty(strings.TrimSpace(*params.LastName)))
		argN++
	}
	if params.Username != nil {
		setClauses = append(setClauses, fmt.Sprintf("username = $%d", argN))
		args = append(args, nullIfEmpty(strings.TrimSpace(*params.Username)))
		argN++
	}
	if params.Role != nil {
		setClauses = append(setClauses, fmt.Sprintf("role = $%d", argN))
		args = append(args, params.Role.String())
		argN++
	}

	if len(setClauses) == 0 {
		return r.GetByID(ctx, id)
	}

	setClauses = append(setClauses, "updated_at = NOW()")
	args = append(args, id)

	q := fmt.Sprintf(`
UPDATE users SET %s
WHERE id = $%d
RETURNING id, telegram_id, COALESCE(username, ''), first_name, COALESCE(last_name, ''),
          COALESCE(avatar_url, ''), role, is_blocked, created_at, updated_at`,
		strings.Join(setClauses, ", "), argN)

	row := r.pool.QueryRow(ctx, q, args...)
	u, err := scanUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return identity.User{}, domainerrors.ErrUserNotFound
	}
	return u, err
}

func (r *UserRepository) SetBlocked(ctx context.Context, id int64, blocked bool) (identity.User, error) {
	const q = `
UPDATE users SET is_blocked = $1, updated_at = NOW()
WHERE id = $2
RETURNING id, telegram_id, COALESCE(username, ''), first_name, COALESCE(last_name, ''),
          COALESCE(avatar_url, ''), role, is_blocked, created_at, updated_at`

	row := r.pool.QueryRow(ctx, q, blocked, id)
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
		&u.Blocked,
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

func scanUserOverview(row pgx.Row) (admin.UserOverview, error) {
	var o admin.UserOverview
	var role string
	err := row.Scan(
		&o.ID,
		&o.TelegramID,
		&o.Username,
		&o.FirstName,
		&o.LastName,
		&o.AvatarURL,
		&role,
		&o.Blocked,
		&o.AppointmentCount,
		&o.PhotoSubmissionCount,
		&o.CreatedAt,
		&o.UpdatedAt,
	)
	if err != nil {
		return admin.UserOverview{}, err
	}
	o.Role = identity.Role(role)
	if !o.Role.Valid() {
		o.Role = identity.RolePatient
	}
	return o, nil
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
