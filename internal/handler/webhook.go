package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/anuarkuanysh/dental_project/internal/port"
)

// WebhookDeps wires interfaces for the webhook handler.
type WebhookDeps struct {
	Downloader   port.FileDownloader
	Sender       port.MessageSender
	Analyzer     port.VisionAnalyzer
	Images       port.ImageProcessor
	Log          *slog.Logger
	ReqTimeout   time.Duration
}

// Webhook returns a Gin handler for Telegram updates (webhook).
func Webhook(deps WebhookDeps) gin.HandlerFunc {
	if deps.Log == nil {
		deps.Log = slog.Default()
	}

	return func(c *gin.Context) {
		ctx := c.Request.Context()
		if deps.ReqTimeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, deps.ReqTimeout)
			defer cancel()
			c.Request = c.Request.WithContext(ctx)
		}

		var upd tgbotapi.Update
		if err := c.ShouldBindJSON(&upd); err != nil {
			deps.Log.Warn("invalid telegram update", "err", err)
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		if upd.Message == nil {
			c.Status(http.StatusOK)
			return
		}

		msg := upd.Message
		chatID := msg.Chat.ID

		if msg.IsCommand() && msg.Command() == "start" {
			_ = deps.Sender.SendText(ctx, chatID, welcomeText())
			c.Status(http.StatusOK)
			return
		}

		if len(msg.Photo) == 0 {
			_ = deps.Sender.SendText(ctx, chatID, helpText())
			c.Status(http.StatusOK)
			return
		}

		photo := msg.Photo[len(msg.Photo)-1]
		raw, mimeHint, err := deps.Downloader.DownloadFile(ctx, photo.FileID)
		if err != nil {
			deps.Log.Error("download failed", "err", err)
			_ = deps.Sender.SendText(ctx, chatID, "Не удалось загрузить изображение. Попробуйте ещё раз.")
			c.Status(http.StatusOK)
			return
		}

		imgBytes, mimeType, err := deps.Images.PrepareForVision(raw, mimeHint)
		if err != nil {
			deps.Log.Warn("image prepare failed", "err", err)
			_ = deps.Sender.SendText(ctx, chatID, "Не удалось обработать изображение. Отправьте фото в формате JPG, PNG или WEBP.")
			c.Status(http.StatusOK)
			return
		}

		analysis, err := deps.Analyzer.AnalyzeTeethImage(ctx, imgBytes, mimeType)
		if err != nil {
			deps.Log.Error("gemini analysis failed", "err", err)
			if errors.Is(err, context.DeadlineExceeded) {
				_ = deps.Sender.SendText(ctx, chatID, "Анализ занял слишком много времени. Попробуйте отправить фото меньшего размера.")
			} else {
				_ = deps.Sender.SendText(ctx, chatID, "Не удалось выполнить анализ. Попробуйте позже.")
			}
			c.Status(http.StatusOK)
			return
		}

		reply := FormatTelegramReply(analysis)
		if err := deps.Sender.SendText(ctx, chatID, reply); err != nil {
			deps.Log.Error("send reply failed", "err", err)
		}

		c.Status(http.StatusOK)
	}
}

func welcomeText() string {
	return "Добро пожаловать в стоматологический AI-ассистент.\n\n" +
		"Отправьте чёткое фото зубов для общих, недиагностических наблюдений.\n" +
		"Чтобы записаться на приём, откройте Mini App через меню бота.\n\n" +
		"Бот не ставит диагнозы и не является медицинской консультацией.\n" +
		"Не используйте ответы для решений о лечении — обратитесь к стоматологу."
}

func helpText() string {
	return "Отправьте фото зубов (не стикер и не документ без сжатого превью).\n\n" +
		"Для записи на приём откройте Mini App в меню бота.\n\n" +
		"Ответы носят информационный характер и не являются медицинской консультацией."
}
