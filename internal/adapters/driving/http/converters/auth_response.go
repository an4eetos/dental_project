package converters

import (
	"time"

	"github.com/anuarkuanysh/dental_project/internal/domain/identity"
	"github.com/anuarkuanysh/dental_project/internal/port"
	authuc "github.com/anuarkuanysh/dental_project/internal/usecase/auth"
)

type UserResponse struct {
	ID         int64  `json:"id"`
	TelegramID int64  `json:"telegram_id"`
	Role       string `json:"role"`
	Username   string `json:"username,omitempty"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name,omitempty"`
	AvatarURL  string `json:"avatar_url,omitempty"`
}

type AuthTelegramResponse struct {
	AccessToken string       `json:"access_token"`
	ExpiresAt   time.Time    `json:"expires_at"`
	User        UserResponse `json:"user"`
}

func ToAuthTelegramResponse(result authuc.LoginResult) AuthTelegramResponse {
	return AuthTelegramResponse{
		AccessToken: result.Token.Value,
		ExpiresAt:   result.Token.ExpiresAt,
		User:        ToUserResponse(result.User),
	}
}

func ToUserResponse(u identity.User) UserResponse {
	return UserResponse{
		ID:         u.ID,
		TelegramID: u.TelegramID,
		Role:       u.Role.String(),
		Username:   u.Username,
		FirstName:  u.FirstName,
		LastName:   u.LastName,
		AvatarURL:  u.AvatarURL,
	}
}

func ToAccessTokenPreview(t port.AccessToken) map[string]any {
	return map[string]any{
		"expires_at": t.ExpiresAt,
	}
}
