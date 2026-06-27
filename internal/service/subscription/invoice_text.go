package subscription

import "unicode/utf8"

const (
	MaxInvoiceTitleLen       = 32
	MaxInvoiceDescriptionLen = 255
	MaxInvoicePayloadLen     = 128
)

// TruncateInvoiceTitle limits text to Telegram invoice title length (32 characters).
func TruncateInvoiceTitle(title string) string {
	return truncateRunes(title, MaxInvoiceTitleLen)
}

// TruncateInvoiceDescription limits text to Telegram invoice description length.
func TruncateInvoiceDescription(description string) string {
	return truncateRunes(description, MaxInvoiceDescriptionLen)
}

func truncateRunes(text string, maxLen int) string {
	if maxLen <= 0 || utf8.RuneCountInString(text) <= maxLen {
		return text
	}
	runes := []rune(text)
	return string(runes[:maxLen])
}
