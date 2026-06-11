package photoreview

import (
	"context"

	"github.com/anuarkuanysh/dental_project/internal/port"
)

// GetImage returns stored image bytes for doctor review.
type GetImage struct {
	Submissions port.PhotoSubmissionRepository
	Users       port.UserRepository
}

type ImageResult struct {
	Data     []byte
	MIMEType string
}

func (uc *GetImage) Execute(ctx context.Context, doctorUserID, submissionID int64) (ImageResult, error) {
	if err := requireDoctor(ctx, uc.Users, doctorUserID); err != nil {
		return ImageResult{}, err
	}

	data, mime, err := uc.Submissions.GetImageData(ctx, submissionID)
	if err != nil {
		return ImageResult{}, err
	}
	return ImageResult{Data: data, MIMEType: mime}, nil
}
