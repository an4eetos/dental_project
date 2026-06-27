package admin

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/anuarkuanysh/dental_project/internal/adapters/driving/http/converters"
	httperrors "github.com/anuarkuanysh/dental_project/internal/adapters/driving/http/errors"
	jwtmiddleware "github.com/anuarkuanysh/dental_project/internal/adapters/driving/http/middleware"
	domainerrors "github.com/anuarkuanysh/dental_project/internal/domain/global/errors"
	"github.com/anuarkuanysh/dental_project/internal/domain/identity"
	adminuc "github.com/anuarkuanysh/dental_project/internal/usecase/admin"
)

type Handler struct {
	StatisticsUC *adminuc.Statistics
	ListUsersUC  *adminuc.ListUsers
	GetUserUC    *adminuc.GetUser
	UpdateUserUC *adminuc.UpdateUser
	SetBlockedUC *adminuc.SetBlocked
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

func (h *Handler) ListUsers(c *gin.Context) {
	userID, ok := jwtmiddleware.MustUserID(c)
	if !ok {
		httperrors.Write(c, httperrors.Unauthorized())
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	role := identity.Role(c.Query("role"))

	users, err := h.ListUsersUC.Execute(c.Request.Context(), adminuc.ListUsersInput{
		AdminUserID: userID,
		Role:        role,
		Search:      c.Query("search"),
		Limit:       limit,
		Offset:      offset,
	})
	if err != nil {
		httperrors.Write(c, err)
		return
	}

	c.JSON(http.StatusOK, converters.ToAdminUserListResponse(users))
}

func (h *Handler) GetUser(c *gin.Context) {
	userID, ok := jwtmiddleware.MustUserID(c)
	if !ok {
		httperrors.Write(c, httperrors.Unauthorized())
		return
	}

	targetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || targetID <= 0 {
		httperrors.Write(c, domainerrors.ErrUserNotFound)
		return
	}

	overview, err := h.GetUserUC.Execute(c.Request.Context(), userID, targetID)
	if err != nil {
		httperrors.Write(c, err)
		return
	}

	c.JSON(http.StatusOK, converters.ToAdminUserOverviewResponse(overview))
}

type updateUserRequest struct {
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	Username  *string `json:"username"`
	Role      *string `json:"role"`
}

func (h *Handler) UpdateUser(c *gin.Context) {
	userID, ok := jwtmiddleware.MustUserID(c)
	if !ok {
		httperrors.Write(c, httperrors.Unauthorized())
		return
	}

	targetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || targetID <= 0 {
		httperrors.Write(c, domainerrors.ErrUserNotFound)
		return
	}

	var req updateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperrors.Write(c, httperrors.Unauthorized())
		return
	}

	input := adminuc.UpdateUserInput{
		AdminUserID:  userID,
		TargetUserID: targetID,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Username:     req.Username,
	}
	if req.Role != nil {
		role := identity.Role(*req.Role)
		input.Role = &role
	}

	updated, err := h.UpdateUserUC.Execute(c.Request.Context(), input)
	if err != nil {
		httperrors.Write(c, err)
		return
	}

	c.JSON(http.StatusOK, converters.ToAdminUserResponse(updated))
}

type setBlockedRequest struct {
	Blocked bool `json:"blocked"`
}

func (h *Handler) SetBlocked(c *gin.Context) {
	userID, ok := jwtmiddleware.MustUserID(c)
	if !ok {
		httperrors.Write(c, httperrors.Unauthorized())
		return
	}

	targetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || targetID <= 0 {
		httperrors.Write(c, domainerrors.ErrUserNotFound)
		return
	}

	var req setBlockedRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperrors.Write(c, httperrors.Unauthorized())
		return
	}

	updated, err := h.SetBlockedUC.Execute(c.Request.Context(), adminuc.SetBlockedInput{
		AdminUserID:  userID,
		TargetUserID: targetID,
		Blocked:      req.Blocked,
	})
	if err != nil {
		httperrors.Write(c, err)
		return
	}

	c.JSON(http.StatusOK, converters.ToAdminUserResponse(updated))
}
