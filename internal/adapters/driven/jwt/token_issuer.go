package jwt

import (
	"context"
	"fmt"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"

	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
	"github.com/anuarkuanysh/dental_project/internal/domain/identity"
	"github.com/anuarkuanysh/dental_project/internal/port"
)

type claims struct {
	UserID int64  `json:"uid"`
	Role   string `json:"role"`
	gojwt.RegisteredClaims
}

// TokenIssuer implements port.TokenIssuer using HS256 JWT.
type TokenIssuer struct {
	Secret []byte
	TTL    time.Duration
}

func NewTokenIssuer(secret string, ttl time.Duration) *TokenIssuer {
	return &TokenIssuer{Secret: []byte(secret), TTL: ttl}
}

func (i *TokenIssuer) Issue(_ context.Context, user identity.User) (port.AccessToken, error) {
	now := time.Now().UTC()
	expires := now.Add(i.TTL)
	role := user.Role
	if !role.Valid() {
		role = identity.RolePatient
	}
	c := claims{
		UserID: user.ID,
		Role:   role.String(),
		RegisteredClaims: gojwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", user.TelegramID),
			IssuedAt:  gojwt.NewNumericDate(now),
			ExpiresAt: gojwt.NewNumericDate(expires),
		},
	}
	token := gojwt.NewWithClaims(gojwt.SigningMethodHS256, c)
	signed, err := token.SignedString(i.Secret)
	if err != nil {
		return port.AccessToken{}, err
	}
	return port.AccessToken{Value: signed, ExpiresAt: expires}, nil
}

func (i *TokenIssuer) ParsePrincipal(_ context.Context, token string) (port.AuthPrincipal, error) {
	parsed, err := gojwt.ParseWithClaims(token, &claims{}, func(t *gojwt.Token) (any, error) {
		if t.Method != gojwt.SigningMethodHS256 {
			return nil, domainerrors.ErrInvalidToken
		}
		return i.Secret, nil
	})
	if err != nil || !parsed.Valid {
		return port.AuthPrincipal{}, domainerrors.ErrInvalidToken
	}
	c, ok := parsed.Claims.(*claims)
	if !ok || c.UserID == 0 {
		return port.AuthPrincipal{}, domainerrors.ErrInvalidToken
	}
	role := identity.Role(c.Role)
	if !role.Valid() {
		role = identity.RolePatient
	}
	return port.AuthPrincipal{UserID: c.UserID, Role: role}, nil
}
