package port

import "context"

// FileDownloader downloads Telegram-hosted files by file ID.
type FileDownloader interface {
	DownloadFile(ctx context.Context, fileID string) ([]byte, string, error)
}

// MessageSender sends plain text messages to a Telegram chat.
type MessageSender interface {
	SendText(ctx context.Context, chatID int64, text string) error
}
