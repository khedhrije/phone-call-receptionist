package port

import "context"

// FileStorage defines the interface for file storage operations.
// Implementations handle saving and loading files from a storage backend.
type FileStorage interface {
	// Create persists file data to storage and returns the storage path.
	Create(ctx context.Context, filename string, data []byte) (string, error)
	// Find retrieves file data from storage by filename.
	Find(ctx context.Context, filename string) ([]byte, error)
	// Delete removes a file from storage by filename.
	Delete(ctx context.Context, filename string) error
}
