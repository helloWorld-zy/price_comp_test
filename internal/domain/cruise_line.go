package domain

import (
	"time"
)

// EntityStatus represents common status for entities
type EntityStatus string

const (
	EntityStatusActive   EntityStatus = "ACTIVE"
	EntityStatusInactive EntityStatus = "INACTIVE"
)

// CruiseLine represents a cruise company
type CruiseLine struct {
	ID        uint64       `json:"id" db:"id"`
	Name      string       `json:"name" db:"name"`
	NameEN    string       `json:"name_en,omitempty" db:"name_en"`
	Aliases   []string     `json:"aliases,omitempty" db:"aliases"`
	Status    EntityStatus `json:"status" db:"status"`
	CreatedAt time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt time.Time    `json:"updated_at" db:"updated_at"`
	CreatedBy *uint64      `json:"created_by,omitempty" db:"created_by"`

	// Loaded relations
	Ships []Ship `json:"ships,omitempty" db:"-"`
}

// IsActive checks if cruise line is active
func (cl *CruiseLine) IsActive() bool {
	return cl.Status == EntityStatusActive
}

// MatchesAlias checks if a given name matches this cruise line's name or aliases
func (cl *CruiseLine) MatchesAlias(name string) bool {
	if cl.Name == name || cl.NameEN == name {
		return true
	}
	for _, alias := range cl.Aliases {
		if alias == name {
			return true
		}
	}
	return false
}
