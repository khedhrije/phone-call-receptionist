package responses

// AppointmentResponse contains the appointment data returned by the API.
type AppointmentResponse struct {
	// ID is the appointment's unique identifier.
	ID string `json:"id"`
	// CallID is the related call identifier.
	CallID string `json:"callId"`
	// CallerPhone is the phone number of the person who booked.
	CallerPhone string `json:"callerPhone"`
	// CallerName is the full name of the person who booked.
	CallerName string `json:"callerName"`
	// CallerEmail is the email address of the person who booked.
	CallerEmail string `json:"callerEmail"`
	// ServiceType is the type of IT service requested.
	ServiceType string `json:"serviceType"`
	// ScheduledAt is the appointment date and time in RFC3339 format.
	ScheduledAt string `json:"scheduledAt"`
	// DurationMins is the appointment duration in minutes.
	DurationMins int `json:"durationMins"`
	// Status is the current appointment status.
	Status string `json:"status"`
	// GoogleEventID is the Google Calendar event identifier.
	GoogleEventID string `json:"googleEventId"`
	// SMSSentAt is the SMS confirmation timestamp in RFC3339 format.
	SMSSentAt string `json:"smsSentAt"`
	// Notes contains additional appointment notes.
	Notes string `json:"notes"`
	// CreatedAt is the creation timestamp in RFC3339 format.
	CreatedAt string `json:"createdAt"`
	// UpdatedAt is the last modification timestamp in RFC3339 format.
	UpdatedAt string `json:"updatedAt"`
}

// TimeSlotResponse contains an available time slot.
type TimeSlotResponse struct {
	// Start is the beginning of the time slot in RFC3339 format.
	Start string `json:"start"`
	// End is the end of the time slot in RFC3339 format.
	End string `json:"end"`
}
