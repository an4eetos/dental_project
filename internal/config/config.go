package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds runtime configuration loaded from the environment.
type Config struct {
	TelegramBotToken       string
	TelegramWebhookSecret  string
	GeminiAPIKey           string
	GeminiModel            string
	HTTPAddr               string
	RequestTimeout         time.Duration
	GeminiHTTPTimeout      time.Duration
	TelegramDownloadTimeout time.Duration
	MaxImageDimension      int
}

// Load reads configuration from environment variables.
func Load() (Config, error) {
	cfg := Config{
		GeminiModel:             getEnv("GEMINI_MODEL", "gemini-1.5-flash"),
		HTTPAddr:                getEnv("HTTP_ADDR", ":8080"),
		RequestTimeout:          getDurationEnv("REQUEST_TIMEOUT", 55*time.Second),
		GeminiHTTPTimeout:       getDurationEnv("GEMINI_HTTP_TIMEOUT", 60*time.Second),
		TelegramDownloadTimeout: getDurationEnv("TELEGRAM_DOWNLOAD_TIMEOUT", 30*time.Second),
		MaxImageDimension:       getIntEnv("MAX_IMAGE_DIMENSION", 1024),
	}

	cfg.TelegramBotToken = os.Getenv("TELEGRAM_BOT_TOKEN")
	if cfg.TelegramBotToken == "" {
		return Config{}, fmt.Errorf("TELEGRAM_BOT_TOKEN is required")
	}

	cfg.GeminiAPIKey = os.Getenv("GEMINI_API_KEY")
	if cfg.GeminiAPIKey == "" {
		return Config{}, fmt.Errorf("GEMINI_API_KEY is required")
	}

	cfg.TelegramWebhookSecret = os.Getenv("TELEGRAM_WEBHOOK_SECRET")

	return cfg, nil
}

func getEnv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

func getIntEnv(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}
