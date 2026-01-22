package repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"cruise-price-compare/internal/domain"
)

// SailingRepository handles sailing data access
type SailingRepository struct {
	db *DB
}

// NewSailingRepository creates a new sailing repository
func NewSailingRepository(db *DB) *SailingRepository {
	return &SailingRepository{db: db}
}

// GetByID retrieves a sailing by ID
func (r *SailingRepository) GetByID(ctx context.Context, id uint64) (*domain.Sailing, error) {
	var row sailingRow
	query := `SELECT id, ship_id, sailing_code, departure_date, return_date, nights, route, ports, description, status, created_at, updated_at, created_by 
              FROM sailing WHERE id = ?`

	if err := r.db.GetContext(ctx, &row, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get sailing by id: %w", err)
	}

	return row.toDomain(), nil
}

// GetByCode retrieves a sailing by sailing code
func (r *SailingRepository) GetByCode(ctx context.Context, code string) (*domain.Sailing, error) {
	var row sailingRow
	query := `SELECT id, ship_id, sailing_code, departure_date, return_date, nights, route, ports, description, status, created_at, updated_at, created_by 
              FROM sailing WHERE sailing_code = ?`

	if err := r.db.GetContext(ctx, &row, query, code); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get sailing by code: %w", err)
	}

	return row.toDomain(), nil
}

// List retrieves sailings with pagination and filters
func (r *SailingRepository) List(ctx context.Context, pagination Pagination, shipID *uint64, status *domain.SailingStatus, fromDate, toDate *time.Time) (PaginatedResult[domain.Sailing], error) {
	var rows []sailingRow
	var total int64

	// Build query
	countQuery := "SELECT COUNT(*) FROM sailing WHERE 1=1"
	selectQuery := `SELECT id, ship_id, sailing_code, departure_date, return_date, nights, route, ports, description, status, created_at, updated_at, created_by FROM sailing WHERE 1=1`
	var args []interface{}

	if shipID != nil {
		countQuery += " AND ship_id = ?"
		selectQuery += " AND ship_id = ?"
		args = append(args, *shipID)
	}

	if status != nil {
		countQuery += " AND status = ?"
		selectQuery += " AND status = ?"
		args = append(args, *status)
	}

	if fromDate != nil {
		countQuery += " AND departure_date >= ?"
		selectQuery += " AND departure_date >= ?"
		args = append(args, *fromDate)
	}

	if toDate != nil {
		countQuery += " AND departure_date <= ?"
		selectQuery += " AND departure_date <= ?"
		args = append(args, *toDate)
	}

	// Count total
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return PaginatedResult[domain.Sailing]{}, fmt.Errorf("failed to count sailings: %w", err)
	}

	// Get paginated results
	selectQuery += " ORDER BY departure_date DESC LIMIT ? OFFSET ?"
	args = append(args, pagination.Limit(), pagination.Offset())

	if err := r.db.SelectContext(ctx, &rows, selectQuery, args...); err != nil {
		return PaginatedResult[domain.Sailing]{}, fmt.Errorf("failed to list sailings: %w", err)
	}

	items := make([]domain.Sailing, len(rows))
	for i, row := range rows {
		items[i] = *row.toDomain()
	}

	return NewPaginatedResult(items, total, pagination), nil
}

// ListByShip retrieves all sailings for a ship
func (r *SailingRepository) ListByShip(ctx context.Context, shipID uint64) ([]domain.Sailing, error) {
	var rows []sailingRow
	query := `SELECT id, ship_id, sailing_code, departure_date, return_date, nights, route, ports, description, status, created_at, updated_at, created_by 
              FROM sailing WHERE ship_id = ? AND status = 'ACTIVE' ORDER BY departure_date`

	if err := r.db.SelectContext(ctx, &rows, query, shipID); err != nil {
		return nil, fmt.Errorf("failed to list sailings by ship: %w", err)
	}

	items := make([]domain.Sailing, len(rows))
	for i, row := range rows {
		items[i] = *row.toDomain()
	}

	return items, nil
}

// ListUpcoming retrieves upcoming sailings
func (r *SailingRepository) ListUpcoming(ctx context.Context, limit int) ([]domain.Sailing, error) {
	var rows []sailingRow
	query := `SELECT id, ship_id, sailing_code, departure_date, return_date, nights, route, ports, description, status, created_at, updated_at, created_by 
              FROM sailing WHERE status = 'ACTIVE' AND departure_date >= CURDATE() ORDER BY departure_date LIMIT ?`

	if err := r.db.SelectContext(ctx, &rows, query, limit); err != nil {
		return nil, fmt.Errorf("failed to list upcoming sailings: %w", err)
	}

	items := make([]domain.Sailing, len(rows))
	for i, row := range rows {
		items[i] = *row.toDomain()
	}

	return items, nil
}

