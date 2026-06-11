package photoreview

import (
	"context"

	photoreview "github.com/anuarkuanysh/dental_project/internal/domain/photo_review"
	"github.com/anuarkuanysh/dental_project/internal/port"
)

// Get returns a single submission for doctor review.
type Get struct {
	Submissions port.PhotoSubmissionRepository
	Users       port.UserRepository
}

func (uc *Get) Execute(ctx context.Context, doctorUserID, submissionID int64) (photoreview.SubmissionWithPatient, error) {
	if err := requireDoctor(ctx, uc.Users, doctorUserID); err != nil {
		return photoreview.SubmissionWithPatient{}, err
	}
	return uc.Submissions.GetByID(ctx, submissionID)
}
