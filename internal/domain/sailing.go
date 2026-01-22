package domain

import (
	"time"
)

// SailingStatus represents the status of a sailing
type SailingStatus string

const (
	SailingStatusActive    SailingStatus = "ACTIVE"
	SailingStatusCancelled SailingStatus = "CANCELLED"
)

// Sailing represents a cruise sailing/voyage
type Sailing struct {
	ID            uint64        `json:"id" db:"id"`
	ShipID        uint64        `json:"ship_id" db:"ship_id"`
	SailingCode   string        `json:"sailing_code,omitempty" db:"sailing_code"`
	DepartureDate time.Time     `json:"departure_date" db:"departure_date"`
	ReturnDate    time.Time     `json:"return_date" db:"return_date"`
	Nights        int           `json:"nights" db:"nights"` // Computed field
	Route         string        `json:"route" db:"route"`
	Ports         []string      `json:"ports,omitempty" db:"ports"`
	Description   string        `json:"description,omitempty" db:"description"`
	Status        SailingStatus `json:"status" db:"status"`
	CreatedAt     time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at" db:"updated_at"`
	CreatedBy     *uint64       `json:"created_by,omitempty" db:"created_by"`

	// Loaded relations
	Ship        *Ship        `json:"ship,omitempty" db:"-"`
	PriceQuotes []PriceQuote `json:"price_quotes,omitempty" db:"-"`
}

// IsActive checks if sailing is active
func (s *Sailing) IsActive() bool {
	return s.Status == SailingStatusActive
}

// IsCancelled checks if sailing is cancelled
func (s *Sailing) IsCancelled() bool {
	return s.Status == SailingStatusCancelled
}

// CalculateNights returns the number of nights for the sailing
func (s *Sailing) CalculateNights() int {
	return int(s.ReturnDate.Sub(s.DepartureDate).Hours() / 24)
}

// IsFuture checks if sailing departure is in the future
func (s *Sailing) IsFuture() bool {
	return s.DepartureDate.After(time.Now())
}
