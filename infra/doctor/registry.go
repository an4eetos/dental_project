package doctor

import "github.com/anuarkuanysh/dental_project/internal/port"

// TelegramIDRegistry marks configured Telegram user IDs as doctors.
type TelegramIDRegistry struct {
	ids map[int64]struct{}
}

func NewTelegramIDRegistry(telegramIDs []int64) *TelegramIDRegistry {
	ids := make(map[int64]struct{}, len(telegramIDs))
	for _, id := range telegramIDs {
		if id > 0 {
			ids[id] = struct{}{}
		}
	}
	return &TelegramIDRegistry{ids: ids}
}

func (r *TelegramIDRegistry) IsDoctor(telegramID int64) bool {
	_, ok := r.ids[telegramID]
	return ok
}

var _ port.DoctorRegistry = (*TelegramIDRegistry)(nil)
