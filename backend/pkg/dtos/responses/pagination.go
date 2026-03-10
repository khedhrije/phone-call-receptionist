package responses

// PaginatedResponse wraps a paginated list of items with metadata.
type PaginatedResponse struct {
	// Items contains the page of results.
	Items interface{} `json:"items"`
	// Total is the total number of matching items across all pages.
	Total int `json:"total"`
	// Page is the current page number (1-based).
	Page int `json:"page"`
	// PageSize is the number of items per page.
	PageSize int `json:"pageSize"`
}

// ErrorResponse contains an error message returned by the API.
type ErrorResponse struct {
	// Error is the error message.
	Error string `json:"error"`
}
