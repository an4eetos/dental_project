package appointment

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/anuarkuanysh/dental_project/internal/adapters/driving/http/converters"
	httperrors "github.com/anuarkuanysh/dental_project/internal/adapters/driving/http/errors"
	jwtmiddleware "github.com/anuarkuanysh/dental_project/internal/adapters/driving/http/middleware"
	appointmentuc "github.com/anuarkuanysh/dental_project/internal/usecase/appointment"
)

type Handler struct {
	CreateUC        *appointmentuc.Create
	ListMineUC      *appointmentuc.ListMine
	ListForDoctorUC *appointmentuc.ListForDoctor
	RespondUC       *appointmentuc.Respond
	SetZoomLinkUC   *appointmentuc.SetZoomLink
	SuggestSlotsUC  *appointmentuc.SuggestSlots
}

type createAppointmentRequest struct {
	PreferredDate string `json:"preferred_date" binding:"required"`
	PreferredTime string `json:"preferred_time" binding:"required"`
}

type respondAppointmentRequest struct {
	Decision      string `json:"decision" binding:"required"`
	PreferredDate string `json:"preferred_date"`
	PreferredTime string `json:"preferred_time"`
	ZoomLink      string `json:"zoom_link"`
	DoctorNotes   string `json:"doctor_notes"`
}

type setZoomLinkRequest struct {
	ZoomLink string `json:"zoom_link" binding:"required"`
}

func (h *Handler) Create(c *gin.Context) {
	userID, ok := jwtmiddleware.MustUserID(c)
	if !ok {
		httperrors.Write(c, httperrors.Unauthorized())
		return
	}

	var req createAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, httperrors.APIError{
			Code:    "validation_error",
			Message: "preferred_date and preferred_time are required",
		})
		return
	}

	appt, err := h.CreateUC.Execute(c.Request.Context(), appointmentuc.CreateInput{
		UserID:        userID,
		PreferredDate: req.PreferredDate,
		PreferredTime: req.PreferredTime,
	})
	if err != nil {
		httperrors.Write(c, err)
		return
	}

	c.JSON(http.StatusCreated, converters.ToAppointmentResponse(appt))
}

func (h *Handler) ListMine(c *gin.Context) {
	userID, ok := jwtmiddleware.MustUserID(c)
	if !ok {
		httperrors.Write(c, httperrors.Unauthorized())
		return
	}

	items, err := h.ListMineUC.Execute(c.Request.Context(), userID)
	if err != nil {
		httperrors.Write(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"appointments": converters.ToAppointmentListResponse(items)})
}

func (h *Handler) ListForDoctor(c *gin.Context) {
	userID, ok := jwtmiddleware.MustUserID(c)
	if !ok {
		httperrors.Write(c, httperrors.Unauthorized())
		return
	}

	items, err := h.ListForDoctorUC.Execute(c.Request.Context(), userID)
	if err != nil {
		httperrors.Write(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"appointments": converters.ToDoctorAppointmentListResponse(items),
	})
}

func (h *Handler) Respond(c *gin.Context) {
	userID, ok := jwtmiddleware.MustUserID(c)
	if !ok {
		httperrors.Write(c, httperrors.Unauthorized())
		return
	}

	appointmentID, err := parseAppointmentID(c)
	if err != nil {
		return
	}

	var req respondAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, httperrors.APIError{
			Code:    "validation_error",
			Message: "decision is required",
		})
		return
	}

	item, err := h.RespondUC.Execute(c.Request.Context(), appointmentuc.RespondInput{
		DoctorUserID:  userID,
		AppointmentID: appointmentID,
		Decision:      req.Decision,
		PreferredDate: req.PreferredDate,
		PreferredTime: req.PreferredTime,
		ZoomLink:      req.ZoomLink,
		DoctorNotes:   req.DoctorNotes,
	})
	if err != nil {
		httperrors.Write(c, err)
		return
	}

	c.JSON(http.StatusOK, converters.ToDoctorAppointmentResponse(item))
}

func (h *Handler) SetZoomLink(c *gin.Context) {
	userID, ok := jwtmiddleware.MustUserID(c)
	if !ok {
		httperrors.Write(c, httperrors.Unauthorized())
		return
	}

	appointmentID, err := parseAppointmentID(c)
	if err != nil {
		return
	}

	var req setZoomLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, httperrors.APIError{
			Code:    "validation_error",
			Message: "zoom_link is required",
		})
		return
	}

	item, err := h.SetZoomLinkUC.Execute(c.Request.Context(), appointmentuc.SetZoomLinkInput{
		DoctorUserID:  userID,
		AppointmentID: appointmentID,
		ZoomLink:      req.ZoomLink,
	})
	if err != nil {
		httperrors.Write(c, err)
		return
	}

	c.JSON(http.StatusOK, converters.ToDoctorAppointmentResponse(item))
}

func (h *Handler) SuggestSlots(c *gin.Context) {
	userID, ok := jwtmiddleware.MustUserID(c)
	if !ok {
		httperrors.Write(c, httperrors.Unauthorized())
		return
	}

	appointmentID, err := parseAppointmentID(c)
	if err != nil {
		return
	}

	result, err := h.SuggestSlotsUC.Execute(c.Request.Context(), userID, appointmentID)
	if err != nil {
		httperrors.Write(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"suggested_text": result.SuggestedText})
}

func parseAppointmentID(c *gin.Context) (int64, error) {
	appointmentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || appointmentID <= 0 {
		c.JSON(http.StatusBadRequest, httperrors.APIError{
			Code:    "validation_error",
			Message: "invalid appointment id",
		})
		return 0, err
	}
	return appointmentID, nil
}
