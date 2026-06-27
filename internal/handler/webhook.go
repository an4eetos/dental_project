package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
	"github.com/anuarkuanysh/dental_project/internal/domain/identity"
	"github.com/anuarkuanysh/dental_project/internal/port"
	photoreviewuc "github.com/anuarkuanysh/dental_project/internal/usecase/photo_review"
	subscriptionuc "github.com/anuarkuanysh/dental_project/internal/usecase/subscription"
)

// WebhookDeps wires interfaces for the webhook handler.
type WebhookDeps struct {
	SubmitPhoto      *photoreviewuc.SubmitFromTelegram
	AnswerPreCheckout *subscriptionuc.AnswerPreCheckout
	ConfirmPayment   *subscriptionuc.ConfirmPayment
	Sender           port.MessageSender
	Log              *slog.Logger
	ReqTimeout       time.Duration
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

		var upd telegramUpdate
		if err := c.ShouldBindJSON(&upd); err != nil {
			deps.Log.Warn("invalid telegram update", "err", err)
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		if upd.PreCheckoutQuery != nil {
			deps.handlePreCheckout(ctx, c, upd.PreCheckoutQuery)
			return
		}

		if upd.Message != nil && upd.Message.SuccessfulPayment != nil {
			deps.handleSuccessfulPayment(ctx, c, upd.Message.SuccessfulPayment)
			return
		}

		if upd.Message != nil {
			deps.handleMessage(ctx, c, upd.Message)
			return
		}

		c.Status(http.StatusOK)
	}
}

func (deps WebhookDeps) handlePreCheckout(ctx context.Context, c *gin.Context, query *preCheckoutQuery) {
	if deps.AnswerPreCheckout == nil {
		c.Status(http.StatusOK)
		return
	}
	if query.Currency != "XTR" {
		deps.Log.Warn("pre-checkout: unsupported currency", "currency", query.Currency)
	}
	err := deps.AnswerPreCheckout.Execute(ctx, subscriptionuc.PreCheckoutInput{
		QueryID: query.ID,
		Payload: query.InvoicePayload,
	})
	if err != nil {
		deps.Log.Warn("pre-checkout failed", "err", err)
	}
	c.Status(http.StatusOK)
}

func (deps WebhookDeps) handleSuccessfulPayment(ctx context.Context, c *gin.Context, payment *successfulPayment) {
	if deps.ConfirmPayment == nil {
		c.Status(http.StatusOK)
		return
	}
	err := deps.ConfirmPayment.Execute(ctx, subscriptionuc.ConfirmPaymentInput{
		Payload:                 payment.InvoicePayload,
		TelegramPaymentChargeID: payment.TelegramPaymentChargeID,
		StarsAmount:             payment.TotalAmount,
	})
	if err != nil {
		deps.Log.Error("confirm payment failed", "err", err)
	}
	c.Status(http.StatusOK)
}

func (deps WebhookDeps) handleMessage(ctx context.Context, c *gin.Context, msg *telegramMessage) {
	chatID := msg.Chat.ID

	if msg.isStartCommand() {
		_ = deps.Sender.SendText(ctx, chatID, welcomeText())
		c.Status(http.StatusOK)
		return
	}

	if len(msg.Photo) == 0 {
		_ = deps.Sender.SendText(ctx, chatID, helpText())
		c.Status(http.StatusOK)
		return
	}

	if msg.From == nil {
		_ = deps.Sender.SendText(ctx, chatID, "Не удалось определить отправителя. Попробуйте ещё раз.")
		c.Status(http.StatusOK)
		return
	}

	photo := msg.Photo[len(msg.Photo)-1]
	err := deps.SubmitPhoto.Execute(ctx, photoreviewuc.SubmitInput{
		Profile: identity.TelegramProfile{
			TelegramID: msg.From.ID,
			Username:   msg.From.UserName,
			FirstName:  msg.From.FirstName,
			LastName:   msg.From.LastName,
		},
		ChatID: chatID,
		FileID: photo.FileID,
	})
	if err != nil {
		deps.Log.Error("photo submission failed", "err", err)
		_ = deps.Sender.SendText(ctx, chatID, submissionErrorText(err))
		c.Status(http.StatusOK)
		return
	}

	c.Status(http.StatusOK)
}

func welcomeText() string {
	return "Добро пожаловать в стоматологический ассистент клиники.\n\n" +
		"Отправьте чёткое фото зубов — врач рассмотрит его и пришлёт ответ в этот чат в течение 48 часов.\n" +
		"Чтобы записаться на приём, откройте Mini App через меню бота.\n\n" +
		"⚠️ Важно:\n" +
		"• Ответ врача — справочная информация, не медицинская консультация.\n" +
		"• Одного фото недостаточно для полной оценки состояния зубов.\n" +
		"• При симптомах, боли, отёке или тревоге за здоровье обратитесь к стоматологу."
}

func helpText() string {
	return "Отправьте фото зубов (не стикер и не документ без сжатого превью).\n\n" +
		"Врач ответит в течение 48 часов.\n" +
		"Для записи на приём откройте Mini App в меню бота."
}

func submissionErrorText(err error) string {
	if errors.Is(err, context.DeadlineExceeded) {
		return "Обработка заняла слишком много времени. Попробуйте отправить фото меньшего размера."
	}
	if errors.Is(err, domainerrors.ErrUserBlocked) {
		return "Ваш аккаунт заблокирован. Обратитесь в клинику для уточнения."
	}
	return "Не удалось принять фото. Попробуйте ещё раз или позже."
}
