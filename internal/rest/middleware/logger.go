package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
)

func (m *Middleware) ZapLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		m.logger.Infof("%s %s %d  %s", c.Request.Method, c.Request.URL.Path, c.Writer.Status(), duration.Truncate(time.Millisecond).String())
	}
}
