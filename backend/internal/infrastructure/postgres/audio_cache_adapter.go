package postgres

import (
	"context"
	"fmt"

	"phone-call-receptionist/backend/internal/domain/model"
	"phone-call-receptionist/backend/internal/domain/port"

	"github.com/rs/zerolog"
)

// audioCacheAdapter implements port.AudioCache using PostgreSQL.
type audioCacheAdapter struct {
	client *Client
	logger *zerolog.Logger
}

// NewAudioCacheAdapter creates a new PostgreSQL audio cache adapter.
func NewAudioCacheAdapter(client *Client, logger *zerolog.Logger) port.AudioCache {
	return &audioCacheAdapter{client: client, logger: logger}
}

// Create persists a new audio cache entry to PostgreSQL.
func (a *audioCacheAdapter) Create(ctx context.Context, cache model.AudioCache) error {
	a.logger.Debug().Str("id", cache.ID).Str("hash", cache.Hash).Msg("[PostgresAudioCache] creating audio cache entry")

	var db AudioCacheDB
	db.FromDomain(cache)

	query := `INSERT INTO audio_cache (id, hash, voice_id, file_path, created_at)
	           VALUES (:id, :hash, :voice_id, :file_path, :created_at)`

	_, err := a.client.DB.NamedExecContext(ctx, query, db)
	if err != nil {
		a.logger.Error().Err(err).Str("id", cache.ID).Str("hash", cache.Hash).Msg("[PostgresAudioCache] failed to create audio cache entry")
		return fmt.Errorf("failed to create audio cache entry: %w", err)
	}

	a.logger.Debug().Str("id", cache.ID).Msg("[PostgresAudioCache] audio cache entry created")
	return nil
}

// FindByHash retrieves an audio cache entry by its content hash from PostgreSQL.
func (a *audioCacheAdapter) FindByHash(ctx context.Context, hash string) (model.AudioCache, error) {
	a.logger.Debug().Str("hash", hash).Msg("[PostgresAudioCache] finding audio cache by hash")

	var db AudioCacheDB
	query := `SELECT id, hash, voice_id, file_path, created_at
	           FROM audio_cache WHERE hash = $1`

	if err := a.client.DB.GetContext(ctx, &db, query, hash); err != nil {
		a.logger.Error().Err(err).Str("hash", hash).Msg("[PostgresAudioCache] failed to find audio cache by hash")
		return model.AudioCache{}, fmt.Errorf("failed to find audio cache by hash: %w", err)
	}

	a.logger.Debug().Str("hash", hash).Str("id", db.ID).Msg("[PostgresAudioCache] audio cache entry found")
	return db.ToDomain(), nil
}

// Delete removes an audio cache entry from PostgreSQL.
func (a *audioCacheAdapter) Delete(ctx context.Context, id string) error {
	a.logger.Debug().Str("id", id).Msg("[PostgresAudioCache] deleting audio cache entry")

	query := `DELETE FROM audio_cache WHERE id = $1`

	_, err := a.client.DB.ExecContext(ctx, query, id)
	if err != nil {
		a.logger.Error().Err(err).Str("id", id).Msg("[PostgresAudioCache] failed to delete audio cache entry")
		return fmt.Errorf("failed to delete audio cache entry: %w", err)
	}

	a.logger.Debug().Str("id", id).Msg("[PostgresAudioCache] audio cache entry deleted")
	return nil
}
