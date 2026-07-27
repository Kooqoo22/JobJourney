package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	adminDto "github.com/Kooqoo22/JobJourney/backend/internal/admin/dto"
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
