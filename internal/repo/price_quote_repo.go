package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"cruise-price-compare/internal/domain"

	"github.com/shopspring/decimal"
)

// PriceQuoteRepository handles price quote data access
type PriceQuoteRepository struct {
	db *DB
}

// NewPriceQuoteRepository creates a new price quote repository
func NewPriceQuoteRepository(db *DB) *PriceQuoteRepository {
	return &PriceQuoteRepository{db: db}
}

// GetByID retrieves a price quote by ID
func (r *PriceQuoteRepository) GetByID(ctx context.Context, id uint64) (*domain.PriceQuote, error) {
	var pq domain.PriceQuote
	query := `SELECT id, sailing_id, cabin_type_id, supplier_id, price, currency, pricing_unit, 
              conditions, guest_count, promotion, cabin_quantity, valid_until, notes, source, 
              source_ref, import_job_id, status, created_at, created_by 
              FROM price_quote WHERE id = ?`

	if err := r.db.GetContext(ctx, &pq, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get price quote by id: %w", err)
	}

	return &pq, nil
}

// List retrieves price quotes with pagination and filters
func (r *PriceQuoteRepository) List(ctx context.Context, pagination Pagination, sailingID, cabinTypeID, supplierID *uint64, status *domain.QuoteStatus) (PaginatedResult[domain.PriceQuote], error) {
	var quotes []domain.PriceQuote
	var total int64

	countQuery := "SELECT COUNT(*) FROM price_quote WHERE 1=1"
	selectQuery := `SELECT id, sailing_id, cabin_type_id, supplier_id, price, currency, pricing_unit, 
                    conditions, guest_count, promotion, cabin_quantity, valid_until, notes, source, 
                    source_ref, import_job_id, status, created_at, created_by FROM price_quote WHERE 1=1`
	var args []interface{}

	if sailingID != nil {
		countQuery += " AND sailing_id = ?"
		selectQuery += " AND sailing_id = ?"
		args = append(args, *sailingID)
	}

	if cabinTypeID != nil {
		countQuery += " AND cabin_type_id = ?"
		selectQuery += " AND cabin_type_id = ?"
		args = append(args, *cabinTypeID)
	}

	if supplierID != nil {
		countQuery += " AND supplier_id = ?"
		selectQuery += " AND supplier_id = ?"
		args = append(args, *supplierID)
	}

	if status != nil {
		countQuery += " AND status = ?"
		selectQuery += " AND status = ?"
		args = append(args, *status)
	}

	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return PaginatedResult[domain.PriceQuote]{}, fmt.Errorf("failed to count price quotes: %w", err)
	}

	selectQuery += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, pagination.Limit(), pagination.Offset())

	if err := r.db.SelectContext(ctx, &quotes, selectQuery, args...); err != nil {
		return PaginatedResult[domain.PriceQuote]{}, fmt.Errorf("failed to list price quotes: %w", err)
	}

	return NewPaginatedResult(quotes, total, pagination), nil
}

// ListBySailing retrieves all active quotes for a sailing
func (r *PriceQuoteRepository) ListBySailing(ctx context.Context, sailingID uint64) ([]domain.PriceQuote, error) {
	var quotes []domain.PriceQuote
	query := `SELECT id, sailing_id, cabin_type_id, supplier_id, price, currency, pricing_unit, 
              conditions, guest_count, promotion, cabin_quantity, valid_until, notes, source, 
              source_ref, import_job_id, status, created_at, created_by 
              FROM price_quote WHERE sailing_id = ? AND status = 'ACTIVE' ORDER BY created_at DESC`

	if err := r.db.SelectContext(ctx, &quotes, query, sailingID); err != nil {
		return nil, fmt.Errorf("failed to list quotes by sailing: %w", err)
	}

	return quotes, nil
}

// ListBySupplier retrieves quotes by supplier with time range
func (r *PriceQuoteRepository) ListBySupplier(ctx context.Context, supplierID uint64, from, to *time.Time) ([]domain.PriceQuote, error) {
	var quotes []domain.PriceQuote
	query := `SELECT id, sailing_id, cabin_type_id, supplier_id, price, currency, pricing_unit, 
              conditions, guest_count, promotion, cabin_quantity, valid_until, notes, source, 
              source_ref, import_job_id, status, created_at, created_by 
              FROM price_quote WHERE supplier_id = ?`
	args := []interface{}{supplierID}

	if from != nil {
		query += " AND created_at >= ?"
		args = append(args, *from)
	}

	if to != nil {
		query += " AND created_at <= ?"
		args = append(args, *to)
	}

	query += " ORDER BY created_at DESC"

	if err := r.db.SelectContext(ctx, &quotes, query, args...); err != nil {
		return nil, fmt.Errorf("failed to list quotes by supplier: %w", err)
	}

	return quotes, nil
}

