-- Migration: 006_sailing.sql
-- Description: Create sailing table for cruise sailings/voyages
-- Created: 2026-01-22

CREATE TABLE IF NOT EXISTS sailing (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    ship_id BIGINT UNSIGNED NOT NULL,
    sailing_code VARCHAR(50) NULL,
    departure_date DATE NOT NULL,
    return_date DATE NOT NULL,
    nights INT GENERATED ALWAYS AS (DATEDIFF(return_date, departure_date)) STORED,
    route VARCHAR(255) NOT NULL,
    ports JSON NULL COMMENT 'Array of port names',
    description TEXT NULL,
    status ENUM('ACTIVE', 'CANCELLED') NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    created_by BIGINT UNSIGNED NULL,
    
    PRIMARY KEY (id),
    UNIQUE KEY idx_sailing_code (sailing_code),
    INDEX idx_sailing_ship_date (ship_id, departure_date),
    INDEX idx_sailing_status (status),
    INDEX idx_sailing_departure (departure_date),
    CONSTRAINT fk_sailing_ship FOREIGN KEY (ship_id) REFERENCES ship(id) ON DELETE RESTRICT,
    CONSTRAINT fk_sailing_created_by FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT chk_sailing_dates CHECK (departure_date < return_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
