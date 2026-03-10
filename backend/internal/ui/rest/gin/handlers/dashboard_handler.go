package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"phone-call-receptionist/backend/internal/domain/api"
	"phone-call-receptionist/backend/pkg/dtos/responses"
)

// DashboardHandler handles dashboard analytics HTTP requests.
type DashboardHandler struct {
	dashboardApi *api.DashboardApi
	logger       *zerolog.Logger
}

// NewDashboardHandler creates a new DashboardHandler with the given dependencies.
func NewDashboardHandler(dashboardApi *api.DashboardApi, logger *zerolog.Logger) *DashboardHandler {
	return &DashboardHandler{dashboardApi: dashboardApi, logger: logger}
}

// Stats godoc
//
//	@Summary		Get dashboard statistics
//	@Description	Returns overview statistics
//	@Tags			Dashboard
//	@Produce		json
//	@Success		200	{object}	responses.DashboardStatsResponse
//	@Security		BearerAuth
//	@Router			/dashboard/stats [get]
func (h *DashboardHandler) Stats(c *gin.Context) {
	stats, err := h.dashboardApi.Stats(c.Request.Context())
	if err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, stats)
}

// Costs godoc
//
//	@Summary		Get cost analytics
//	@Description	Returns cost breakdown for a time range
//	@Tags			Dashboard
//	@Produce		json
//	@Param			from	query	string	true	"From date (RFC3339)"
//	@Param			to		query	string	true	"To date (RFC3339)"
//	@Success		200	{object}	responses.CostAnalyticsResponse
//	@Security		BearerAuth
//	@Router			/dashboard/costs [get]
func (h *DashboardHandler) Costs(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")
	if from == "" || to == "" {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse{Error: "from and to are required"})
		return
	}

	analytics, err := h.dashboardApi.CostAnalytics(c.Request.Context(), from, to)
	if err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, analytics)
}

// Volume godoc
//
//	@Summary		Get call volume
//	@Description	Returns call volume for a time range
//	@Tags			Dashboard
//	@Produce		json
//	@Param			from	query	string	true	"From date (RFC3339)"
//	@Param			to		query	string	true	"To date (RFC3339)"
//	@Success		200	{object}	responses.CallVolumeResponse
//	@Security		BearerAuth
//	@Router			/dashboard/volume [get]
func (h *DashboardHandler) Volume(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")
	if from == "" || to == "" {
		c.JSON(http.StatusBadRequest, responses.ErrorResponse{Error: "from and to are required"})
		return
	}

	volume, err := h.dashboardApi.CallVolume(c.Request.Context(), from, to)
	if err != nil {
		HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, volume)
}
