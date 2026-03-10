package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"phone-call-receptionist/backend/internal/domain/port"
	"phone-call-receptionist/backend/pkg/dtos/responses"
)

// TwilioSignature validates the X-Twilio-Signature header on webhook requests.
func TwilioSignature(voiceCaller port.VoiceCaller, logger *zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		signature := c.GetHeader("X-Twilio-Signature")
		if signature == "" {
			logger.Warn().Msg("Missing Twilio signature")
			// In development, allow requests without signature
			c.Next()
			return
		}

		if err := c.Request.ParseForm(); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest,
				responses.ErrorResponse{Error: "failed to parse form"})
			return
		}

		params := make(map[string]string)
		for key, values := range c.Request.PostForm {
			if len(values) > 0 {
				params[key] = values[0]
			}
		}

		url := "https://" + c.Request.Host + c.Request.URL.String()

		if !voiceCaller.ValidateSignature(url, params, signature) {
			logger.Warn().Str("url", url).Msg("Invalid Twilio signature")
			c.AbortWithStatusJSON(http.StatusForbidden,
				responses.ErrorResponse{Error: "invalid signature"})
			return
		}

		c.Next()
	}
}
