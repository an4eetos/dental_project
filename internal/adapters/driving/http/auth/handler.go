package auth

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	httperrors "github.com/anuarkuanysh/dental_project/internal/adapters/driving/http/errors"
	"github.com/anuarkuanysh/dental_project/internal/adapters/driving/http/converters"
	authuc "github.com/anuarkuanysh/dental_project/internal/usecase/auth"
)

type Handler struct {
	Login *authuc.TelegramLogin
	Log   *slog.Logger
}

type telegramAuthRequest struct {
	InitData string `json:"init_data" binding:"required"`
}

func (h *Handler) Telegram(c *gin.Context) {
	log := h.Log
	if log == nil {
		log = slog.Default()
	}

	var req telegramAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn("telegram auth: invalid request body",
			"origin", c.GetHeader("Origin"),
			"err", err,
		)
		c.JSON(http.StatusBadRequest, httperrors.APIError{
			Code:    "validation_error",
			Message: "init_data is required",
		})
		return
	}

	log.Info("telegram auth: attempt",
		"origin", c.GetHeader("Origin"),
		"init_data_len", len(req.InitData),
	)

	result, err := h.Login.Execute(c.Request.Context(), req.InitData)
	if err != nil {
		log.Warn("telegram auth: failed",
			"origin", c.GetHeader("Origin"),
			"err", err.Error(),
		)
		httperrors.Write(c, err)
		return
	}

	log.Info("telegram auth: ok",
		"user_id", result.User.ID,
		"telegram_id", result.User.TelegramID,
		"role", result.User.Role,
	)
	c.JSON(http.StatusOK, converters.ToAuthTelegramResponse(result))
}
