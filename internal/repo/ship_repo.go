package repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"cruise-price-compare/internal/domain"
)

// ShipRepository handles ship data access
type ShipRepository struct {
	db *DB
}

// NewShipRepository creates a new ship repository
func NewShipRepository(db *DB) *ShipRepository {
	return &ShipRepository{db: db}
}

// GetByID retrieves a ship by ID
func (r *ShipRepository) GetByID(ctx context.Context, id uint64) (*domain.Ship, error) {
	var row shipRow
	query := `SELECT id, cruise_line_id, name, aliases, status, created_at, updated_at, created_by 
              FROM ship WHERE id = ?`

	if err := r.db.GetContext(ctx, &row, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get ship by id: %w", err)
	}

	return row.toDomain(), nil
}

// List retrieves ships with pagination
func (r *ShipRepository) List(ctx context.Context, pagination Pagination, cruiseLineID *uint64, status *domain.EntityStatus) (PaginatedResult[domain.Ship], error) {
	var rows []shipRow
	var total int64

	// Build query
	countQuery := "SELECT COUNT(*) FROM ship WHERE 1=1"
	selectQuery := `SELECT id, cruise_line_id, name, aliases, status, created_at, updated_at, created_by FROM ship WHERE 1=1`
	var args []interface{}

	if cruiseLineID != nil {
		countQuery += " AND cruise_line_id = ?"
		selectQuery += " AND cruise_line_id = ?"
		args = append(args, *cruiseLineID)
	}

	if status != nil {
		countQuery += " AND status = ?"
		selectQuery += " AND status = ?"
		args = append(args, *status)
	}

	// Count total
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return PaginatedResult[domain.Ship]{}, fmt.Errorf("failed to count ships: %w", err)
	}

	// Get paginated results
	selectQuery += " ORDER BY name LIMIT ? OFFSET ?"
	args = append(args, pagination.Limit(), pagination.Offset())

	if err := r.db.SelectContext(ctx, &rows, selectQuery, args...); err != nil {
		return PaginatedResult[domain.Ship]{}, fmt.Errorf("failed to list ships: %w", err)
	}

	items := make([]domain.Ship, len(rows))
	for i, row := range rows {
		items[i] = *row.toDomain()
	}

	return NewPaginatedResult(items, total, pagination), nil
}

// ListByCruiseLine retrieves all ships for a cruise line
func (r *ShipRepository) ListByCruiseLine(ctx context.Context, cruiseLineID uint64) ([]domain.Ship, error) {
	var rows []shipRow
	query := `SELECT id, cruise_line_id, name, aliases, status, created_at, updated_at, created_by 
              FROM ship WHERE cruise_line_id = ? AND status = 'ACTIVE' ORDER BY name`

	if err := r.db.SelectContext(ctx, &rows, query, cruiseLineID); err != nil {
		return nil, fmt.Errorf("failed to list ships by cruise line: %w", err)
	}

	items := make([]domain.Ship, len(rows))
	for i, row := range rows {
		items[i] = *row.toDomain()
	}

	return items, nil
}

// Create creates a new ship
func (r *ShipRepository) Create(ctx context.Context, ship *domain.Ship) error {
	aliasesJSON, err := json.Marshal(ship.Aliases)
	if err != nil {
		return fmt.Errorf("failed to marshal aliases: %w", err)
	}

	query := `INSERT INTO ship (cruise_line_id, name, aliases, status, created_by) 
              VALUES (?, ?, ?, ?, ?)`

	result, err := r.db.ExecContext(ctx, query, ship.CruiseLineID, ship.Name, aliasesJSON, ship.Status, ship.CreatedBy)
	if err != nil {
		return fmt.Errorf("failed to create ship: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	ship.ID = uint64(id)

	return nil
}

// Update updates a ship
func (r *ShipRepository) Update(ctx context.Context, ship *domain.Ship) error {
	aliasesJSON, err := json.Marshal(ship.Aliases)
	if err != nil {
		return fmt.Errorf("failed to marshal aliases: %w", err)
	}

	query := `UPDATE ship SET cruise_line_id = ?, name = ?, aliases = ?, status = ? WHERE id = ?`

	_, err = r.db.ExecContext(ctx, query, ship.CruiseLineID, ship.Name, aliasesJSON, ship.Status, ship.ID)
	if err != nil {
		return fmt.Errorf("failed to update ship: %w", err)
	}

	return nil
}

// Delete deletes a ship
func (r *ShipRepository) Delete(ctx context.Context, id uint64) error {
	query := `DELETE FROM ship WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete ship: %w", err)
	}

	return nil
}

// ExistsByName checks if a ship name exists for a cruise line
func (r *ShipRepository) ExistsByName(ctx context.Context, cruiseLineID uint64, name string, excludeID *uint64) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM ship WHERE cruise_line_id = ? AND name = ?`
	args := []interface{}{cruiseLineID, name}

	if excludeID != nil {
		query += " AND id != ?"
		args = append(args, *excludeID)
	}

	if err := r.db.GetContext(ctx, &count, query, args...); err != nil {
		return false, fmt.Errorf("failed to check ship exists: %w", err)
	}

	return count > 0, nil
}

// shipRow is the database row structure for ship
type shipRow struct {
	ID           uint64        `db:"id"`
	CruiseLineID uint64        `db:"cruise_line_id"`
	Name         string        `db:"name"`
	Aliases      []byte        `db:"aliases"`
	Status       string        `db:"status"`
	CreatedAt    sql.NullTime  `db:"created_at"`
	UpdatedAt    sql.NullTime  `db:"updated_at"`
	CreatedBy    sql.NullInt64 `db:"created_by"`
}

func (r *shipRow) toDomain() *domain.Ship {
	ship := &domain.Ship{
		ID:           r.ID,
		CruiseLineID: r.CruiseLineID,
		Name:         r.Name,
		Status:       domain.EntityStatus(r.Status),
	}

	if r.Aliases != nil {
		_ = json.Unmarshal(r.Aliases, &ship.Aliases)
	}

	if r.CreatedAt.Valid {
		ship.CreatedAt = r.CreatedAt.Time
	}

	if r.UpdatedAt.Valid {
		ship.UpdatedAt = r.UpdatedAt.Time
	}

	if r.CreatedBy.Valid {
		createdBy := uint64(r.CreatedBy.Int64)
		ship.CreatedBy = &createdBy
	}

	return ship
}
