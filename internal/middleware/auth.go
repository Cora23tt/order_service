package middleware

import (
	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct{}

func NewAuthMiddleware() *AuthMiddleware {
	return &AuthMiddleware{}
}

func (m *AuthMiddleware) Authorize(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Проверка JWT и роли
		c.Next()
	}
}
