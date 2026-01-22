package obs

import (
	"context"
	"encoding/json"
	"time"

	"cruise-price-compare/internal/domain"
	"cruise-price-compare/internal/repo"

	"github.com/gin-gonic/gin"
)

// AuditService handles audit logging
type AuditService struct {
	repo   *repo.AuditLogRepository
	logger *Logger
}

// NewAuditService creates a new audit service
func NewAuditService(repo *repo.AuditLogRepository, logger *Logger) *AuditService {
	return &AuditService{
		repo:   repo,
		logger: logger,
	}
}

// LogCreate logs a create action
func (s *AuditService) LogCreate(ctx context.Context, userID uint64, supplierID *uint64, entityType string, entityID uint64, entity interface{}) error {
	return s.log(ctx, userID, supplierID, domain.AuditActionCreate, entityType, entityID, nil, entity)
}

// LogUpdate logs an update action
func (s *AuditService) LogUpdate(ctx context.Context, userID uint64, supplierID *uint64, entityType string, entityID uint64, oldEntity, newEntity interface{}) error {
	return s.log(ctx, userID, supplierID, domain.AuditActionUpdate, entityType, entityID, oldEntity, newEntity)
}

// LogDelete logs a delete action
func (s *AuditService) LogDelete(ctx context.Context, userID uint64, supplierID *uint64, entityType string, entityID uint64, entity interface{}) error {
	return s.log(ctx, userID, supplierID, domain.AuditActionDelete, entityType, entityID, entity, nil)
}

// LogLogin logs a login action
func (s *AuditService) LogLogin(ctx context.Context, userID uint64) error {
	return s.log(ctx, userID, nil, domain.AuditActionLogin, domain.EntityTypeUser, userID, nil, nil)
}

// LogVoid logs a void action
func (s *AuditService) LogVoid(ctx context.Context, userID uint64, supplierID *uint64, entityType string, entityID uint64, entity interface{}) error {
	return s.log(ctx, userID, supplierID, domain.AuditActionVoid, entityType, entityID, entity, nil)
}

// LogImport logs an import action
func (s *AuditService) LogImport(ctx context.Context, userID uint64, supplierID *uint64, entityID uint64, summary interface{}) error {
	return s.log(ctx, userID, supplierID, domain.AuditActionImport, domain.EntityTypeImportJob, entityID, nil, summary)
}

// log creates an audit log entry
func (s *AuditService) log(ctx context.Context, userID uint64, supplierID *uint64, action domain.AuditAction, entityType string, entityID uint64, oldEntity, newEntity interface{}) error {
	var oldValue, newValue json.RawMessage
	var err error

	if oldEntity != nil {
		oldValue, err = json.Marshal(oldEntity)
		if err != nil {
			s.logger.WithContext(ctx).WithError(err).Error("failed to marshal old entity for audit")
		}
	}

	if newEntity != nil {
		newValue, err = json.Marshal(newEntity)
		if err != nil {
			s.logger.WithContext(ctx).WithError(err).Error("failed to marshal new entity for audit")
		}
	}

	log := &domain.AuditLog{
		UserID:     userID,
		SupplierID: supplierID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		OldValue:   oldValue,
		NewValue:   newValue,
		TraceID:    GetTraceIDFromContext(ctx),
		CreatedAt:  time.Now(),
	}

	if err := s.repo.Create(ctx, log); err != nil {
		s.logger.WithContext(ctx).WithError(err).Error("failed to create audit log")
		return err
	}

	return nil
}

// LogFromGinContext logs an action using gin context for additional info
func (s *AuditService) LogFromGinContext(c *gin.Context, userID uint64, supplierID *uint64, action domain.AuditAction, entityType string, entityID uint64, oldEntity, newEntity interface{}) error {
	var oldValue, newValue json.RawMessage
	var err error

	if oldEntity != nil {
		oldValue, err = json.Marshal(oldEntity)
		if err != nil {
			s.logger.WithContext(c.Request.Context()).WithError(err).Error("failed to marshal old entity for audit")
		}
	}

	if newEntity != nil {
		newValue, err = json.Marshal(newEntity)
		if err != nil {
			s.logger.WithContext(c.Request.Context()).WithError(err).Error("failed to marshal new entity for audit")
		}
	}

	log := &domain.AuditLog{
		UserID:     userID,
		SupplierID: supplierID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		OldValue:   oldValue,
		NewValue:   newValue,
		TraceID:    GetTraceID(c),
		IPAddress:  c.ClientIP(),
		UserAgent:  c.GetHeader("User-Agent"),
		CreatedAt:  time.Now(),
	}

	if err := s.repo.Create(c.Request.Context(), log); err != nil {
		s.logger.WithContext(c.Request.Context()).WithError(err).Error("failed to create audit log")
		return err
	}

	return nil
}
