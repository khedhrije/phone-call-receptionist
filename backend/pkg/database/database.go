// Package database provides PostgreSQL connection management and migration utilities.
package database

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// Connect establishes a connection to the PostgreSQL database.
func Connect(host string, port string, name string, user string, password string, sslMode string) (*sqlx.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
		host, port, name, user, password, sslMode,
	)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	return db, nil
}
