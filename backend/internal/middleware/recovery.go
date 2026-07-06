package middleware

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Kooqoo22/JobJourney/backend/pkg/utils"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				slog.Error("panic recovered",
					"error", recovered,
					"path", c.Request.URL.Path,
					"request_id", c.GetString(ContextRequestID),
				)
				c.AbortWithStatusJSON(http.StatusInternalServerError, utils.NewMessage("internal server error"))
			}
		}()
		c.Next()
	}
}
