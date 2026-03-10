package requests

// SearchKnowledgeRequest contains the fields for a semantic search query.
type SearchKnowledgeRequest struct {
	// Query is the natural language search query.
	Query string `json:"query" binding:"required"`
	// TopK is the number of results to return.
	TopK int `json:"topK" binding:"omitempty,min=1,max=20"`
}
