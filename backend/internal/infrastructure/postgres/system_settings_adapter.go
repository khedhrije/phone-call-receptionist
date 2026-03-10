package postgres

import (
	"context"
	"fmt"

	"phone-call-receptionist/backend/internal/domain/model"
	"phone-call-receptionist/backend/internal/domain/port"

	"github.com/rs/zerolog"
)

// systemSettingsAdapter implements port.SystemSettings using PostgreSQL.
type systemSettingsAdapter struct {
	client *Client
	logger *zerolog.Logger
}

// NewSystemSettingsAdapter creates a new PostgreSQL system settings adapter.
func NewSystemSettingsAdapter(client *Client, logger *zerolog.Logger) port.SystemSettings {
	return &systemSettingsAdapter{client: client, logger: logger}
}

// Find retrieves the current system settings from PostgreSQL.
func (a *systemSettingsAdapter) Find(ctx context.Context) (model.SystemSettings, error) {
	var db SystemSettingsDB
	query := `SELECT id, default_llm_provider, default_voice_id, top_k, max_call_duration_secs, updated_at
	           FROM system_settings WHERE id = 1`

	if err := a.client.DB.GetContext(ctx, &db, query); err != nil {
		return model.SystemSettings{}, fmt.Errorf("failed to find system settings: %w", err)
	}
	return db.ToDomain(), nil
}

// Update modifies the system settings in PostgreSQL.
func (a *systemSettingsAdapter) Update(ctx context.Context, settings model.SystemSettings) error {
	var db SystemSettingsDB
	db.FromDomain(settings)

	query := `UPDATE system_settings SET default_llm_provider = :default_llm_provider,
	           default_voice_id = :default_voice_id, top_k = :top_k,
	           max_call_duration_secs = :max_call_duration_secs, updated_at = :updated_at
	           WHERE id = :id`

	_, err := a.client.DB.NamedExecContext(ctx, query, db)
	if err != nil {
		return fmt.Errorf("failed to update system settings: %w", err)
	}
	return nil
}
