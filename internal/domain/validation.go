package domain

import (
	"errors"
	"fmt"
	"regexp"
	"time"
	"unicode/utf8"
)

// Validation errors
var (
	ErrFieldRequired       = errors.New("field is required")
	ErrFieldTooShort       = errors.New("field is too short")
	ErrFieldTooLong        = errors.New("field is too long")
	ErrFieldInvalidFormat  = errors.New("field has invalid format")
	ErrFieldInvalidValue   = errors.New("field has invalid value")
	ErrFieldMustBePositive = errors.New("field must be positive")
	ErrDateOrderInvalid    = errors.New("start date must be before end date")
	ErrRoleSupplierMissing = errors.New("vendor role requires supplier_id")
)

// ValidationError represents a validation error with field context
type ValidationError struct {
	Field   string
	Message string
	Err     error
}

func (ve *ValidationError) Error() string {
	if ve.Message != "" {
		return fmt.Sprintf("%s: %s", ve.Field, ve.Message)
	}
	return fmt.Sprintf("%s: %v", ve.Field, ve.Err)
}

func (ve *ValidationError) Unwrap() error {
	return ve.Err
}

// NewValidationError creates a new validation error
func NewValidationError(field string, err error) *ValidationError {
	return &ValidationError{Field: field, Err: err}
}

// NewValidationErrorMsg creates a new validation error with message
func NewValidationErrorMsg(field, message string) *ValidationError {
	return &ValidationError{Field: field, Message: message}
}

// ValidationErrors is a collection of validation errors
type ValidationErrors []*ValidationError

func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return "no validation errors"
	}
	if len(ve) == 1 {
		return ve[0].Error()
	}
	return fmt.Sprintf("%s (and %d more errors)", ve[0].Error(), len(ve)-1)
}

// HasErrors returns true if there are validation errors
func (ve ValidationErrors) HasErrors() bool {
	return len(ve) > 0
}

// Add adds a validation error
func (ve *ValidationErrors) Add(field string, err error) {
	*ve = append(*ve, NewValidationError(field, err))
}

// AddMsg adds a validation error with message
func (ve *ValidationErrors) AddMsg(field, message string) {
	*ve = append(*ve, NewValidationErrorMsg(field, message))
}

// Validator provides validation utilities
type Validator struct {
	errors ValidationErrors
}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{}
}

// Errors returns the validation errors
func (v *Validator) Errors() ValidationErrors {
	return v.errors
}

// HasErrors returns true if there are validation errors
func (v *Validator) HasErrors() bool {
	return len(v.errors) > 0
}

// Required validates that a string field is not empty
func (v *Validator) Required(field, value string) bool {
	if value == "" {
		v.errors.Add(field, ErrFieldRequired)
		return false
	}
	return true
}

// RequiredPtr validates that a pointer field is not nil
func (v *Validator) RequiredPtr(field string, value interface{}) bool {
	if value == nil {
		v.errors.Add(field, ErrFieldRequired)
		return false
	}
	return true
}

// MinLength validates minimum string length
func (v *Validator) MinLength(field, value string, min int) bool {
	if utf8.RuneCountInString(value) < min {
		v.errors.AddMsg(field, fmt.Sprintf("must be at least %d characters", min))
		return false
	}
	return true
}

// MaxLength validates maximum string length
func (v *Validator) MaxLength(field, value string, max int) bool {
	if utf8.RuneCountInString(value) > max {
		v.errors.AddMsg(field, fmt.Sprintf("must be at most %d characters", max))
		return false
	}
	return true
}

// LengthRange validates string length is within range
func (v *Validator) LengthRange(field, value string, min, max int) bool {
	length := utf8.RuneCountInString(value)
	if length < min || length > max {
		v.errors.AddMsg(field, fmt.Sprintf("must be between %d and %d characters", min, max))
		return false
	}
	return true
}

// Positive validates that a number is positive
func (v *Validator) Positive(field string, value float64) bool {
	if value <= 0 {
		v.errors.Add(field, ErrFieldMustBePositive)
		return false
	}
	return true
}

// PositiveInt validates that an integer is positive
func (v *Validator) PositiveInt(field string, value int64) bool {
	if value <= 0 {
		v.errors.Add(field, ErrFieldMustBePositive)
		return false
	}
	return true
}

