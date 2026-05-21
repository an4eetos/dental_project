package app

import (
	"testing"
	"time"

	"go.uber.org/fx"

	"github.com/anuarkuanysh/dental_project/infra/config"
)

func testConfig() config.Config {
	return config.Config{
		TelegramBotToken:        "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
		GeminiAPIKey:            "test-gemini-key",
		DatabaseURL:             "postgres://localhost:5432/dental?sslmode=disable",
		JWTSecret:               "abcdefghijklmnopqrstuvwxyz123456",
		HTTPAddr:                ":8080",
		RequestTimeout:          55 * time.Second,
		GeminiHTTPTimeout:       60 * time.Second,
		TelegramDownloadTimeout: 30 * time.Second,
		MaxImageDimension:       1024,
		JWTTTL:                  24 * time.Hour,
		TelegramAuthMaxAge:      24 * time.Hour,
	}
}

func TestModule_FxGraphBuilds(t *testing.T) {
	t.Parallel()

	err := fx.ValidateApp(
		Module(),
		fx.Replace(testConfig()),
		fx.NopLogger,
	)
	if err != nil {
		t.Fatalf("fx graph: %v", err)
	}
}
