package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"phone-call-receptionist/backend/internal/domain/api"
	"phone-call-receptionist/backend/internal/domain/model"
	"phone-call-receptionist/backend/internal/domain/port"
	"phone-call-receptionist/backend/pkg/dtos/requests"
	"phone-call-receptionist/backend/pkg/dtos/responses"
)

// LeadHandler handles lead-related HTTP requests.
type LeadHandler struct {
	leadApi *api.LeadApi
	logger  *zerolog.Logger
}

// NewLeadHandler creates a new LeadHandler with the given dependencies.
func NewLeadHandler(leadApi *api.LeadApi, logger *zerolog.Logger) *LeadHandler {
	return &LeadHandler{leadApi: leadApi, logger: logger}
}

// List godoc
//
//	@Summary		List leads
//	@Description	Returns paginated list of leads
//	@Tags			Leads
//	@Produce		json
//	@Param			status		query	string	false	"Filter by status"
//	@Param			page		query	int		false	"Page number"
//	@Param			pageSize	query	int		false	"Items per page"
//	@Success		200	{object}	responses.PaginatedResponse
//	@Security		BearerAuth
//	@Router			/leads [get]
func (h *LeadHandler) List(c *gin.Context) {
	h.logger.Info().Msg("[LeadHandler] List request received")

	var pagination requests.PaginationRequest
	c.ShouldBindQuery(&pagination)

	filters := port.LeadFilters{
		Status: c.Query("status"),
		Limit:  pagination.Limit(),
		Offset: pagination.Offset(),
	}

	h.logger.Info().
		Str("status", filters.Status).
		Int("limit", filters.Limit).
		Int("offset", filters.Offset).
		Msg("[LeadHandler] List fetching leads with filters")

	leads, total, err := h.leadApi.List(c.Request.Context(), filters)
	if err != nil {
		h.logger.Error().Err(err).Msg("[LeadHandler] List failed")
		HandleError(c, err)
		return
	}

	var items []responses.LeadResponse
	for _, l := range leads {
		items = append(items, toLeadResponse(l))
	}

	h.logger.Info().Int("total", total).Int("count", len(items)).Msg("[LeadHandler] List succeeded")

	c.JSON(http.StatusOK, responses.PaginatedResponse{
		Items: items, Total: total,
		Page: pagination.Page, PageSize: pagination.Limit(),
	})
}

// FindByID godoc
//
//	@Summary		Get lead detail
//	@Description	Returns lead by ID
//	@Tags			Leads
//	@Produce		json
//	@Param			id	path	string	true	"Lead ID"
//	@Success		200	{object}	responses.LeadResponse
//	@Security		BearerAuth
//	@Router			/leads/{id} [get]
func (h *LeadHandler) FindByID(c *gin.Context) {
	leadID := c.Param("id")
	h.logger.Info().Str("leadID", leadID).Msg("[LeadHandler] FindByID request received")

	lead, err := h.leadApi.FindByID(c.Request.Context(), leadID)
	if err != nil {
		h.logger.Error().Err(err).Str("leadID", leadID).Msg("[LeadHandler] FindByID failed")
		HandleError(c, err)
		return
	}

	h.logger.Info().Str("leadID", leadID).Msg("[LeadHandler] FindByID succeeded")

	c.JSON(http.StatusOK, toLeadResponse(lead))
}

// Update godoc
//
//	@Summary		Update lead
//	@Description	Update lead status and notes
//	@Tags			Leads
//	@Accept			json
//	@Produce		json
//	@Param			id		path	string					true	"Lead ID"
//	@Param			request	body	requests.UpdateLeadRequest	true	"Lead data"
//	@Success		200	{object}	responses.LeadResponse
//	@Security		BearerAuth
//	@Router			/leads/{id} [put]
func (h *LeadHandler) Update(c *gin.Context) {
	leadID := c.Param("id")
	h.logger.Info().Str("leadID", leadID).Msg("[LeadHandler] Update request received")

	var req requests.UpdateLeadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error().Err(err).Msg("[LeadHandler] Update failed to bind request body")
		c.JSON(http.StatusBadRequest, responses.ErrorResponse{Error: err.Error()})
		return
	}

	h.logger.Info().Str("leadID", leadID).Str("status", req.Status).Msg("[LeadHandler] Update processing")

	lead, err := h.leadApi.UpdateStatus(c.Request.Context(), leadID, req.Status, req.Notes)
	if err != nil {
		h.logger.Error().Err(err).Str("leadID", leadID).Msg("[LeadHandler] Update failed")
		HandleError(c, err)
		return
	}

	h.logger.Info().Str("leadID", leadID).Msg("[LeadHandler] Update succeeded")

	c.JSON(http.StatusOK, toLeadResponse(lead))
}

func toLeadResponse(l model.Lead) responses.LeadResponse {
	return responses.LeadResponse{
		ID: l.ID, CallID: l.CallID, Phone: l.Phone, Name: l.Name,
		Email: l.Email, Status: l.Status, Notes: l.Notes,
		CreatedAt: l.CreatedAt, UpdatedAt: l.UpdatedAt,
	}
}
