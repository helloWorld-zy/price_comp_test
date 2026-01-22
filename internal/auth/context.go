package auth

import (
	"net/http"
	"strings"

	"cruise-price-compare/internal/domain"

	"github.com/gin-gonic/gin"
)

const (
	// ContextKeyUser is the key for storing user in context
	ContextKeyUser = "user"
	// ContextKeyClaims is the key for storing claims in context
	ContextKeyClaims = "claims"
	// ContextKeySupplierID is the key for storing supplier ID in context
	ContextKeySupplierID = "supplier_id"
)

// UserContextMiddleware extracts user info from JWT and injects into context
type UserContextMiddleware struct {
	jwtService *JWTService
}

// NewUserContextMiddleware creates a new user context middleware
func NewUserContextMiddleware(jwtService *JWTService) *UserContextMiddleware {
	return &UserContextMiddleware{jwtService: jwtService}
}

// Handler returns the middleware handler function
func (m *UserContextMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		// Check Bearer prefix
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.Next()
			return
		}

		tokenString := parts[1]

		// Validate token
		claims, err := m.jwtService.ValidateToken(tokenString)
		if err != nil {
			// Token is invalid, but we don't abort - let route handlers decide
			c.Next()
			return
		}

		// Create user from claims
		user := &domain.User{
			ID:       claims.UserID,
			Username: claims.Username,
			Role:     domain.UserRole(claims.Role),
		}

		if claims.SupplierID > 0 {
			supplierID := claims.SupplierID
			user.SupplierID = &supplierID
		}

		// Store in context
		c.Set(ContextKeyUser, user)
		c.Set(ContextKeyClaims, claims)
		if claims.SupplierID > 0 {
			c.Set(ContextKeySupplierID, claims.SupplierID)
		}

		c.Next()
	}
}

// RequireAuth returns a middleware that requires authentication
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, exists := GetUserFromContext(c)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authentication required",
				"code":  "UNAUTHORIZED",
			})
			return
		}
		c.Next()
	}
}

// GetUserFromContext retrieves the user from the gin context
func GetUserFromContext(c *gin.Context) (*domain.User, bool) {
	val, exists := c.Get(ContextKeyUser)
	if !exists {
		return nil, false
	}

	user, ok := val.(*domain.User)
	if !ok {
		return nil, false
	}

	return user, true
}

// GetClaimsFromContext retrieves the claims from the gin context
func GetClaimsFromContext(c *gin.Context) (*Claims, bool) {
	val, exists := c.Get(ContextKeyClaims)
	if !exists {
		return nil, false
	}

	claims, ok := val.(*Claims)
	if !ok {
		return nil, false
	}

	return claims, true
}

// GetSupplierIDFromContext retrieves the supplier ID from the gin context
func GetSupplierIDFromContext(c *gin.Context) (uint64, bool) {
	val, exists := c.Get(ContextKeySupplierID)
	if !exists {
		return 0, false
	}

	supplierID, ok := val.(uint64)
	if !ok {
		return 0, false
	}

	return supplierID, true
}

// MustGetUser retrieves the user from context or panics
func MustGetUser(c *gin.Context) *domain.User {
	user, exists := GetUserFromContext(c)
	if !exists {
		panic("user not found in context")
	}
	return user
}

// UserContext represents user context with role and supplier info
type UserContext struct {
	UserID     uint64
	Username   string
	Role       domain.UserRole
	SupplierID uint64
}

// GetUserContext retrieves user context from gin context
func GetUserContext(c *gin.Context) *UserContext {
	user, exists := GetUserFromContext(c)
	if !exists {
		return nil
	}

	ctx := &UserContext{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
	}

	if user.SupplierID != nil {
		ctx.SupplierID = *user.SupplierID
	}

	return ctx
}
