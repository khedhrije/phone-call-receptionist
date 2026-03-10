package requests

// PaginationRequest contains common pagination parameters.
type PaginationRequest struct {
	// Page is the page number (1-based).
	Page int `form:"page" binding:"omitempty,min=1"`
	// PageSize is the number of items per page.
	PageSize int `form:"pageSize" binding:"omitempty,min=1,max=100"`
}

// Limit returns the SQL LIMIT value from the pagination params.
func (p PaginationRequest) Limit() int {
	if p.PageSize <= 0 {
		return 20
	}
	return p.PageSize
}

// Offset returns the SQL OFFSET value from the pagination params.
func (p PaginationRequest) Offset() int {
	if p.Page <= 1 {
		return 0
	}
	return (p.Page - 1) * p.Limit()
}
