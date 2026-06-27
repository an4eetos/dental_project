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
	Blocked    bool   `json:"blocked,omitempty"`
}

type AuthTelegramResponse struct {
	AccessToken  string                      `json:"access_token"`
	ExpiresAt    time.Time                   `json:"expires_at"`
	User         UserResponse                `json:"user"`
	Subscription *SubscriptionStatusResponse `json:"subscription,omitempty"`
}

func ToAuthTelegramResponse(result authuc.LoginResult, subscription *SubscriptionStatusResponse) AuthTelegramResponse {
	resp := AuthTelegramResponse{
		AccessToken: result.Token.Value,
		ExpiresAt:   result.Token.ExpiresAt,
		User:        ToUserResponse(result.User),
		Subscription: subscription,
	}
	return resp
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
		Blocked:    u.Blocked,
	}
}

func ToAccessTokenPreview(t port.AccessToken) map[string]any {
	return map[string]any{
		"expires_at": t.ExpiresAt,
	}
}
