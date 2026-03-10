// Package helpers provides shared utility functions for context management,
// JWT handling, and password hashing.
package helpers

import "context"

type contextKey string

const (
	requestIDKey contextKey = "requestID"
	userIDKey    contextKey = "userID"
	userEmailKey contextKey = "userEmail"
	userRoleKey  contextKey = "userRole"
)

// UserContext holds the authenticated user's information extracted from context.
type UserContext struct {
	// UserID is the authenticated user's unique identifier.
	UserID string
	// Email is the authenticated user's email address.
	Email string
	// Role is the authenticated user's role.
	Role string
}

// InjectRequestID stores a request ID in the given context.
func InjectRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// ExtractRequestID retrieves the request ID from the given context.
func ExtractRequestID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(requestIDKey).(string)
	return id, ok
}

// InjectUser stores user information in the given context.
func InjectUser(ctx context.Context, userID string, email string, role string) context.Context {
	ctx = context.WithValue(ctx, userIDKey, userID)
	ctx = context.WithValue(ctx, userEmailKey, email)
	ctx = context.WithValue(ctx, userRoleKey, role)
	return ctx
}

// ExtractUser retrieves the authenticated user's information from the given context.
func ExtractUser(ctx context.Context) (UserContext, bool) {
	userID, ok := ctx.Value(userIDKey).(string)
	if !ok {
		return UserContext{}, false
	}
	email, _ := ctx.Value(userEmailKey).(string)
	role, _ := ctx.Value(userRoleKey).(string)
	return UserContext{UserID: userID, Email: email, Role: role}, true
}
