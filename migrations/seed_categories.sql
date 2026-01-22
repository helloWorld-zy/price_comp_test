-- Seed: seed_categories.sql
-- Description: Insert default cabin categories
-- Created: 2026-01-22

INSERT INTO cabin_category (name, name_en, sort_order, is_default) VALUES
('内舱', 'Interior', 1, 1),
('海景', 'Ocean View', 2, 1),
('阳台', 'Balcony', 3, 1),
('套房', 'Suite', 4, 1)
ON DUPLICATE KEY UPDATE 
    name_en = VALUES(name_en),
    sort_order = VALUES(sort_order),
    is_default = VALUES(is_default);
