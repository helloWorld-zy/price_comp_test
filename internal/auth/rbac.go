package auth

import (
	"net/http"

	"cruise-price-compare/internal/domain"

	"github.com/gin-gonic/gin"
)

// RBACConfig holds RBAC configuration
type RBACConfig struct {
	// Route permissions: map[method:path][]allowedRoles
	RoutePermissions map[string][]domain.UserRole
}

// RBACMiddleware provides role-based access control
type RBACMiddleware struct {
	config RBACConfig
}

// NewRBACMiddleware creates a new RBAC middleware
func NewRBACMiddleware(config RBACConfig) *RBACMiddleware {
	return &RBACMiddleware{config: config}
}

// RequireRole returns a middleware that requires specific roles
func RequireRole(roles ...domain.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := GetUserFromContext(c)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized",
				"code":  "UNAUTHORIZED",
			})
			return
		}

		// Check if user has any of the required roles
		hasRole := false
		for _, role := range roles {
			if user.Role == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "forbidden: insufficient permissions",
				"code":  "FORBIDDEN",
			})
			return
		}

		c.Next()
	}
}

// RequireAdmin is a shorthand for RequireRole(domain.UserRoleAdmin)
func RequireAdmin() gin.HandlerFunc {
	return RequireRole(domain.UserRoleAdmin)
}

// RequireVendor is a shorthand for RequireRole(domain.UserRoleVendor)
func RequireVendor() gin.HandlerFunc {
	return RequireRole(domain.UserRoleVendor)
}

// RequireAdminOrVendor allows both admin and vendor roles
func RequireAdminOrVendor() gin.HandlerFunc {
	return RequireRole(domain.UserRoleAdmin, domain.UserRoleVendor)
}

// RequireSupplierAccess ensures user can access supplier-specific resources
func RequireSupplierAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := GetUserFromContext(c)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized",
				"code":  "UNAUTHORIZED",
			})
			return
		}

		// Admin has access to all suppliers
		if user.IsAdmin() {
			c.Next()
			return
		}

		// Vendor must have a supplier ID
		if user.SupplierID == nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "vendor user must be associated with a supplier",
				"code":  "NO_SUPPLIER",
			})
			return
		}

		c.Next()
	}
}

// CheckSupplierOwnership verifies user can access a specific supplier's resources
func CheckSupplierOwnership(user *domain.User, supplierID uint64) bool {
	if user.IsAdmin() {
		return true
	}
	return user.SupplierID != nil && *user.SupplierID == supplierID
}
