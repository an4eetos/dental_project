package middleware

import (
	"github.com/gin-gonic/gin"

	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
	"github.com/anuarkuanysh/dental_project/internal/domain/identity"
	httperrors "github.com/anuarkuanysh/dental_project/internal/adapters/driving/http/errors"
	"github.com/anuarkuanysh/dental_project/internal/port"
)

// RequireSubscription aborts unless the user has an active subscription.
// Doctors and admins are always allowed.
func RequireSubscription(checker port.SubscriptionChecker, users port.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, ok := MustRole(c)
		if ok && (role == identity.RoleDoctor || role == identity.RoleAdmin) {
			c.Next()
			return
		}

		userID, ok := MustUserID(c)
		if !ok {
			httperrors.Write(c, domainerrors.ErrUnauthorized)
			return
		}

		user, err := users.GetByID(c.Request.Context(), userID)
		if err != nil {
			httperrors.Write(c, err)
			return
		}

		status, err := checker.Check(c.Request.Context(), user)
		if err != nil {
			httperrors.Write(c, err)
			return
		}
		if !status.Active {
			httperrors.Write(c, domainerrors.ErrSubscriptionRequired)
			return
		}
		c.Next()
	}
}
