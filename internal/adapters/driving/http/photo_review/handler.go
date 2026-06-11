package photoreview

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/anuarkuanysh/dental_project/internal/adapters/driving/http/converters"
	httperrors "github.com/anuarkuanysh/dental_project/internal/adapters/driving/http/errors"
	jwtmiddleware "github.com/anuarkuanysh/dental_project/internal/adapters/driving/http/middleware"
	photoreviewuc "github.com/anuarkuanysh/dental_project/internal/usecase/photo_review"
)

type Handler struct {
	ListPendingUC   *photoreviewuc.ListPending
	ListAnsweredUC  *photoreviewuc.ListAnswered
	GetUC           *photoreviewuc.Get
	GetImageUC      *photoreviewuc.GetImage
	GenerateDraftUC *photoreviewuc.GenerateDraft
	RespondUC       *photoreviewuc.Respond
}

type respondRequest struct {
	Response string `json:"response" binding:"required"`
}

func (h *Handler) ListPending(c *gin.Context) {
	userID, ok := jwtmiddleware.MustUserID(c)
	if !ok {
		httperrors.Write(c, httperrors.Unauthorized())
		return
	}

	items, err := h.ListPendingUC.Execute(c.Request.Context(), userID)
	if err != nil {
		httperrors.Write(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"submissions": converters.ToSubmissionListResponse(items)})
}

func (h *Handler) ListAnswered(c *gin.Context) {
	userID, ok := jwtmiddleware.MustUserID(c)
	if !ok {
		httperrors.Write(c, httperrors.Unauthorized())
		return
	}

	items, err := h.ListAnsweredUC.Execute(c.Request.Context(), userID)
	if err != nil {
		httperrors.Write(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"submissions": converters.ToSubmissionListResponse(items)})
}

func (h *Handler) Get(c *gin.Context) {
	userID, ok := jwtmiddleware.MustUserID(c)
	if !ok {
		httperrors.Write(c, httperrors.Unauthorized())
		return
	}

	submissionID, ok := parseSubmissionID(c)
	if !ok {
		return
	}

	item, err := h.GetUC.Execute(c.Request.Context(), userID, submissionID)
	if err != nil {
		httperrors.Write(c, err)
		return
	}

	c.JSON(http.StatusOK, converters.ToSubmissionResponse(item))
}

func (h *Handler) GetImage(c *gin.Context) {
	userID, ok := jwtmiddleware.MustUserID(c)
	if !ok {
		httperrors.Write(c, httperrors.Unauthorized())
		return
	}

	submissionID, ok := parseSubmissionID(c)
	if !ok {
		return
	}

	result, err := h.GetImageUC.Execute(c.Request.Context(), userID, submissionID)
	if err != nil {
		httperrors.Write(c, err)
		return
	}

	c.Data(http.StatusOK, result.MIMEType, result.Data)
}

func (h *Handler) GenerateDraft(c *gin.Context) {
	userID, ok := jwtmiddleware.MustUserID(c)
	if !ok {
		httperrors.Write(c, httperrors.Unauthorized())
		return
	}

	submissionID, ok := parseSubmissionID(c)
	if !ok {
		return
	}

	result, err := h.GenerateDraftUC.Execute(c.Request.Context(), userID, submissionID)
	if err != nil {
		httperrors.Write(c, err)
		return
	}

	c.JSON(http.StatusOK, converters.ToGenerateDraftResponse(result.Submission, result.DraftText))
}

func (h *Handler) Respond(c *gin.Context) {
	userID, ok := jwtmiddleware.MustUserID(c)
	if !ok {
		httperrors.Write(c, httperrors.Unauthorized())
		return
	}

	submissionID, ok := parseSubmissionID(c)
	if !ok {
		return
	}

	var req respondRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, httperrors.APIError{
			Code:    "validation_error",
			Message: "response is required",
		})
		return
	}

	item, err := h.RespondUC.Execute(c.Request.Context(), photoreviewuc.RespondInput{
		DoctorUserID: userID,
		SubmissionID: submissionID,
		Response:     req.Response,
	})
	if err != nil {
		httperrors.Write(c, err)
		return
	}

	c.JSON(http.StatusOK, converters.ToSubmissionResponse(item))
}

func parseSubmissionID(c *gin.Context) (int64, bool) {
	raw := c.Param("id")
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, httperrors.APIError{
			Code:    "validation_error",
			Message: "invalid submission id",
		})
		return 0, false
	}
	return id, true
}
