package port

import (
	"context"

	"phone-call-receptionist/backend/internal/domain/model"
)

// Calendar defines the interface for calendar operations.
// Implementations manage appointment scheduling via external calendar services.
type Calendar interface {
	// CheckAvailability returns available time slots within the given time range.
	// The from and to parameters are RFC3339 formatted timestamps.
	CheckAvailability(ctx context.Context, from string, to string) ([]model.TimeSlot, error)
	// CreateEvent creates a calendar event for the given appointment.
	// Returns the external event ID and any error.
	CreateEvent(ctx context.Context, appt model.Appointment) (string, error)
	// UpdateEvent modifies an existing calendar event.
	UpdateEvent(ctx context.Context, eventID string, appt model.Appointment) error
	// DeleteEvent removes a calendar event by its external event ID.
	DeleteEvent(ctx context.Context, eventID string) error
}
