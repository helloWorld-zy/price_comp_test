package domain

import (
	"encoding/json"
	"time"
)

// AuditAction represents the type of audit action
type AuditAction string

const (
	AuditActionCreate AuditAction = "CREATE"
	AuditActionUpdate AuditAction = "UPDATE"
	AuditActionDelete AuditAction = "DELETE"
	AuditActionLogin  AuditAction = "LOGIN"
	AuditActionLogout AuditAction = "LOGOUT"
	AuditActionImport AuditAction = "IMPORT"
	AuditActionExport AuditAction = "EXPORT"
	AuditActionVoid   AuditAction = "VOID"
)

// AuditLog represents an audit log entry
type AuditLog struct {
	ID         uint64          `json:"id" db:"id"`
	UserID     uint64          `json:"user_id" db:"user_id"`
	SupplierID *uint64         `json:"supplier_id,omitempty" db:"supplier_id"`
	Action     AuditAction     `json:"action" db:"action"`
	EntityType string          `json:"entity_type" db:"entity_type"`
	EntityID   uint64          `json:"entity_id" db:"entity_id"`
	OldValue   json.RawMessage `json:"old_value,omitempty" db:"old_value"`
	NewValue   json.RawMessage `json:"new_value,omitempty" db:"new_value"`
	TraceID    string          `json:"trace_id,omitempty" db:"trace_id"`
	IPAddress  string          `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent  string          `json:"user_agent,omitempty" db:"user_agent"`
	CreatedAt  time.Time       `json:"created_at" db:"created_at"`

	// Loaded relations
	User *User `json:"user,omitempty" db:"-"`
}

// EntityTypes for audit logging
const (
	EntityTypeUser          = "user"
	EntityTypeCruiseLine    = "cruise_line"
	EntityTypeShip          = "ship"
	EntityTypeSailing       = "sailing"
	EntityTypeCabinCategory = "cabin_category"
	EntityTypeCabinType     = "cabin_type"
	EntityTypeSupplier      = "supplier"
	EntityTypePriceQuote    = "price_quote"
	EntityTypeImportJob     = "import_job"
)
