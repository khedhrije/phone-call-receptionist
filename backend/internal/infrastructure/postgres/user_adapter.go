package postgres

import (
	"context"
	"fmt"
	"strings"

	"phone-call-receptionist/backend/internal/domain/model"
	"phone-call-receptionist/backend/internal/domain/port"

	"github.com/rs/zerolog"
)

// userAdapter implements port.User using PostgreSQL.
type userAdapter struct {
	client *Client
	logger *zerolog.Logger
}

// NewUserAdapter creates a new PostgreSQL user adapter.
func NewUserAdapter(client *Client, logger *zerolog.Logger) port.User {
	return &userAdapter{client: client, logger: logger}
}

// Create persists a new user to PostgreSQL.
func (a *userAdapter) Create(ctx context.Context, user model.User) error {
	var db UserDB
	db.FromDomain(user)

	query := `INSERT INTO users (id, email, password_hash, role, is_blocked, created_at, updated_at)
	           VALUES (:id, :email, :password_hash, :role, :is_blocked, :created_at, :updated_at)`

	_, err := a.client.DB.NamedExecContext(ctx, query, db)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// FindByID retrieves a user by their unique identifier from PostgreSQL.
func (a *userAdapter) FindByID(ctx context.Context, id string) (model.User, error) {
	var db UserDB
	query := `SELECT id, email, password_hash, role, is_blocked, created_at, updated_at
	           FROM users WHERE id = $1`

	if err := a.client.DB.GetContext(ctx, &db, query, id); err != nil {
		return model.User{}, fmt.Errorf("failed to find user by id: %w", err)
	}
	return db.ToDomain(), nil
}

// FindByEmail retrieves a user by their email address from PostgreSQL.
func (a *userAdapter) FindByEmail(ctx context.Context, email string) (model.User, error) {
	var db UserDB
	query := `SELECT id, email, password_hash, role, is_blocked, created_at, updated_at
	           FROM users WHERE email = $1`

	if err := a.client.DB.GetContext(ctx, &db, query, email); err != nil {
		return model.User{}, fmt.Errorf("failed to find user by email: %w", err)
	}
	return db.ToDomain(), nil
}

// List retrieves users matching the given filters from PostgreSQL.
func (a *userAdapter) List(ctx context.Context, filters port.UserFilters) ([]model.User, int, error) {
	var conditions []string
	args := make(map[string]interface{})

	if filters.Role != "" {
		conditions = append(conditions, "role = :role")
		args["role"] = filters.Role
	}
	if filters.IsBlocked != nil {
		conditions = append(conditions, "is_blocked = :is_blocked")
		args["is_blocked"] = *filters.IsBlocked
	}

	where := ""
	if len(conditions) > 0 {
		where = " WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total matching records.
	countQuery := "SELECT COUNT(*) FROM users" + where
	countStmt, countArgs, err := a.client.DB.BindNamed(countQuery, args)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to bind count query: %w", err)
	}

	var total int
	if err := a.client.DB.GetContext(ctx, &total, countStmt, countArgs...); err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Fetch paginated results.
	selectQuery := `SELECT id, email, password_hash, role, is_blocked, created_at, updated_at
	                 FROM users` + where + ` ORDER BY created_at DESC LIMIT :limit OFFSET :offset`

	args["limit"] = filters.Limit
	args["offset"] = filters.Offset

	selectStmt, selectArgs, err := a.client.DB.BindNamed(selectQuery, args)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to bind select query: %w", err)
	}

	var rows []UserDB
	if err := a.client.DB.SelectContext(ctx, &rows, selectStmt, selectArgs...); err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	users := make([]model.User, len(rows))
	for i, row := range rows {
		users[i] = row.ToDomain()
	}
	return users, total, nil
}

// Update modifies an existing user's data in PostgreSQL.
func (a *userAdapter) Update(ctx context.Context, user model.User) error {
	var db UserDB
	db.FromDomain(user)

	query := `UPDATE users SET email = :email, password_hash = :password_hash, role = :role,
	           is_blocked = :is_blocked, updated_at = :updated_at WHERE id = :id`

	_, err := a.client.DB.NamedExecContext(ctx, query, db)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

// Delete removes a user from PostgreSQL.
func (a *userAdapter) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`

	_, err := a.client.DB.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}
