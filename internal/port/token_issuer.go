package port

import (
	"context"
	"time"

	"github.com/anuarkuanysh/dental_project/internal/domain/identity"
)

// AccessToken is a signed bearer token for API access.
type AccessToken struct {
	Value     string
	ExpiresAt time.Time
}

// AuthPrincipal is extracted from a valid access token.
type AuthPrincipal struct {
	UserID int64
	Role   identity.Role
}

// TokenIssuer creates and validates JWT access tokens.
type TokenIssuer interface {
	Issue(ctx context.Context, user identity.User) (AccessToken, error)
	ParsePrincipal(ctx context.Context, token string) (AuthPrincipal, error)
}
