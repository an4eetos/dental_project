package port

import (
	"context"
	"time"

	photoreview "github.com/anuarkuanysh/dental_project/internal/domain/photo_review"
	"github.com/anuarkuanysh/dental_project/internal/model"
)

// PhotoSubmissionRepository persists patient photo submissions for clinic review.
type PhotoSubmissionRepository interface {
	Create(ctx context.Context, params photoreview.CreateParams) (photoreview.Submission, error)
	GetByID(ctx context.Context, id int64) (photoreview.SubmissionWithPatient, error)
	ListByStatus(ctx context.Context, status photoreview.Status) ([]photoreview.SubmissionWithPatient, error)
	SaveAIDraft(ctx context.Context, id int64, draft model.Analysis) error
	MarkAnswered(ctx context.Context, id int64, doctorUserID int64, response string, at time.Time) error
	GetImageData(ctx context.Context, id int64) ([]byte, string, error)
	DeletePendingOlderThan(ctx context.Context, before time.Time) (int64, error)
}
