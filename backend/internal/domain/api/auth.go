// Package api provides business logic services that orchestrate domain operations.
package api

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	domainErrors "phone-call-receptionist/backend/internal/domain/errors"
	"phone-call-receptionist/backend/internal/domain/model"
	"phone-call-receptionist/backend/internal/domain/port"
	"phone-call-receptionist/backend/pkg/helpers"
)

// AuthApi provides business operations for user authentication.
type AuthApi struct {
	userPort    port.User
	jwtSecret   string
	expiryHours int
	logger      *zerolog.Logger
}

// NewAuthApi creates a new AuthApi with the given dependencies.
func NewAuthApi(userPort port.User, jwtSecret string, expiryHours int, logger *zerolog.Logger) *AuthApi {
	return &AuthApi{
		userPort:    userPort,
		jwtSecret:   jwtSecret,
		expiryHours: expiryHours,
		logger:      logger,
	}
}

// SignUp creates a new user account with the given email and password.
// Returns the JWT token and user model.
func (a *AuthApi) SignUp(ctx context.Context, email string, password string) (string, model.User, error) {
	_, err := a.userPort.FindByEmail(ctx, email)
	if err == nil {
		return "", model.User{}, domainErrors.NewConflict("user with this email already exists")
	}

	hash, err := helpers.HashPassword(password)
	if err != nil {
		return "", model.User{}, fmt.Errorf("failed to hash password: %w", err)
	}

	now := time.Now().Format(time.RFC3339)
	user := model.User{
		ID:           uuid.New().String(),
		Email:        email,
		PasswordHash: hash,
		Role:         "user",
		IsBlocked:    false,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := a.userPort.Create(ctx, user); err != nil {
		return "", model.User{}, fmt.Errorf("failed to create user: %w", err)
	}

	token, err := helpers.GenerateToken(user.ID, user.Email, user.Role, a.jwtSecret, a.expiryHours)
	if err != nil {
		return "", model.User{}, fmt.Errorf("failed to generate token: %w", err)
	}

	a.logger.Info().Str("userId", user.ID).Msg("User signed up")
	return token, user, nil
}

// SignIn authenticates a user with the given email and password.
// Returns the JWT token and user model.
func (a *AuthApi) SignIn(ctx context.Context, email string, password string) (string, model.User, error) {
	user, err := a.userPort.FindByEmail(ctx, email)
	if err != nil {
		return "", model.User{}, domainErrors.NewValidation("email", "invalid email or password")
	}

	if user.IsBlocked {
		return "", model.User{}, domainErrors.NewForbidden("account is blocked")
	}

	if !helpers.CheckPassword(password, user.PasswordHash) {
		return "", model.User{}, domainErrors.NewValidation("password", "invalid email or password")
	}

	token, err := helpers.GenerateToken(user.ID, user.Email, user.Role, a.jwtSecret, a.expiryHours)
	if err != nil {
		return "", model.User{}, fmt.Errorf("failed to generate token: %w", err)
	}

	a.logger.Info().Str("userId", user.ID).Msg("User signed in")
	return token, user, nil
}

// RefreshToken validates the given token and issues a new one.
func (a *AuthApi) RefreshToken(ctx context.Context, tokenString string) (string, error) {
	claims, err := helpers.ValidateToken(tokenString, a.jwtSecret)
	if err != nil {
		return "", domainErrors.NewValidation("token", "invalid or expired token")
	}

	user, err := a.userPort.FindByID(ctx, claims.UserID)
	if err != nil {
		return "", fmt.Errorf("failed to find user: %w", err)
	}

	if user.IsBlocked {
		return "", domainErrors.NewForbidden("account is blocked")
	}

	token, err := helpers.GenerateToken(user.ID, user.Email, user.Role, a.jwtSecret, a.expiryHours)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return token, nil
}

// Me retrieves the current user's profile by their ID.
func (a *AuthApi) Me(ctx context.Context, userID string) (model.User, error) {
	user, err := a.userPort.FindByID(ctx, userID)
	if err != nil {
		return model.User{}, fmt.Errorf("failed to find user: %w", err)
	}
	return user, nil
}

// ChangePassword updates the user's password after verifying the current one.
func (a *AuthApi) ChangePassword(ctx context.Context, userID string, currentPassword string, newPassword string) error {
	user, err := a.userPort.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}

	if !helpers.CheckPassword(currentPassword, user.PasswordHash) {
		return domainErrors.NewValidation("currentPassword", "current password is incorrect")
	}

	hash, err := helpers.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user.PasswordHash = hash
	user.UpdatedAt = time.Now().Format(time.RFC3339)

	if err := a.userPort.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	a.logger.Info().Str("userId", userID).Msg("Password changed")
	return nil
}
