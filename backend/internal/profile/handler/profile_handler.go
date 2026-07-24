package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Kooqoo22/JobJourney/backend/internal/middleware"
	"github.com/Kooqoo22/JobJourney/backend/internal/profile/dto"
	"github.com/Kooqoo22/JobJourney/backend/pkg/utils"
)

type ProfileHandler struct {
	usecase ProfileUsecaseIface
}

func New(uc ProfileUsecaseIface) *ProfileHandler {
	return &ProfileHandler{usecase: uc}
}

func (h *ProfileHandler) GetProfile(c *gin.Context) {
	userID := c.GetInt64(middleware.ContextUserID)
	resp, err := h.usecase.GetProfile(c.Request.Context(), userID)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, utils.NewSuccess("profile retrieved", resp))
}

func (h *ProfileHandler) ChangePassword(c *gin.Context) {
	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}
	userID := c.GetInt64(middleware.ContextUserID)
	if err := h.usecase.ChangePassword(c.Request.Context(), userID, req); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, utils.NewMessage("password changed successfully"))
}

func (h *ProfileHandler) UpdateProfile(c *gin.Context) {
	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}
	userID := c.GetInt64(middleware.ContextUserID)
	resp, err := h.usecase.UpdateProfile(c.Request.Context(), userID, req)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, utils.NewSuccess("profile updated", resp))
}
