package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"cruise-price-compare/internal/domain"
)

// CabinCategoryRepository handles cabin category data access
type CabinCategoryRepository struct {
	db *DB
}

// NewCabinCategoryRepository creates a new cabin category repository
func NewCabinCategoryRepository(db *DB) *CabinCategoryRepository {
	return &CabinCategoryRepository{db: db}
}

// GetByID retrieves a cabin category by ID
func (r *CabinCategoryRepository) GetByID(ctx context.Context, id uint64) (*domain.CabinCategory, error) {
	var cc domain.CabinCategory
	query := `SELECT id, name, name_en, sort_order, is_default, created_at 
              FROM cabin_category WHERE id = ?`

	if err := r.db.GetContext(ctx, &cc, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get cabin category by id: %w", err)
	}

	return &cc, nil
}

// GetByName retrieves a cabin category by name
func (r *CabinCategoryRepository) GetByName(ctx context.Context, name string) (*domain.CabinCategory, error) {
	var cc domain.CabinCategory
	query := `SELECT id, name, name_en, sort_order, is_default, created_at 
              FROM cabin_category WHERE name = ?`

	if err := r.db.GetContext(ctx, &cc, query, name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get cabin category by name: %w", err)
	}

	return &cc, nil
}

// List retrieves all cabin categories
func (r *CabinCategoryRepository) List(ctx context.Context) ([]domain.CabinCategory, error) {
	var categories []domain.CabinCategory
	query := `SELECT id, name, name_en, sort_order, is_default, created_at 
              FROM cabin_category ORDER BY sort_order, name`

	if err := r.db.SelectContext(ctx, &categories, query); err != nil {
		return nil, fmt.Errorf("failed to list cabin categories: %w", err)
	}

	return categories, nil
}

// ListDefaults retrieves default cabin categories
func (r *CabinCategoryRepository) ListDefaults(ctx context.Context) ([]domain.CabinCategory, error) {
	var categories []domain.CabinCategory
	query := `SELECT id, name, name_en, sort_order, is_default, created_at 
              FROM cabin_category WHERE is_default = 1 ORDER BY sort_order`

	if err := r.db.SelectContext(ctx, &categories, query); err != nil {
		return nil, fmt.Errorf("failed to list default cabin categories: %w", err)
	}

	return categories, nil
}

// Create creates a new cabin category
func (r *CabinCategoryRepository) Create(ctx context.Context, cc *domain.CabinCategory) error {
	query := `INSERT INTO cabin_category (name, name_en, sort_order, is_default) 
              VALUES (?, ?, ?, ?)`

	result, err := r.db.ExecContext(ctx, query, cc.Name, cc.NameEN, cc.SortOrder, cc.IsDefault)
	if err != nil {
		return fmt.Errorf("failed to create cabin category: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	cc.ID = uint64(id)

	return nil
}

// Update updates a cabin category
func (r *CabinCategoryRepository) Update(ctx context.Context, cc *domain.CabinCategory) error {
	query := `UPDATE cabin_category SET name = ?, name_en = ?, sort_order = ?, is_default = ? WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, cc.Name, cc.NameEN, cc.SortOrder, cc.IsDefault, cc.ID)
	if err != nil {
		return fmt.Errorf("failed to update cabin category: %w", err)
	}

	return nil
}

// Delete deletes a cabin category
func (r *CabinCategoryRepository) Delete(ctx context.Context, id uint64) error {
	query := `DELETE FROM cabin_category WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete cabin category: %w", err)
	}

	return nil
}

// ExistsByName checks if a cabin category name exists
func (r *CabinCategoryRepository) ExistsByName(ctx context.Context, name string, excludeID *uint64) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM cabin_category WHERE name = ?`
	args := []interface{}{name}

	if excludeID != nil {
		query += " AND id != ?"
		args = append(args, *excludeID)
	}

	if err := r.db.GetContext(ctx, &count, query, args...); err != nil {
		return false, fmt.Errorf("failed to check cabin category exists: %w", err)
	}

	return count > 0, nil
}
