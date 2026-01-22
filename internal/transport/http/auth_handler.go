package http

import (
	"net/http"

	"cruise-price-compare/internal/auth"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService *auth.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *auth.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login handles POST /auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondValidationError(c, "invalid request body", err.Error())
		return
	}

	result, err := h.authService.Login(c.Request.Context(), &auth.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	})

	if err != nil {
		switch err {
		case auth.ErrInvalidCredentials:
			RespondUnauthorized(c, "invalid username or password")
		case auth.ErrUserInactive:
			RespondForbidden(c, "user account is inactive")
		default:
			RespondInternalError(c, err)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":   result.User,
		"tokens": result.Tokens,
	})
}

// RefreshRequest represents a token refresh request
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Refresh handles POST /auth/refresh
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondValidationError(c, "invalid request body", err.Error())
		return
	}

	tokens, err := h.authService.Refresh(c.Request.Context(), &auth.RefreshRequest{
		RefreshToken: req.RefreshToken,
	})

	if err != nil {
		switch err {
		case auth.ErrInvalidRefreshToken:
			RespondUnauthorized(c, "invalid refresh token")
		case auth.ErrUserInactive:
			RespondForbidden(c, "user account is inactive")
		default:
			RespondInternalError(c, err)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tokens": tokens,
	})
}

// GetCurrentUser handles GET /auth/me
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	claims, exists := auth.GetClaimsFromContext(c)
	if !exists {
		RespondUnauthorized(c, "no valid token")
		return
	}

	user, err := h.authService.GetCurrentUser(c.Request.Context(), claims)
	if err != nil {
		RespondInternalError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

// Logout handles POST /auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// In a stateless JWT system, logout is handled client-side
	// We just return success
	c.JSON(http.StatusOK, gin.H{
		"message": "logged out successfully",
	})
}

// ChangePasswordRequest represents a password change request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
}

// ChangePassword handles PUT /auth/password
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondValidationError(c, "invalid request body", err.Error())
		return
	}

	user, exists := auth.GetUserFromContext(c)
	if !exists {
		RespondUnauthorized(c, "no valid token")
		return
	}

	err := h.authService.ChangePassword(c.Request.Context(), user.ID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		switch err {
		case auth.ErrInvalidCredentials:
			RespondBadRequest(c, "current password is incorrect")
		default:
			RespondInternalError(c, err)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "password changed successfully",
	})
}
