package photoreview

import (
	"context"

	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
	photoreview "github.com/anuarkuanysh/dental_project/internal/domain/photo_review"
	"github.com/anuarkuanysh/dental_project/internal/port"
	photoreviewservice "github.com/anuarkuanysh/dental_project/internal/service/photo_review"
)

// GenerateDraft runs Gemini vision analysis and stores an AI draft for the doctor.
type GenerateDraft struct {
	Submissions port.PhotoSubmissionRepository
	Users       port.UserRepository
	Analyzer    port.VisionAnalyzer
}

type GenerateDraftResult struct {
	Submission photoreview.SubmissionWithPatient
	DraftText  string
}

func (uc *GenerateDraft) Execute(ctx context.Context, doctorUserID, submissionID int64) (GenerateDraftResult, error) {
	if err := requireDoctor(ctx, uc.Users, doctorUserID); err != nil {
		return GenerateDraftResult{}, err
	}

	item, err := uc.Submissions.GetByID(ctx, submissionID)
	if err != nil {
		return GenerateDraftResult{}, err
	}
	if item.Submission.Status != photoreview.StatusPending {
		return GenerateDraftResult{}, domainerrors.ErrSubmissionAlreadyAnswered
	}

	imgBytes, mime, err := uc.Submissions.GetImageData(ctx, submissionID)
	if err != nil {
		return GenerateDraftResult{}, err
	}

	analysis, err := uc.Analyzer.AnalyzeTeethImage(ctx, imgBytes, mime)
	if err != nil {
		return GenerateDraftResult{}, domainerrors.ErrDraftGenerationFailed
	}

	if err := uc.Submissions.SaveAIDraft(ctx, submissionID, *analysis); err != nil {
		return GenerateDraftResult{}, err
	}

	item.Submission.AIDraft = analysis
	return GenerateDraftResult{
		Submission: item,
		DraftText:  photoreviewservice.FormatAnalysisDraft(analysis),
	}, nil
}
