-- Migration: 009_import_job.sql
-- Description: Create import_job table for tracking import tasks
-- Created: 2026-01-22

CREATE TABLE IF NOT EXISTS import_job (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    type ENUM('FILE_UPLOAD', 'TEXT_INPUT', 'TEMPLATE_IMPORT', 'ADMIN_LLM_GENERATE') NOT NULL,
    status ENUM('PENDING', 'RUNNING', 'NEEDS_CONFIRMATION', 'SUCCEEDED', 'FAILED') NOT NULL DEFAULT 'PENDING',
    file_name VARCHAR(255) NULL,
    file_hash VARCHAR(64) NULL COMMENT 'SHA-256 hash of file',
    file_size BIGINT NULL,
    file_path VARCHAR(500) NULL COMMENT 'Storage path of uploaded file',
    raw_text LONGTEXT NULL,
    idempotency_key VARCHAR(100) NULL,
    model_version VARCHAR(50) NULL COMMENT 'LLM model version used',
    prompt_version VARCHAR(50) NULL COMMENT 'Prompt template version',
    result_summary JSON NULL COMMENT 'Summary of import results',
    error_message TEXT NULL,
    started_at TIMESTAMP NULL,
    completed_at TIMESTAMP NULL,
    duration_ms BIGINT GENERATED ALWAYS AS (
        CASE 
            WHEN started_at IS NOT NULL AND completed_at IS NOT NULL 
            THEN TIMESTAMPDIFF(MICROSECOND, started_at, completed_at) / 1000
            ELSE NULL
        END
    ) STORED,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by BIGINT UNSIGNED NOT NULL,
    
    PRIMARY KEY (id),
    UNIQUE KEY idx_import_job_idempotency (idempotency_key),
    INDEX idx_import_job_status (status),
    INDEX idx_import_job_type (type),
    INDEX idx_import_job_created (created_at),
    INDEX idx_import_job_created_by (created_by),
    CONSTRAINT fk_import_job_created_by FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Add foreign key from price_quote to import_job
ALTER TABLE price_quote 
    ADD CONSTRAINT fk_quote_import_job FOREIGN KEY (import_job_id) REFERENCES import_job(id) ON DELETE SET NULL;
