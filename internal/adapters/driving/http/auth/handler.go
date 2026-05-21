package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"

	httperrors "github.com/anuarkuanysh/dental_project/internal/adapters/driving/http/errors"
	"github.com/anuarkuanysh/dental_project/internal/adapters/driving/http/converters"
	authuc "github.com/anuarkuanysh/dental_project/internal/usecase/auth"
)

type Handler struct {
	Login *authuc.TelegramLogin
}

type telegramAuthRequest struct {
	InitData string `json:"init_data" binding:"required"`
}

func (h *Handler) Telegram(c *gin.Context) {
	var req telegramAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, httperrors.APIError{
			Code:    "validation_error",
			Message: "init_data is required",
		})
		return
	}

	result, err := h.Login.Execute(c.Request.Context(), req.InitData)
	if err != nil {
		httperrors.Write(c, err)
		return
	}

	c.JSON(http.StatusOK, converters.ToAuthTelegramResponse(result))
}
