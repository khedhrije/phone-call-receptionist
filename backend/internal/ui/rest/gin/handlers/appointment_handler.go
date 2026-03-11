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

// AppointmentHandler handles appointment-related HTTP requests.
type AppointmentHandler struct {
	apptApi *api.AppointmentApi
	logger  *zerolog.Logger
}

// NewAppointmentHandler creates a new AppointmentHandler with the given dependencies.
func NewAppointmentHandler(apptApi *api.AppointmentApi, logger *zerolog.Logger) *AppointmentHandler {
	return &AppointmentHandler{apptApi: apptApi, logger: logger}
}

// Create godoc
//
//	@Summary		Book an appointment
//	@Description	Create a new appointment
//	@Tags			Appointments
//	@Accept			json
//	@Produce		json
//	@Param			request	body		requests.CreateAppointmentRequest	true	"Appointment data"
//	@Success		201		{object}	responses.AppointmentResponse
//	@Security		BearerAuth
//	@Router			/appointments [post]
func (h *AppointmentHandler) Create(c *gin.Context) {
	h.logger.Info().Msg("[AppointmentHandler] Create request received")

	var req requests.CreateAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error().Err(err).Msg("[AppointmentHandler] Create failed to bind request body")
		c.JSON(http.StatusBadRequest, responses.ErrorResponse{Error: err.Error()})
		return
	}

	h.logger.Info().
		Str("callerPhone", req.CallerPhone).
		Str("serviceType", req.ServiceType).
		Str("scheduledAt", req.ScheduledAt).
		Msg("[AppointmentHandler] Create processing appointment")

	appt, err := h.apptApi.Book(c.Request.Context(), "", req.CallerPhone, req.CallerName, req.CallerEmail, req.ServiceType, req.ScheduledAt, req.DurationMins, req.Notes)
	if err != nil {
		h.logger.Error().Err(err).Str("callerPhone", req.CallerPhone).Msg("[AppointmentHandler] Create failed")
		HandleError(c, err)
		return
	}

	h.logger.Info().Str("appointmentID", appt.ID).Msg("[AppointmentHandler] Create succeeded")

	c.JSON(http.StatusCreated, toApptResponse(appt))
}

// List godoc
//
//	@Summary		List appointments
//	@Description	Returns paginated list of appointments
//	@Tags			Appointments
//	@Produce		json
//	@Param			status		query	string	false	"Filter by status"
//	@Param			page		query	int		false	"Page number"
//	@Param			pageSize	query	int		false	"Items per page"
//	@Success		200	{object}	responses.PaginatedResponse
//	@Security		BearerAuth
//	@Router			/appointments [get]
func (h *AppointmentHandler) List(c *gin.Context) {
	h.logger.Info().Msg("[AppointmentHandler] List request received")

	var pagination requests.PaginationRequest
	c.ShouldBindQuery(&pagination)

	filters := port.AppointmentFilters{
		Status: c.Query("status"),
		Limit:  pagination.Limit(),
		Offset: pagination.Offset(),
	}

	h.logger.Info().
		Str("status", filters.Status).
		Int("limit", filters.Limit).
		Int("offset", filters.Offset).
		Msg("[AppointmentHandler] List fetching appointments with filters")

	appts, total, err := h.apptApi.List(c.Request.Context(), filters)
	if err != nil {
		h.logger.Error().Err(err).Msg("[AppointmentHandler] List failed")
		HandleError(c, err)
		return
	}

	var items []responses.AppointmentResponse
	for _, a := range appts {
		items = append(items, toApptResponse(a))
	}

	h.logger.Info().Int("total", total).Int("count", len(items)).Msg("[AppointmentHandler] List succeeded")

	c.JSON(http.StatusOK, responses.PaginatedResponse{
		Items: items, Total: total,
		Page: pagination.Page, PageSize: pagination.Limit(),
	})
}

// FindByID godoc
//
//	@Summary		Get appointment detail
//	@Description	Returns appointment by ID
//	@Tags			Appointments
//	@Produce		json
//	@Param			id	path	string	true	"Appointment ID"
//	@Success		200	{object}	responses.AppointmentResponse
//	@Security		BearerAuth
//	@Router			/appointments/{id} [get]
func (h *AppointmentHandler) FindByID(c *gin.Context) {
	apptID := c.Param("id")
	h.logger.Info().Str("appointmentID", apptID).Msg("[AppointmentHandler] FindByID request received")

	appt, err := h.apptApi.FindByID(c.Request.Context(), apptID)
	if err != nil {
		h.logger.Error().Err(err).Str("appointmentID", apptID).Msg("[AppointmentHandler] FindByID failed")
		HandleError(c, err)
		return
	}

	h.logger.Info().Str("appointmentID", apptID).Msg("[AppointmentHandler] FindByID succeeded")

	c.JSON(http.StatusOK, toApptResponse(appt))
}

