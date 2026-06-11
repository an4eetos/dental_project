package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/anuarkuanysh/dental_project/internal/adapters/driving/http/converters"
	httperrors "github.com/anuarkuanysh/dental_project/internal/adapters/driving/http/errors"
	jwtmiddleware "github.com/anuarkuanysh/dental_project/internal/adapters/driving/http/middleware"
	adminuc "github.com/anuarkuanysh/dental_project/internal/usecase/admin"
)

type Handler struct {
	StatisticsUC *adminuc.Statistics
}

func (h *Handler) GetStatistics(c *gin.Context) {
	userID, ok := jwtmiddleware.MustUserID(c)
	if !ok {
		httperrors.Write(c, httperrors.Unauthorized())
		return
	}

	stats, err := h.StatisticsUC.Execute(c.Request.Context(), userID)
	if err != nil {
		httperrors.Write(c, err)
		return
	}

	c.JSON(http.StatusOK, converters.ToStatisticsResponse(stats))
}
