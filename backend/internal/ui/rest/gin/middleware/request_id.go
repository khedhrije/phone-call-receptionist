// Package middleware provides HTTP middleware for the Gin router.
package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"phone-call-receptionist/backend/pkg/helpers"
)

// RequestID generates a unique request ID for each request and sets it in context and response header.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := uuid.New().String()
		ctx := helpers.InjectRequestID(c.Request.Context(), requestID)
		c.Request = c.Request.WithContext(ctx)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}
