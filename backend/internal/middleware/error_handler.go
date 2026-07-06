package middleware

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Kooqoo22/JobJourney/backend/pkg/utils"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err
		status, body := utils.MapError(err)

		if status == http.StatusInternalServerError {
			slog.Error("request failed",
				"error", err.Error(),
				"path", c.Request.URL.Path,
				"request_id", c.GetString(ContextRequestID),
			)
		}

		c.JSON(status, body)
	}
}
