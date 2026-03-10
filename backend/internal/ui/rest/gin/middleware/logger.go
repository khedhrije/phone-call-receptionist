package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"phone-call-receptionist/backend/pkg/helpers"
)

// Logger logs HTTP request details using zerolog.
func Logger(logger *zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		requestID, _ := helpers.ExtractRequestID(c.Request.Context())

		event := logger.Info()
		if status >= 500 {
			event = logger.Error()
		} else if status >= 400 {
			event = logger.Warn()
		}

		event.
			Str("requestId", requestID).
			Str("method", c.Request.Method).
			Str("path", path).
			Int("status", status).
			Dur("latency", latency).
			Str("clientIp", c.ClientIP()).
			Msg("HTTP request")
	}
}
