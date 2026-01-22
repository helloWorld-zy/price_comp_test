-- Migration: 008_price_quote.sql
-- Description: Create price_quote table for quote records (append-only)
-- Created: 2026-01-22

CREATE TABLE IF NOT EXISTS price_quote (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    sailing_id BIGINT UNSIGNED NOT NULL,
    cabin_type_id BIGINT UNSIGNED NOT NULL,
    supplier_id BIGINT UNSIGNED NOT NULL,
    price DECIMAL(12, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'CNY',
    pricing_unit ENUM('PER_PERSON', 'PER_CABIN', 'TOTAL') NOT NULL,
    conditions TEXT NULL,
    guest_count INT NULL,
    promotion TEXT NULL,
    cabin_quantity INT NULL,
    valid_until DATE NULL,
    notes TEXT NULL,
    source ENUM('MANUAL', 'FILE_IMPORT', 'TEXT_IMPORT', 'TEMPLATE_IMPORT') NOT NULL,
    source_ref VARCHAR(255) NULL COMMENT 'Reference to source (file ID, text snippet)',
    import_job_id BIGINT UNSIGNED NULL,
    status ENUM('ACTIVE', 'VOIDED', 'CORRECTED') NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by BIGINT UNSIGNED NOT NULL,
    
    PRIMARY KEY (id),
    INDEX idx_quote_sailing_cabin_supplier (sailing_id, cabin_type_id, supplier_id),
    INDEX idx_quote_supplier_created (supplier_id, created_at),
    INDEX idx_quote_sailing (sailing_id),
    INDEX idx_quote_status (status),
    INDEX idx_quote_created (created_at),
    CONSTRAINT fk_quote_sailing FOREIGN KEY (sailing_id) REFERENCES sailing(id) ON DELETE RESTRICT,
    CONSTRAINT fk_quote_cabin_type FOREIGN KEY (cabin_type_id) REFERENCES cabin_type(id) ON DELETE RESTRICT,
    CONSTRAINT fk_quote_supplier FOREIGN KEY (supplier_id) REFERENCES supplier(id) ON DELETE RESTRICT,
    CONSTRAINT fk_quote_created_by FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE RESTRICT,
    CONSTRAINT chk_quote_price CHECK (price > 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
