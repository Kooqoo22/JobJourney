package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	adminDto "github.com/Kooqoo22/JobJourney/backend/internal/admin/dto"
	"github.com/Kooqoo22/JobJourney/backend/internal/middleware"
	"github.com/Kooqoo22/JobJourney/backend/pkg/utils"
)

type AdminHandler struct {
	usecase AdminUsecaseIface
}

func New(uc AdminUsecaseIface) *AdminHandler {
	return &AdminHandler{usecase: uc}
}

func (h *AdminHandler) ListUsers(c *gin.Context) {
	var q adminDto.ListUsersQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.Error(err)
		return
	}
	tz := c.GetHeader("X-Timezone")
	if tz == "" {
		tz = "Asia/Jakarta"
	}
	users, meta, err := h.usecase.ListUsers(c.Request.Context(), q, tz)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, utils.NewList("users retrieved", users, meta))
}

func (h *AdminHandler) BanUser(c *gin.Context) {
	var param adminDto.UserIDParam
	if err := c.ShouldBindUri(&param); err != nil {
		c.Error(err)
		return
	}
	var req adminDto.BanUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}
	adminID := c.GetInt64(middleware.ContextUserID)
	if err := h.usecase.BanUser(c.Request.Context(), adminID, param.ID, req.Reason); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, utils.NewMessage("user banned"))
}

func (h *AdminHandler) UnbanUser(c *gin.Context) {
	var param adminDto.UserIDParam
	if err := c.ShouldBindUri(&param); err != nil {
		c.Error(err)
		return
	}
	adminID := c.GetInt64(middleware.ContextUserID)
	if err := h.usecase.UnbanUser(c.Request.Context(), adminID, param.ID); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, utils.NewMessage("user unbanned"))
}

func (h *AdminHandler) DeleteUser(c *gin.Context) {
	var param adminDto.UserIDParam
	if err := c.ShouldBindUri(&param); err != nil {
		c.Error(err)
		return
	}
	adminID := c.GetInt64(middleware.ContextUserID)
	if err := h.usecase.DeleteUser(c.Request.Context(), adminID, param.ID); err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, utils.NewMessage("user deleted"))
}
