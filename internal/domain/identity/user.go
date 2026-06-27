package identity

import "time"

// User is a Telegram-authenticated account (patient or doctor).
type User struct {
	ID         int64
	TelegramID int64
	Username   string
	FirstName  string
	LastName   string
	AvatarURL  string
	Role       Role
	Blocked    bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// PatientSummary is embedded patient info on doctor-facing appointment views.
type PatientSummary struct {
	ID         int64
	TelegramID int64
	Username   string
	FirstName  string
	LastName   string
}

// TelegramProfile holds data parsed from Telegram initData.
type TelegramProfile struct {
	TelegramID int64
	Username   string
	FirstName  string
	LastName   string
	AvatarURL  string
}
