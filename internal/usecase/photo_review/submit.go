package photoreview

import (
	"context"

	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
	"github.com/anuarkuanysh/dental_project/internal/domain/identity"
	photoreview "github.com/anuarkuanysh/dental_project/internal/domain/photo_review"
	"github.com/anuarkuanysh/dental_project/internal/port"
)

const patientAckMessage = "Фото получено.\n\n" +
	"Врач клиники рассмотрит его и пришлёт ответ в этот чат в течение 48 часов.\n\n" +
	"⚠️ Ответ врача — справочная информация, не медицинская консультация. " +
	"При симптомах, боли или тревоге за здоровье обратитесь к стоматологу лично."

// SubmitFromTelegram persists a patient photo and acknowledges receipt.
type SubmitFromTelegram struct {
	Users       port.UserRepository
	Submissions port.PhotoSubmissionRepository
	Images      port.ImageProcessor
	Downloader  port.FileDownloader
	Sender      port.MessageSender
	Doctors     port.DoctorRegistry
	Admins      port.AdminRegistry
}

type SubmitInput struct {
	Profile identity.TelegramProfile
	ChatID  int64
	FileID  string
}

func (uc *SubmitFromTelegram) Execute(ctx context.Context, in SubmitInput) error {
	role := identity.RolePatient
	if uc.Doctors.IsDoctor(in.Profile.TelegramID) {
		role = identity.RoleDoctor
	}
	if uc.Admins.IsAdmin(in.Profile.TelegramID) {
		role = identity.RoleAdmin
	}

	user, err := uc.Users.UpsertByTelegramID(ctx, port.UpsertUserParams{
		Profile: in.Profile,
		Role:    role,
	})
	if err != nil {
		return err
	}
	if user.Blocked {
		return domainerrors.ErrUserBlocked
	}

	raw, mimeHint, err := uc.Downloader.DownloadFile(ctx, in.FileID)
	if err != nil {
		return err
	}

	imgBytes, mimeType, err := uc.Images.PrepareForVision(raw, mimeHint)
	if err != nil {
		return err
	}

	_, err = uc.Submissions.Create(ctx, photoreview.CreateParams{
		UserID:         user.ID,
		TelegramFileID: in.FileID,
		ImageData:      imgBytes,
		ImageMIME:      mimeType,
	})
	if err != nil {
		return err
	}

	return uc.Sender.SendText(ctx, in.ChatID, patientAckMessage)
}
