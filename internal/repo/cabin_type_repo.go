package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"cruise-price-compare/internal/domain"
)

// CabinTypeRepository handles cabin type data access
type CabinTypeRepository struct {
	db *DB
}

// NewCabinTypeRepository creates a new cabin type repository
func NewCabinTypeRepository(db *DB) *CabinTypeRepository {
	return &CabinTypeRepository{db: db}
}

// GetByID retrieves a cabin type by ID
func (r *CabinTypeRepository) GetByID(ctx context.Context, id uint64) (*domain.CabinType, error) {
	var ct domain.CabinType
	query := `SELECT id, ship_id, category_id, name, code, description, sort_order, is_enabled, created_at, updated_at 
              FROM cabin_type WHERE id = ?`

	if err := r.db.GetContext(ctx, &ct, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get cabin type by id: %w", err)
	}

	return &ct, nil
}

// List retrieves cabin types with pagination
func (r *CabinTypeRepository) List(ctx context.Context, pagination Pagination, shipID *uint64, categoryID *uint64, enabledOnly bool) (PaginatedResult[domain.CabinType], error) {
	var cabinTypes []domain.CabinType
	var total int64

	// Build query
	countQuery := "SELECT COUNT(*) FROM cabin_type WHERE 1=1"
	selectQuery := `SELECT id, ship_id, category_id, name, code, description, sort_order, is_enabled, created_at, updated_at FROM cabin_type WHERE 1=1`
	var args []interface{}

	if shipID != nil {
		countQuery += " AND ship_id = ?"
		selectQuery += " AND ship_id = ?"
		args = append(args, *shipID)
	}

	if categoryID != nil {
		countQuery += " AND category_id = ?"
		selectQuery += " AND category_id = ?"
		args = append(args, *categoryID)
	}

	if enabledOnly {
		countQuery += " AND is_enabled = 1"
		selectQuery += " AND is_enabled = 1"
	}

	// Count total
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return PaginatedResult[domain.CabinType]{}, fmt.Errorf("failed to count cabin types: %w", err)
	}

	// Get paginated results
	selectQuery += " ORDER BY category_id, sort_order, name LIMIT ? OFFSET ?"
	args = append(args, pagination.Limit(), pagination.Offset())

	if err := r.db.SelectContext(ctx, &cabinTypes, selectQuery, args...); err != nil {
		return PaginatedResult[domain.CabinType]{}, fmt.Errorf("failed to list cabin types: %w", err)
	}

	return NewPaginatedResult(cabinTypes, total, pagination), nil
}

// ListByShip retrieves all cabin types for a ship
func (r *CabinTypeRepository) ListByShip(ctx context.Context, shipID uint64) ([]domain.CabinType, error) {
	var cabinTypes []domain.CabinType
	query := `SELECT id, ship_id, category_id, name, code, description, sort_order, is_enabled, created_at, updated_at 
              FROM cabin_type WHERE ship_id = ? AND is_enabled = 1 ORDER BY category_id, sort_order, name`

	if err := r.db.SelectContext(ctx, &cabinTypes, query, shipID); err != nil {
		return nil, fmt.Errorf("failed to list cabin types by ship: %w", err)
	}

	return cabinTypes, nil
}

// ListByShipAndCategory retrieves cabin types for a ship and category
func (r *CabinTypeRepository) ListByShipAndCategory(ctx context.Context, shipID, categoryID uint64) ([]domain.CabinType, error) {
	var cabinTypes []domain.CabinType
	query := `SELECT id, ship_id, category_id, name, code, description, sort_order, is_enabled, created_at, updated_at 
              FROM cabin_type WHERE ship_id = ? AND category_id = ? AND is_enabled = 1 ORDER BY sort_order, name`

	if err := r.db.SelectContext(ctx, &cabinTypes, query, shipID, categoryID); err != nil {
		return nil, fmt.Errorf("failed to list cabin types: %w", err)
	}

	return cabinTypes, nil
}

// Create creates a new cabin type
func (r *CabinTypeRepository) Create(ctx context.Context, ct *domain.CabinType) error {
	query := `INSERT INTO cabin_type (ship_id, category_id, name, code, description, sort_order, is_enabled) 
              VALUES (?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db.ExecContext(ctx, query, ct.ShipID, ct.CategoryID, ct.Name, ct.Code, ct.Description, ct.SortOrder, ct.IsEnabled)
	if err != nil {
		return fmt.Errorf("failed to create cabin type: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	ct.ID = uint64(id)

	return nil
}

// Update updates a cabin type
func (r *CabinTypeRepository) Update(ctx context.Context, ct *domain.CabinType) error {
	query := `UPDATE cabin_type SET ship_id = ?, category_id = ?, name = ?, code = ?, description = ?, sort_order = ?, is_enabled = ? WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, ct.ShipID, ct.CategoryID, ct.Name, ct.Code, ct.Description, ct.SortOrder, ct.IsEnabled, ct.ID)
	if err != nil {
		return fmt.Errorf("failed to update cabin type: %w", err)
	}

	return nil
}

// Delete deletes a cabin type
func (r *CabinTypeRepository) Delete(ctx context.Context, id uint64) error {
	query := `DELETE FROM cabin_type WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete cabin type: %w", err)
	}

	return nil
}

// ExistsByName checks if a cabin type name exists for a ship and category
func (r *CabinTypeRepository) ExistsByName(ctx context.Context, shipID, categoryID uint64, name string, excludeID *uint64) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM cabin_type WHERE ship_id = ? AND category_id = ? AND name = ?`
	args := []interface{}{shipID, categoryID, name}

	if excludeID != nil {
		query += " AND id != ?"
		args = append(args, *excludeID)
	}

	if err := r.db.GetContext(ctx, &count, query, args...); err != nil {
		return false, fmt.Errorf("failed to check cabin type exists: %w", err)
	}

	return count > 0, nil
}

// ExistsByCode checks if a cabin type code exists for a ship
func (r *CabinTypeRepository) ExistsByCode(ctx context.Context, shipID uint64, code string, excludeID *uint64) (bool, error) {
	if code == "" {
		return false, nil
	}

	var count int
	query := `SELECT COUNT(*) FROM cabin_type WHERE ship_id = ? AND code = ?`
	args := []interface{}{shipID, code}

	if excludeID != nil {
		query += " AND id != ?"
		args = append(args, *excludeID)
	}

	if err := r.db.GetContext(ctx, &count, query, args...); err != nil {
		return false, fmt.Errorf("failed to check cabin type code exists: %w", err)
	}

	return count > 0, nil
}
