package domain

import (
	"time"
)

// SupplierVisibility represents the visibility of a supplier's quotes
type SupplierVisibility string

const (
	SupplierVisibilityPrivate SupplierVisibility = "PRIVATE"
	SupplierVisibilityPublic  SupplierVisibility = "PUBLIC"
)

// Supplier represents a price quote supplier
type Supplier struct {
	ID          uint64             `json:"id" db:"id"`
	Name        string             `json:"name" db:"name"`
	Aliases     []string           `json:"aliases,omitempty" db:"aliases"`
	ContactInfo string             `json:"contact_info,omitempty" db:"contact_info"`
	Visibility  SupplierVisibility `json:"visibility" db:"visibility"`
	Status      EntityStatus       `json:"status" db:"status"`
	CreatedAt   time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" db:"updated_at"`
	CreatedBy   *uint64            `json:"created_by,omitempty" db:"created_by"`

	// Loaded relations
	Users       []User       `json:"users,omitempty" db:"-"`
	PriceQuotes []PriceQuote `json:"price_quotes,omitempty" db:"-"`
}

// IsActive checks if supplier is active
func (s *Supplier) IsActive() bool {
	return s.Status == EntityStatusActive
}

// IsPublic checks if supplier quotes are publicly visible
func (s *Supplier) IsPublic() bool {
	return s.Visibility == SupplierVisibilityPublic
}

// MatchesAlias checks if a given name matches this supplier's name or aliases
func (s *Supplier) MatchesAlias(name string) bool {
	if s.Name == name {
		return true
	}
	for _, alias := range s.Aliases {
		if alias == name {
			return true
		}
	}
	return false
}
