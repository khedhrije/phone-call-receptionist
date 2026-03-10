// Package requests contains request DTOs with validation tags for API endpoints.
package requests

// SignUpRequest contains the fields required to create a new user account.
type SignUpRequest struct {
	// Email is the user's email address.
	Email string `json:"email" binding:"required,email"`
	// Password is the user's chosen password.
	Password string `json:"password" binding:"required,min=8"`
}

// SignInRequest contains the fields required to authenticate a user.
type SignInRequest struct {
	// Email is the user's email address.
	Email string `json:"email" binding:"required,email"`
	// Password is the user's password.
	Password string `json:"password" binding:"required"`
}

// ChangePasswordRequest contains the fields required to change a user's password.
type ChangePasswordRequest struct {
	// CurrentPassword is the user's current password.
	CurrentPassword string `json:"currentPassword" binding:"required"`
	// NewPassword is the user's desired new password.
	NewPassword string `json:"newPassword" binding:"required,min=8"`
}

// UpdateProfileRequest contains the fields that can be updated on a user profile.
type UpdateProfileRequest struct {
	// Email is the user's new email address.
	Email string `json:"email" binding:"omitempty,email"`
}
