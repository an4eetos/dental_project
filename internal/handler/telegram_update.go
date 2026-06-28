package handler

import "strings"

// telegramUpdate mirrors Telegram webhook JSON (payment fields not in older bot API structs).
type telegramUpdate struct {
	UpdateID         int               `json:"update_id"`
	Message          *telegramMessage  `json:"message"`
	PreCheckoutQuery *preCheckoutQuery `json:"pre_checkout_query"`
}

type telegramUser struct {
	ID        int64  `json:"id"`
	UserName  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type telegramChat struct {
	ID int64 `json:"id"`
}

type telegramPhotoSize struct {
	FileID string `json:"file_id"`
}

type telegramVideo struct {
	FileID   string `json:"file_id"`
	MimeType string `json:"mime_type"`
}

type telegramDocument struct {
	FileID   string `json:"file_id"`
	MimeType string `json:"mime_type"`
}

type successfulPayment struct {
	Currency                string `json:"currency"`
	TotalAmount             int    `json:"total_amount"`
	InvoicePayload          string `json:"invoice_payload"`
	TelegramPaymentChargeID string `json:"telegram_payment_charge_id"`
}

type telegramMessage struct {
	MessageID         int                `json:"message_id"`
	Chat              telegramChat       `json:"chat"`
	From              *telegramUser      `json:"from"`
	Text              string             `json:"text"`
	Entities          []messageEntity    `json:"entities"`
	Photo             []telegramPhotoSize `json:"photo"`
	Video             *telegramVideo      `json:"video"`
	Document          *telegramDocument   `json:"document"`
	SuccessfulPayment *successfulPayment  `json:"successful_payment"`
}

func (m *telegramMessage) submissionMedia() (fileID, mediaType string, ok bool) {
	if m == nil {
		return "", "", false
	}
	if len(m.Photo) > 0 {
		return m.Photo[len(m.Photo)-1].FileID, "photo", true
	}
	if m.Video != nil && m.Video.FileID != "" {
		return m.Video.FileID, "video", true
	}
	if m.Document != nil && m.Document.FileID != "" && isVideoMIME(m.Document.MimeType) {
		return m.Document.FileID, "video", true
	}
	return "", "", false
}

func isVideoMIME(mime string) bool {
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(mime)), "video/")
}

type messageEntity struct {
	Type   string `json:"type"`
	Offset int    `json:"offset"`
	Length int    `json:"length"`
}

func (m *telegramMessage) isStartCommand() bool {
	if m == nil || m.Text == "" {
		return false
	}
	for _, e := range m.Entities {
		if e.Type == "bot_command" && e.Offset == 0 && e.Length > 0 {
			cmd := m.Text[e.Offset : e.Offset+e.Length]
			return cmd == "/start"
		}
	}
	return false
}

type preCheckoutQuery struct {
	ID             string       `json:"id"`
	From           telegramUser `json:"from"`
	Currency       string       `json:"currency"`
	TotalAmount    int          `json:"total_amount"`
	InvoicePayload string       `json:"invoice_payload"`
}
