package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// HealthHandler handles health check requests.
type HealthHandler struct {
	logger *zerolog.Logger
}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler(logger *zerolog.Logger) *HealthHandler {
	return &HealthHandler{logger: logger}
}

// Check godoc
//
//	@Summary		Health check
//	@Description	Check if the service is running
//	@Tags			Health
//	@Produce		json
//	@Success		200	{object}	map[string]string
//	@Router			/health [get]
func (h *HealthHandler) Check(c *gin.Context) {
	h.logger.Info().Msg("[HealthHandler] Check request received")
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}
