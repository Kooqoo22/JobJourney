package middleware

import (
	"slices"

	"github.com/gin-gonic/gin"

	"github.com/Kooqoo22/JobJourney/backend/pkg/utils"
)

func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !slices.Contains(roles, c.GetString(ContextRole)) {
			c.Error(utils.ErrForbidden("insufficient permissions"))
			c.Abort()
			return
		}
		c.Next()
	}
}
