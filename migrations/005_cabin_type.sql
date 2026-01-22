-- Migration: 005_cabin_type.sql
-- Description: Create cabin_type table for specific cabin types per ship
-- Created: 2026-01-22

CREATE TABLE IF NOT EXISTS cabin_type (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    ship_id BIGINT UNSIGNED NOT NULL,
    category_id BIGINT UNSIGNED NOT NULL,
    name VARCHAR(100) NOT NULL,
    code VARCHAR(20) NULL,
    description TEXT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    is_enabled TINYINT(1) NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    PRIMARY KEY (id),
    UNIQUE KEY idx_cabin_type_ship_cat_name (ship_id, category_id, name),
    UNIQUE KEY idx_cabin_type_ship_code (ship_id, code),
    INDEX idx_cabin_type_ship_cat (ship_id, category_id),
    INDEX idx_cabin_type_enabled (is_enabled),
    CONSTRAINT fk_cabin_type_ship FOREIGN KEY (ship_id) REFERENCES ship(id) ON DELETE CASCADE,
    CONSTRAINT fk_cabin_type_category FOREIGN KEY (category_id) REFERENCES cabin_category(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
