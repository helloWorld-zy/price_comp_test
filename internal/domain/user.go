package domain

import (
	"time"
)

// UserRole represents the role of a user
type UserRole string

const (
	UserRoleAdmin  UserRole = "ADMIN"
	UserRoleVendor UserRole = "VENDOR"
)

// UserStatus represents the status of a user
type UserStatus string

const (
	UserStatusActive   UserStatus = "ACTIVE"
	UserStatusInactive UserStatus = "INACTIVE"
)

// User represents a system user
type User struct {
	ID           uint64     `json:"id" db:"id"`
	Username     string     `json:"username" db:"username"`
	PasswordHash string     `json:"-" db:"password_hash"` // Never expose password hash
	Role         UserRole   `json:"role" db:"role"`
	SupplierID   *uint64    `json:"supplier_id,omitempty" db:"supplier_id"`
	Status       UserStatus `json:"status" db:"status"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`

	// Loaded relations (not always populated)
	Supplier *Supplier `json:"supplier,omitempty" db:"-"`
}

// IsAdmin checks if user has admin role
func (u *User) IsAdmin() bool {
	return u.Role == UserRoleAdmin
}

// IsVendor checks if user has vendor role
func (u *User) IsVendor() bool {
	return u.Role == UserRoleVendor
}

// IsActive checks if user is active
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

// CanAccessSupplier checks if user can access a specific supplier's data
func (u *User) CanAccessSupplier(supplierID uint64) bool {
	if u.IsAdmin() {
		return true
	}
	if u.SupplierID != nil && *u.SupplierID == supplierID {
		return true
	}
	return false
}
