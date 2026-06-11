package admin

import (
	"context"

	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
	admindomain "github.com/anuarkuanysh/dental_project/internal/domain/admin"
	"github.com/anuarkuanysh/dental_project/internal/domain/identity"
	"github.com/anuarkuanysh/dental_project/internal/port"
)

// Statistics returns aggregated clinic metrics for admins.
type Statistics struct {
	Stats port.StatsRepository
	Users port.UserRepository
}

func (uc *Statistics) Execute(ctx context.Context, adminUserID int64) (admindomain.Statistics, error) {
	user, err := uc.Users.GetByID(ctx, adminUserID)
	if err != nil {
		return admindomain.Statistics{}, err
	}
	if user.Role != identity.RoleAdmin {
		return admindomain.Statistics{}, domainerrors.ErrForbidden
	}
	return uc.Stats.GetStatistics(ctx)
}
