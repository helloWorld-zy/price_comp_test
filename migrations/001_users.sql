-- Migration: 001_users.sql
-- Description: Create users table for authentication
-- Created: 2026-01-22

CREATE TABLE IF NOT EXISTS users (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    username VARCHAR(50) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role ENUM('ADMIN', 'VENDOR') NOT NULL,
    supplier_id BIGINT UNSIGNED NULL,
    status ENUM('ACTIVE', 'INACTIVE') NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    PRIMARY KEY (id),
    UNIQUE KEY idx_users_username (username),
    INDEX idx_users_supplier (supplier_id),
    INDEX idx_users_role (role)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Insert default admin user (password: admin123)
-- Password hash generated with bcrypt cost 10
INSERT INTO users (username, password_hash, role, status) VALUES 
('admin', '$2a$10$rBvJWMRWHFZ3YYmJ4yGq6.K8.Vy3WU1v5qH1nKJ9mMHxkp5WJ1L0W', 'ADMIN', 'ACTIVE')
ON DUPLICATE KEY UPDATE username = username;
