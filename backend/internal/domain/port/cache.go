package port

import (
	"context"
	"time"
)

// Cache defines the interface for key-value cache operations.
// Implementations provide temporary storage with time-to-live support.
type Cache interface {
	// Find retrieves a cached value by its key.
	// Returns the value bytes and any error. Returns an error if not found.
	Find(ctx context.Context, key string) ([]byte, error)
	// Create persists a value in the cache with the given TTL.
	Create(ctx context.Context, key string, value []byte, ttl time.Duration) error
	// Delete removes a cached value by its key.
	Delete(ctx context.Context, key string) error
}
