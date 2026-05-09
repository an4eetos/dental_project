package telegrambot

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/anuarkuanysh/dental_project/internal/port"
)

// Client implements Telegram file download and messaging using the Bot API.
type Client struct {
	bot    *tgbotapi.BotAPI
	httpCl *http.Client
}

var (
	_ port.FileDownloader = (*Client)(nil)
	_ port.MessageSender  = (*Client)(nil)
)

// New creates a Telegram Bot API client wrapper.
func New(bot *tgbotapi.BotAPI, httpClient *http.Client) *Client {
	return &Client{bot: bot, httpCl: httpClient}
}

// DownloadFile resolves the file URL via Telegram and downloads bytes.
func (c *Client) DownloadFile(ctx context.Context, fileID string) ([]byte, string, error) {
	fileCfg := tgbotapi.FileConfig{FileID: fileID}
	file, err := c.bot.GetFile(fileCfg)
	if err != nil {
		return nil, "", fmt.Errorf("get file: %w", err)
	}

	link := file.Link(c.bot.Token)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, link, nil)
	if err != nil {
		return nil, "", fmt.Errorf("build request: %w", err)
	}

	resp, err := c.httpCl.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, "", fmt.Errorf("download status %d: %s", resp.StatusCode, string(body))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("read body: %w", err)
	}

	ext := path.Ext(file.FilePath)
	mime := mimeFromExt(ext)
	return data, mime, nil
}

// SendText sends a UTF-8 text message (respects Telegram length limits by truncation upstream if needed).
func (c *Client) SendText(ctx context.Context, chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = ""

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	_, err := c.bot.Send(msg)
	if err != nil {
		return fmt.Errorf("send message: %w", err)
	}
	return nil
}

func mimeFromExt(ext string) string {
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}
