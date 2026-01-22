-- Migration: 011_audit_log.sql
-- Description: Create audit_log table for operation auditing
-- Created: 2026-01-22

CREATE TABLE IF NOT EXISTS audit_log (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id BIGINT UNSIGNED NOT NULL,
    supplier_id BIGINT UNSIGNED NULL,
    action VARCHAR(50) NOT NULL COMMENT 'Action type: CREATE, UPDATE, DELETE, LOGIN, etc.',
    entity_type VARCHAR(50) NOT NULL COMMENT 'Entity type: user, cruise_line, sailing, etc.',
    entity_id BIGINT UNSIGNED NOT NULL,
    old_value JSON NULL COMMENT 'Previous state (for updates/deletes)',
    new_value JSON NULL COMMENT 'New state (for creates/updates)',
    trace_id VARCHAR(36) NULL COMMENT 'Request trace ID for correlation',
    ip_address VARCHAR(45) NULL COMMENT 'Client IP address',
    user_agent VARCHAR(255) NULL COMMENT 'Client user agent',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    PRIMARY KEY (id),
    INDEX idx_audit_entity (entity_type, entity_id),
    INDEX idx_audit_user (user_id),
    INDEX idx_audit_action (action),
    INDEX idx_audit_created (created_at),
    INDEX idx_audit_trace (trace_id),
    CONSTRAINT fk_audit_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
