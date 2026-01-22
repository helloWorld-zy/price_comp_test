-- Migration: 010_parse_job.sql
-- Description: Create parse_job table for LLM parsing tasks
-- Created: 2026-01-22

CREATE TABLE IF NOT EXISTS parse_job (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    import_job_id BIGINT UNSIGNED NOT NULL,
    status ENUM('PENDING', 'RUNNING', 'SUCCEEDED', 'FAILED') NOT NULL DEFAULT 'PENDING',
    parsed_data JSON NULL COMMENT 'Intermediate parsed structure from LLM',
    confidence DECIMAL(3, 2) NULL COMMENT 'Confidence score 0-1',
    warnings JSON NULL COMMENT 'Array of warning messages',
    page_info JSON NULL COMMENT 'Page extraction info for PDFs',
    error_message TEXT NULL,
    started_at TIMESTAMP NULL,
    completed_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    PRIMARY KEY (id),
    INDEX idx_parse_job_import (import_job_id),
    INDEX idx_parse_job_status (status),
    CONSTRAINT fk_parse_job_import FOREIGN KEY (import_job_id) REFERENCES import_job(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
