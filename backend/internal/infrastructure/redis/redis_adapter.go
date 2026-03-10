// Package redis implements the Cache port using Redis.
package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	"phone-call-receptionist/backend/internal/domain/errors"
	"phone-call-receptionist/backend/internal/domain/port"
)

// Adapter implements port.Cache using Redis.
type Adapter struct {
	client *redis.Client
	logger *zerolog.Logger
}

// NewRedisAdapter creates a new Redis cache adapter.
func NewRedisAdapter(client *redis.Client, logger *zerolog.Logger) port.Cache {
	return &Adapter{
		client: client,
		logger: logger,
	}
}

// Find retrieves a cached value by its key.
// Returns a NotFoundError if the key does not exist.
func (a *Adapter) Find(ctx context.Context, key string) ([]byte, error) {
	val, err := a.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, errors.NewNotFound("cache", key)
		}
		return nil, fmt.Errorf("failed to find cache key %s: %w", key, err)
	}

	a.logger.Debug().Str("key", key).Msg("Cache hit")
	return val, nil
}

// Store persists a value in the cache with the given TTL.
func (a *Adapter) Store(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	err := a.client.Set(ctx, key, value, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to store cache key %s: %w", key, err)
	}

	a.logger.Debug().Str("key", key).Dur("ttl", ttl).Msg("Cache stored")
	return nil
}

// Delete removes a cached value by its key.
func (a *Adapter) Delete(ctx context.Context, key string) error {
	err := a.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete cache key %s: %w", key, err)
	}

	a.logger.Debug().Str("key", key).Msg("Cache deleted")
	return nil
}
