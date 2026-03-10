package responses

// LeadResponse contains the lead data returned by the API.
type LeadResponse struct {
	// ID is the lead's unique identifier.
	ID string `json:"id"`
	// CallID is the related call identifier.
	CallID string `json:"callId"`
	// Phone is the lead's phone number.
	Phone string `json:"phone"`
	// Name is the lead's full name.
	Name string `json:"name"`
	// Email is the lead's email address.
	Email string `json:"email"`
	// Status is the current lead status.
	Status string `json:"status"`
	// Notes contains additional notes about the lead.
	Notes string `json:"notes"`
	// CreatedAt is the creation timestamp in RFC3339 format.
	CreatedAt string `json:"createdAt"`
	// UpdatedAt is the last modification timestamp in RFC3339 format.
	UpdatedAt string `json:"updatedAt"`
}
