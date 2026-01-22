package domain

import (
	"time"
)

// CabinType represents a specific cabin type for a ship (小类)
type CabinType struct {
	ID          uint64    `json:"id" db:"id"`
	ShipID      uint64    `json:"ship_id" db:"ship_id"`
	CategoryID  uint64    `json:"category_id" db:"category_id"`
	Name        string    `json:"name" db:"name"`
	Code        string    `json:"code,omitempty" db:"code"`
	Description string    `json:"description,omitempty" db:"description"`
	SortOrder   int       `json:"sort_order" db:"sort_order"`
	IsEnabled   bool      `json:"is_enabled" db:"is_enabled"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`

	// Loaded relations
	Ship     *Ship          `json:"ship,omitempty" db:"-"`
	Category *CabinCategory `json:"category,omitempty" db:"-"`
}

// IsActive checks if cabin type is enabled
func (ct *CabinType) IsActive() bool {
	return ct.IsEnabled
}
