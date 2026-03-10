// Package responses contains response DTOs with JSON tags for API endpoints.
package responses

// AuthResponse contains the authentication token returned after sign-in or sign-up.
type AuthResponse struct {
	// Token is the JWT authentication token.
	Token string `json:"token"`
	// User is the authenticated user's profile.
	User UserResponse `json:"user"`
}

// UserResponse contains the user profile data returned by the API.
type UserResponse struct {
	// ID is the user's unique identifier.
	ID string `json:"id"`
	// Email is the user's email address.
	Email string `json:"email"`
	// Role is the user's permission role.
	Role string `json:"role"`
	// IsBlocked indicates whether the user account is blocked.
	IsBlocked bool `json:"isBlocked"`
	// CreatedAt is the account creation timestamp in RFC3339 format.
	CreatedAt string `json:"createdAt"`
	// UpdatedAt is the last modification timestamp in RFC3339 format.
	UpdatedAt string `json:"updatedAt"`
}