// DateBefore validates that date1 is before date2
func (v *Validator) DateBefore(field string, date1, date2 time.Time) bool {
	if !date1.Before(date2) {
		v.errors.Add(field, ErrDateOrderInvalid)
		return false
	}
	return true
}

// Pattern validates string against regex pattern
func (v *Validator) Pattern(field, value, pattern string) bool {
	matched, err := regexp.MatchString(pattern, value)
	if err != nil || !matched {
		v.errors.Add(field, ErrFieldInvalidFormat)
		return false
	}
	return true
}

// OneOf validates value is one of allowed values
func (v *Validator) OneOf(field, value string, allowed []string) bool {
	for _, a := range allowed {
		if value == a {
			return true
		}
	}
	v.errors.AddMsg(field, fmt.Sprintf("must be one of: %v", allowed))
	return false
}

// ValidateUser validates a user entity
func ValidateUser(u *User) ValidationErrors {
	v := NewValidator()

	v.Required("username", u.Username)
	if u.Username != "" {
		v.LengthRange("username", u.Username, 2, 50)
	}

	v.Required("password_hash", u.PasswordHash)

	v.OneOf("role", string(u.Role), []string{string(UserRoleAdmin), string(UserRoleVendor)})

	// Vendor must have supplier_id
	if u.Role == UserRoleVendor && u.SupplierID == nil {
		v.errors.Add("supplier_id", ErrRoleSupplierMissing)
	}

	return v.Errors()
}

// ValidateCruiseLine validates a cruise line entity
func ValidateCruiseLine(cl *CruiseLine) ValidationErrors {
	v := NewValidator()

	v.Required("name", cl.Name)
	if cl.Name != "" {
		v.LengthRange("name", cl.Name, 2, 100)
	}

	if cl.NameEN != "" {
		v.MaxLength("name_en", cl.NameEN, 100)
	}

	return v.Errors()
}

// ValidateShip validates a ship entity
func ValidateShip(s *Ship) ValidationErrors {
	v := NewValidator()

	v.Required("name", s.Name)
	if s.Name != "" {
		v.LengthRange("name", s.Name, 2, 100)
	}

	v.PositiveInt("cruise_line_id", int64(s.CruiseLineID))

	return v.Errors()
}

// ValidateSailing validates a sailing entity
func ValidateSailing(s *Sailing) ValidationErrors {
	v := NewValidator()

	v.PositiveInt("ship_id", int64(s.ShipID))
	v.Required("route", s.Route)

	if !s.DepartureDate.IsZero() && !s.ReturnDate.IsZero() {
		v.DateBefore("departure_date", s.DepartureDate, s.ReturnDate)
	}

	return v.Errors()
}

// ValidateCabinType validates a cabin type entity
func ValidateCabinType(ct *CabinType) ValidationErrors {
	v := NewValidator()

	v.PositiveInt("ship_id", int64(ct.ShipID))
	v.PositiveInt("category_id", int64(ct.CategoryID))
	v.Required("name", ct.Name)
	if ct.Name != "" {
		v.LengthRange("name", ct.Name, 1, 100)
	}

	return v.Errors()
}

// ValidateSupplier validates a supplier entity
func ValidateSupplier(s *Supplier) ValidationErrors {
	v := NewValidator()

	v.Required("name", s.Name)
	if s.Name != "" {
		v.LengthRange("name", s.Name, 2, 100)
	}

	return v.Errors()
}

// ValidatePriceQuote validates a price quote entity
func ValidatePriceQuote(pq *PriceQuote) ValidationErrors {
	v := NewValidator()

	v.PositiveInt("sailing_id", int64(pq.SailingID))
	v.PositiveInt("cabin_type_id", int64(pq.CabinTypeID))
	v.PositiveInt("supplier_id", int64(pq.SupplierID))

	price, _ := pq.Price.Float64()
	v.Positive("price", price)

	v.Required("currency", pq.Currency)
	if pq.Currency != "" {
		v.LengthRange("currency", pq.Currency, 3, 3)
	}

	v.OneOf("pricing_unit", string(pq.PricingUnit), []string{
		string(PricingUnitPerPerson),
		string(PricingUnitPerCabin),
		string(PricingUnitTotal),
	})

	v.OneOf("source", string(pq.Source), []string{
		string(QuoteSourceManual),
		string(QuoteSourceFileImport),
		string(QuoteSourceTextImport),
		string(QuoteSourceTemplateImport),
	})

	return v.Errors()
}
