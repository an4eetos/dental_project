package photoreview

import (
	"context"

	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
	photoreview "github.com/anuarkuanysh/dental_project/internal/domain/photo_review"
	"github.com/anuarkuanysh/dental_project/internal/domain/identity"
	"github.com/anuarkuanysh/dental_project/internal/port"
)

// ListPending returns unanswered photo submissions for clinic doctors.
type ListPending struct {
	Submissions port.PhotoSubmissionRepository
	Users       port.UserRepository
}

func (uc *ListPending) Execute(ctx context.Context, doctorUserID int64) ([]photoreview.SubmissionWithPatient, error) {
	if err := requireDoctor(ctx, uc.Users, doctorUserID); err != nil {
		return nil, err
	}
	return uc.Submissions.ListByStatus(ctx, photoreview.StatusPending)
}

func requireDoctor(ctx context.Context, users port.UserRepository, userID int64) error {
	user, err := users.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user.Role != identity.RoleDoctor {
		return domainerrors.ErrForbidden
	}
	return nil
}
