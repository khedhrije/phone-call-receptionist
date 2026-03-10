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

	eventID, err := a.calendar.CreateEvent(ctx, appt)
	if err != nil {
		a.logger.Warn().Err(err).Msg("Failed to create calendar event")
	} else {
		appt.GoogleEventID = eventID
		appt.Status = "confirmed"
	}

	if err := a.apptPort.Create(ctx, appt); err != nil {
		return model.Appointment{}, fmt.Errorf("failed to create appointment: %w", err)
	}

	go a.sendConfirmationSMS(context.Background(), appt)

	a.broadcaster.Broadcast(map[string]interface{}{
		"type":        "appointment_booked",
		"appointmentId": appt.ID,
		"callerName":  name,
		"scheduledAt": scheduledAt,
	})

	a.logger.Info().Str("appointmentId", appt.ID).Msg("Appointment booked")
	return appt, nil
}

func (a *AppointmentApi) sendConfirmationSMS(ctx context.Context, appt model.Appointment) {
	message := fmt.Sprintf("Hi %s, your %s appointment is confirmed for %s. Reply CANCEL to cancel.",
		appt.CallerName, appt.ServiceType, appt.ScheduledAt)

	sid, err := a.voiceCaller.SendSMS(ctx, appt.CallerPhone, message)
	if err != nil {
		a.logger.Error().Err(err).Str("phone", appt.CallerPhone).Msg("Failed to send confirmation SMS")
		return
	}

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
		a.logger.Error().Err(err).Msg("Failed to log SMS")
	}

	appt.SMSSentAt = now
	appt.UpdatedAt = now
	if err := a.apptPort.Update(ctx, appt); err != nil {
		a.logger.Error().Err(err).Msg("Failed to update appointment SMS timestamp")
	}
}

// Reschedule moves an appointment to a new time.
func (a *AppointmentApi) Reschedule(ctx context.Context, id string, newScheduledAt string) (model.Appointment, error) {
	appt, err := a.apptPort.FindByID(ctx, id)
	if err != nil {
		return model.Appointment{}, fmt.Errorf("failed to find appointment: %w", err)
	}

	if appt.GoogleEventID != "" {
		appt.ScheduledAt = newScheduledAt
		if err := a.calendar.UpdateEvent(ctx, appt.GoogleEventID, appt); err != nil {
			a.logger.Warn().Err(err).Msg("Failed to update calendar event")
		}
	}

	appt.ScheduledAt = newScheduledAt
	appt.Status = "rescheduled"
	appt.UpdatedAt = time.Now().Format(time.RFC3339)

	if err := a.apptPort.Update(ctx, appt); err != nil {
		return model.Appointment{}, fmt.Errorf("failed to update appointment: %w", err)
	}

	go a.sendRescheduleSMS(context.Background(), appt)

	a.logger.Info().Str("appointmentId", id).Msg("Appointment rescheduled")
	return appt, nil
}

func (a *AppointmentApi) sendRescheduleSMS(ctx context.Context, appt model.Appointment) {
	message := fmt.Sprintf("Hi %s, your appointment has been rescheduled to %s.",
		appt.CallerName, appt.ScheduledAt)
	a.voiceCaller.SendSMS(ctx, appt.CallerPhone, message)
}

// Cancel cancels an appointment and removes the calendar event.
func (a *AppointmentApi) Cancel(ctx context.Context, id string) error {
	appt, err := a.apptPort.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to find appointment: %w", err)
	}

	if appt.GoogleEventID != "" {
		if err := a.calendar.DeleteEvent(ctx, appt.GoogleEventID); err != nil {
			a.logger.Warn().Err(err).Msg("Failed to delete calendar event")
		}
	}

	appt.Status = "cancelled"
	appt.UpdatedAt = time.Now().Format(time.RFC3339)

	if err := a.apptPort.Update(ctx, appt); err != nil {
		return fmt.Errorf("failed to update appointment: %w", err)
	}

	go func() {
		message := fmt.Sprintf("Hi %s, your appointment has been cancelled.", appt.CallerName)
		a.voiceCaller.SendSMS(context.Background(), appt.CallerPhone, message)
	}()

	a.logger.Info().Str("appointmentId", id).Msg("Appointment cancelled")
	return nil
}

// Availability returns available time slots within the given time range.
func (a *AppointmentApi) Availability(ctx context.Context, from string, to string) ([]model.TimeSlot, error) {
	slots, err := a.calendar.CheckAvailability(ctx, from, to)
	if err != nil {
		return nil, fmt.Errorf("failed to check availability: %w", err)
	}
	return slots, nil
}

// List retrieves a paginated list of appointments.
func (a *AppointmentApi) List(ctx context.Context, filters port.AppointmentFilters) ([]model.Appointment, int, error) {
	appts, total, err := a.apptPort.List(ctx, filters)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list appointments: %w", err)
	}
	return appts, total, nil
}

// FindByID retrieves a specific appointment by its ID.
func (a *AppointmentApi) FindByID(ctx context.Context, id string) (model.Appointment, error) {
	appt, err := a.apptPort.FindByID(ctx, id)
	if err != nil {
		return model.Appointment{}, fmt.Errorf("failed to find appointment: %w", err)
	}
	return appt, nil
}
