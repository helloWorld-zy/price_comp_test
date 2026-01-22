-- Migration: 003_ship.sql
-- Description: Create ship table for cruise ships
-- Created: 2026-01-22

CREATE TABLE IF NOT EXISTS ship (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    cruise_line_id BIGINT UNSIGNED NOT NULL,
    name VARCHAR(100) NOT NULL,
    aliases JSON NULL COMMENT 'Array of alias strings for LLM matching',
    status ENUM('ACTIVE', 'INACTIVE') NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    created_by BIGINT UNSIGNED NULL,
    
    PRIMARY KEY (id),
    UNIQUE KEY idx_ship_line_name (cruise_line_id, name),
    INDEX idx_ship_cruise_line (cruise_line_id),
    INDEX idx_ship_status (status),
    CONSTRAINT fk_ship_cruise_line FOREIGN KEY (cruise_line_id) REFERENCES cruise_line(id) ON DELETE RESTRICT,
    CONSTRAINT fk_ship_created_by FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
