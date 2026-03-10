package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"phone-call-receptionist/backend/internal/infrastructure/ws"
)

// WSHandler handles WebSocket connection upgrades.
type WSHandler struct {
	hub    *ws.Hub
	logger *zerolog.Logger
}

// NewWSHandler creates a new WSHandler with the given dependencies.
func NewWSHandler(hub *ws.Hub, logger *zerolog.Logger) *WSHandler {
	return &WSHandler{hub: hub, logger: logger}
}

// HandleWebSocket godoc
//
//	@Summary		WebSocket connection
//	@Description	Upgrade to WebSocket for real-time events
//	@Tags			WebSocket
//	@Success		101
//	@Router			/ws [get]
func (h *WSHandler) HandleWebSocket(c *gin.Context) {
	h.hub.HandleWebSocket(c.Writer, c.Request)
}
