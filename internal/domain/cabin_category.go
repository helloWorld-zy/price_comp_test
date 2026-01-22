package domain

import (
	"time"
)

// CabinCategory represents a cabin type category (大类)
type CabinCategory struct {
	ID        uint64    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	NameEN    string    `json:"name_en,omitempty" db:"name_en"`
	SortOrder int       `json:"sort_order" db:"sort_order"`
	IsDefault bool      `json:"is_default" db:"is_default"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`

	// Loaded relations
	CabinTypes []CabinType `json:"cabin_types,omitempty" db:"-"`
}

// Default cabin categories
const (
	CabinCategoryInterior  = "内舱"
	CabinCategoryOceanView = "海景"
	CabinCategoryBalcony   = "阳台"
	CabinCategorySuite     = "套房"
)

// DefaultCabinCategories returns the default cabin category names
func DefaultCabinCategories() []string {
	return []string{
		CabinCategoryInterior,
		CabinCategoryOceanView,
		CabinCategoryBalcony,
		CabinCategorySuite,
	}
}
