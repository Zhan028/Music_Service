package middleware

import (
	"time"

	"github.com/Zhan028/Music_Service/api_gateway/internal/logger"
	"github.com/gin-gonic/gin"
)

func GinLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		start := time.Now()

		c.Next()

		duration := time.Since(start)
		statusCode := c.Writer.Status()
		method := c.Request.Method
		path := c.Request.URL.Path
		ip := c.ClientIP()

		logger.InfoLogger.Printf("%s %s | %d | %s | %v", method, path, statusCode, ip, duration)
	}
}
