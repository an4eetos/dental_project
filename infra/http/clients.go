package http

import (
	"net/http"
	"time"
)

// TelegramHTTPClient downloads media from Telegram (webhook flow).
type TelegramHTTPClient struct {
	Client *http.Client
}

// GeminiHTTPClient calls the Gemini API.
type GeminiHTTPClient struct {
	Client *http.Client
}

// NewTelegramHTTPClient builds a client with the given request timeout.
func NewTelegramHTTPClient(timeout time.Duration) TelegramHTTPClient {
	return TelegramHTTPClient{Client: &http.Client{Timeout: timeout}}
}

// NewGeminiHTTPClient builds a client with the given request timeout.
func NewGeminiHTTPClient(timeout time.Duration) GeminiHTTPClient {
	return GeminiHTTPClient{Client: &http.Client{Timeout: timeout}}
}
