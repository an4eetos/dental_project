package photoreview

import (
	"context"
	"strings"

	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
	photoreview "github.com/anuarkuanysh/dental_project/internal/domain/photo_review"
	"github.com/anuarkuanysh/dental_project/internal/port"
)

const doctorReplyPrefix = "Ответ врача клиники по вашему фото:\n\n"
const doctorReplySuffix = "\n\n⚠️ Справочная информация, не медицинская консультация. " +
	"При симптомах, боли или тревоге за здоровье обратитесь к стоматологу лично."

// Respond saves a doctor answer and delivers it to the patient via the bot.
type Respond struct {
	Submissions port.PhotoSubmissionRepository
	Users       port.UserRepository
	Sender      port.MessageSender
	Clock       port.Clock
}

type RespondInput struct {
	DoctorUserID int64
	SubmissionID int64
	Response     string
}

func (uc *Respond) Execute(ctx context.Context, in RespondInput) (photoreview.SubmissionWithPatient, error) {
	if err := requireDoctor(ctx, uc.Users, in.DoctorUserID); err != nil {
		return photoreview.SubmissionWithPatient{}, err
	}

	response := strings.TrimSpace(in.Response)
	if response == "" {
		return photoreview.SubmissionWithPatient{}, domainerrors.ErrEmptyDoctorResponse
	}

	item, err := uc.Submissions.GetByID(ctx, in.SubmissionID)
	if err != nil {
		return photoreview.SubmissionWithPatient{}, err
	}
	if item.Submission.Status != photoreview.StatusPending {
		return photoreview.SubmissionWithPatient{}, domainerrors.ErrSubmissionAlreadyAnswered
	}

	now := uc.Clock.Now()
	if err := uc.Submissions.MarkAnswered(ctx, in.SubmissionID, in.DoctorUserID, response, now); err != nil {
		return photoreview.SubmissionWithPatient{}, err
	}

	message := doctorReplyPrefix + response + doctorReplySuffix
	if err := uc.Sender.SendText(ctx, item.Patient.TelegramID, message); err != nil {
		return photoreview.SubmissionWithPatient{}, err
	}

	item.Submission.Status = photoreview.StatusAnswered
	item.Submission.DoctorResponse = response
	item.Submission.RespondedBy = in.DoctorUserID
	item.Submission.RespondedAt = &now
	return item, nil
}
