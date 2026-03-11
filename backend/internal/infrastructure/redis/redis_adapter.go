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
	a.logger.Debug().Str("key", key).Msg("[RedisAdapter] finding cache entry")

	val, err := a.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			a.logger.Debug().Str("key", key).Msg("[RedisAdapter] cache miss")
			return nil, errors.NewNotFound("cache", key)
		}
		a.logger.Error().Err(err).Str("key", key).Msg("[RedisAdapter] failed to find cache entry")
		return nil, fmt.Errorf("failed to find cache key %s: %w", key, err)
	}

	a.logger.Debug().Str("key", key).Int("bytes", len(val)).Msg("[RedisAdapter] cache hit")
	return val, nil
}

// Create persists a value in the cache with the given TTL.
func (a *Adapter) Create(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	a.logger.Debug().Str("key", key).Int("bytes", len(value)).Dur("ttl", ttl).Msg("[RedisAdapter] storing cache entry")

	err := a.client.Set(ctx, key, value, ttl).Err()
	if err != nil {
		a.logger.Error().Err(err).Str("key", key).Msg("[RedisAdapter] failed to store cache entry")
		return fmt.Errorf("failed to store cache key %s: %w", key, err)
	}

	a.logger.Debug().Str("key", key).Dur("ttl", ttl).Msg("[RedisAdapter] cache entry stored")
	return nil
}

// Delete removes a cached value by its key.
func (a *Adapter) Delete(ctx context.Context, key string) error {
	a.logger.Debug().Str("key", key).Msg("[RedisAdapter] deleting cache entry")

	err := a.client.Del(ctx, key).Err()
	if err != nil {
		a.logger.Error().Err(err).Str("key", key).Msg("[RedisAdapter] failed to delete cache entry")
		return fmt.Errorf("failed to delete cache key %s: %w", key, err)
	}

	a.logger.Debug().Str("key", key).Msg("[RedisAdapter] cache entry deleted")
	return nil
}
