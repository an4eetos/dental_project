package photoreview

import (
	"time"

	"github.com/anuarkuanysh/dental_project/internal/domain/identity"
	"github.com/anuarkuanysh/dental_project/internal/model"
)

// Submission is a patient photo awaiting or after clinic review.
type Submission struct {
	ID             int64
	UserID         int64
	TelegramFileID string
	MediaType      MediaType
	ImageMIME      string
	Status         Status
	AIDraft        *model.Analysis
	DoctorResponse string
	RespondedBy    int64
	RespondedAt    *time.Time
	CreatedAt      time.Time
}

// SubmissionWithPatient joins submission with patient profile for doctor inbox views.
type SubmissionWithPatient struct {
	Submission Submission
	Patient    identity.PatientSummary
}

// CreateParams is input for persisting a new submission from the bot.
type CreateParams struct {
	UserID         int64
	TelegramFileID string
	MediaType      MediaType
	ImageData      []byte
	ImageMIME      string
}
