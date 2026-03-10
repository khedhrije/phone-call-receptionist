package requests

// CreateAppointmentRequest contains the fields required to book a new appointment.
type CreateAppointmentRequest struct {
	// CallerPhone is the phone number of the person booking.
	CallerPhone string `json:"callerPhone" binding:"required"`
	// CallerName is the full name of the person booking.
	CallerName string `json:"callerName" binding:"required"`
	// CallerEmail is the email address of the person booking.
	CallerEmail string `json:"callerEmail" binding:"required,email"`
	// ServiceType is the type of IT service requested.
	ServiceType string `json:"serviceType" binding:"required"`
	// ScheduledAt is the requested appointment time in RFC3339 format.
	ScheduledAt string `json:"scheduledAt" binding:"required"`
	// DurationMins is the appointment duration in minutes.
	DurationMins int `json:"durationMins" binding:"required,min=15,max=480"`
	// Notes contains additional appointment notes.
	Notes string `json:"notes"`
}

// RescheduleAppointmentRequest contains the fields required to reschedule an appointment.
type RescheduleAppointmentRequest struct {
	// ScheduledAt is the new appointment time in RFC3339 format.
	ScheduledAt string `json:"scheduledAt" binding:"required"`
}

// AvailabilityRequest contains the time range for checking appointment availability.
type AvailabilityRequest struct {
	// From is the start of the time range in RFC3339 format.
	From string `form:"from" binding:"required"`
	// To is the end of the time range in RFC3339 format.
	To string `form:"to" binding:"required"`
}
