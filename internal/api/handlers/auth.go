package handlers

import (
	"net/http"

	"github.com/pratik-mahalle/infraudit/internal/api/dto"
	"github.com/pratik-mahalle/infraudit/internal/api/middleware"
	"github.com/pratik-mahalle/infraudit/internal/domain/user"
	"github.com/pratik-mahalle/infraudit/internal/pkg/errors"
	"github.com/pratik-mahalle/infraudit/internal/pkg/logger"
	"github.com/pratik-mahalle/infraudit/internal/pkg/utils"
)

// AuthHandler handles authentication-related requests.
// With Supabase Auth, login/register/refresh are handled by the frontend.
// This handler only provides the /me endpoint and logout.
type AuthHandler struct {
	userService user.Service
	logger      *logger.Logger
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(
	userService user.Service,
	log *logger.Logger,
) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		logger:      log,
	}
}

// Me returns the current user's profile information
// @Summary Get current user
// @Description Get authenticated user's profile from the profiles table
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

	u, err := h.userService.GetByID(r.Context(), userID)
	if err != nil {
		// If profile not found, try to create it from the Supabase auth ID
		authID, authOk := middleware.GetAuthID(r)
		email, emailOk := middleware.GetUserEmail(r)
		if authOk && emailOk {
			u, err = h.userService.EnsureProfile(r.Context(), authID, email, "")
			if err != nil {
				h.logger.ErrorWithErr(err, "Failed to get or create user profile")
				utils.WriteError(w, errors.Internal("Failed to get user profile", err))
				return
			}
		} else {
			h.logger.ErrorWithErr(err, "Failed to get user")
			utils.WriteError(w, errors.Internal("Failed to get user", err))
			return
		}
	}

	userDTO := &dto.UserDTO{
		ID:       u.ID,
		Email:    u.Email,
		Username: u.Username,
		FullName: u.FullName,
		Role:     u.Role,
		PlanType: u.PlanType,
	}

	utils.WriteSuccess(w, http.StatusOK, userDTO)
}

// Logout handles user logout (clears any server-side cookies)
// @Summary User logout
// @Description Logout current user by clearing cookies
// @Tags Auth
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
