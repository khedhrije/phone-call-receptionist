package model

// Lead represents a potential customer captured from an inbound call.
type Lead struct {
	// ID is the unique identifier for the lead.
	ID string
	// CallID is the identifier of the inbound call that created this lead.
	// Empty string means the lead was created manually.
	CallID string
	// Phone is the lead's phone number (unique).
	Phone string
	// Name is the lead's full name.
	Name string
	// Email is the lead's email address.
	Email string
	// Status is the current lead status (new, contacted, qualified, converted, lost).
	Status string
	// Notes contains additional notes about the lead.
	Notes string
	// CreatedAt is the creation timestamp in RFC3339 format.
	CreatedAt string
	// UpdatedAt is the last modification timestamp in RFC3339 format.
	UpdatedAt string
}
