package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"

	"github.com/anuarkuanysh/dental_project/internal/config"
	"github.com/anuarkuanysh/dental_project/internal/gemini"
	"github.com/anuarkuanysh/dental_project/internal/handler"
	"github.com/anuarkuanysh/dental_project/internal/imageproc"
	"github.com/anuarkuanysh/dental_project/internal/middleware"
	"github.com/anuarkuanysh/dental_project/internal/telegrambot"
)

func main() {
	_ = godotenv.Load()

	log := newLogger()

	cfg, err := config.Load()
	if err != nil {
		log.Error("config", "err", err)
		os.Exit(1)
	}

	bot, err := tgbotapi.NewBotAPI(cfg.TelegramBotToken)
	if err != nil {
		log.Error("telegram bot api", "err", err)
		os.Exit(1)
	}
	bot.Debug = false
	log.Info("authorized bot", "username", bot.Self.UserName)

	telegramHTTP := &http.Client{Timeout: cfg.TelegramDownloadTimeout}
	geminiHTTP := &http.Client{Timeout: cfg.GeminiHTTPTimeout}

	tgClient := telegrambot.New(bot, telegramHTTP)
	imgProc := imageproc.New(cfg.MaxImageDimension)
	geminiClient := gemini.New(geminiHTTP, cfg.GeminiAPIKey, cfg.GeminiModel, log)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(middleware.Recovery(log), middleware.RequestID(), middleware.SlogLogger(log))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.POST("/webhook",
		middleware.WebhookSecret(cfg.TelegramWebhookSecret),
		handler.Webhook(handler.WebhookDeps{
			Downloader: tgClient,
			Sender:     tgClient,
			Analyzer:   geminiClient,
			Images:     imgProc,
			Log:        log,
			ReqTimeout: cfg.RequestTimeout,
		}),
	)

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       cfg.RequestTimeout + 5*time.Second,
		WriteTimeout:      cfg.RequestTimeout + 5*time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		log.Info("listening", "addr", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server", "err", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("shutdown", "err", err)
	}
	log.Info("stopped")
}

func newLogger() *slog.Logger {
	level := slog.LevelInfo
	if os.Getenv("LOG_LEVEL") == "debug" {
		level = slog.LevelDebug
	}
	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	return slog.New(h)
}