// Create creates a new sailing
func (r *SailingRepository) Create(ctx context.Context, sailing *domain.Sailing) error {
	portsJSON, err := json.Marshal(sailing.Ports)
	if err != nil {
		return fmt.Errorf("failed to marshal ports: %w", err)
	}

	query := `INSERT INTO sailing (ship_id, sailing_code, departure_date, return_date, route, ports, description, status, created_by) 
              VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db.ExecContext(ctx, query, sailing.ShipID, sailing.SailingCode, sailing.DepartureDate, sailing.ReturnDate,
		sailing.Route, portsJSON, sailing.Description, sailing.Status, sailing.CreatedBy)
	if err != nil {
		return fmt.Errorf("failed to create sailing: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	sailing.ID = uint64(id)

	return nil
}

// Update updates a sailing
func (r *SailingRepository) Update(ctx context.Context, sailing *domain.Sailing) error {
	portsJSON, err := json.Marshal(sailing.Ports)
	if err != nil {
		return fmt.Errorf("failed to marshal ports: %w", err)
	}

	query := `UPDATE sailing SET ship_id = ?, sailing_code = ?, departure_date = ?, return_date = ?, route = ?, ports = ?, description = ?, status = ? WHERE id = ?`

	_, err = r.db.ExecContext(ctx, query, sailing.ShipID, sailing.SailingCode, sailing.DepartureDate, sailing.ReturnDate,
		sailing.Route, portsJSON, sailing.Description, sailing.Status, sailing.ID)
	if err != nil {
		return fmt.Errorf("failed to update sailing: %w", err)
	}

	return nil
}

// Delete deletes a sailing
func (r *SailingRepository) Delete(ctx context.Context, id uint64) error {
	query := `DELETE FROM sailing WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete sailing: %w", err)
	}

	return nil
}

// ExistsByCode checks if a sailing code exists
func (r *SailingRepository) ExistsByCode(ctx context.Context, code string, excludeID *uint64) (bool, error) {
	if code == "" {
		return false, nil
	}

	var count int
	query := `SELECT COUNT(*) FROM sailing WHERE sailing_code = ?`
	args := []interface{}{code}

	if excludeID != nil {
		query += " AND id != ?"
		args = append(args, *excludeID)
	}

	if err := r.db.GetContext(ctx, &count, query, args...); err != nil {
		return false, fmt.Errorf("failed to check sailing code exists: %w", err)
	}

	return count > 0, nil
}

// sailingRow is the database row structure for sailing
type sailingRow struct {
	ID            uint64         `db:"id"`
	ShipID        uint64         `db:"ship_id"`
	SailingCode   sql.NullString `db:"sailing_code"`
	DepartureDate time.Time      `db:"departure_date"`
	ReturnDate    time.Time      `db:"return_date"`
	Nights        int            `db:"nights"`
	Route         string         `db:"route"`
	Ports         []byte         `db:"ports"`
	Description   sql.NullString `db:"description"`
	Status        string         `db:"status"`
	CreatedAt     sql.NullTime   `db:"created_at"`
	UpdatedAt     sql.NullTime   `db:"updated_at"`
	CreatedBy     sql.NullInt64  `db:"created_by"`
}

func (r *sailingRow) toDomain() *domain.Sailing {
	sailing := &domain.Sailing{
		ID:            r.ID,
		ShipID:        r.ShipID,
		DepartureDate: r.DepartureDate,
		ReturnDate:    r.ReturnDate,
		Nights:        r.Nights,
		Route:         r.Route,
		Status:        domain.SailingStatus(r.Status),
	}

	if r.SailingCode.Valid {
		sailing.SailingCode = r.SailingCode.String
	}

	if r.Ports != nil {
		_ = json.Unmarshal(r.Ports, &sailing.Ports)
	}

	if r.Description.Valid {
		sailing.Description = r.Description.String
	}

	if r.CreatedAt.Valid {
		sailing.CreatedAt = r.CreatedAt.Time
	}

	if r.UpdatedAt.Valid {
		sailing.UpdatedAt = r.UpdatedAt.Time
	}

	if r.CreatedBy.Valid {
		createdBy := uint64(r.CreatedBy.Int64)
		sailing.CreatedBy = &createdBy
	}

	return sailing
}
