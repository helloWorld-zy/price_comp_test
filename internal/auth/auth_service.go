package auth

import (
	"context"
	"errors"
	"fmt"

	"cruise-price-compare/internal/domain"
	"cruise-price-compare/internal/repo"
)

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrUserInactive        = errors.New("user account is inactive")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
)

// AuthService handles authentication operations
type AuthService struct {
	userRepo        *repo.UserRepository
	jwtService      *JWTService
	passwordService *PasswordService
}

// NewAuthService creates a new auth service
func NewAuthService(userRepo *repo.UserRepository, jwtService *JWTService, passwordService *PasswordService) *AuthService {
	return &AuthService{
		userRepo:        userRepo,
		jwtService:      jwtService,
		passwordService: passwordService,
	}
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	User   *domain.User `json:"user"`
	Tokens *TokenPair   `json:"tokens"`
}

// Login authenticates a user and returns tokens
func (s *AuthService) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	// Find user
	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	// Check status
	if !user.IsActive() {
		return nil, ErrUserInactive
	}

	// Verify password
	valid, err := s.passwordService.VerifyPassword(req.Password, user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("failed to verify password: %w", err)
	}
	if !valid {
		return nil, ErrInvalidCredentials
	}

	// Generate tokens
	var supplierID uint64
	if user.SupplierID != nil {
		supplierID = *user.SupplierID
	}

	tokens, err := s.jwtService.GenerateTokenPair(user.ID, user.Username, string(user.Role), supplierID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return &LoginResponse{
		User:   user,
		Tokens: tokens,
	}, nil
}

// RefreshRequest represents a token refresh request
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// Refresh refreshes an access token
func (s *AuthService) Refresh(ctx context.Context, req *RefreshRequest) (*TokenPair, error) {
	// Validate refresh token and get user ID
	claims, err := s.jwtService.ValidateToken(req.RefreshToken)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// Check status
	if !user.IsActive() {
		return nil, ErrUserInactive
	}

	// Generate new tokens
	var supplierID uint64
	if user.SupplierID != nil {
		supplierID = *user.SupplierID
	}

	tokens, err := s.jwtService.RefreshAccessToken(req.RefreshToken, user.Username, string(user.Role), supplierID)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	return tokens, nil
}

// GetCurrentUser returns the current user from token claims
func (s *AuthService) GetCurrentUser(ctx context.Context, claims *Claims) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	return user, nil
}

// ChangePassword changes a user's password
func (s *AuthService) ChangePassword(ctx context.Context, userID uint64, currentPassword, newPassword string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	// Verify current password
	valid, err := s.passwordService.VerifyPassword(currentPassword, user.PasswordHash)
	if err != nil {
		return fmt.Errorf("failed to verify password: %w", err)
	}
	if !valid {
		return ErrInvalidCredentials
	}

	// Hash new password
	newHash, err := s.passwordService.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	if err := s.userRepo.UpdatePassword(ctx, userID, newHash); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}
