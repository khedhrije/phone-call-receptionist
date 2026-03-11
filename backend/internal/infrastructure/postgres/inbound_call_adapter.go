package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"phone-call-receptionist/backend/internal/domain/model"
	"phone-call-receptionist/backend/internal/domain/port"

	"github.com/rs/zerolog"
)

// inboundCallAdapter implements port.InboundCall using PostgreSQL.
type inboundCallAdapter struct {
	client *Client
	logger *zerolog.Logger
}

// NewInboundCallAdapter creates a new PostgreSQL inbound call adapter.
func NewInboundCallAdapter(client *Client, logger *zerolog.Logger) port.InboundCall {
	return &inboundCallAdapter{client: client, logger: logger}
}

// Create persists a new inbound call record to PostgreSQL.
func (a *inboundCallAdapter) Create(ctx context.Context, call model.InboundCall) error {
	a.logger.Debug().Str("id", call.ID).Str("callerPhone", call.CallerPhone).Msg("[PostgresInboundCall] creating inbound call")

	var db InboundCallDB
	if err := db.FromDomain(call); err != nil {
		a.logger.Error().Err(err).Str("id", call.ID).Msg("[PostgresInboundCall] failed to convert to db entity")
		return fmt.Errorf("failed to convert inbound call to db entity: %w", err)
	}

	query := `INSERT INTO inbound_calls (id, twilio_call_sid, caller_phone, status, transcript, rag_queries,
	           duration_seconds, twilio_cost_usd, llm_cost_usd, total_cost_usd, created_at, ended_at)
	           VALUES (:id, :twilio_call_sid, :caller_phone, :status, :transcript, :rag_queries,
	           :duration_seconds, :twilio_cost_usd, :llm_cost_usd, :total_cost_usd, :created_at, :ended_at)`

	_, err := a.client.DB.NamedExecContext(ctx, query, db)
	if err != nil {
		a.logger.Error().Err(err).Str("id", call.ID).Msg("[PostgresInboundCall] failed to create inbound call")
		return fmt.Errorf("failed to create inbound call: %w", err)
	}

	a.logger.Debug().Str("id", call.ID).Msg("[PostgresInboundCall] inbound call created")
	return nil
}

// FindByID retrieves an inbound call by its unique identifier from PostgreSQL.
func (a *inboundCallAdapter) FindByID(ctx context.Context, id string) (model.InboundCall, error) {
	a.logger.Debug().Str("id", id).Msg("[PostgresInboundCall] finding inbound call by ID")

	var db InboundCallDB
	query := `SELECT id, twilio_call_sid, caller_phone, status, transcript, rag_queries,
	           duration_seconds, twilio_cost_usd, llm_cost_usd, total_cost_usd, created_at, ended_at
	           FROM inbound_calls WHERE id = $1`

	if err := a.client.DB.GetContext(ctx, &db, query, id); err != nil {
		a.logger.Error().Err(err).Str("id", id).Msg("[PostgresInboundCall] failed to find inbound call by ID")
		return model.InboundCall{}, fmt.Errorf("failed to find inbound call by id: %w", err)
	}

	call, err := db.ToDomain()
	if err != nil {
		a.logger.Error().Err(err).Str("id", id).Msg("[PostgresInboundCall] failed to convert to domain")
		return model.InboundCall{}, fmt.Errorf("failed to convert inbound call to domain: %w", err)
	}

	a.logger.Debug().Str("id", id).Msg("[PostgresInboundCall] inbound call found")
	return call, nil
}

// FindByTwilioSID retrieves an inbound call by its Twilio call SID from PostgreSQL.
func (a *inboundCallAdapter) FindByTwilioSID(ctx context.Context, sid string) (model.InboundCall, error) {
	a.logger.Debug().Str("sid", sid).Msg("[PostgresInboundCall] finding inbound call by Twilio SID")

	var db InboundCallDB
	query := `SELECT id, twilio_call_sid, caller_phone, status, transcript, rag_queries,
	           duration_seconds, twilio_cost_usd, llm_cost_usd, total_cost_usd, created_at, ended_at
	           FROM inbound_calls WHERE twilio_call_sid = $1`

	if err := a.client.DB.GetContext(ctx, &db, query, sid); err != nil {
		a.logger.Error().Err(err).Str("sid", sid).Msg("[PostgresInboundCall] failed to find inbound call by Twilio SID")
		return model.InboundCall{}, fmt.Errorf("failed to find inbound call by twilio sid: %w", err)
	}

	call, err := db.ToDomain()
	if err != nil {
		a.logger.Error().Err(err).Str("sid", sid).Msg("[PostgresInboundCall] failed to convert to domain")
		return model.InboundCall{}, fmt.Errorf("failed to convert inbound call to domain: %w", err)
	}

	a.logger.Debug().Str("sid", sid).Str("id", call.ID).Msg("[PostgresInboundCall] inbound call found by Twilio SID")
	return call, nil
}

