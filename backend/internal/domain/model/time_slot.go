package model

// TimeSlot represents an available time window for appointment booking.
type TimeSlot struct {
	// Start is the beginning of the time slot in RFC3339 format.
	Start string
	// End is the end of the time slot in RFC3339 format.
	End string
}
