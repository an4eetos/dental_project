package photoreview

import (
	"context"

	photoreview "github.com/anuarkuanysh/dental_project/internal/domain/photo_review"
	"github.com/anuarkuanysh/dental_project/internal/port"
)

// ListAnswered returns completed photo reviews for clinic doctors.
type ListAnswered struct {
	Submissions port.PhotoSubmissionRepository
	Users       port.UserRepository
}

func (uc *ListAnswered) Execute(ctx context.Context, doctorUserID int64) ([]photoreview.SubmissionWithPatient, error) {
	if err := requireDoctor(ctx, uc.Users, doctorUserID); err != nil {
		return nil, err
	}
	return uc.Submissions.ListByStatus(ctx, photoreview.StatusAnswered)
}