// List retrieves inbound calls matching the given filters from PostgreSQL.
func (a *inboundCallAdapter) List(ctx context.Context, filters port.CallFilters) ([]model.InboundCall, int, error) {
	a.logger.Debug().Int("limit", filters.Limit).Int("offset", filters.Offset).Msg("[PostgresInboundCall] listing inbound calls")
	var conditions []string
	args := make(map[string]interface{})

	if filters.Status != "" {
		conditions = append(conditions, "status = :status")
		args["status"] = filters.Status
	}
	if filters.CallerPhone != "" {
		conditions = append(conditions, "caller_phone = :caller_phone")
		args["caller_phone"] = filters.CallerPhone
	}
	if filters.From != "" {
		if t, err := time.Parse(time.RFC3339, filters.From); err == nil {
			conditions = append(conditions, "created_at >= :from_time")
			args["from_time"] = t
		}
	}
	if filters.To != "" {
		if t, err := time.Parse(time.RFC3339, filters.To); err == nil {
			conditions = append(conditions, "created_at <= :to_time")
			args["to_time"] = t
		}
	}

	where := ""
	if len(conditions) > 0 {
		where = " WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total matching records.
	countQuery := "SELECT COUNT(*) FROM inbound_calls" + where
	countStmt, countArgs, err := a.client.DB.BindNamed(countQuery, args)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to bind count query: %w", err)
	}

	var total int
	if err := a.client.DB.GetContext(ctx, &total, countStmt, countArgs...); err != nil {
		a.logger.Error().Err(err).Msg("[PostgresInboundCall] failed to count inbound calls")
		return nil, 0, fmt.Errorf("failed to count inbound calls: %w", err)
	}

	// Fetch paginated results.
	selectQuery := `SELECT id, twilio_call_sid, caller_phone, status, transcript, rag_queries,
	                 duration_seconds, twilio_cost_usd, llm_cost_usd, total_cost_usd, created_at, ended_at
	                 FROM inbound_calls` + where + ` ORDER BY created_at DESC LIMIT :limit OFFSET :offset`

	args["limit"] = filters.Limit
	args["offset"] = filters.Offset

	selectStmt, selectArgs, err := a.client.DB.BindNamed(selectQuery, args)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to bind select query: %w", err)
	}

	var rows []InboundCallDB
	if err := a.client.DB.SelectContext(ctx, &rows, selectStmt, selectArgs...); err != nil {
		a.logger.Error().Err(err).Msg("[PostgresInboundCall] failed to list inbound calls")
		return nil, 0, fmt.Errorf("failed to list inbound calls: %w", err)
	}

	calls := make([]model.InboundCall, len(rows))
	for i, row := range rows {
		call, err := row.ToDomain()
		if err != nil {
			a.logger.Error().Err(err).Msg("[PostgresInboundCall] failed to convert to domain")
			return nil, 0, fmt.Errorf("failed to convert inbound call to domain: %w", err)
		}
		calls[i] = call
	}

	a.logger.Debug().Int("count", len(calls)).Int("total", total).Msg("[PostgresInboundCall] inbound calls listed")
	return calls, total, nil
}

// Update modifies an existing inbound call's data in PostgreSQL.
func (a *inboundCallAdapter) Update(ctx context.Context, call model.InboundCall) error {
	a.logger.Debug().Str("id", call.ID).Str("status", call.Status).Msg("[PostgresInboundCall] updating inbound call")

	var db InboundCallDB
	if err := db.FromDomain(call); err != nil {
		a.logger.Error().Err(err).Str("id", call.ID).Msg("[PostgresInboundCall] failed to convert to db entity")
		return fmt.Errorf("failed to convert inbound call to db entity: %w", err)
	}

	query := `UPDATE inbound_calls SET twilio_call_sid = :twilio_call_sid, caller_phone = :caller_phone,
	           status = :status, transcript = :transcript, rag_queries = :rag_queries,
	           duration_seconds = :duration_seconds, twilio_cost_usd = :twilio_cost_usd,
	           llm_cost_usd = :llm_cost_usd, total_cost_usd = :total_cost_usd,
	           ended_at = :ended_at WHERE id = :id`

	_, err := a.client.DB.NamedExecContext(ctx, query, db)
	if err != nil {
		a.logger.Error().Err(err).Str("id", call.ID).Msg("[PostgresInboundCall] failed to update inbound call")
		return fmt.Errorf("failed to update inbound call: %w", err)
	}

	a.logger.Debug().Str("id", call.ID).Msg("[PostgresInboundCall] inbound call updated")
	return nil
}
