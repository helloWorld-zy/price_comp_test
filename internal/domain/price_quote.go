package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

// PricingUnit represents the pricing unit type
type PricingUnit string

const (
	PricingUnitPerPerson PricingUnit = "PER_PERSON"
	PricingUnitPerCabin  PricingUnit = "PER_CABIN"
	PricingUnitTotal     PricingUnit = "TOTAL"
)

// QuoteSource represents the source of a price quote
type QuoteSource string

const (
	QuoteSourceManual         QuoteSource = "MANUAL"
	QuoteSourceFileImport     QuoteSource = "FILE_IMPORT"
	QuoteSourceTextImport     QuoteSource = "TEXT_IMPORT"
	QuoteSourceTemplateImport QuoteSource = "TEMPLATE_IMPORT"
)

// QuoteStatus represents the status of a price quote
type QuoteStatus string

const (
	QuoteStatusActive    QuoteStatus = "ACTIVE"
	QuoteStatusVoided    QuoteStatus = "VOIDED"
	QuoteStatusCorrected QuoteStatus = "CORRECTED"
)

// PriceQuote represents a price quote record (append-only)
type PriceQuote struct {
	ID            uint64          `json:"id" db:"id"`
	SailingID     uint64          `json:"sailing_id" db:"sailing_id"`
	CabinTypeID   uint64          `json:"cabin_type_id" db:"cabin_type_id"`
	SupplierID    uint64          `json:"supplier_id" db:"supplier_id"`
	Price         decimal.Decimal `json:"price" db:"price"`
	Currency      string          `json:"currency" db:"currency"`
	PricingUnit   PricingUnit     `json:"pricing_unit" db:"pricing_unit"`
	Conditions    string          `json:"conditions,omitempty" db:"conditions"`
	GuestCount    *int            `json:"guest_count,omitempty" db:"guest_count"`
	Promotion     string          `json:"promotion,omitempty" db:"promotion"`
	CabinQuantity *int            `json:"cabin_quantity,omitempty" db:"cabin_quantity"`
	ValidUntil    *time.Time      `json:"valid_until,omitempty" db:"valid_until"`
	Notes         string          `json:"notes,omitempty" db:"notes"`
	Source        QuoteSource     `json:"source" db:"source"`
	SourceRef     string          `json:"source_ref,omitempty" db:"source_ref"`
	ImportJobID   *uint64         `json:"import_job_id,omitempty" db:"import_job_id"`
	Status        QuoteStatus     `json:"status" db:"status"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
	CreatedBy     uint64          `json:"created_by" db:"created_by"`

	// Loaded relations
	Sailing   *Sailing   `json:"sailing,omitempty" db:"-"`
	CabinType *CabinType `json:"cabin_type,omitempty" db:"-"`
	Supplier  *Supplier  `json:"supplier,omitempty" db:"-"`
	ImportJob *ImportJob `json:"import_job,omitempty" db:"-"`
}

// IsActive checks if quote is active
func (pq *PriceQuote) IsActive() bool {
	return pq.Status == QuoteStatusActive
}

// IsVoided checks if quote is voided
func (pq *PriceQuote) IsVoided() bool {
	return pq.Status == QuoteStatusVoided
}

// IsValid checks if quote is still valid (not expired)
func (pq *PriceQuote) IsValid() bool {
	if !pq.IsActive() {
		return false
	}
	if pq.ValidUntil != nil && pq.ValidUntil.Before(time.Now()) {
		return false
	}
	return true
}

// PricePerPerson calculates the price per person based on pricing unit
func (pq *PriceQuote) PricePerPerson(defaultGuestCount int) decimal.Decimal {
	switch pq.PricingUnit {
	case PricingUnitPerPerson:
		return pq.Price
	case PricingUnitPerCabin:
		guests := defaultGuestCount
		if pq.GuestCount != nil && *pq.GuestCount > 0 {
			guests = *pq.GuestCount
		}
		if guests <= 0 {
			guests = 2 // Default to 2 guests per cabin
		}
		return pq.Price.Div(decimal.NewFromInt(int64(guests)))
	case PricingUnitTotal:
		guests := defaultGuestCount
		if pq.GuestCount != nil && *pq.GuestCount > 0 {
			guests = *pq.GuestCount
		}
		if guests <= 0 {
			guests = 2
		}
		return pq.Price.Div(decimal.NewFromInt(int64(guests)))
	default:
		return pq.Price
	}
}
