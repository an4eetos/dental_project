package port

// AdminRegistry marks configured Telegram user IDs as clinic admins.
type AdminRegistry interface {
	IsAdmin(telegramID int64) bool
}
