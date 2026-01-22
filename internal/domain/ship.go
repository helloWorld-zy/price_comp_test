package domain

import (
	"time"
)

// Ship represents a cruise ship
type Ship struct {
	ID           uint64       `json:"id" db:"id"`
	CruiseLineID uint64       `json:"cruise_line_id" db:"cruise_line_id"`
	Name         string       `json:"name" db:"name"`
	Aliases      []string     `json:"aliases,omitempty" db:"aliases"`
	Status       EntityStatus `json:"status" db:"status"`
	CreatedAt    time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at" db:"updated_at"`
	CreatedBy    *uint64      `json:"created_by,omitempty" db:"created_by"`

	// Loaded relations
	CruiseLine *CruiseLine `json:"cruise_line,omitempty" db:"-"`
	Sailings   []Sailing   `json:"sailings,omitempty" db:"-"`
	CabinTypes []CabinType `json:"cabin_types,omitempty" db:"-"`
}

// IsActive checks if ship is active
func (s *Ship) IsActive() bool {
	return s.Status == EntityStatusActive
}

// MatchesAlias checks if a given name matches this ship's name or aliases
func (s *Ship) MatchesAlias(name string) bool {
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
