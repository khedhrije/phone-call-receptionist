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
	var db AudioCacheDB
	db.FromDomain(cache)

	query := `INSERT INTO audio_cache (id, hash, voice_id, file_path, created_at)
	           VALUES (:id, :hash, :voice_id, :file_path, :created_at)`

	_, err := a.client.DB.NamedExecContext(ctx, query, db)
	if err != nil {
		return fmt.Errorf("failed to create audio cache entry: %w", err)
	}
	return nil
}

// FindByHash retrieves an audio cache entry by its content hash from PostgreSQL.
func (a *audioCacheAdapter) FindByHash(ctx context.Context, hash string) (model.AudioCache, error) {
	var db AudioCacheDB
	query := `SELECT id, hash, voice_id, file_path, created_at
	           FROM audio_cache WHERE hash = $1`

	if err := a.client.DB.GetContext(ctx, &db, query, hash); err != nil {
		return model.AudioCache{}, fmt.Errorf("failed to find audio cache by hash: %w", err)
	}
	return db.ToDomain(), nil
}

// Delete removes an audio cache entry from PostgreSQL.
func (a *audioCacheAdapter) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM audio_cache WHERE id = $1`

	_, err := a.client.DB.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete audio cache entry: %w", err)
	}
	return nil
}
