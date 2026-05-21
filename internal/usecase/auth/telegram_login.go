package auth

import (
	"context"

	"github.com/anuarkuanysh/dental_project/internal/domain/identity"
	"github.com/anuarkuanysh/dental_project/internal/port"
)

// TelegramLogin authenticates a Mini App user via initData.
type TelegramLogin struct {
	Validator port.TelegramInitDataValidator
	Users     port.UserRepository
	Tokens    port.TokenIssuer
	Doctors   port.DoctorRegistry
}

// LoginResult is returned after successful authentication.
type LoginResult struct {
	Token port.AccessToken
	User  identity.User
}

func (uc *TelegramLogin) Execute(ctx context.Context, initData string) (LoginResult, error) {
	profile, err := uc.Validator.Validate(ctx, initData)
	if err != nil {
		return LoginResult{}, err
	}

	role := identity.RolePatient
	if uc.Doctors.IsDoctor(profile.TelegramID) {
		role = identity.RoleDoctor
	}

	user, err := uc.Users.UpsertByTelegramID(ctx, port.UpsertUserParams{
		Profile: profile,
		Role:    role,
	})
	if err != nil {
		return LoginResult{}, err
	}

	token, err := uc.Tokens.Issue(ctx, user)
	if err != nil {
		return LoginResult{}, err
	}

	return LoginResult{Token: token, User: user}, nil
}
