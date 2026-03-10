package postgres

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
)

// Client manages the PostgreSQL database connection pool.
type Client struct {
	// DB is the sqlx database connection.
	DB *sqlx.DB
	logger *zerolog.Logger
}

// NewClient creates a new PostgreSQL client with the given connection parameters.
func NewClient(host, port, name, user, password, sslMode string, logger *zerolog.Logger) (*Client, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
		host, port, name, user, password, sslMode,
	)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	logger.Info().Msg("Connected to PostgreSQL")

	return &Client{DB: db, logger: logger}, nil
}

// Close closes the database connection.
func (c *Client) Close() error {
	return c.DB.Close()
}
