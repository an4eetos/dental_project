package content

import (
	"context"
	"strings"

	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
	"github.com/anuarkuanysh/dental_project/internal/port"
	adminuc "github.com/anuarkuanysh/dental_project/internal/usecase/admin"
)

var allowedImageMIME = map[string]struct{}{
	"image/jpeg": {},
	"image/png":  {},
	"image/webp": {},
	"image/gif":  {},
}

var allowedVideoMIME = map[string]struct{}{
	"video/mp4":       {},
	"video/webm":      {},
	"video/quicktime": {},
}

type UploadMediaInput struct {
	AdminUserID   int64
	ContentItemID *int64
	MIMEType      string
	Data          []byte
	MaxBytes      int
}

type UploadMediaResult struct {
	MediaID int64
}

type UploadMedia struct {
	Media   port.ContentMediaRepository
	Images  port.ImageProcessor
	Users   port.UserRepository
}

func (uc *UploadMedia) Execute(ctx context.Context, in UploadMediaInput) (UploadMediaResult, error) {
	if err := adminuc.RequireAdmin(ctx, uc.Users, in.AdminUserID); err != nil {
		return UploadMediaResult{}, err
	}
	if in.MaxBytes > 0 && len(in.Data) > in.MaxBytes {
		return UploadMediaResult{}, domainerrors.ErrContentMediaTooLarge
	}
	if len(in.Data) == 0 {
		return UploadMediaResult{}, domainerrors.ErrInvalidContentMedia
	}

	mime := strings.ToLower(strings.TrimSpace(in.MIMEType))
	data := in.Data

	if _, ok := allowedImageMIME[mime]; ok {
		processed, processedMIME, err := uc.Images.PrepareForVision(data, mime)
		if err != nil {
			return UploadMediaResult{}, domainerrors.ErrInvalidContentMedia
		}
		data = processed
		mime = processedMIME
	} else if _, ok := allowedVideoMIME[mime]; !ok {
		return UploadMediaResult{}, domainerrors.ErrInvalidContentMedia
	}

	id, err := uc.Media.Create(ctx, port.CreateContentMediaParams{
		ContentItemID: in.ContentItemID,
		MIMEType:      mime,
		Data:          data,
	})
	if err != nil {
		return UploadMediaResult{}, err
	}
	return UploadMediaResult{MediaID: id}, nil
}
