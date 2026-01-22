-- Migration: 004_cabin_category.sql
-- Description: Create cabin_category table for cabin type categories
-- Created: 2026-01-22

CREATE TABLE IF NOT EXISTS cabin_category (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    name VARCHAR(50) NOT NULL,
    name_en VARCHAR(50) NULL,
    sort_order INT NOT NULL DEFAULT 0,
    is_default TINYINT(1) NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    PRIMARY KEY (id),
    UNIQUE KEY idx_cabin_category_name (name),
    INDEX idx_cabin_category_sort (sort_order)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
