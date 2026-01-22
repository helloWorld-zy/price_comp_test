-- Migration: 002_cruise_line.sql
-- Description: Create cruise_line table for cruise companies
-- Created: 2026-01-22

CREATE TABLE IF NOT EXISTS cruise_line (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL,
    name_en VARCHAR(100) NULL,
    aliases JSON NULL COMMENT 'Array of alias strings for LLM matching',
    status ENUM('ACTIVE', 'INACTIVE') NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    created_by BIGINT UNSIGNED NULL,
    
    PRIMARY KEY (id),
    UNIQUE KEY idx_cruise_line_name (name),
    INDEX idx_cruise_line_status (status),
    CONSTRAINT fk_cruise_line_created_by FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
