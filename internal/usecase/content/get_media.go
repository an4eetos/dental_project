package content

import (
	"context"

	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
	contentdomain "github.com/anuarkuanysh/dental_project/internal/domain/content"
	"github.com/anuarkuanysh/dental_project/internal/port"
)

type MediaResult struct {
	Data     []byte
	MIMEType string
}

// GetMedia streams uploaded media when the user can access the parent content item.
type GetMedia struct {
	Content port.ContentRepository
	Media   port.ContentMediaRepository
	Checker port.SubscriptionChecker
	Users   port.UserRepository
}

func (uc *GetMedia) Execute(ctx context.Context, userID, mediaID int64) (MediaResult, error) {
	user, err := uc.Users.GetByID(ctx, userID)
	if err != nil {
		return MediaResult{}, err
	}

	rec, err := uc.Media.GetByID(ctx, mediaID)
	if err != nil {
		return MediaResult{}, err
	}
	if rec.ContentItemID == nil {
		return MediaResult{}, domainerrors.ErrContentMediaNotFound
	}

	item, err := uc.Content.GetByID(ctx, *rec.ContentItemID)
	if err != nil {
		return MediaResult{}, err
	}
	if !item.Published {
		return MediaResult{}, domainerrors.ErrContentMediaNotFound
	}

	entitled, err := userHasEntitlement(ctx, uc.Checker, user)
	if err != nil {
		return MediaResult{}, err
	}
	if !contentdomain.IsAccessible(item.Access, entitled) {
		return MediaResult{}, domainerrors.ErrSubscriptionRequired
	}

	return MediaResult{Data: rec.Data, MIMEType: rec.MIMEType}, nil
}
