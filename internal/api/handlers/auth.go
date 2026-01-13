package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/pratik-mahalle/infraudit/internal/api/dto"
	"github.com/pratik-mahalle/infraudit/internal/api/middleware"
	"github.com/pratik-mahalle/infraudit/internal/auth"
	"github.com/pratik-mahalle/infraudit/internal/config"
	"github.com/pratik-mahalle/infraudit/internal/domain/user"
	"github.com/pratik-mahalle/infraudit/internal/pkg/errors"
	"github.com/pratik-mahalle/infraudit/internal/pkg/logger"
	"github.com/pratik-mahalle/infraudit/internal/pkg/utils"
	"github.com/pratik-mahalle/infraudit/internal/pkg/validator"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	userService user.Service
	config      *config.Config
	logger      *logger.Logger
	validator   *validator.Validator
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(
	userService user.Service,
	cfg *config.Config,
	log *logger.Logger,
	val *validator.Validator,
) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		config:      cfg,
		logger:      log,
		validator:   val,
	}
}

// Login handles user login
// @Summary User login
// @Description Authenticate user with email and password
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login credentials"
// @Success 200 {object} dto.AuthResponse "Successfully authenticated"
// @Failure 400 {object} utils.ErrorResponse "Invalid request"
// @Failure 401 {object} utils.ErrorResponse "Invalid credentials"
// @Router /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, errors.BadRequest("Invalid request body"))
		return
	}

	// Validate request
	if validationErrs := h.validator.Validate(req); len(validationErrs) > 0 {
		utils.WriteError(w, errors.ValidationError("Validation failed", validationErrs))
		return
	}

	h.logger.Infof("Login attempt for email: %s", req.Email)

	// Authenticate user
	authenticatedUser, err := h.userService.Authenticate(r.Context(), req.Email, req.Password)
	if err != nil {
		h.logger.WithFields(map[string]interface{}{
			"email": req.Email,
		}).Warn("Authentication failed")
		// Check if it's an AppError
		if appErr, ok := err.(*errors.AppError); ok {
			utils.WriteError(w, appErr)
		} else {
			utils.WriteError(w, errors.Unauthorized("Invalid credentials"))
		}
		return
	}

	// Generate tokens
	tokens, err := auth.MintTokens(
		authenticatedUser.ID,
		authenticatedUser.Email,
		h.config.Auth.JWTSecret,
		h.config.Auth.AccessTokenExpiry,
		h.config.Auth.RefreshTokenExpiry,
	)
	if err != nil {
		h.logger.ErrorWithErr(err, "Failed to generate tokens")
		utils.WriteError(w, errors.Internal("Failed to generate tokens", err))
		return
	}

	// Set cookies
	http.SetCookie(w, &http.Cookie{
		Name:     "accessToken",
		Value:    tokens.AccessToken,
		HttpOnly: true,
		Secure:   h.config.Server.Environment == "production",
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   int(h.config.Auth.AccessTokenExpiry.Seconds()),
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refreshToken",
		Value:    tokens.RefreshToken,
		HttpOnly: true,
		Secure:   h.config.Server.Environment == "production",
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   int(h.config.Auth.RefreshTokenExpiry.Seconds()),
	})

	// Return response
	response := dto.AuthResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		User: &dto.UserDTO{
			ID:       authenticatedUser.ID,
			Email:    authenticatedUser.Email,
			Username: authenticatedUser.Username,
			FullName: authenticatedUser.FullName,
			Role:     authenticatedUser.Role,
			PlanType: authenticatedUser.PlanType,
		},
	}

	h.logger.WithFields(map[string]interface{}{
		"user_id": authenticatedUser.ID,
		"email":   authenticatedUser.Email,
	}).Info("User logged in successfully")

	utils.WriteSuccess(w, http.StatusOK, response)
}

