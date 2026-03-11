package api

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"phone-call-receptionist/backend/internal/domain/model"
	"phone-call-receptionist/backend/internal/domain/port"
)

// AppointmentApi provides business operations for appointment management.
type AppointmentApi struct {
	apptPort    port.Appointment
	calendar    port.Calendar
	voiceCaller port.VoiceCaller
	smsLogPort  port.SMSLog
	broadcaster port.EventBroadcaster
	logger      *zerolog.Logger
}

// NewAppointmentApi creates a new AppointmentApi with the given dependencies.
func NewAppointmentApi(
	apptPort port.Appointment,
	calendar port.Calendar,
	voiceCaller port.VoiceCaller,
	smsLogPort port.SMSLog,
	broadcaster port.EventBroadcaster,
	logger *zerolog.Logger,
) *AppointmentApi {
	return &AppointmentApi{
		apptPort:    apptPort,
		calendar:    calendar,
		voiceCaller: voiceCaller,
		smsLogPort:  smsLogPort,
		broadcaster: broadcaster,
		logger:      logger,
	}
}

// Book creates a new appointment with Google Calendar event and SMS confirmation.
func (a *AppointmentApi) Book(ctx context.Context, callID string, phone string, name string, email string, serviceType string, scheduledAt string, durationMins int, notes string) (model.Appointment, error) {
	a.logger.Info().Str("name", name).Str("phone", phone).Str("service", serviceType).Str("scheduledAt", scheduledAt).Msg("[AppointmentApi] Book started")

	now := time.Now().Format(time.RFC3339)
	appt := model.Appointment{
		ID:           uuid.New().String(),
		CallID:       callID,
		CallerPhone:  phone,
		CallerName:   name,
		CallerEmail:  email,
		ServiceType:  serviceType,
		ScheduledAt:  scheduledAt,
		DurationMins: durationMins,
		Status:       "pending",
		Notes:        notes,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	a.logger.Debug().Str("appointmentId", appt.ID).Msg("[AppointmentApi] Creating calendar event")
	eventID, err := a.calendar.CreateEvent(ctx, appt)
	if err != nil {
		a.logger.Warn().Err(err).Str("appointmentId", appt.ID).Msg("[AppointmentApi] Failed to create calendar event")
	} else {
		appt.GoogleEventID = eventID
		appt.Status = "confirmed"
		a.logger.Info().Str("appointmentId", appt.ID).Str("googleEventId", eventID).Msg("[AppointmentApi] Calendar event created")
	}

	if err := a.apptPort.Create(ctx, appt); err != nil {
		a.logger.Error().Err(err).Str("appointmentId", appt.ID).Msg("[AppointmentApi] Failed to create appointment record")
		return model.Appointment{}, fmt.Errorf("failed to create appointment: %w", err)
	}
	a.logger.Info().Str("appointmentId", appt.ID).Str("status", appt.Status).Msg("[AppointmentApi] Appointment record created")

	go a.sendConfirmationSMS(context.Background(), appt)

	a.broadcaster.Broadcast(ctx, map[string]interface{}{
		"type":        "appointment_booked",
		"appointmentId": appt.ID,
		"callerName":  name,
		"scheduledAt": scheduledAt,
	})

	a.logger.Info().Str("appointmentId", appt.ID).Msg("[AppointmentApi] Appointment booked successfully")
	return appt, nil
}

func (a *AppointmentApi) sendConfirmationSMS(ctx context.Context, appt model.Appointment) {
	a.logger.Info().Str("appointmentId", appt.ID).Str("phone", appt.CallerPhone).Msg("[AppointmentApi] Sending confirmation SMS")

	message := fmt.Sprintf("Bonjour %s, votre rendez-vous %s est confirmé pour le %s. Répondez ANNULER pour annuler.",
		appt.CallerName, appt.ServiceType, appt.ScheduledAt)

	sid, err := a.voiceCaller.SendSMS(ctx, appt.CallerPhone, message)
	if err != nil {
		a.logger.Error().Err(err).Str("phone", appt.CallerPhone).Msg("[AppointmentApi] Failed to send confirmation SMS")
		return
	}
	a.logger.Info().Str("smsSid", sid).Str("phone", appt.CallerPhone).Msg("[AppointmentApi] Confirmation SMS sent")

	now := time.Now().Format(time.RFC3339)
	smsLog := model.SMSLog{
		ID:        uuid.New().String(),
		CallID:    appt.CallID,
		ToPhone:   appt.CallerPhone,
		Message:   message,
		TwilioSID: sid,
		Status:    "sent",
		CreatedAt: now,
	}

	if err := a.smsLogPort.Create(ctx, smsLog); err != nil {
		a.logger.Error().Err(err).Msg("[AppointmentApi] Failed to log SMS")
	}

	appt.SMSSentAt = now
	appt.UpdatedAt = now
	if err := a.apptPort.Update(ctx, appt); err != nil {
		a.logger.Error().Err(err).Msg("[AppointmentApi] Failed to update appointment SMS timestamp")
	}
}

// Reschedule moves an appointment to a new time.
func (a *AppointmentApi) Reschedule(ctx context.Context, id string, newScheduledAt string) (model.Appointment, error) {
	a.logger.Info().Str("appointmentId", id).Str("newScheduledAt", newScheduledAt).Msg("[AppointmentApi] Reschedule started")

	appt, err := a.apptPort.FindByID(ctx, id)
	if err != nil {
		a.logger.Error().Err(err).Str("appointmentId", id).Msg("[AppointmentApi] Failed to find appointment")
		return model.Appointment{}, fmt.Errorf("failed to find appointment: %w", err)
	}

	if appt.GoogleEventID != "" {
		appt.ScheduledAt = newScheduledAt
		if err := a.calendar.UpdateEvent(ctx, appt.GoogleEventID, appt); err != nil {
			a.logger.Warn().Err(err).Str("googleEventId", appt.GoogleEventID).Msg("[AppointmentApi] Failed to update calendar event")
		} else {
			a.logger.Info().Str("googleEventId", appt.GoogleEventID).Msg("[AppointmentApi] Calendar event updated")
		}
	}

	appt.ScheduledAt = newScheduledAt
	appt.Status = "rescheduled"
	appt.UpdatedAt = time.Now().Format(time.RFC3339)

	if err := a.apptPort.Update(ctx, appt); err != nil {
		a.logger.Error().Err(err).Str("appointmentId", id).Msg("[AppointmentApi] Failed to update appointment")
		return model.Appointment{}, fmt.Errorf("failed to update appointment: %w", err)
	}

	go a.sendRescheduleSMS(context.Background(), appt)

	a.logger.Info().Str("appointmentId", id).Msg("[AppointmentApi] Appointment rescheduled successfully")
	return appt, nil
}

func (a *AppointmentApi) sendRescheduleSMS(ctx context.Context, appt model.Appointment) {
	a.logger.Info().Str("appointmentId", appt.ID).Str("phone", appt.CallerPhone).Msg("[AppointmentApi] Sending reschedule SMS")
	message := fmt.Sprintf("Bonjour %s, votre rendez-vous a été reprogrammé au %s.",
		appt.CallerName, appt.ScheduledAt)
	_, err := a.voiceCaller.SendSMS(ctx, appt.CallerPhone, message)
	if err != nil {
		a.logger.Error().Err(err).Str("phone", appt.CallerPhone).Msg("[AppointmentApi] Failed to send reschedule SMS")
	}
}

// Cancel cancels an appointment and removes the calendar event.
func (a *AppointmentApi) Cancel(ctx context.Context, id string) error {
	a.logger.Info().Str("appointmentId", id).Msg("[AppointmentApi] Cancel started")

	appt, err := a.apptPort.FindByID(ctx, id)
	if err != nil {
		a.logger.Error().Err(err).Str("appointmentId", id).Msg("[AppointmentApi] Failed to find appointment")
		return fmt.Errorf("failed to find appointment: %w", err)
	}

	if appt.GoogleEventID != "" {
		if err := a.calendar.DeleteEvent(ctx, appt.GoogleEventID); err != nil {
			a.logger.Warn().Err(err).Str("googleEventId", appt.GoogleEventID).Msg("[AppointmentApi] Failed to delete calendar event")
		} else {
			a.logger.Info().Str("googleEventId", appt.GoogleEventID).Msg("[AppointmentApi] Calendar event deleted")
		}
	}

	appt.Status = "cancelled"
	appt.UpdatedAt = time.Now().Format(time.RFC3339)

	if err := a.apptPort.Update(ctx, appt); err != nil {
		a.logger.Error().Err(err).Str("appointmentId", id).Msg("[AppointmentApi] Failed to update appointment")
		return fmt.Errorf("failed to update appointment: %w", err)
	}

	go func() {
		a.logger.Info().Str("appointmentId", appt.ID).Str("phone", appt.CallerPhone).Msg("[AppointmentApi] Sending cancellation SMS")
		message := fmt.Sprintf("Bonjour %s, votre rendez-vous a été annulé.", appt.CallerName)
		a.voiceCaller.SendSMS(context.Background(), appt.CallerPhone, message)
	}()

	a.logger.Info().Str("appointmentId", id).Msg("[AppointmentApi] Appointment cancelled successfully")
	return nil
}

// Availability returns available time slots within the given time range.
func (a *AppointmentApi) Availability(ctx context.Context, from string, to string) ([]model.TimeSlot, error) {
	a.logger.Debug().Str("from", from).Str("to", to).Msg("[AppointmentApi] Checking availability")
	slots, err := a.calendar.CheckAvailability(ctx, from, to)
	if err != nil {
		a.logger.Error().Err(err).Msg("[AppointmentApi] Failed to check availability")
		return nil, fmt.Errorf("failed to check availability: %w", err)
	}
	a.logger.Debug().Int("slots", len(slots)).Msg("[AppointmentApi] Availability retrieved")
	return slots, nil
}

// List retrieves a paginated list of appointments.
func (a *AppointmentApi) List(ctx context.Context, filters port.AppointmentFilters) ([]model.Appointment, int, error) {
	a.logger.Debug().Int("limit", filters.Limit).Int("offset", filters.Offset).Msg("[AppointmentApi] Listing appointments")
	appts, total, err := a.apptPort.List(ctx, filters)
	if err != nil {
		a.logger.Error().Err(err).Msg("[AppointmentApi] Failed to list appointments")
		return nil, 0, fmt.Errorf("failed to list appointments: %w", err)
	}
	a.logger.Debug().Int("total", total).Int("returned", len(appts)).Msg("[AppointmentApi] Appointments listed")
	return appts, total, nil
}

// FindByID retrieves a specific appointment by its ID.
func (a *AppointmentApi) FindByID(ctx context.Context, id string) (model.Appointment, error) {
	a.logger.Debug().Str("appointmentId", id).Msg("[AppointmentApi] Finding appointment by ID")
	appt, err := a.apptPort.FindByID(ctx, id)
	if err != nil {
		a.logger.Error().Err(err).Str("appointmentId", id).Msg("[AppointmentApi] Failed to find appointment")
		return model.Appointment{}, fmt.Errorf("failed to find appointment: %w", err)
	}
	return appt, nil
}