// Reschedule godoc
//
//	@Summary		Reschedule appointment
//	@Description	Change appointment time
//	@Tags			Appointments
//	@Accept			json
//	@Produce		json
//	@Param			id		path	string								true	"Appointment ID"
//	@Param			request	body	requests.RescheduleAppointmentRequest	true	"New time"
//	@Success		200	{object}	responses.AppointmentResponse
//	@Security		BearerAuth
//	@Router			/appointments/{id} [put]
func (h *AppointmentHandler) Reschedule(c *gin.Context) {
	apptID := c.Param("id")
	h.logger.Info().Str("appointmentID", apptID).Msg("[AppointmentHandler] Reschedule request received")

	var req requests.RescheduleAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error().Err(err).Msg("[AppointmentHandler] Reschedule failed to bind request body")
		c.JSON(http.StatusBadRequest, responses.ErrorResponse{Error: err.Error()})
		return
	}

	h.logger.Info().Str("appointmentID", apptID).Str("scheduledAt", req.ScheduledAt).Msg("[AppointmentHandler] Reschedule processing")

	appt, err := h.apptApi.Reschedule(c.Request.Context(), apptID, req.ScheduledAt)
	if err != nil {
		h.logger.Error().Err(err).Str("appointmentID", apptID).Msg("[AppointmentHandler] Reschedule failed")
		HandleError(c, err)
		return
	}

	h.logger.Info().Str("appointmentID", apptID).Msg("[AppointmentHandler] Reschedule succeeded")

	c.JSON(http.StatusOK, toApptResponse(appt))
}

// Cancel godoc
//
//	@Summary		Cancel appointment
//	@Description	Cancel an appointment
//	@Tags			Appointments
//	@Produce		json
//	@Param			id	path	string	true	"Appointment ID"
//	@Success		204
//	@Security		BearerAuth
//	@Router			/appointments/{id} [delete]
func (h *AppointmentHandler) Cancel(c *gin.Context) {
	apptID := c.Param("id")
	h.logger.Info().Str("appointmentID", apptID).Msg("[AppointmentHandler] Cancel request received")

	if err := h.apptApi.Cancel(c.Request.Context(), apptID); err != nil {
		h.logger.Error().Err(err).Str("appointmentID", apptID).Msg("[AppointmentHandler] Cancel failed")
		HandleError(c, err)
		return
	}

	h.logger.Info().Str("appointmentID", apptID).Msg("[AppointmentHandler] Cancel succeeded")

	c.Status(http.StatusNoContent)
}

// Availability godoc
//
//	@Summary		Check availability
//	@Description	Returns available time slots
//	@Tags			Appointments
//	@Produce		json
//	@Param			from	query	string	true	"From date (RFC3339)"
//	@Param			to		query	string	true	"To date (RFC3339)"
//	@Success		200	{array}	responses.TimeSlotResponse
//	@Security		BearerAuth
//	@Router			/appointments/availability [get]
func (h *AppointmentHandler) Availability(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")

	h.logger.Info().Str("from", from).Str("to", to).Msg("[AppointmentHandler] Availability request received")

	if from == "" || to == "" {
		h.logger.Error().Msg("[AppointmentHandler] Availability missing from/to query parameters")
		c.JSON(http.StatusBadRequest, responses.ErrorResponse{Error: "from and to query parameters are required"})
		return
	}

	slots, err := h.apptApi.Availability(c.Request.Context(), from, to)
	if err != nil {
		h.logger.Error().Err(err).Str("from", from).Str("to", to).Msg("[AppointmentHandler] Availability failed")
		HandleError(c, err)
		return
	}

	var items []responses.TimeSlotResponse
	for _, s := range slots {
		items = append(items, responses.TimeSlotResponse{Start: s.Start, End: s.End})
	}

	h.logger.Info().Int("slotCount", len(items)).Msg("[AppointmentHandler] Availability succeeded")

	c.JSON(http.StatusOK, items)
}

func toApptResponse(a model.Appointment) responses.AppointmentResponse {
	return responses.AppointmentResponse{
		ID: a.ID, CallID: a.CallID, CallerPhone: a.CallerPhone,
		CallerName: a.CallerName, CallerEmail: a.CallerEmail,
		ServiceType: a.ServiceType, ScheduledAt: a.ScheduledAt,
		DurationMins: a.DurationMins, Status: a.Status,
		GoogleEventID: a.GoogleEventID, SMSSentAt: a.SMSSentAt,
		Notes: a.Notes, CreatedAt: a.CreatedAt, UpdatedAt: a.UpdatedAt,
	}
}
