package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"phone-call-receptionist/backend/pkg/dtos/responses"
)

// Recovery recovers from panics and returns a 500 error response.
func Recovery(logger *zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error().
					Interface("panic", r).
					Str("stack", string(debug.Stack())).
					Msg("Panic recovered")

				c.AbortWithStatusJSON(http.StatusInternalServerError,
					responses.ErrorResponse{Error: "internal server error"})
			}
		}()
		c.Next()
	}
}
