package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds runtime configuration loaded from the environment.
type Config struct {
	TelegramBotToken        string
	TelegramWebhookSecret   string
	GeminiAPIKey            string
	GeminiModel             string
	HTTPAddr                string
	RequestTimeout          time.Duration
	GeminiHTTPTimeout       time.Duration
	TelegramDownloadTimeout time.Duration
	MaxImageDimension       int

	DatabaseURL      string
	JWTSecret        string
	JWTTTL           time.Duration
	TelegramAuthMaxAge time.Duration
	CORSAllowOrigins    []string
	DoctorTelegramIDs   []int64
	AdminTelegramIDs    []int64
	PredictionExamplesPath string

	ReminderPollInterval time.Duration
	AppointmentTimezone  string
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
		JWTTTL:                  getDurationEnv("JWT_TTL", 24*time.Hour),
		TelegramAuthMaxAge:      getDurationEnv("TELEGRAM_AUTH_MAX_AGE", 24*time.Hour),
		PredictionExamplesPath:  getEnv("PREDICTION_EXAMPLES_PATH", "data/prediction_examples.xlsx"),
		ReminderPollInterval:    getDurationEnv("REMINDER_POLL_INTERVAL", 5*time.Minute),
		AppointmentTimezone:     getEnv("APPOINTMENT_TIMEZONE", "UTC"),
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

	cfg.DatabaseURL = os.Getenv("DATABASE_URL")
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	cfg.JWTSecret = os.Getenv("JWT_SECRET")
	if len(cfg.JWTSecret) < 32 {
		return Config{}, fmt.Errorf("JWT_SECRET is required and must be at least 32 characters")
	}

	cfg.CORSAllowOrigins = parseCSV(os.Getenv("CORS_ALLOW_ORIGINS"))
	cfg.DoctorTelegramIDs = parseTelegramIDs(os.Getenv("DOCTOR_TELEGRAM_IDS"))
	cfg.AdminTelegramIDs = parseTelegramIDs(os.Getenv("ADMIN_TELEGRAM_IDS"))

	return cfg, nil
}

func parseTelegramIDs(raw string) []int64 {
	parts := parseCSV(raw)
	out := make([]int64, 0, len(parts))
	for _, p := range parts {
		id, err := strconv.ParseInt(p, 10, 64)
		if err != nil || id <= 0 {
			continue
		}
		out = append(out, id)
	}
	return out
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

func parseCSV(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
