-- Migration: 007_supplier.sql
-- Description: Create supplier table for price quote suppliers
-- Created: 2026-01-22

CREATE TABLE IF NOT EXISTS supplier (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL,
    aliases JSON NULL COMMENT 'Array of alias strings for matching',
    contact_info TEXT NULL,
    visibility ENUM('PRIVATE', 'PUBLIC') NOT NULL DEFAULT 'PRIVATE',
    status ENUM('ACTIVE', 'INACTIVE') NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    created_by BIGINT UNSIGNED NULL,
    
    PRIMARY KEY (id),
    UNIQUE KEY idx_supplier_name (name),
    INDEX idx_supplier_status (status),
    INDEX idx_supplier_visibility (visibility),
    CONSTRAINT fk_supplier_created_by FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Add supplier_id foreign key to users table
ALTER TABLE users 
    ADD CONSTRAINT fk_users_supplier FOREIGN KEY (supplier_id) REFERENCES supplier(id) ON DELETE SET NULL;
