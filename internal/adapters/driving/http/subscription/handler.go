package subscription

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/anuarkuanysh/dental_project/internal/adapters/driving/http/converters"
	httperrors "github.com/anuarkuanysh/dental_project/internal/adapters/driving/http/errors"
	jwtmiddleware "github.com/anuarkuanysh/dental_project/internal/adapters/driving/http/middleware"
	subscriptionuc "github.com/anuarkuanysh/dental_project/internal/usecase/subscription"
	"github.com/anuarkuanysh/dental_project/internal/port"
)

type Handler struct {
	GetStatus     *subscriptionuc.GetStatus
	CreateInvoiceUC *subscriptionuc.CreateInvoice
	Users         port.UserRepository
	Log           *slog.Logger
}

func (h *Handler) GetMe(c *gin.Context) {
	userID, ok := jwtmiddleware.MustUserID(c)
	if !ok {
		httperrors.Write(c, httperrors.Unauthorized())
		return
	}
	user, err := h.Users.GetByID(c.Request.Context(), userID)
	if err != nil {
		httperrors.Write(c, err)
		return
	}
	result, err := h.GetStatus.Execute(c.Request.Context(), user)
	if err != nil {
		httperrors.Write(c, err)
		return
	}
	c.JSON(http.StatusOK, converters.ToSubscriptionStatusResponse(result))
}

func (h *Handler) CreateInvoice(c *gin.Context) {
	userID, ok := jwtmiddleware.MustUserID(c)
	if !ok {
		httperrors.Write(c, httperrors.Unauthorized())
		return
	}
	result, err := h.CreateInvoiceUC.Execute(c.Request.Context(), userID)
	if err != nil {
		if h.Log != nil {
			h.Log.Warn("subscription invoice failed", "user_id", userID, "err", err)
		}
		httperrors.Write(c, err)
		return
	}
	c.JSON(http.StatusOK, converters.CreateInvoiceResponse{InvoiceLink: result.InvoiceLink})
}
