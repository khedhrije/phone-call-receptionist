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
func (a *AuthApi) SignUp(ctx context.Context, email string, password string) (string, model.User, error) {
	a.logger.Info().Str("email", email).Msg("[AuthApi] SignUp started")

	_, err := a.userPort.FindByEmail(ctx, email)
	if err == nil {
		a.logger.Warn().Str("email", email).Msg("[AuthApi] SignUp failed: email already exists")
		return "", model.User{}, domainErrors.NewConflict("user with this email already exists")
	}

	hash, err := helpers.HashPassword(password)
	if err != nil {
		a.logger.Error().Err(err).Msg("[AuthApi] Failed to hash password")
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
		a.logger.Error().Err(err).Str("email", email).Msg("[AuthApi] Failed to create user")
		return "", model.User{}, fmt.Errorf("failed to create user: %w", err)
	}

	token, err := helpers.GenerateToken(user.ID, user.Email, user.Role, a.jwtSecret, a.expiryHours)
	if err != nil {
		a.logger.Error().Err(err).Msg("[AuthApi] Failed to generate token")
		return "", model.User{}, fmt.Errorf("failed to generate token: %w", err)
	}

	a.logger.Info().Str("userId", user.ID).Str("email", email).Msg("[AuthApi] User signed up successfully")
	return token, user, nil
}

// SignIn authenticates a user with the given email and password.
func (a *AuthApi) SignIn(ctx context.Context, email string, password string) (string, model.User, error) {
	a.logger.Info().Str("email", email).Msg("[AuthApi] SignIn started")

	user, err := a.userPort.FindByEmail(ctx, email)
	if err != nil {
		a.logger.Warn().Str("email", email).Msg("[AuthApi] SignIn failed: user not found")
		return "", model.User{}, domainErrors.NewValidation("email", "invalid email or password")
	}

	if user.IsBlocked {
		a.logger.Warn().Str("userId", user.ID).Str("email", email).Msg("[AuthApi] SignIn failed: account blocked")
		return "", model.User{}, domainErrors.NewForbidden("account is blocked")
	}

	if !helpers.CheckPassword(password, user.PasswordHash) {
		a.logger.Warn().Str("email", email).Msg("[AuthApi] SignIn failed: invalid password")
		return "", model.User{}, domainErrors.NewValidation("password", "invalid email or password")
	}

	token, err := helpers.GenerateToken(user.ID, user.Email, user.Role, a.jwtSecret, a.expiryHours)
	if err != nil {
		a.logger.Error().Err(err).Msg("[AuthApi] Failed to generate token")
		return "", model.User{}, fmt.Errorf("failed to generate token: %w", err)
	}

	a.logger.Info().Str("userId", user.ID).Str("email", email).Msg("[AuthApi] User signed in successfully")
	return token, user, nil
}

// RefreshToken validates the given token and issues a new one.
func (a *AuthApi) RefreshToken(ctx context.Context, tokenString string) (string, error) {
	a.logger.Debug().Msg("[AuthApi] RefreshToken started")

	claims, err := helpers.ValidateToken(tokenString, a.jwtSecret)
	if err != nil {
		a.logger.Warn().Msg("[AuthApi] RefreshToken failed: invalid token")
		return "", domainErrors.NewValidation("token", "invalid or expired token")
	}

	user, err := a.userPort.FindByID(ctx, claims.UserID)
	if err != nil {
		a.logger.Error().Err(err).Str("userId", claims.UserID).Msg("[AuthApi] RefreshToken failed: user not found")
		return "", fmt.Errorf("failed to find user: %w", err)
	}

	if user.IsBlocked {
		a.logger.Warn().Str("userId", user.ID).Msg("[AuthApi] RefreshToken failed: account blocked")
		return "", domainErrors.NewForbidden("account is blocked")
	}

	token, err := helpers.GenerateToken(user.ID, user.Email, user.Role, a.jwtSecret, a.expiryHours)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	a.logger.Debug().Str("userId", user.ID).Msg("[AuthApi] Token refreshed")
	return token, nil
}

// Me retrieves the current user's profile by their ID.
func (a *AuthApi) Me(ctx context.Context, userID string) (model.User, error) {
	a.logger.Debug().Str("userId", userID).Msg("[AuthApi] Fetching user profile")
	user, err := a.userPort.FindByID(ctx, userID)
	if err != nil {
		a.logger.Error().Err(err).Str("userId", userID).Msg("[AuthApi] Failed to find user")
		return model.User{}, fmt.Errorf("failed to find user: %w", err)
	}
	return user, nil
}

// ChangePassword updates the user's password after verifying the current one.
func (a *AuthApi) ChangePassword(ctx context.Context, userID string, currentPassword string, newPassword string) error {
	a.logger.Info().Str("userId", userID).Msg("[AuthApi] ChangePassword started")

	user, err := a.userPort.FindByID(ctx, userID)
	if err != nil {
		a.logger.Error().Err(err).Str("userId", userID).Msg("[AuthApi] Failed to find user for password change")
		return fmt.Errorf("failed to find user: %w", err)
	}

	if !helpers.CheckPassword(currentPassword, user.PasswordHash) {
		a.logger.Warn().Str("userId", userID).Msg("[AuthApi] ChangePassword failed: incorrect current password")
		return domainErrors.NewValidation("currentPassword", "current password is incorrect")
	}

	hash, err := helpers.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user.PasswordHash = hash
	user.UpdatedAt = time.Now().Format(time.RFC3339)

	if err := a.userPort.Update(ctx, user); err != nil {
		a.logger.Error().Err(err).Str("userId", userID).Msg("[AuthApi] Failed to update password")
		return fmt.Errorf("failed to update password: %w", err)
	}

	a.logger.Info().Str("userId", userID).Msg("[AuthApi] Password changed successfully")
	return nil
}
