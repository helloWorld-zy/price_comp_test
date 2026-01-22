package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"cruise-price-compare/internal/domain"
)

// AuditLogRepository handles audit log data access
type AuditLogRepository struct {
	db *DB
}

// NewAuditLogRepository creates a new audit log repository
func NewAuditLogRepository(db *DB) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

// Create creates a new audit log entry
func (r *AuditLogRepository) Create(ctx context.Context, log *domain.AuditLog) error {
	query := `INSERT INTO audit_log (user_id, supplier_id, action, entity_type, entity_id, 
              old_value, new_value, trace_id, ip_address, user_agent) 
              VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db.ExecContext(ctx, query, log.UserID, log.SupplierID, log.Action,
		log.EntityType, log.EntityID, log.OldValue, log.NewValue, log.TraceID,
		log.IPAddress, log.UserAgent)
	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	log.ID = uint64(id)

	return nil
}

// List retrieves audit logs with pagination and filters
func (r *AuditLogRepository) List(ctx context.Context, pagination Pagination, userID *uint64, entityType *string, entityID *uint64, action *domain.AuditAction, from, to *time.Time) (PaginatedResult[domain.AuditLog], error) {
	var logs []domain.AuditLog
	var total int64

	countQuery := "SELECT COUNT(*) FROM audit_log WHERE 1=1"
	selectQuery := `SELECT id, user_id, supplier_id, action, entity_type, entity_id, 
                    old_value, new_value, trace_id, ip_address, user_agent, created_at 
                    FROM audit_log WHERE 1=1`
	var args []interface{}

	if userID != nil {
		countQuery += " AND user_id = ?"
		selectQuery += " AND user_id = ?"
		args = append(args, *userID)
	}

	if entityType != nil {
		countQuery += " AND entity_type = ?"
		selectQuery += " AND entity_type = ?"
		args = append(args, *entityType)
	}

	if entityID != nil {
		countQuery += " AND entity_id = ?"
		selectQuery += " AND entity_id = ?"
		args = append(args, *entityID)
	}

	if action != nil {
		countQuery += " AND action = ?"
		selectQuery += " AND action = ?"
		args = append(args, *action)
	}

	if from != nil {
		countQuery += " AND created_at >= ?"
		selectQuery += " AND created_at >= ?"
		args = append(args, *from)
	}

	if to != nil {
		countQuery += " AND created_at <= ?"
		selectQuery += " AND created_at <= ?"
		args = append(args, *to)
	}

	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return PaginatedResult[domain.AuditLog]{}, fmt.Errorf("failed to count audit logs: %w", err)
	}

	selectQuery += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, pagination.Limit(), pagination.Offset())

	if err := r.db.SelectContext(ctx, &logs, selectQuery, args...); err != nil {
		return PaginatedResult[domain.AuditLog]{}, fmt.Errorf("failed to list audit logs: %w", err)
	}

	return NewPaginatedResult(logs, total, pagination), nil
}

// ListByEntity retrieves audit logs for a specific entity
func (r *AuditLogRepository) ListByEntity(ctx context.Context, entityType string, entityID uint64, limit int) ([]domain.AuditLog, error) {
	var logs []domain.AuditLog
	query := `SELECT id, user_id, supplier_id, action, entity_type, entity_id, 
              old_value, new_value, trace_id, ip_address, user_agent, created_at 
              FROM audit_log WHERE entity_type = ? AND entity_id = ? 
              ORDER BY created_at DESC LIMIT ?`

	if err := r.db.SelectContext(ctx, &logs, query, entityType, entityID, limit); err != nil {
		return nil, fmt.Errorf("failed to list audit logs by entity: %w", err)
	}

	return logs, nil
}

// CreateFromEntity creates an audit log entry from entity changes
func (r *AuditLogRepository) CreateFromEntity(ctx context.Context, userID uint64, supplierID *uint64, action domain.AuditAction, entityType string, entityID uint64, oldEntity, newEntity interface{}, traceID, ipAddress, userAgent string) error {
	var oldValue, newValue json.RawMessage

	if oldEntity != nil {
		data, err := json.Marshal(oldEntity)
		if err != nil {
			return fmt.Errorf("failed to marshal old entity: %w", err)
		}
		oldValue = data
	}

	if newEntity != nil {
		data, err := json.Marshal(newEntity)
		if err != nil {
			return fmt.Errorf("failed to marshal new entity: %w", err)
		}
		newValue = data
	}

	log := &domain.AuditLog{
		UserID:     userID,
		SupplierID: supplierID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		OldValue:   oldValue,
		NewValue:   newValue,
		TraceID:    traceID,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
	}

	return r.Create(ctx, log)
}