// Register handles user registration
// @Summary User registration
// @Description Register a new user account
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Registration details"
// @Success 201 {object} dto.AuthResponse "User successfully registered"
// @Failure 400 {object} utils.ErrorResponse "Invalid request or validation error"
// @Router /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, errors.BadRequest("Invalid request body"))
		return
	}

	// Validate request
	if validationErrs := h.validator.Validate(req); len(validationErrs) > 0 {
		utils.WriteError(w, errors.ValidationError("Validation failed", validationErrs))
		return
	}

	h.logger.Infof("Registration attempt for email: %s", req.Email)

	// Check if user already exists
	existingUser, err := h.userService.GetByEmail(r.Context(), req.Email)
	if err == nil && existingUser != nil {
		h.logger.WithFields(map[string]interface{}{
			"email": req.Email,
		}).Warn("Registration attempt with existing email")
		utils.WriteError(w, errors.Conflict("Email already registered"))
		return
	}

	// Create user with password
	newUser, err := h.userService.Create(r.Context(), req.Email, req.Password)
	if err != nil {
		h.logger.ErrorWithErr(err, "Failed to create user")
		// Check if it's a duplicate email error (fallback check)
		errMsg := err.Error()
		if errMsg != "" {
			// Check for SQLite unique constraint error
			if strings.Contains(errMsg, "UNIQUE constraint failed") ||
				strings.Contains(errMsg, "duplicate key") ||
				strings.Contains(errMsg, "already exists") {
				utils.WriteError(w, errors.Conflict("Email already registered"))
				return
			}
		}
		utils.WriteError(w, errors.Internal("Failed to create user", err))
		return
	}

	// Generate tokens
	tokens, err := auth.MintTokens(
		newUser.ID,
		newUser.Email,
		h.config.Auth.JWTSecret,
		h.config.Auth.AccessTokenExpiry,
		h.config.Auth.RefreshTokenExpiry,
	)
	if err != nil {
		h.logger.ErrorWithErr(err, "Failed to generate tokens")
		utils.WriteError(w, errors.Internal("Failed to generate tokens", err))
		return
	}

	// Set cookies
	http.SetCookie(w, &http.Cookie{
		Name:     "accessToken",
		Value:    tokens.AccessToken,
		HttpOnly: true,
		Secure:   h.config.Server.Environment == "production",
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   int(h.config.Auth.AccessTokenExpiry.Seconds()),
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refreshToken",
		Value:    tokens.RefreshToken,
		HttpOnly: true,
		Secure:   h.config.Server.Environment == "production",
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   int(h.config.Auth.RefreshTokenExpiry.Seconds()),
	})

	// Return response
	response := dto.AuthResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		User: &dto.UserDTO{
			ID:       newUser.ID,
			Email:    newUser.Email,
			Username: newUser.Username,
			FullName: newUser.FullName,
			Role:     newUser.Role,
			PlanType: newUser.PlanType,
		},
	}

	utils.WriteSuccess(w, http.StatusCreated, response)
}

// Logout handles user logout
// @Summary User logout
// @Description Logout current user
// @Tags auth
// @Success 200 {object} utils.SuccessResponse
// @Router /api/auth/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Clear cookies
	http.SetCookie(w, &http.Cookie{
		Name:     "accessToken",
		Value:    "",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   -1,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refreshToken",
		Value:    "",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   -1,
	})

	utils.WriteSuccessWithMessage(w, http.StatusOK, "Logged out successfully", nil)
}

// Me returns the current user's information
// @Summary Get current user
// @Description Get authenticated user's information
// @Tags Auth
// @Produce json
// @Success 200 {object} dto.UserDTO "User information"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized"
// @Security BearerAuth
// @Router /auth/me [get]
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r)
	if !ok {
		utils.WriteError(w, errors.Unauthorized("User not authenticated"))
		return
	}

	user, err := h.userService.GetByID(r.Context(), userID)
	if err != nil {
		h.logger.ErrorWithErr(err, "Failed to get user")
		utils.WriteError(w, errors.Internal("Failed to get user", err))
		return
	}

	userDTO := &dto.UserDTO{
		ID:       user.ID,
		Email:    user.Email,
		Username: user.Username,
		FullName: user.FullName,
		Role:     user.Role,
		PlanType: user.PlanType,
	}

	utils.WriteSuccess(w, http.StatusOK, userDTO)
}

// RefreshToken handles token refresh
// @Summary Refresh access token
// @Description Refresh access token using refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} dto.AuthResponse "New tokens generated"
// @Failure 400 {object} utils.ErrorResponse "Invalid request"
// @Failure 401 {object} utils.ErrorResponse "Invalid refresh token"
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req dto.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, errors.BadRequest("Invalid request body"))
		return
	}

	// Validate request
	if validationErrs := h.validator.Validate(req); len(validationErrs) > 0 {
		utils.WriteError(w, errors.ValidationError("Validation failed", validationErrs))
		return
	}

	// Parse refresh token
	claims, err := auth.ParseClaims(req.RefreshToken, h.config.Auth.JWTSecret)
	if err != nil {
		utils.WriteError(w, errors.Unauthorized("Invalid refresh token"))
		return
	}

	// Get user
	user, err := h.userService.GetByID(r.Context(), claims.UserID)
	if err != nil {
		h.logger.ErrorWithErr(err, "Failed to get user")
		utils.WriteError(w, errors.Unauthorized("Invalid refresh token"))
		return
	}

	// Generate new tokens
	tokens, err := auth.MintTokens(
		user.ID,
		user.Email,
		h.config.Auth.JWTSecret,
		h.config.Auth.AccessTokenExpiry,
		h.config.Auth.RefreshTokenExpiry,
	)
	if err != nil {
		h.logger.ErrorWithErr(err, "Failed to generate tokens")
		utils.WriteError(w, errors.Internal("Failed to generate tokens", err))
		return
	}

	response := dto.AuthResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		User: &dto.UserDTO{
			ID:       user.ID,
			Email:    user.Email,
			Username: user.Username,
			FullName: user.FullName,
			Role:     user.Role,
			PlanType: user.PlanType,
		},
	}

	utils.WriteSuccess(w, http.StatusOK, response)
}
