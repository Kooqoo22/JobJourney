package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/Kooqoo22/JobJourney/backend/pkg/token"
	"github.com/Kooqoo22/JobJourney/backend/pkg/utils"
)

func Auth(tm *token.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := stripBearer(c.GetHeader("Authorization"))
		if raw == "" {
			c.Error(utils.ErrUnauthorized("missing authorization token"))
			c.Abort()
			return
		}

		claims, err := tm.Parse(raw)
		if err != nil {
			c.Error(utils.ErrUnauthorized("invalid or expired token"))
			c.Abort()
			return
		}

		c.Set(ContextUserID, claims.UserID)
		c.Set(ContextRole, claims.Role)
		c.Next()
	}
}

func stripBearer(header string) string {
	const prefix = "Bearer "
	if len(header) > len(prefix) && strings.EqualFold(header[:len(prefix)], prefix) {
		return strings.TrimSpace(header[len(prefix):])
	}
	return ""
}
