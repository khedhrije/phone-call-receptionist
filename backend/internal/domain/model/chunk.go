package model

// Chunk represents a text chunk from a knowledge document with its embedding.
type Chunk struct {
	// ID is the unique identifier for the chunk.
	ID string
	// DocumentID is the identifier of the parent knowledge document.
	DocumentID string
	// Content is the text content of the chunk.
	Content string
	// PageNumber is the page number where the chunk originated.
	PageNumber int
	// ChunkIndex is the sequential index of this chunk within the document.
	ChunkIndex int
	// Embedding is the vector embedding of the chunk content.
	Embedding []float32
	// CreatedAt is the creation timestamp in RFC3339 format.
	CreatedAt string
}
