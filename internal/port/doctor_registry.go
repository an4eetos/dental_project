package port

// DoctorRegistry identifies Telegram accounts that act as clinic doctors.
type DoctorRegistry interface {
	IsDoctor(telegramID int64) bool
}
