package appointment

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	httperrors "github.com/anuarkuanysh/dental_project/internal/adapters/driving/http/errors"
	jwtmiddleware "github.com/anuarkuanysh/dental_project/internal/adapters/driving/http/middleware"
	"github.com/anuarkuanysh/dental_project/internal/adapters/driving/http/converters"
	appointmentuc "github.com/anuarkuanysh/dental_project/internal/usecase/appointment"
)

type Handler struct {
	CreateUC        *appointmentuc.Create
	ListMineUC      *appointmentuc.ListMine
	ListForDoctorUC *appointmentuc.ListForDoctor
	OfferUC         *appointmentuc.Offer
}

type createAppointmentRequest struct {
	PreferredDate string `json:"preferred_date" binding:"required"`
	PreferredTime string `json:"preferred_time" binding:"required"`
}

type offerAppointmentRequest struct {
	PreferredDate string `json:"preferred_date" binding:"required"`
	PreferredTime string `json:"preferred_time" binding:"required"`
	ZoomLink      string `json:"zoom_link" binding:"required"`
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

func (h *Handler) Offer(c *gin.Context) {
	userID, ok := jwtmiddleware.MustUserID(c)
	if !ok {
		httperrors.Write(c, httperrors.Unauthorized())
		return
	}

	appointmentID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || appointmentID <= 0 {
		c.JSON(http.StatusBadRequest, httperrors.APIError{
			Code:    "validation_error",
			Message: "invalid appointment id",
		})
		return
	}

	var req offerAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, httperrors.APIError{
			Code:    "validation_error",
			Message: "preferred_date, preferred_time and zoom_link are required",
		})
		return
	}

	item, err := h.OfferUC.Execute(c.Request.Context(), appointmentuc.OfferInput{
		DoctorUserID:  userID,
		AppointmentID: appointmentID,
		PreferredDate: req.PreferredDate,
		PreferredTime: req.PreferredTime,
		ZoomLink:      req.ZoomLink,
	})
	if err != nil {
		httperrors.Write(c, err)
		return
	}

	c.JSON(http.StatusOK, converters.ToDoctorAppointmentResponse(item))
}
