package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Kooqoo22/JobJourney/backend/internal/application/dto"
	"github.com/Kooqoo22/JobJourney/backend/internal/middleware"
	"github.com/Kooqoo22/JobJourney/backend/pkg/utils"
)

type ApplicationHandler struct {
	usecase ApplicationUsecaseIface
}

func New(uc ApplicationUsecaseIface) *ApplicationHandler {
	return &ApplicationHandler{usecase: uc}
}

func (h *ApplicationHandler) ListApplications(c *gin.Context) {
	var q dto.ListApplicationsQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.Error(err)
		return
	}
	userID := c.GetInt64(middleware.ContextUserID)
	tz := c.GetHeader("X-Timezone")
	if tz == "" {
		tz = "Asia/Jakarta"
	}
	apps, meta, err := h.usecase.ListApplications(c.Request.Context(), userID, tz, q)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, utils.NewList("applications retrieved", apps, meta))
}

func (h *ApplicationHandler) CreateApplication(c *gin.Context) {
	var req dto.CreateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}
	userID := c.GetInt64(middleware.ContextUserID)
	tz := c.GetHeader("X-Timezone")
	if tz == "" {
		tz = "Asia/Jakarta"
	}
	resp, err := h.usecase.CreateApplication(c.Request.Context(), userID, tz, req)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, utils.NewSuccess("application created", resp))
}
