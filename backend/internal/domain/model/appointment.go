package model

// Appointment represents a scheduled appointment created during or after a call.
type Appointment struct {
	// ID is the unique identifier for the appointment.
	ID string
	// CallID is the identifier of the inbound call that initiated this appointment.
	// Empty string means created manually, not from a call.
	CallID string
	// CallerPhone is the phone number of the person who booked.
	CallerPhone string
	// CallerName is the full name of the person who booked.
	CallerName string
	// CallerEmail is the email address of the person who booked.
	CallerEmail string
	// ServiceType is the type of IT service requested.
	ServiceType string
	// ScheduledAt is the appointment date and time in RFC3339 format.
	ScheduledAt string
	// DurationMins is the appointment duration in minutes.
	DurationMins int
	// Status is the current appointment status (pending, confirmed, rescheduled, cancelled).
	Status string
	// GoogleEventID is the Google Calendar event identifier.
	// Empty string means no calendar event was created.
	GoogleEventID string
	// SMSSentAt is the timestamp when the SMS confirmation was sent in RFC3339 format.
	// Empty string means no SMS was sent.
	SMSSentAt string
	// Notes contains additional notes about the appointment.
	Notes string
	// CreatedAt is the creation timestamp in RFC3339 format.
	CreatedAt string
	// UpdatedAt is the last modification timestamp in RFC3339 format.
	UpdatedAt string
}
