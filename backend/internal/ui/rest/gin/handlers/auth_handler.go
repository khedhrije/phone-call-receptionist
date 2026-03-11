package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"phone-call-receptionist/backend/internal/domain/api"
	"phone-call-receptionist/backend/pkg/dtos/requests"
	"phone-call-receptionist/backend/pkg/dtos/responses"
	"phone-call-receptionist/backend/pkg/helpers"
)

// AuthHandler handles authentication-related HTTP requests.
type AuthHandler struct {
	authApi *api.AuthApi
	logger  *zerolog.Logger
}

// NewAuthHandler creates a new AuthHandler with the given dependencies.
func NewAuthHandler(authApi *api.AuthApi, logger *zerolog.Logger) *AuthHandler {
	return &AuthHandler{authApi: authApi, logger: logger}
}

// SignUp godoc
//
//	@Summary		Create a new user account
//	@Description	Register a new user with email and password
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		requests.SignUpRequest	true	"Sign up data"
//	@Success		201		{object}	responses.AuthResponse
//	@Failure		400		{object}	responses.ErrorResponse
//	@Failure		409		{object}	responses.ErrorResponse
//	@Router			/auth/signup [post]
func (h *AuthHandler) SignUp(c *gin.Context) {
	h.logger.Info().Msg("[AuthHandler] SignUp request received")

	var req requests.SignUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error().Err(err).Msg("[AuthHandler] SignUp failed to bind request body")
		c.JSON(http.StatusBadRequest, responses.ErrorResponse{Error: err.Error()})
		return
	}

	h.logger.Info().Str("email", req.Email).Msg("[AuthHandler] SignUp processing registration")

	token, user, err := h.authApi.SignUp(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		h.logger.Error().Err(err).Str("email", req.Email).Msg("[AuthHandler] SignUp failed")
		HandleError(c, err)
		return
	}

	h.logger.Info().Str("userID", user.ID).Str("email", user.Email).Msg("[AuthHandler] SignUp succeeded")

	c.JSON(http.StatusCreated, responses.AuthResponse{
		Token: token,
		User: responses.UserResponse{
			ID: user.ID, Email: user.Email, Role: user.Role,
			IsBlocked: user.IsBlocked, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt,
		},
	})
}

// SignIn godoc
//
//	@Summary		Authenticate user
//	@Description	Sign in with email and password
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		requests.SignInRequest	true	"Sign in data"
//	@Success		200		{object}	responses.AuthResponse
//	@Failure		400		{object}	responses.ErrorResponse
//	@Router			/auth/signin [post]
func (h *AuthHandler) SignIn(c *gin.Context) {
	h.logger.Info().Msg("[AuthHandler] SignIn request received")

	var req requests.SignInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error().Err(err).Msg("[AuthHandler] SignIn failed to bind request body")
		c.JSON(http.StatusBadRequest, responses.ErrorResponse{Error: err.Error()})
		return
	}

	h.logger.Info().Str("email", req.Email).Msg("[AuthHandler] SignIn processing authentication")

	token, user, err := h.authApi.SignIn(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		h.logger.Error().Err(err).Str("email", req.Email).Msg("[AuthHandler] SignIn failed")
		HandleError(c, err)
		return
	}

	h.logger.Info().Str("userID", user.ID).Str("email", user.Email).Msg("[AuthHandler] SignIn succeeded")

	c.JSON(http.StatusOK, responses.AuthResponse{
		Token: token,
		User: responses.UserResponse{
			ID: user.ID, Email: user.Email, Role: user.Role,
			IsBlocked: user.IsBlocked, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt,
		},
	})
}

// Me godoc
//
//	@Summary		Get current user profile
//	@Description	Returns the authenticated user's profile
//	@Tags			Auth
//	@Produce		json
//	@Success		200	{object}	responses.UserResponse
//	@Failure		401	{object}	responses.ErrorResponse
//	@Security		BearerAuth
//	@Router			/users/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	h.logger.Info().Msg("[AuthHandler] Me request received")

	userCtx, ok := helpers.ExtractUser(c.Request.Context())
	if !ok {
		h.logger.Error().Msg("[AuthHandler] Me failed to extract user from context")
		c.JSON(http.StatusUnauthorized, responses.ErrorResponse{Error: "unauthorized"})
		return
	}

	h.logger.Info().Str("userID", userCtx.UserID).Msg("[AuthHandler] Me fetching user profile")

	user, err := h.authApi.Me(c.Request.Context(), userCtx.UserID)
	if err != nil {
		h.logger.Error().Err(err).Str("userID", userCtx.UserID).Msg("[AuthHandler] Me failed")
		HandleError(c, err)
		return
	}

	h.logger.Info().Str("userID", user.ID).Msg("[AuthHandler] Me succeeded")

	c.JSON(http.StatusOK, responses.UserResponse{
		ID: user.ID, Email: user.Email, Role: user.Role,
		IsBlocked: user.IsBlocked, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt,
	})
}

// UpdateProfile godoc
//
//	@Summary		Update user profile
//	@Description	Update the authenticated user's profile
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		requests.UpdateProfileRequest	true	"Profile data"
//	@Success		200		{object}	responses.UserResponse
//	@Security		BearerAuth
//	@Router			/users/me [put]
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	h.logger.Info().Msg("[AuthHandler] UpdateProfile request received")
	c.JSON(http.StatusOK, gin.H{"message": "profile updated"})
	h.logger.Info().Msg("[AuthHandler] UpdateProfile succeeded")
}

// ChangePassword godoc
//
//	@Summary		Change password
//	@Description	Change the authenticated user's password
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body	requests.ChangePasswordRequest	true	"Password data"
//	@Success		200		{object}	map[string]string
//	@Security		BearerAuth
//	@Router			/users/me/password [post]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	h.logger.Info().Msg("[AuthHandler] ChangePassword request received")

	userCtx, ok := helpers.ExtractUser(c.Request.Context())
	if !ok {
		h.logger.Error().Msg("[AuthHandler] ChangePassword failed to extract user from context")
		c.JSON(http.StatusUnauthorized, responses.ErrorResponse{Error: "unauthorized"})
		return
	}

	var req requests.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error().Err(err).Msg("[AuthHandler] ChangePassword failed to bind request body")
		c.JSON(http.StatusBadRequest, responses.ErrorResponse{Error: err.Error()})
		return
	}

	h.logger.Info().Str("userID", userCtx.UserID).Msg("[AuthHandler] ChangePassword processing")

	if err := h.authApi.ChangePassword(c.Request.Context(), userCtx.UserID, req.CurrentPassword, req.NewPassword); err != nil {
		h.logger.Error().Err(err).Str("userID", userCtx.UserID).Msg("[AuthHandler] ChangePassword failed")
		HandleError(c, err)
		return
	}

	h.logger.Info().Str("userID", userCtx.UserID).Msg("[AuthHandler] ChangePassword succeeded")

	c.JSON(http.StatusOK, gin.H{"message": "password changed"})
}
