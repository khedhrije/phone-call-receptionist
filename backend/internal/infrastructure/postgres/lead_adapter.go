package postgres

import (
	"context"
	"fmt"
	"strings"

	"phone-call-receptionist/backend/internal/domain/model"
	"phone-call-receptionist/backend/internal/domain/port"

	"github.com/rs/zerolog"
)

// leadAdapter implements port.Lead using PostgreSQL.
type leadAdapter struct {
	client *Client
	logger *zerolog.Logger
}

// NewLeadAdapter creates a new PostgreSQL lead adapter.
func NewLeadAdapter(client *Client, logger *zerolog.Logger) port.Lead {
	return &leadAdapter{client: client, logger: logger}
}

// Create persists a new lead to PostgreSQL.
func (a *leadAdapter) Create(ctx context.Context, lead model.Lead) error {
	var db LeadDB
	db.FromDomain(lead)

	query := `INSERT INTO leads (id, call_id, phone, name, email, status, notes, created_at, updated_at)
	           VALUES (:id, :call_id, :phone, :name, :email, :status, :notes, :created_at, :updated_at)`

	_, err := a.client.DB.NamedExecContext(ctx, query, db)
	if err != nil {
		return fmt.Errorf("failed to create lead: %w", err)
	}
	return nil
}

// FindByID retrieves a lead by its unique identifier from PostgreSQL.
func (a *leadAdapter) FindByID(ctx context.Context, id string) (model.Lead, error) {
	var db LeadDB
	query := `SELECT id, call_id, phone, name, email, status, notes, created_at, updated_at
	           FROM leads WHERE id = $1`

	if err := a.client.DB.GetContext(ctx, &db, query, id); err != nil {
		return model.Lead{}, fmt.Errorf("failed to find lead by id: %w", err)
	}
	return db.ToDomain(), nil
}

// FindByPhone retrieves a lead by their phone number from PostgreSQL.
func (a *leadAdapter) FindByPhone(ctx context.Context, phone string) (model.Lead, error) {
	var db LeadDB
	query := `SELECT id, call_id, phone, name, email, status, notes, created_at, updated_at
	           FROM leads WHERE phone = $1`

	if err := a.client.DB.GetContext(ctx, &db, query, phone); err != nil {
		return model.Lead{}, fmt.Errorf("failed to find lead by phone: %w", err)
	}
	return db.ToDomain(), nil
}

// List retrieves leads matching the given filters from PostgreSQL.
func (a *leadAdapter) List(ctx context.Context, filters port.LeadFilters) ([]model.Lead, int, error) {
	var conditions []string
	args := make(map[string]interface{})

	if filters.Status != "" {
		conditions = append(conditions, "status = :status")
		args["status"] = filters.Status
	}

	where := ""
	if len(conditions) > 0 {
		where = " WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total matching records.
	countQuery := "SELECT COUNT(*) FROM leads" + where
	countStmt, countArgs, err := a.client.DB.BindNamed(countQuery, args)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to bind count query: %w", err)
	}

	var total int
	if err := a.client.DB.GetContext(ctx, &total, countStmt, countArgs...); err != nil {
		return nil, 0, fmt.Errorf("failed to count leads: %w", err)
	}

	// Fetch paginated results.
	selectQuery := `SELECT id, call_id, phone, name, email, status, notes, created_at, updated_at
	                 FROM leads` + where + ` ORDER BY created_at DESC LIMIT :limit OFFSET :offset`

	args["limit"] = filters.Limit
	args["offset"] = filters.Offset

	selectStmt, selectArgs, err := a.client.DB.BindNamed(selectQuery, args)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to bind select query: %w", err)
	}

	var rows []LeadDB
	if err := a.client.DB.SelectContext(ctx, &rows, selectStmt, selectArgs...); err != nil {
		return nil, 0, fmt.Errorf("failed to list leads: %w", err)
	}

	leads := make([]model.Lead, len(rows))
	for i, row := range rows {
		leads[i] = row.ToDomain()
	}
	return leads, total, nil
}

// Update modifies an existing lead's data in PostgreSQL.
func (a *leadAdapter) Update(ctx context.Context, lead model.Lead) error {
	var db LeadDB
	db.FromDomain(lead)

	query := `UPDATE leads SET call_id = :call_id, phone = :phone, name = :name, email = :email,
	           status = :status, notes = :notes, updated_at = :updated_at WHERE id = :id`

	_, err := a.client.DB.NamedExecContext(ctx, query, db)
	if err != nil {
		return fmt.Errorf("failed to update lead: %w", err)
	}
	return nil
}
