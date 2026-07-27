package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Kooqoo22/JobJourney/backend/internal/middleware"
	"github.com/Kooqoo22/JobJourney/backend/pkg/utils"
)

type StatsHandler struct {
	usecase StatsUsecaseIface
}

func New(uc StatsUsecaseIface) *StatsHandler {
	return &StatsHandler{usecase: uc}
}

func (h *StatsHandler) GetAnalytics(c *gin.Context) {
	userID := c.GetInt64(middleware.ContextUserID)
	period := c.Query("period")
	resp, err := h.usecase.GetAnalytics(c.Request.Context(), userID, period)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, utils.NewSuccess("statistics retrieved", resp))
}

func (h *StatsHandler) GetSummary(c *gin.Context) {
	userID := c.GetInt64(middleware.ContextUserID)
	tz := c.GetHeader("X-Timezone")
	if tz == "" {
		tz = "Asia/Jakarta"
	}
	resp, err := h.usecase.GetSummary(c.Request.Context(), userID, tz)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, utils.NewSuccess("summary retrieved", resp))
}
