package responses

// KnowledgeDocumentResponse contains the knowledge document data returned by the API.
type KnowledgeDocumentResponse struct {
	// ID is the document's unique identifier.
	ID string `json:"id"`
	// Filename is the original file name.
	Filename string `json:"filename"`
	// MimeType is the file MIME type.
	MimeType string `json:"mimeType"`
	// ChunkCount is the number of chunks the document was split into.
	ChunkCount int `json:"chunkCount"`
	// Status is the current indexing status.
	Status string `json:"status"`
	// IndexedAt is the indexing completion timestamp in RFC3339 format.
	IndexedAt string `json:"indexedAt"`
	// CreatedAt is the creation timestamp in RFC3339 format.
	CreatedAt string `json:"createdAt"`
	// UpdatedAt is the last modification timestamp in RFC3339 format.
	UpdatedAt string `json:"updatedAt"`
}

// SearchResultResponse contains a single search result from the knowledge base.
type SearchResultResponse struct {
	// ChunkID is the matched chunk's identifier.
	ChunkID string `json:"chunkId"`
	// DocumentID is the source document's identifier.
	DocumentID string `json:"documentId"`
	// Content is the matched text content.
	Content string `json:"content"`
	// PageNumber is the source page number.
	PageNumber int `json:"pageNumber"`
	// Score is the similarity score.
	Score float64 `json:"score"`
}

// KnowledgeSearchResponse contains the search results and generated answer.
type KnowledgeSearchResponse struct {
	// Answer is the LLM-generated answer from the retrieved context.
	Answer string `json:"answer"`
	// Sources is the list of source chunks used to generate the answer.
	Sources []SearchResultResponse `json:"sources"`
	// Provider is the LLM provider that generated the answer.
	Provider string `json:"provider"`
	// Tokens is the number of tokens consumed.
	Tokens int `json:"tokens"`
}
