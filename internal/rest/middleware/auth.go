package middleware

import (
	"errors"
	"net/http"
	"strings"

	pkgErrors "github.com/Cora23tt/order_service/pkg/errors"

	"github.com/gin-gonic/gin"
)

type AuthValidator interface {
	Validate(username, password string) (bool, error)
}

func (m *Middleware) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")

		token := strings.TrimPrefix(tokenString, "Bearer ")

		UserID, err := m.authService.GetUserID(c, token)
		if err != nil {
			statusCode := http.StatusUnauthorized
			errorMessage := "unauthorized"

			if errors.Is(err, pkgErrors.ErrNotFound) {
				m.logger.Errorw("user not found", "error", err)
			} else {
				m.logger.Errorw("failed to get user id", "error", err)
			}

			c.JSON(statusCode, gin.H{"error": errorMessage})
			c.Abort()
			return
		}

		c.Set("userID", UserID)
		c.Next()
	}

}
