package middleware

import (
	"github.com/gin-gonic/gin"

	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
	"github.com/anuarkuanysh/dental_project/internal/domain/identity"
	httperrors "github.com/anuarkuanysh/dental_project/internal/adapters/driving/http/errors"
	"github.com/anuarkuanysh/dental_project/internal/port"
)

// RequireNotBlocked rejects authenticated users whose account is blocked.
func RequireNotBlocked(users port.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, ok := MustRole(c)
		if ok && role == identity.RoleAdmin {
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
		if user.Blocked {
			httperrors.Write(c, domainerrors.ErrUserBlocked)
			return
		}
		c.Next()
	}
}
