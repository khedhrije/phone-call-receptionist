package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthHandler handles health check requests.
type HealthHandler struct{}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
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
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}
