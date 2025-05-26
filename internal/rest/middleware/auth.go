package middleware

import (
	"net/http"
	"strings"

	"slices"

	"github.com/gin-gonic/gin"
)

type AuthValidator interface {
	Validate(token string) (userID int64, role string, err error)
}

func (m *Middleware) AuthWithRoles(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			c.Abort()
			return
		}

		userID, role, err := m.validator.Validate(tokenString)
		if err != nil {
			m.logger.Errorw("token parse error", "err", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		if !slices.Contains(allowedRoles, role) {
			m.logger.Warnw("forbidden access", "userID", userID, "role", role)
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			c.Abort()
			return
		}

		c.Set("userID", userID)
		c.Set("role", role)
		c.Next()
	}
}
