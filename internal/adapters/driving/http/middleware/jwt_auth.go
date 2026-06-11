package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
	"github.com/anuarkuanysh/dental_project/internal/domain/identity"
	httperrors "github.com/anuarkuanysh/dental_project/internal/adapters/driving/http/errors"
	"github.com/anuarkuanysh/dental_project/internal/port"
)

const (
	ContextUserIDKey = "user_id"
	ContextRoleKey   = "user_role"
)

// JWTAuth validates Bearer tokens and stores principal in context.
func JWTAuth(tokens port.TokenIssuer) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			httperrors.Write(c, domainerrors.ErrUnauthorized)
			return
		}
		raw := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
		principal, err := tokens.ParsePrincipal(c.Request.Context(), raw)
		if err != nil {
			httperrors.Write(c, err)
			return
		}
		c.Set(ContextUserIDKey, principal.UserID)
		c.Set(ContextRoleKey, principal.Role)
		c.Next()
	}
}

// RequireDoctor aborts unless the authenticated user has the doctor role.
func RequireDoctor() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, ok := MustRole(c)
		if !ok || role != identity.RoleDoctor {
			httperrors.Write(c, domainerrors.ErrForbidden)
			return
		}
		c.Next()
	}
}

// RequireAdmin aborts unless the authenticated user has the admin role.
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, ok := MustRole(c)
		if !ok || role != identity.RoleAdmin {
			httperrors.Write(c, domainerrors.ErrForbidden)
			return
		}
		c.Next()
	}
}

// MustUserID reads authenticated user id from context.
func MustUserID(c *gin.Context) (int64, bool) {
	v, ok := c.Get(ContextUserIDKey)
	if !ok {
		return 0, false
	}
	id, ok := v.(int64)
	return id, ok && id > 0
}

// MustRole reads authenticated role from context.
func MustRole(c *gin.Context) (identity.Role, bool) {
	v, ok := c.Get(ContextRoleKey)
	if !ok {
		return "", false
	}
	role, ok := v.(identity.Role)
	return role, ok && role.Valid()
}

// CORS allows configured origins for the Mini App.
func CORS(log *slog.Logger, allowOrigins []string) gin.HandlerFunc {
	if log == nil {
		log = slog.Default()
	}
	allowAll := len(allowOrigins) == 0
	allowed := make(map[string]struct{}, len(allowOrigins))
	for _, o := range allowOrigins {
		allowed[o] = struct{}{}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		originAllowed := origin == "" || allowAll
		if origin != "" {
			if allowAll {
				c.Header("Access-Control-Allow-Origin", origin)
				originAllowed = true
			} else if _, ok := allowed[origin]; ok {
				c.Header("Access-Control-Allow-Origin", origin)
				originAllowed = true
			} else {
				log.Warn("cors: origin not allowed",
					"origin", origin,
					"method", c.Request.Method,
					"path", c.Request.URL.Path,
					"allowed", allowOrigins,
				)
			}
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-ID")
			c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		}
		if c.Request.Method == http.MethodOptions {
			if origin != "" {
				log.Debug("cors: preflight",
					"origin", origin,
					"path", c.Request.URL.Path,
					"allowed", originAllowed,
				)
			}
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
