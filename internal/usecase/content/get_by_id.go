package content

import (
	"context"

	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
	"github.com/anuarkuanysh/dental_project/internal/port"
)

// GetByID returns a published content item when the user is entitled.
type GetByID struct {
	Content port.ContentRepository
	Checker port.SubscriptionChecker
	Users   port.UserRepository
}

func (uc *GetByID) Execute(ctx context.Context, userID, contentID int64) (ItemView, error) {
	user, err := uc.Users.GetByID(ctx, userID)
	if err != nil {
		return ItemView{}, err
	}

	item, err := uc.Content.GetByID(ctx, contentID)
	if err != nil {
		return ItemView{}, err
	}
	if !item.Published {
		return ItemView{}, domainerrors.ErrContentNotFound
	}

	entitled, err := userHasEntitlement(ctx, uc.Checker, user)
	if err != nil {
		return ItemView{}, err
	}

	view := toItemView(item, entitled)
	if view.Locked {
		return ItemView{}, domainerrors.ErrSubscriptionRequired
	}
	return view, nil
}