// Create creates a new price quote (append-only)
func (r *PriceQuoteRepository) Create(ctx context.Context, pq *domain.PriceQuote) error {
	query := `INSERT INTO price_quote (sailing_id, cabin_type_id, supplier_id, price, currency, 
              pricing_unit, conditions, guest_count, promotion, cabin_quantity, valid_until, 
              notes, source, source_ref, import_job_id, status, created_by) 
              VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db.ExecContext(ctx, query, pq.SailingID, pq.CabinTypeID, pq.SupplierID,
		pq.Price, pq.Currency, pq.PricingUnit, pq.Conditions, pq.GuestCount, pq.Promotion,
		pq.CabinQuantity, pq.ValidUntil, pq.Notes, pq.Source, pq.SourceRef, pq.ImportJobID,
		pq.Status, pq.CreatedBy)
	if err != nil {
		return fmt.Errorf("failed to create price quote: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	pq.ID = uint64(id)

	return nil
}

// VoidQuote marks a quote as voided (no updates, append new status)
func (r *PriceQuoteRepository) VoidQuote(ctx context.Context, id uint64) error {
	query := `UPDATE price_quote SET status = 'VOIDED' WHERE id = ? AND status = 'ACTIVE'`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to void quote: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if affected == 0 {
		return errors.New("quote not found or already voided")
	}

	return nil
}

// GetLatestPrice gets the latest active price for a sailing + cabin type + supplier combination
func (r *PriceQuoteRepository) GetLatestPrice(ctx context.Context, sailingID, cabinTypeID, supplierID uint64) (*domain.PriceQuote, error) {
	var pq domain.PriceQuote
	query := `SELECT id, sailing_id, cabin_type_id, supplier_id, price, currency, pricing_unit, 
              conditions, guest_count, promotion, cabin_quantity, valid_until, notes, source, 
              source_ref, import_job_id, status, created_at, created_by 
              FROM price_quote 
              WHERE sailing_id = ? AND cabin_type_id = ? AND supplier_id = ? AND status = 'ACTIVE'
              ORDER BY created_at DESC LIMIT 1`

	if err := r.db.GetContext(ctx, &pq, query, sailingID, cabinTypeID, supplierID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get latest price: %w", err)
	}

	return &pq, nil
}

// GetPriceHistory gets price history for a sailing + cabin type + supplier
func (r *PriceQuoteRepository) GetPriceHistory(ctx context.Context, sailingID, cabinTypeID, supplierID uint64, limit int) ([]domain.PriceQuote, error) {
	var quotes []domain.PriceQuote
	query := `SELECT id, sailing_id, cabin_type_id, supplier_id, price, currency, pricing_unit, 
              conditions, guest_count, promotion, cabin_quantity, valid_until, notes, source, 
              source_ref, import_job_id, status, created_at, created_by 
              FROM price_quote 
              WHERE sailing_id = ? AND cabin_type_id = ? AND supplier_id = ?
              ORDER BY created_at DESC LIMIT ?`

	if err := r.db.SelectContext(ctx, &quotes, query, sailingID, cabinTypeID, supplierID, limit); err != nil {
		return nil, fmt.Errorf("failed to get price history: %w", err)
	}

	return quotes, nil
}

// GetComparisonData retrieves latest prices for comparison view
func (r *PriceQuoteRepository) GetComparisonData(ctx context.Context, sailingID uint64) ([]ComparisonRow, error) {
	var rows []ComparisonRow
	query := `SELECT pq.cabin_type_id, pq.supplier_id, pq.price, pq.currency, pq.pricing_unit, pq.created_at
              FROM price_quote pq
              INNER JOIN (
                  SELECT cabin_type_id, supplier_id, MAX(created_at) as max_created
                  FROM price_quote 
                  WHERE sailing_id = ? AND status = 'ACTIVE'
                  GROUP BY cabin_type_id, supplier_id
              ) latest ON pq.cabin_type_id = latest.cabin_type_id 
                       AND pq.supplier_id = latest.supplier_id 
                       AND pq.created_at = latest.max_created
              WHERE pq.sailing_id = ? AND pq.status = 'ACTIVE'`

	if err := r.db.SelectContext(ctx, &rows, query, sailingID, sailingID); err != nil {
		return nil, fmt.Errorf("failed to get comparison data: %w", err)
	}

	return rows, nil
}

// ComparisonRow represents a row in the comparison view
type ComparisonRow struct {
	CabinTypeID uint64          `db:"cabin_type_id"`
	SupplierID  uint64          `db:"supplier_id"`
	Price       decimal.Decimal `db:"price"`
	Currency    string          `db:"currency"`
	PricingUnit string          `db:"pricing_unit"`
	CreatedAt   time.Time       `db:"created_at"`
}
