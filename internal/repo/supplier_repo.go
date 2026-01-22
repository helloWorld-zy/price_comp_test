package repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"cruise-price-compare/internal/domain"
)

// SupplierRepository handles supplier data access
type SupplierRepository struct {
	db *DB
}

// NewSupplierRepository creates a new supplier repository
func NewSupplierRepository(db *DB) *SupplierRepository {
	return &SupplierRepository{db: db}
}

// GetByID retrieves a supplier by ID
func (r *SupplierRepository) GetByID(ctx context.Context, id uint64) (*domain.Supplier, error) {
	var row supplierRow
	query := `SELECT id, name, aliases, contact_info, visibility, status, created_at, updated_at, created_by 
              FROM supplier WHERE id = ?`

	if err := r.db.GetContext(ctx, &row, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get supplier by id: %w", err)
	}

	return row.toDomain(), nil
}

// List retrieves suppliers with pagination
func (r *SupplierRepository) List(ctx context.Context, pagination Pagination, status *domain.EntityStatus) (PaginatedResult[domain.Supplier], error) {
	var rows []supplierRow
	var total int64

	countQuery := "SELECT COUNT(*) FROM supplier WHERE 1=1"
	selectQuery := `SELECT id, name, aliases, contact_info, visibility, status, created_at, updated_at, created_by FROM supplier WHERE 1=1`
	var args []interface{}

	if status != nil {
		countQuery += " AND status = ?"
		selectQuery += " AND status = ?"
		args = append(args, *status)
	}

	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return PaginatedResult[domain.Supplier]{}, fmt.Errorf("failed to count suppliers: %w", err)
	}

	selectQuery += " ORDER BY name LIMIT ? OFFSET ?"
	args = append(args, pagination.Limit(), pagination.Offset())

	if err := r.db.SelectContext(ctx, &rows, selectQuery, args...); err != nil {
		return PaginatedResult[domain.Supplier]{}, fmt.Errorf("failed to list suppliers: %w", err)
	}

	items := make([]domain.Supplier, len(rows))
	for i, row := range rows {
		items[i] = *row.toDomain()
	}

	return NewPaginatedResult(items, total, pagination), nil
}

// ListAll retrieves all active suppliers
func (r *SupplierRepository) ListAll(ctx context.Context) ([]domain.Supplier, error) {
	var rows []supplierRow
	query := `SELECT id, name, aliases, contact_info, visibility, status, created_at, updated_at, created_by 
              FROM supplier WHERE status = 'ACTIVE' ORDER BY name`

	if err := r.db.SelectContext(ctx, &rows, query); err != nil {
		return nil, fmt.Errorf("failed to list all suppliers: %w", err)
	}

	items := make([]domain.Supplier, len(rows))
	for i, row := range rows {
		items[i] = *row.toDomain()
	}

	return items, nil
}

// Create creates a new supplier
func (r *SupplierRepository) Create(ctx context.Context, supplier *domain.Supplier) error {
	aliasesJSON, err := json.Marshal(supplier.Aliases)
	if err != nil {
		return fmt.Errorf("failed to marshal aliases: %w", err)
	}

	query := `INSERT INTO supplier (name, aliases, contact_info, visibility, status, created_by) 
              VALUES (?, ?, ?, ?, ?, ?)`

	result, err := r.db.ExecContext(ctx, query, supplier.Name, aliasesJSON, supplier.ContactInfo, supplier.Visibility, supplier.Status, supplier.CreatedBy)
	if err != nil {
		return fmt.Errorf("failed to create supplier: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	supplier.ID = uint64(id)

	return nil
}

// Update updates a supplier
func (r *SupplierRepository) Update(ctx context.Context, supplier *domain.Supplier) error {
	aliasesJSON, err := json.Marshal(supplier.Aliases)
	if err != nil {
		return fmt.Errorf("failed to marshal aliases: %w", err)
	}

	query := `UPDATE supplier SET name = ?, aliases = ?, contact_info = ?, visibility = ?, status = ? WHERE id = ?`

	_, err = r.db.ExecContext(ctx, query, supplier.Name, aliasesJSON, supplier.ContactInfo, supplier.Visibility, supplier.Status, supplier.ID)
	if err != nil {
		return fmt.Errorf("failed to update supplier: %w", err)
	}

	return nil
}

// Delete deletes a supplier
func (r *SupplierRepository) Delete(ctx context.Context, id uint64) error {
	query := `DELETE FROM supplier WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete supplier: %w", err)
	}

	return nil
}

// ExistsByName checks if a supplier name exists
func (r *SupplierRepository) ExistsByName(ctx context.Context, name string, excludeID *uint64) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM supplier WHERE name = ?`
	args := []interface{}{name}

	if excludeID != nil {
		query += " AND id != ?"
		args = append(args, *excludeID)
	}

	if err := r.db.GetContext(ctx, &count, query, args...); err != nil {
		return false, fmt.Errorf("failed to check supplier exists: %w", err)
	}

	return count > 0, nil
}

type supplierRow struct {
	ID          uint64         `db:"id"`
	Name        string         `db:"name"`
	Aliases     []byte         `db:"aliases"`
	ContactInfo sql.NullString `db:"contact_info"`
	Visibility  string         `db:"visibility"`
	Status      string         `db:"status"`
	CreatedAt   sql.NullTime   `db:"created_at"`
	UpdatedAt   sql.NullTime   `db:"updated_at"`
	CreatedBy   sql.NullInt64  `db:"created_by"`
}

func (r *supplierRow) toDomain() *domain.Supplier {
	s := &domain.Supplier{
		ID:         r.ID,
		Name:       r.Name,
		Visibility: domain.SupplierVisibility(r.Visibility),
		Status:     domain.EntityStatus(r.Status),
	}

	if r.Aliases != nil {
		_ = json.Unmarshal(r.Aliases, &s.Aliases)
	}

	if r.ContactInfo.Valid {
		s.ContactInfo = r.ContactInfo.String
	}

	if r.CreatedAt.Valid {
		s.CreatedAt = r.CreatedAt.Time
	}

	if r.UpdatedAt.Valid {
		s.UpdatedAt = r.UpdatedAt.Time
	}

	if r.CreatedBy.Valid {
		createdBy := uint64(r.CreatedBy.Int64)
		s.CreatedBy = &createdBy
	}

	return s
}
