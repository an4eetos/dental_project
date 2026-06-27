package content

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/anuarkuanysh/dental_project/internal/adapters/driving/http/converters"
	httperrors "github.com/anuarkuanysh/dental_project/internal/adapters/driving/http/errors"
	jwtmiddleware "github.com/anuarkuanysh/dental_project/internal/adapters/driving/http/middleware"
	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
	contentuc "github.com/anuarkuanysh/dental_project/internal/usecase/content"
)

type Handler struct {
	ListPublishedUC *contentuc.ListPublished
	GetByIDUC       *contentuc.GetByID
	GetMediaUC      *contentuc.GetMedia
}

func (h *Handler) List(c *gin.Context) {
	userID, ok := jwtmiddleware.MustUserID(c)
	if !ok {
		httperrors.Write(c, httperrors.Unauthorized())
		return
	}

	items, err := h.ListPublishedUC.Execute(c.Request.Context(), userID)
	if err != nil {
		httperrors.Write(c, err)
		return
	}

	c.JSON(http.StatusOK, converters.ToContentListResponse(items))
}

func (h *Handler) Get(c *gin.Context) {
	userID, ok := jwtmiddleware.MustUserID(c)
	if !ok {
		httperrors.Write(c, httperrors.Unauthorized())
		return
	}

	contentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || contentID <= 0 {
		httperrors.Write(c, domainerrors.ErrContentNotFound)
		return
	}

	item, err := h.GetByIDUC.Execute(c.Request.Context(), userID, contentID)
	if err != nil {
		httperrors.Write(c, err)
		return
	}

	c.JSON(http.StatusOK, converters.ToContentItemResponse(item))
}

func (h *Handler) GetMedia(c *gin.Context) {
	userID, ok := jwtmiddleware.MustUserID(c)
	if !ok {
		httperrors.Write(c, httperrors.Unauthorized())
		return
	}

	mediaID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || mediaID <= 0 {
		httperrors.Write(c, domainerrors.ErrContentMediaNotFound)
		return
	}

	result, err := h.GetMediaUC.Execute(c.Request.Context(), userID, mediaID)
	if err != nil {
		httperrors.Write(c, err)
		return
	}

	c.Data(http.StatusOK, result.MIMEType, result.Data)
}
