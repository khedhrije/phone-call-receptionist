package port

import (
	"context"

	"phone-call-receptionist/backend/internal/domain/model"
)

// VectorDB defines the interface for vector database operations.
// Implementations handle storage and similarity search of text chunk embeddings.
type VectorDB interface {
	// Create persists a chunk with its embedding to the vector database.
	Create(ctx context.Context, chunk model.Chunk) error
	// Search finds the most similar chunks to the given embedding vector.
	Search(ctx context.Context, embedding []float32, topK int) ([]model.Chunk, error)
	// DeleteByDocumentID removes all chunks belonging to a specific document.
	DeleteByDocumentID(ctx context.Context, documentID string) error
}
