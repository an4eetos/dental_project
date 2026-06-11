package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
	photoreview "github.com/anuarkuanysh/dental_project/internal/domain/photo_review"
	"github.com/anuarkuanysh/dental_project/internal/model"
)

// PhotoSubmissionRepository implements port.PhotoSubmissionRepository.
type PhotoSubmissionRepository struct {
	pool *pgxpool.Pool
}

func NewPhotoSubmissionRepository(pool *pgxpool.Pool) *PhotoSubmissionRepository {
	return &PhotoSubmissionRepository{pool: pool}
}

func (r *PhotoSubmissionRepository) Create(ctx context.Context, params photoreview.CreateParams) (photoreview.Submission, error) {
	const q = `
INSERT INTO photo_submissions (user_id, telegram_file_id, image_data, image_mime, status, created_at)
VALUES ($1, $2, $3, $4, $5, NOW())
RETURNING id, user_id, telegram_file_id, image_mime, status, ai_draft, doctor_response,
          COALESCE(responded_by, 0), responded_at, created_at`

	row := r.pool.QueryRow(ctx, q,
		params.UserID,
		params.TelegramFileID,
		params.ImageData,
		params.ImageMIME,
		photoreview.StatusPending.String(),
	)
	return scanSubmission(row)
}

func (r *PhotoSubmissionRepository) GetByID(ctx context.Context, id int64) (photoreview.SubmissionWithPatient, error) {
	const q = `
SELECT
    s.id, s.user_id, s.telegram_file_id, s.image_mime, s.status, s.ai_draft, s.doctor_response,
    COALESCE(s.responded_by, 0), s.responded_at, s.created_at,
    u.id, u.telegram_id, COALESCE(u.username, ''), u.first_name, COALESCE(u.last_name, '')
FROM photo_submissions s
JOIN users u ON u.id = s.user_id
WHERE s.id = $1`

	row := r.pool.QueryRow(ctx, q, id)
	item, err := scanSubmissionWithPatient(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return photoreview.SubmissionWithPatient{}, domainerrors.ErrSubmissionNotFound
	}
	return item, err
}

func (r *PhotoSubmissionRepository) ListByStatus(ctx context.Context, status photoreview.Status) ([]photoreview.SubmissionWithPatient, error) {
	const q = `
SELECT
    s.id, s.user_id, s.telegram_file_id, s.image_mime, s.status, s.ai_draft, s.doctor_response,
    COALESCE(s.responded_by, 0), s.responded_at, s.created_at,
    u.id, u.telegram_id, COALESCE(u.username, ''), u.first_name, COALESCE(u.last_name, '')
FROM photo_submissions s
JOIN users u ON u.id = s.user_id
WHERE s.status = $1
ORDER BY s.created_at DESC`

	rows, err := r.pool.Query(ctx, q, status.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []photoreview.SubmissionWithPatient
	for rows.Next() {
		item, err := scanSubmissionWithPatient(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *PhotoSubmissionRepository) SaveAIDraft(ctx context.Context, id int64, draft model.Analysis) error {
	raw, err := json.Marshal(draft)
	if err != nil {
		return err
	}

	const q = `UPDATE photo_submissions SET ai_draft = $2 WHERE id = $1`
	tag, err := r.pool.Exec(ctx, q, id, raw)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domainerrors.ErrSubmissionNotFound
	}
	return nil
}

func (r *PhotoSubmissionRepository) MarkAnswered(ctx context.Context, id int64, doctorUserID int64, response string, at time.Time) error {
	const q = `
UPDATE photo_submissions
SET status = $2, doctor_response = $3, responded_by = $4, responded_at = $5
WHERE id = $1 AND status = $6`

	tag, err := r.pool.Exec(ctx, q,
		id,
		photoreview.StatusAnswered.String(),
		response,
		doctorUserID,
		at,
		photoreview.StatusPending.String(),
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		existing, err := r.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if existing.Submission.Status == photoreview.StatusAnswered {
			return domainerrors.ErrSubmissionAlreadyAnswered
		}
		return domainerrors.ErrSubmissionNotFound
	}
	return nil
}

func (r *PhotoSubmissionRepository) GetImageData(ctx context.Context, id int64) ([]byte, string, error) {
	const q = `SELECT image_data, image_mime FROM photo_submissions WHERE id = $1`
	var data []byte
	var mime string
	err := r.pool.QueryRow(ctx, q, id).Scan(&data, &mime)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, "", domainerrors.ErrSubmissionNotFound
	}
	return data, mime, err
}

func scanSubmissionWithPatient(row pgxRowScanner) (photoreview.SubmissionWithPatient, error) {
	var item photoreview.SubmissionWithPatient
	var status string
	var draftRaw []byte
	var respondedAt *time.Time

	err := row.Scan(
		&item.Submission.ID,
		&item.Submission.UserID,
		&item.Submission.TelegramFileID,
		&item.Submission.ImageMIME,
		&status,
		&draftRaw,
		&item.Submission.DoctorResponse,
		&item.Submission.RespondedBy,
		&respondedAt,
		&item.Submission.CreatedAt,
		&item.Patient.ID,
		&item.Patient.TelegramID,
		&item.Patient.Username,
		&item.Patient.FirstName,
		&item.Patient.LastName,
	)
	if err != nil {
		return photoreview.SubmissionWithPatient{}, err
	}

	item.Submission.Status = photoreview.Status(status)
	item.Submission.RespondedAt = respondedAt
	if len(draftRaw) > 0 {
		var draft model.Analysis
		if err := json.Unmarshal(draftRaw, &draft); err == nil {
			item.Submission.AIDraft = &draft
		}
	}
	return item, nil
}

func scanSubmission(row pgxRowScanner) (photoreview.Submission, error) {
	var sub photoreview.Submission
	var status string
	var draftRaw []byte
	var respondedAt *time.Time

	err := row.Scan(
		&sub.ID,
		&sub.UserID,
		&sub.TelegramFileID,
		&sub.ImageMIME,
		&status,
		&draftRaw,
		&sub.DoctorResponse,
		&sub.RespondedBy,
		&respondedAt,
		&sub.CreatedAt,
	)
	if err != nil {
		return photoreview.Submission{}, err
	}

	sub.Status = photoreview.Status(status)
	sub.RespondedAt = respondedAt
	if len(draftRaw) > 0 {
		var draft model.Analysis
		if err := json.Unmarshal(draftRaw, &draft); err == nil {
			sub.AIDraft = &draft
		}
	}
	return sub, nil
}
