package repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"cruise-price-compare/internal/domain"
)

// CruiseLineRepository handles cruise line data access
type CruiseLineRepository struct {
	db *DB
}

// NewCruiseLineRepository creates a new cruise line repository
func NewCruiseLineRepository(db *DB) *CruiseLineRepository {
	return &CruiseLineRepository{db: db}
}

// GetByID retrieves a cruise line by ID
func (r *CruiseLineRepository) GetByID(ctx context.Context, id uint64) (*domain.CruiseLine, error) {
	var cl cruiseLineRow
	query := `SELECT id, name, name_en, aliases, status, created_at, updated_at, created_by 
              FROM cruise_line WHERE id = ?`

	if err := r.db.GetContext(ctx, &cl, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get cruise line by id: %w", err)
	}

	return cl.toDomain(), nil
}

// GetByName retrieves a cruise line by name
func (r *CruiseLineRepository) GetByName(ctx context.Context, name string) (*domain.CruiseLine, error) {
	var cl cruiseLineRow
	query := `SELECT id, name, name_en, aliases, status, created_at, updated_at, created_by 
              FROM cruise_line WHERE name = ?`

	if err := r.db.GetContext(ctx, &cl, query, name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get cruise line by name: %w", err)
	}

	return cl.toDomain(), nil
}

// List retrieves all cruise lines with pagination
func (r *CruiseLineRepository) List(ctx context.Context, pagination Pagination, status *domain.EntityStatus) (PaginatedResult[domain.CruiseLine], error) {
	var rows []cruiseLineRow
	var total int64

	// Build query
	countQuery := "SELECT COUNT(*) FROM cruise_line"
	selectQuery := `SELECT id, name, name_en, aliases, status, created_at, updated_at, created_by FROM cruise_line`
	var args []interface{}

	if status != nil {
		countQuery += " WHERE status = ?"
		selectQuery += " WHERE status = ?"
		args = append(args, *status)
	}

	// Count total
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return PaginatedResult[domain.CruiseLine]{}, fmt.Errorf("failed to count cruise lines: %w", err)
	}

	// Get paginated results
	selectQuery += " ORDER BY name LIMIT ? OFFSET ?"
	args = append(args, pagination.Limit(), pagination.Offset())

	if err := r.db.SelectContext(ctx, &rows, selectQuery, args...); err != nil {
		return PaginatedResult[domain.CruiseLine]{}, fmt.Errorf("failed to list cruise lines: %w", err)
	}

	items := make([]domain.CruiseLine, len(rows))
	for i, row := range rows {
		items[i] = *row.toDomain()
	}

	return NewPaginatedResult(items, total, pagination), nil
}

// ListAll retrieves all active cruise lines
func (r *CruiseLineRepository) ListAll(ctx context.Context) ([]domain.CruiseLine, error) {
	var rows []cruiseLineRow
	query := `SELECT id, name, name_en, aliases, status, created_at, updated_at, created_by 
              FROM cruise_line WHERE status = 'ACTIVE' ORDER BY name`

	if err := r.db.SelectContext(ctx, &rows, query); err != nil {
		return nil, fmt.Errorf("failed to list all cruise lines: %w", err)
	}

	items := make([]domain.CruiseLine, len(rows))
	for i, row := range rows {
		items[i] = *row.toDomain()
	}

	return items, nil
}

// Create creates a new cruise line
func (r *CruiseLineRepository) Create(ctx context.Context, cl *domain.CruiseLine) error {
	aliasesJSON, err := json.Marshal(cl.Aliases)
	if err != nil {
		return fmt.Errorf("failed to marshal aliases: %w", err)
	}

	query := `INSERT INTO cruise_line (name, name_en, aliases, status, created_by) 
              VALUES (?, ?, ?, ?, ?)`

	result, err := r.db.ExecContext(ctx, query, cl.Name, cl.NameEN, aliasesJSON, cl.Status, cl.CreatedBy)
	if err != nil {
		return fmt.Errorf("failed to create cruise line: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	cl.ID = uint64(id)

	return nil
}

// Update updates a cruise line
func (r *CruiseLineRepository) Update(ctx context.Context, cl *domain.CruiseLine) error {
	aliasesJSON, err := json.Marshal(cl.Aliases)
	if err != nil {
		return fmt.Errorf("failed to marshal aliases: %w", err)
	}

	query := `UPDATE cruise_line SET name = ?, name_en = ?, aliases = ?, status = ? WHERE id = ?`

	_, err = r.db.ExecContext(ctx, query, cl.Name, cl.NameEN, aliasesJSON, cl.Status, cl.ID)
	if err != nil {
		return fmt.Errorf("failed to update cruise line: %w", err)
	}

	return nil
}

// Delete deletes a cruise line
func (r *CruiseLineRepository) Delete(ctx context.Context, id uint64) error {
	query := `DELETE FROM cruise_line WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete cruise line: %w", err)
	}

	return nil
}

// ExistsByName checks if a cruise line name exists
func (r *CruiseLineRepository) ExistsByName(ctx context.Context, name string, excludeID *uint64) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM cruise_line WHERE name = ?`
	args := []interface{}{name}

	if excludeID != nil {
		query += " AND id != ?"
		args = append(args, *excludeID)
	}

	if err := r.db.GetContext(ctx, &count, query, args...); err != nil {
		return false, fmt.Errorf("failed to check cruise line exists: %w", err)
	}

	return count > 0, nil
}

// cruiseLineRow is the database row structure for cruise_line
type cruiseLineRow struct {
	ID        uint64         `db:"id"`
	Name      string         `db:"name"`
	NameEN    sql.NullString `db:"name_en"`
	Aliases   []byte         `db:"aliases"`
	Status    string         `db:"status"`
	CreatedAt sql.NullTime   `db:"created_at"`
	UpdatedAt sql.NullTime   `db:"updated_at"`
	CreatedBy sql.NullInt64  `db:"created_by"`
}

func (r *cruiseLineRow) toDomain() *domain.CruiseLine {
	cl := &domain.CruiseLine{
		ID:     r.ID,
		Name:   r.Name,
		Status: domain.EntityStatus(r.Status),
	}

	if r.NameEN.Valid {
		cl.NameEN = r.NameEN.String
	}

	if r.Aliases != nil {
		_ = json.Unmarshal(r.Aliases, &cl.Aliases)
	}

	if r.CreatedAt.Valid {
		cl.CreatedAt = r.CreatedAt.Time
	}

	if r.UpdatedAt.Valid {
		cl.UpdatedAt = r.UpdatedAt.Time
	}

	if r.CreatedBy.Valid {
		createdBy := uint64(r.CreatedBy.Int64)
		cl.CreatedBy = &createdBy
	}

	return cl
}
