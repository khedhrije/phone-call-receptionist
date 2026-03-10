package model

// KnowledgeDocument represents an uploaded document in the knowledge base.
type KnowledgeDocument struct {
	// ID is the unique identifier for the document.
	ID string
	// Filename is the original name of the uploaded file.
	Filename string
	// MimeType is the MIME type of the uploaded file.
	MimeType string
	// FilePath is the storage path of the uploaded file.
	FilePath string
	// ChunkCount is the number of chunks the document was split into.
	ChunkCount int
	// Status is the current indexing status (pending, indexing, indexed, failed).
	Status string
	// IndexedAt is the timestamp when indexing completed in RFC3339 format.
	// Empty string means not yet indexed.
	IndexedAt string
	// CreatedAt is the creation timestamp in RFC3339 format.
	CreatedAt string
	// UpdatedAt is the last modification timestamp in RFC3339 format.
	UpdatedAt string
}
