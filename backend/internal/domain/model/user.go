// Package model contains pure domain entities without any infrastructure dependencies.
package model

// User represents a user account in the system.
type User struct {
	// ID is the unique identifier for the user.
	ID string
	// Email is the user's email address used for authentication.
	Email string
	// PasswordHash is the bcrypt-hashed password.
	PasswordHash string
	// Role defines the user's permission level (super_admin, admin, user).
	Role string
	// IsBlocked indicates whether the user account is blocked.
	IsBlocked bool
	// CreatedAt is the creation timestamp in RFC3339 format.
	CreatedAt string
	// UpdatedAt is the last modification timestamp in RFC3339 format.
	UpdatedAt string
}
