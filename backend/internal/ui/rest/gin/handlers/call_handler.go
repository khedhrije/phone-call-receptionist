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

// CallHandler handles call-related HTTP requests.
type CallHandler struct {
	voiceCallApi *api.VoiceCallApi
	dashboardApi *api.DashboardApi
	logger       *zerolog.Logger
}

// NewCallHandler creates a new CallHandler with the given dependencies.
func NewCallHandler(voiceCallApi *api.VoiceCallApi, dashboardApi *api.DashboardApi, logger *zerolog.Logger) *CallHandler {
	return &CallHandler{voiceCallApi: voiceCallApi, dashboardApi: dashboardApi, logger: logger}
}

// List godoc
//
//	@Summary		List inbound calls
//	@Description	Returns paginated list of inbound calls
//	@Tags			Calls
//	@Produce		json
//	@Param			status		query	string	false	"Filter by status"
//	@Param			from		query	string	false	"From date (RFC3339)"
//	@Param			to			query	string	false	"To date (RFC3339)"
//	@Param			page		query	int		false	"Page number"
//	@Param			pageSize	query	int		false	"Items per page"
//	@Success		200	{object}	responses.PaginatedResponse
//	@Security		BearerAuth
//	@Router			/calls [get]
func (h *CallHandler) List(c *gin.Context) {
	h.logger.Info().Msg("[CallHandler] List request received")

	var pagination requests.PaginationRequest
	c.ShouldBindQuery(&pagination)

	filters := port.CallFilters{
		Status: c.Query("status"),
		From:   c.Query("from"),
		To:     c.Query("to"),
		Limit:  pagination.Limit(),
		Offset: pagination.Offset(),
	}

	h.logger.Info().
		Str("status", filters.Status).
		Str("from", filters.From).
		Str("to", filters.To).
		Int("limit", filters.Limit).
		Int("offset", filters.Offset).
		Msg("[CallHandler] List fetching calls with filters")

	calls, total, err := h.voiceCallApi.CallHistory(c.Request.Context(), filters)
	if err != nil {
		h.logger.Error().Err(err).Msg("[CallHandler] List failed to fetch calls")
		HandleError(c, err)
		return
	}

	var items []responses.CallResponse
	for _, call := range calls {
		items = append(items, toCallResponse(call))
	}

	h.logger.Info().Int("total", total).Int("count", len(items)).Msg("[CallHandler] List succeeded")

	c.JSON(http.StatusOK, responses.PaginatedResponse{
		Items: items, Total: total,
		Page: pagination.Page, PageSize: pagination.Limit(),
	})
}

// Detail godoc
//
//	@Summary		Get call detail
//	@Description	Returns full call detail with transcript
//	@Tags			Calls
//	@Produce		json
//	@Param			id	path	string	true	"Call ID"
//	@Success		200	{object}	responses.CallResponse
//	@Security		BearerAuth
//	@Router			/calls/{id} [get]
func (h *CallHandler) Detail(c *gin.Context) {
	callID := c.Param("id")
	h.logger.Info().Str("callID", callID).Msg("[CallHandler] Detail request received")

	call, err := h.voiceCallApi.CallDetail(c.Request.Context(), callID)
	if err != nil {
		h.logger.Error().Err(err).Str("callID", callID).Msg("[CallHandler] Detail failed")
		HandleError(c, err)
		return
	}

	h.logger.Info().Str("callID", callID).Msg("[CallHandler] Detail succeeded")

	c.JSON(http.StatusOK, toCallResponse(call))
}

// RAGQueries godoc
//
//	@Summary		Get RAG queries for a call
//	@Description	Returns RAG queries executed during a call
//	@Tags			Calls
//	@Produce		json
//	@Param			id	path	string	true	"Call ID"
//	@Success		200	{array}	responses.RAGQueryResponse
//	@Security		BearerAuth
//	@Router			/calls/{id}/rag-queries [get]
func (h *CallHandler) RAGQueries(c *gin.Context) {
	callID := c.Param("id")
	h.logger.Info().Str("callID", callID).Msg("[CallHandler] RAGQueries request received")

	call, err := h.voiceCallApi.CallDetail(c.Request.Context(), callID)
	if err != nil {
		h.logger.Error().Err(err).Str("callID", callID).Msg("[CallHandler] RAGQueries failed to fetch call")
		HandleError(c, err)
		return
	}

	var items []responses.RAGQueryResponse
	for _, q := range call.RAGQueries {
		items = append(items, responses.RAGQueryResponse{
			Query: q.Query, Chunks: q.Chunks, Response: q.Response,
			Provider: q.Provider, Tokens: q.Tokens, At: q.At,
		})
	}

	h.logger.Info().Str("callID", callID).Int("queryCount", len(items)).Msg("[CallHandler] RAGQueries succeeded")

	c.JSON(http.StatusOK, items)
}

// Stats godoc
//
//	@Summary		Get call statistics
//	@Description	Returns call cost and volume stats
//	@Tags			Calls
//	@Produce		json
//	@Success		200	{object}	responses.DashboardStatsResponse
//	@Security		BearerAuth
//	@Router			/calls/stats [get]
func (h *CallHandler) Stats(c *gin.Context) {
	h.logger.Info().Msg("[CallHandler] Stats request received")

	stats, err := h.dashboardApi.Stats(c.Request.Context())
	if err != nil {
		h.logger.Error().Err(err).Msg("[CallHandler] Stats failed")
		HandleError(c, err)
		return
	}

	h.logger.Info().Msg("[CallHandler] Stats succeeded")

	c.JSON(http.StatusOK, stats)
}

func toCallResponse(call model.InboundCall) responses.CallResponse {
	var transcript []responses.TranscriptEntryResponse
	for _, t := range call.Transcript {
		transcript = append(transcript, responses.TranscriptEntryResponse{
			Speaker: t.Speaker, Text: t.Text, At: t.At,
		})
	}

	var ragQueries []responses.RAGQueryResponse
	for _, q := range call.RAGQueries {
		ragQueries = append(ragQueries, responses.RAGQueryResponse{
			Query: q.Query, Chunks: q.Chunks, Response: q.Response,
			Provider: q.Provider, Tokens: q.Tokens, At: q.At,
		})
	}

	return responses.CallResponse{
		ID: call.ID, TwilioCallSID: call.TwilioCallSID, CallerPhone: call.CallerPhone,
		Status: call.Status, Transcript: transcript, RAGQueries: ragQueries,
		DurationSeconds: call.DurationSeconds, TwilioCostUSD: call.TwilioCostUSD,
		LLMCostUSD: call.LLMCostUSD, TotalCostUSD: call.TotalCostUSD,
		CreatedAt: call.CreatedAt, EndedAt: call.EndedAt,
	}
}
