package port

import "context"

// Embedding defines the interface for text embedding generation.
// Implementations convert text into vector representations for similarity search.
type Embedding interface {
	// Embed converts the given text into a vector embedding.
	Embed(ctx context.Context, text string) ([]float32, error)
}
