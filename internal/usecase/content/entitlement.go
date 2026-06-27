package content

import (
	"context"

	"github.com/anuarkuanysh/dental_project/internal/domain/identity"
	"github.com/anuarkuanysh/dental_project/internal/port"
)

func userHasEntitlement(ctx context.Context, checker port.SubscriptionChecker, user identity.User) (bool, error) {
	if user.Role == identity.RoleDoctor || user.Role == identity.RoleAdmin {
		return true, nil
	}
	status, err := checker.Check(ctx, user)
	if err != nil {
		return false, err
	}
	return status.Active, nil
}
