package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"phone-call-receptionist/backend/internal/domain/model"
	"phone-call-receptionist/backend/internal/domain/port"
	"phone-call-receptionist/backend/pkg/dtos/requests"
	"phone-call-receptionist/backend/pkg/dtos/responses"
)

// SettingsHandler handles system settings HTTP requests.
type SettingsHandler struct {
	settingsPort port.SystemSettings
	logger       *zerolog.Logger
}

// NewSettingsHandler creates a new SettingsHandler with the given dependencies.
func NewSettingsHandler(settingsPort port.SystemSettings, logger *zerolog.Logger) *SettingsHandler {
	return &SettingsHandler{settingsPort: settingsPort, logger: logger}
}

// Find godoc
//
//	@Summary		Get system settings
//	@Description	Returns the current system settings
//	@Tags			Settings
//	@Produce		json
//	@Success		200	{object}	responses.SystemSettingsResponse
//	@Security		BearerAuth
//	@Router			/settings [get]
func (h *SettingsHandler) Find(c *gin.Context) {
	settings, err := h.settingsPort.Find(c.Request.Context())
	if err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, responses.SystemSettingsResponse{
		DefaultLLMProvider:  settings.DefaultLLMProvider,
		DefaultVoiceID:      settings.DefaultVoiceID,
		TopK:                settings.TopK,
		MaxCallDurationSecs: settings.MaxCallDurationSecs,
		UpdatedAt:           settings.UpdatedAt,
	})
}

// Update godoc
//
//	@Summary		Update system settings
//	@Description	Update the system settings
//	@Tags			Settings
//	@Accept			json
//	@Produce		json
//	@Param			request	body	requests.UpdateSettingsRequest	true	"Settings data"
//	@Success		200	{object}	responses.SystemSettingsResponse
//	@Security		BearerAuth
//	@Router			/settings [put]
func (h *SettingsHandler) Update(c *gin.Context) {
	var req requests.UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse{Error: err.Error()})
		return
	}

	current, err := h.settingsPort.Find(c.Request.Context())
	if err != nil {
		HandleError(c, err)
		return
	}

	if req.DefaultLLMProvider != "" {
		current.DefaultLLMProvider = req.DefaultLLMProvider
	}
	if req.DefaultVoiceID != "" {
		current.DefaultVoiceID = req.DefaultVoiceID
	}
	if req.TopK > 0 {
		current.TopK = req.TopK
	}
	if req.MaxCallDurationSecs > 0 {
		current.MaxCallDurationSecs = req.MaxCallDurationSecs
	}
	current.UpdatedAt = time.Now().Format(time.RFC3339)

	if err := h.settingsPort.Update(c.Request.Context(), model.SystemSettings{
		DefaultLLMProvider:  current.DefaultLLMProvider,
		DefaultVoiceID:      current.DefaultVoiceID,
		TopK:                current.TopK,
		MaxCallDurationSecs: current.MaxCallDurationSecs,
		UpdatedAt:           current.UpdatedAt,
	}); err != nil {
		HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, responses.SystemSettingsResponse{
		DefaultLLMProvider:  current.DefaultLLMProvider,
		DefaultVoiceID:      current.DefaultVoiceID,
		TopK:                current.TopK,
		MaxCallDurationSecs: current.MaxCallDurationSecs,
		UpdatedAt:           current.UpdatedAt,
	})
}
