package admin

import (
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/anuarkuanysh/dental_project/internal/adapters/driving/http/converters"
	httperrors "github.com/anuarkuanysh/dental_project/internal/adapters/driving/http/errors"
	jwtmiddleware "github.com/anuarkuanysh/dental_project/internal/adapters/driving/http/middleware"
	contentdomain "github.com/anuarkuanysh/dental_project/internal/domain/content"
	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
	admincontentuc "github.com/anuarkuanysh/dental_project/internal/usecase/admin/content"
)

type ContentHandler struct {
	ListUC        *admincontentuc.List
	GetUC         *admincontentuc.Get
	CreateUC      *admincontentuc.Create
	UpdateUC      *admincontentuc.Update
	DeleteUC      *admincontentuc.Delete
	ReorderUC     *admincontentuc.Reorder
	UploadMediaUC *admincontentuc.UploadMedia
	MaxMediaBytes int
}

type contentBlockRequest struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

type saveContentRequest struct {
	Title       string                `json:"title"`
	Description string                `json:"description"`
	Access      string                `json:"access"`
	Published   bool                  `json:"published"`
	Blocks      []contentBlockRequest `json:"blocks"`
}

type reorderContentRequest struct {
	IDs []int64 `json:"ids"`
}

func (h *ContentHandler) List(c *gin.Context) {
	userID, ok := jwtmiddleware.MustUserID(c)
	if !ok {
		httperrors.Write(c, httperrors.Unauthorized())
		return
	}

	items, err := h.ListUC.Execute(c.Request.Context(), userID)
	if err != nil {
		httperrors.Write(c, err)
		return
	}

	c.JSON(http.StatusOK, converters.ToAdminContentListResponse(items))
}

func (h *ContentHandler) Get(c *gin.Context) {
	userID, ok := jwtmiddleware.MustUserID(c)
	if !ok {
		httperrors.Write(c, httperrors.Unauthorized())
		return
	}

	contentID, err := parseContentID(c)
	if err != nil {
		httperrors.Write(c, err)
		return
	}

	item, err := h.GetUC.Execute(c.Request.Context(), userID, contentID)
	if err != nil {
		httperrors.Write(c, err)
		return
	}

	c.JSON(http.StatusOK, converters.ToAdminContentItemResponse(item))
}

func (h *ContentHandler) Create(c *gin.Context) {
	userID, ok := jwtmiddleware.MustUserID(c)
	if !ok {
		httperrors.Write(c, httperrors.Unauthorized())
		return
	}

	var req saveContentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperrors.Write(c, domainerrors.ErrInvalidContentBlocks)
		return
	}

	blocks, err := parseBlocks(req.Blocks)
	if err != nil {
		httperrors.Write(c, err)
		return
	}

	item, err := h.CreateUC.Execute(c.Request.Context(), admincontentuc.CreateInput{
		AdminUserID: userID,
		Title:       req.Title,
		Description: req.Description,
		Access:      req.Access,
		Published:   req.Published,
		Blocks:      blocks,
	})
	if err != nil {
		httperrors.Write(c, err)
		return
	}

	c.JSON(http.StatusCreated, converters.ToAdminContentItemResponse(item))
}

func (h *ContentHandler) Update(c *gin.Context) {
	userID, ok := jwtmiddleware.MustUserID(c)
	if !ok {
		httperrors.Write(c, httperrors.Unauthorized())
		return
	}

	contentID, err := parseContentID(c)
	if err != nil {
		httperrors.Write(c, err)
		return
	}

	var req saveContentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperrors.Write(c, domainerrors.ErrInvalidContentBlocks)
		return
	}

	blocks, err := parseBlocks(req.Blocks)
	if err != nil {
		httperrors.Write(c, err)
		return
	}

	item, err := h.UpdateUC.Execute(c.Request.Context(), admincontentuc.UpdateInput{
		AdminUserID: userID,
		ContentID:   contentID,
		Title:       req.Title,
		Description: req.Description,
		Access:      req.Access,
		Published:   req.Published,
		Blocks:      blocks,
	})
	if err != nil {
		httperrors.Write(c, err)
		return
	}

	c.JSON(http.StatusOK, converters.ToAdminContentItemResponse(item))
}

func (h *ContentHandler) Delete(c *gin.Context) {
	userID, ok := jwtmiddleware.MustUserID(c)
	if !ok {
		httperrors.Write(c, httperrors.Unauthorized())
		return
	}

	contentID, err := parseContentID(c)
	if err != nil {
		httperrors.Write(c, err)
		return
	}

	if err := h.DeleteUC.Execute(c.Request.Context(), userID, contentID); err != nil {
		httperrors.Write(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *ContentHandler) Reorder(c *gin.Context) {
	userID, ok := jwtmiddleware.MustUserID(c)
	if !ok {
		httperrors.Write(c, httperrors.Unauthorized())
		return
	}

	var req reorderContentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperrors.Write(c, domainerrors.ErrInvalidContentBlocks)
		return
	}

	if err := h.ReorderUC.Execute(c.Request.Context(), admincontentuc.ReorderInput{
		AdminUserID: userID,
		IDs:         req.IDs,
	}); err != nil {
		httperrors.Write(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *ContentHandler) UploadMedia(c *gin.Context) {
	userID, ok := jwtmiddleware.MustUserID(c)
	if !ok {
		httperrors.Write(c, httperrors.Unauthorized())
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		httperrors.Write(c, domainerrors.ErrInvalidContentMedia)
		return
	}

	if h.MaxMediaBytes > 0 && file.Size > int64(h.MaxMediaBytes) {
		httperrors.Write(c, domainerrors.ErrContentMediaTooLarge)
		return
	}

	f, err := file.Open()
	if err != nil {
		httperrors.Write(c, domainerrors.ErrInvalidContentMedia)
		return
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		httperrors.Write(c, domainerrors.ErrInvalidContentMedia)
		return
	}
	if h.MaxMediaBytes > 0 && len(data) > h.MaxMediaBytes {
		httperrors.Write(c, domainerrors.ErrContentMediaTooLarge)
		return
	}

	var contentItemID *int64
	if raw := c.PostForm("content_item_id"); raw != "" {
		id, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || id <= 0 {
			httperrors.Write(c, domainerrors.ErrContentNotFound)
			return
		}
		contentItemID = &id
	}

	mimeType := file.Header.Get("Content-Type")
	result, err := h.UploadMediaUC.Execute(c.Request.Context(), admincontentuc.UploadMediaInput{
		AdminUserID:   userID,
		ContentItemID: contentItemID,
		MIMEType:      mimeType,
		Data:          data,
		MaxBytes:      h.MaxMediaBytes,
	})
	if err != nil {
		httperrors.Write(c, err)
		return
	}

	c.JSON(http.StatusCreated, converters.UploadMediaResponse{MediaID: result.MediaID})
}

func parseContentID(c *gin.Context) (int64, error) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, domainerrors.ErrContentNotFound
	}
	return id, nil
}

func parseBlocks(raw []contentBlockRequest) ([]contentdomain.Block, error) {
	if len(raw) == 0 {
		return nil, domainerrors.ErrInvalidContentBlocks
	}
	blocks := make([]contentdomain.Block, 0, len(raw))
	for _, item := range raw {
		data, err := marshalBlockData(item.Data)
		if err != nil {
			return nil, domainerrors.ErrInvalidContentBlocks
		}
		blocks = append(blocks, contentdomain.Block{
			Type: contentdomain.BlockType(item.Type),
			Data: data,
		})
	}
	return blocks, nil
}
