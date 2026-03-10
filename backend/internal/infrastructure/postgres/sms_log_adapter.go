package postgres

import (
	"context"
	"fmt"

	"phone-call-receptionist/backend/internal/domain/model"
	"phone-call-receptionist/backend/internal/domain/port"

	"github.com/rs/zerolog"
)

// smsLogAdapter implements port.SMSLog using PostgreSQL.
type smsLogAdapter struct {
	client *Client
	logger *zerolog.Logger
}

// NewSMSLogAdapter creates a new PostgreSQL SMS log adapter.
func NewSMSLogAdapter(client *Client, logger *zerolog.Logger) port.SMSLog {
	return &smsLogAdapter{client: client, logger: logger}
}

// Create persists a new SMS log entry to PostgreSQL.
func (a *smsLogAdapter) Create(ctx context.Context, log model.SMSLog) error {
	var db SMSLogDB
	db.FromDomain(log)

	query := `INSERT INTO sms_logs (id, call_id, to_phone, message, twilio_sid, status, created_at)
	           VALUES (:id, :call_id, :to_phone, :message, :twilio_sid, :status, :created_at)`

	_, err := a.client.DB.NamedExecContext(ctx, query, db)
	if err != nil {
		return fmt.Errorf("failed to create sms log: %w", err)
	}
	return nil
}

// FindByID retrieves an SMS log entry by its unique identifier from PostgreSQL.
func (a *smsLogAdapter) FindByID(ctx context.Context, id string) (model.SMSLog, error) {
	var db SMSLogDB
	query := `SELECT id, call_id, to_phone, message, twilio_sid, status, created_at
	           FROM sms_logs WHERE id = $1`

	if err := a.client.DB.GetContext(ctx, &db, query, id); err != nil {
		return model.SMSLog{}, fmt.Errorf("failed to find sms log by id: %w", err)
	}
	return db.ToDomain(), nil
}

// ListByCallID retrieves all SMS log entries for a given call from PostgreSQL.
func (a *smsLogAdapter) ListByCallID(ctx context.Context, callID string) ([]model.SMSLog, error) {
	query := `SELECT id, call_id, to_phone, message, twilio_sid, status, created_at
	           FROM sms_logs WHERE call_id = $1 ORDER BY created_at ASC`

	var rows []SMSLogDB
	if err := a.client.DB.SelectContext(ctx, &rows, query, callID); err != nil {
		return nil, fmt.Errorf("failed to list sms logs by call id: %w", err)
	}

	logs := make([]model.SMSLog, len(rows))
	for i, row := range rows {
		logs[i] = row.ToDomain()
	}
	return logs, nil
}
