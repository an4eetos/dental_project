package converters

import (
	"time"

	admindomain "github.com/anuarkuanysh/dental_project/internal/domain/admin"
	"github.com/anuarkuanysh/dental_project/internal/domain/identity"
)

type AdminUserResponse struct {
	ID         int64     `json:"id"`
	TelegramID int64     `json:"telegram_id"`
	Role       string    `json:"role"`
	Username   string    `json:"username,omitempty"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name,omitempty"`
	AvatarURL  string    `json:"avatar_url,omitempty"`
	Blocked    bool      `json:"blocked"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type AdminUserOverviewResponse struct {
	AdminUserResponse
	AppointmentCount     int64 `json:"appointment_count"`
	PhotoSubmissionCount int64 `json:"photo_submission_count"`
}

type AdminUserListResponse struct {
	Users []AdminUserResponse `json:"users"`
}

func ToAdminUserResponse(u identity.User) AdminUserResponse {
	return AdminUserResponse{
		ID:         u.ID,
		TelegramID: u.TelegramID,
		Role:       u.Role.String(),
		Username:   u.Username,
		FirstName:  u.FirstName,
		LastName:   u.LastName,
		AvatarURL:  u.AvatarURL,
		Blocked:    u.Blocked,
		CreatedAt:  u.CreatedAt,
		UpdatedAt:  u.UpdatedAt,
	}
}

func ToAdminUserOverviewResponse(o admindomain.UserOverview) AdminUserOverviewResponse {
	return AdminUserOverviewResponse{
		AdminUserResponse: AdminUserResponse{
			ID:         o.ID,
			TelegramID: o.TelegramID,
			Role:       o.Role.String(),
			Username:   o.Username,
			FirstName:  o.FirstName,
			LastName:   o.LastName,
			AvatarURL:  o.AvatarURL,
			Blocked:    o.Blocked,
			CreatedAt:  o.CreatedAt,
			UpdatedAt:  o.UpdatedAt,
		},
		AppointmentCount:     o.AppointmentCount,
		PhotoSubmissionCount: o.PhotoSubmissionCount,
	}
}

func ToAdminUserListResponse(users []identity.User) AdminUserListResponse {
	out := make([]AdminUserResponse, 0, len(users))
	for _, u := range users {
		out = append(out, ToAdminUserResponse(u))
	}
	return AdminUserListResponse{Users: out}
}
