package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"cruise-price-compare/internal/domain"
	"cruise-price-compare/internal/repo"

	"github.com/shopspring/decimal"
)

// QuoteService handles quote business logic
type QuoteService struct {
	quoteRepo    *repo.PriceQuoteRepository
	sailingRepo  *repo.SailingRepository
	cabinRepo    *repo.CabinTypeRepository
	supplierRepo *repo.SupplierRepository
	auditService *AuditService
}

// NewQuoteService creates a new quote service
func NewQuoteService(
	quoteRepo *repo.PriceQuoteRepository,
	sailingRepo *repo.SailingRepository,
	cabinRepo *repo.CabinTypeRepository,
	supplierRepo *repo.SupplierRepository,
	auditService *AuditService,
) *QuoteService {
	return &QuoteService{
		quoteRepo:    quoteRepo,
		sailingRepo:  sailingRepo,
		cabinRepo:    cabinRepo,
		supplierRepo: supplierRepo,
		auditService: auditService,
	}
}

// CreateQuoteInput represents the input for creating a quote
type CreateQuoteInput struct {
	SailingID      uint64
	CabinTypeID    uint64
	Price          string
	Currency       string
	PricingUnit    domain.PricingUnit
	Conditions     string
	GuestCount     *int
	Promotion      string
	CabinQuantity  *int
	ValidUntil     *time.Time
	Notes          string
	IdempotencyKey string
	SupplierID     uint64 // From auth context
	UserID         uint64 // From auth context
}

// CreateQuote creates a new quote (manual entry)
func (s *QuoteService) CreateQuote(ctx context.Context, input CreateQuoteInput) (*domain.PriceQuote, error) {
	// Validate price
	price, err := decimal.NewFromString(input.Price)
	if err != nil {
		return nil, fmt.Errorf("invalid price format: %w", err)
	}
	if price.LessThanOrEqual(decimal.Zero) {
		return nil, errors.New("price must be greater than zero")
	}

	// Validate sailing exists
	sailing, err := s.sailingRepo.GetByID(ctx, input.SailingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sailing: %w", err)
	}
	if sailing == nil {
		return nil, errors.New("sailing not found")
	}

	// Validate cabin type exists
	cabinType, err := s.cabinRepo.GetByID(ctx, input.CabinTypeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cabin type: %w", err)
	}
	if cabinType == nil {
		return nil, errors.New("cabin type not found")
	}

	// Validate supplier exists
	supplier, err := s.supplierRepo.GetByID(ctx, input.SupplierID)
	if err != nil {
		return nil, fmt.Errorf("failed to get supplier: %w", err)
	}
	if supplier == nil {
		return nil, errors.New("supplier not found")
	}

	// Validate currency (basic check)
	if input.Currency == "" {
		input.Currency = "CNY"
	}

	// Validate pricing unit
	if input.PricingUnit == "" {
		return nil, errors.New("pricing unit is required")
	}

	// Create quote
	quote := &domain.PriceQuote{
		SailingID:     input.SailingID,
		CabinTypeID:   input.CabinTypeID,
		SupplierID:    input.SupplierID,
		Price:         price,
		Currency:      input.Currency,
		PricingUnit:   input.PricingUnit,
		Conditions:    input.Conditions,
		GuestCount:    input.GuestCount,
		Promotion:     input.Promotion,
		CabinQuantity: input.CabinQuantity,
		ValidUntil:    input.ValidUntil,
		Notes:         input.Notes,
		Source:        domain.QuoteSourceManual,
		SourceRef:     input.IdempotencyKey,
		Status:        domain.QuoteStatusActive,
		CreatedBy:     input.UserID,
	}

	if err := s.quoteRepo.Create(ctx, quote); err != nil {
		return nil, fmt.Errorf("failed to create quote: %w", err)
	}

	// Audit log
	if s.auditService != nil {
		s.auditService.LogCreate(ctx, input.UserID, input.SupplierID, "PriceQuote", quote.ID, quote)
	}

	return quote, nil
}

// ListQuotesInput represents the input for listing quotes
type ListQuotesInput struct {
	Pagination   repo.Pagination
	SailingID    *uint64
	CabinTypeID  *uint64
	SupplierID   *uint64
	Status       *domain.QuoteStatus
	UserID       uint64 // From auth context
	UserRole     domain.UserRole
	UserSupplier uint64 // From auth context (if vendor)
}

// ListQuotes retrieves quotes with filters
func (s *QuoteService) ListQuotes(ctx context.Context, input ListQuotesInput) (repo.PaginatedResult[domain.PriceQuote], error) {
	// If vendor role, filter by their supplier
	supplierID := input.SupplierID
	if input.UserRole == domain.UserRoleVendor {
		supplierID = &input.UserSupplier
	}

	result, err := s.quoteRepo.List(ctx, input.Pagination, input.SailingID, input.CabinTypeID, supplierID, input.Status)
	if err != nil {
		return repo.PaginatedResult[domain.PriceQuote]{}, fmt.Errorf("failed to list quotes: %w", err)
	}

	return result, nil
}

// GetQuote retrieves a single quote by ID
func (s *QuoteService) GetQuote(ctx context.Context, id uint64, userRole domain.UserRole, userSupplier uint64) (*domain.PriceQuote, error) {
	quote, err := s.quoteRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get quote: %w", err)
	}
	if quote == nil {
		return nil, errors.New("quote not found")
	}

	// Vendor can only see their own supplier's quotes
	if userRole == domain.UserRoleVendor && quote.SupplierID != userSupplier {
		return nil, errors.New("forbidden: cannot access other supplier's quotes")
	}

	return quote, nil
}

// VoidQuote marks a quote as voided
func (s *QuoteService) VoidQuote(ctx context.Context, id uint64, reason string, userID uint64, userRole domain.UserRole, userSupplier uint64) (*domain.PriceQuote, error) {
	// Get the quote first
	quote, err := s.quoteRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get quote: %w", err)
	}
	if quote == nil {
		return nil, errors.New("quote not found")
	}

	// Vendor can only void their own supplier's quotes
	if userRole == domain.UserRoleVendor && quote.SupplierID != userSupplier {
		return nil, errors.New("forbidden: cannot void other supplier's quotes")
	}

	// Cannot void if already voided
	if quote.Status != domain.QuoteStatusActive {
		return nil, errors.New("quote is not active")
	}

	// Void the quote
	if err := s.quoteRepo.VoidQuote(ctx, id); err != nil {
		return nil, fmt.Errorf("failed to void quote: %w", err)
	}

	// Reload to get updated status
	quote, err = s.quoteRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to reload quote: %w", err)
	}

	// Audit log
	if s.auditService != nil {
		s.auditService.LogUpdate(ctx, userID, quote.SupplierID, "PriceQuote", quote.ID, map[string]interface{}{"reason": reason}, quote)
	}

	return quote, nil
}

// BatchCreateQuotesInput represents input for batch quote creation
type BatchCreateQuotesInput struct {
	Quotes      []CreateQuoteInput
	ImportJobID *uint64
	SupplierID  uint64
	UserID      uint64
}

// BatchCreateQuotes creates multiple quotes
func (s *QuoteService) BatchCreateQuotes(ctx context.Context, input BatchCreateQuotesInput) ([]domain.PriceQuote, []error) {
	quotes := make([]domain.PriceQuote, 0, len(input.Quotes))
	errors := make([]error, 0)

	for _, q := range input.Quotes {
		q.SupplierID = input.SupplierID
		q.UserID = input.UserID

		quote, err := s.CreateQuote(ctx, q)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		quotes = append(quotes, *quote)
	}

	return quotes, errors
}
