package requests

// UpdateLeadRequest contains the fields that can be updated on a lead.
type UpdateLeadRequest struct {
	// Status is the new lead status (new, contacted, qualified, converted, lost).
	Status string `json:"status" binding:"required,oneof=new contacted qualified converted lost"`
	// Notes contains additional notes about the lead.
	Notes string `json:"notes"`
}
