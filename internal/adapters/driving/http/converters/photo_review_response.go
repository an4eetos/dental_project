package converters

import (
	"time"

	photoreview "github.com/anuarkuanysh/dental_project/internal/domain/photo_review"
	"github.com/anuarkuanysh/dental_project/internal/model"
)

type submissionPatientResponse struct {
	ID         int64  `json:"id"`
	TelegramID int64  `json:"telegram_id"`
	Username   string `json:"username,omitempty"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name,omitempty"`
}

type submissionResponse struct {
	ID             int64                    `json:"id"`
	MediaType      string                   `json:"media_type"`
	Status         string                   `json:"status"`
	CreatedAt      string                   `json:"created_at"`
	RespondedAt    string                   `json:"responded_at,omitempty"`
	DoctorResponse string                   `json:"doctor_response,omitempty"`
	AIDraft        *model.Analysis          `json:"ai_draft,omitempty"`
	Patient        submissionPatientResponse `json:"patient"`
}

type generateDraftResponse struct {
	Submission submissionResponse `json:"submission"`
	DraftText  string             `json:"draft_text"`
}

func ToSubmissionResponse(item photoreview.SubmissionWithPatient) submissionResponse {
	resp := submissionResponse{
		ID:             item.Submission.ID,
		MediaType:      item.Submission.MediaType.String(),
		Status:         item.Submission.Status.String(),
		CreatedAt:      formatTime(item.Submission.CreatedAt),
		DoctorResponse: item.Submission.DoctorResponse,
		AIDraft:        item.Submission.AIDraft,
		Patient: submissionPatientResponse{
			ID:         item.Patient.ID,
			TelegramID: item.Patient.TelegramID,
			Username:   item.Patient.Username,
			FirstName:  item.Patient.FirstName,
			LastName:   item.Patient.LastName,
		},
	}
	if item.Submission.RespondedAt != nil {
		resp.RespondedAt = formatTime(*item.Submission.RespondedAt)
	}
	return resp
}

func ToSubmissionListResponse(items []photoreview.SubmissionWithPatient) []submissionResponse {
	out := make([]submissionResponse, 0, len(items))
	for _, item := range items {
		out = append(out, ToSubmissionResponse(item))
	}
	return out
}

func ToGenerateDraftResponse(submission photoreview.SubmissionWithPatient, draftText string) generateDraftResponse {
	return generateDraftResponse{
		Submission: ToSubmissionResponse(submission),
		DraftText:  draftText,
	}
}

func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}
