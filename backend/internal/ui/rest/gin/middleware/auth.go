package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"phone-call-receptionist/backend/pkg/dtos/responses"
	"phone-call-receptionist/backend/pkg/helpers"
)

// Auth validates JWT tokens from the Authorization header and injects user context.
func Auth(jwtSecret string, logger *zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, responses.ErrorResponse{Error: "authorization header required"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, responses.ErrorResponse{Error: "invalid authorization format"})
			return
		}

		claims, err := helpers.ValidateToken(parts[1], jwtSecret)
		if err != nil {
			logger.Debug().Err(err).Msg("Token validation failed")
			c.AbortWithStatusJSON(http.StatusUnauthorized, responses.ErrorResponse{Error: "invalid or expired token"})
			return
		}

		ctx := helpers.InjectUser(c.Request.Context(), claims.UserID, claims.Email, claims.Role)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
