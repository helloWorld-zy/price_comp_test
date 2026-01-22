package http

import (
	"net/http"
	"time"

	"cruise-price-compare/internal/auth"
	"cruise-price-compare/internal/domain"
	"cruise-price-compare/internal/service"

	"github.com/gin-gonic/gin"
)

// QuoteHandler handles quote-related HTTP requests
type QuoteHandler struct {
	quoteService *service.QuoteService
}

// NewQuoteHandler creates a new quote handler
func NewQuoteHandler(quoteService *service.QuoteService) *QuoteHandler {
	return &QuoteHandler{quoteService: quoteService}
}

// CreateQuote handles POST /api/v1/quotes
func (h *QuoteHandler) CreateQuote(c *gin.Context) {
	userCtx := auth.GetUserContext(c)
	if userCtx == nil {
		RespondError(c, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "User not authenticated")
		return
	}

	var req struct {
		SailingID      uint64  `json:"sailing_id" binding:"required"`
		CabinTypeID    uint64  `json:"cabin_type_id" binding:"required"`
		Price          string  `json:"price" binding:"required"`
		Currency       string  `json:"currency"`
		PricingUnit    string  `json:"pricing_unit" binding:"required"`
		Conditions     string  `json:"conditions"`
		GuestCount     *int    `json:"guest_count"`
		Promotion      string  `json:"promotion"`
		CabinQuantity  *int    `json:"cabin_quantity"`
		ValidUntil     *string `json:"valid_until"` // YYYY-MM-DD
		Notes          string  `json:"notes"`
		IdempotencyKey string  `json:"idempotency_key"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "ERR_INVALID_REQUEST", err.Error())
		return
	}

	// Parse valid_until
	var validUntil *time.Time
	if req.ValidUntil != nil && *req.ValidUntil != "" {
		t, err := time.Parse("2006-01-02", *req.ValidUntil)
		if err != nil {
			RespondError(c, http.StatusBadRequest, "ERR_INVALID_DATE", "Invalid valid_until date format")
			return
		}
		validUntil = &t
	}

	input := service.CreateQuoteInput{
		SailingID:      req.SailingID,
		CabinTypeID:    req.CabinTypeID,
		Price:          req.Price,
		Currency:       req.Currency,
		PricingUnit:    domain.PricingUnit(req.PricingUnit),
		Conditions:     req.Conditions,
		GuestCount:     req.GuestCount,
		Promotion:      req.Promotion,
		CabinQuantity:  req.CabinQuantity,
		ValidUntil:     validUntil,
		Notes:          req.Notes,
		IdempotencyKey: req.IdempotencyKey,
		SupplierID:     userCtx.SupplierID,
		UserID:         userCtx.UserID,
	}

	quote, err := h.quoteService.CreateQuote(c.Request.Context(), input)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "ERR_CREATE_QUOTE", err.Error())
		return
	}

	c.JSON(http.StatusCreated, quote)
}

// ListQuotes handles GET /api/v1/quotes
func (h *QuoteHandler) ListQuotes(c *gin.Context) {
	userCtx := auth.GetUserContext(c)
	if userCtx == nil {
		RespondError(c, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "User not authenticated")
		return
	}

	pagination := ParsePagination(c)

	// Parse filters
	sailingID := ParseUint64Query(c, "sailing_id")
	cabinTypeID := ParseUint64Query(c, "cabin_type_id")
	supplierID := ParseUint64Query(c, "supplier_id")

	var status *domain.QuoteStatus
	statusParam := c.Query("status")
	if statusParam != "" {
		s := domain.QuoteStatus(statusParam)
		status = &s
	}

	input := service.ListQuotesInput{
		Pagination:   pagination,
		SailingID:    sailingID,
		CabinTypeID:  cabinTypeID,
		SupplierID:   supplierID,
		Status:       status,
		UserID:       userCtx.UserID,
		UserRole:     userCtx.Role,
		UserSupplier: userCtx.SupplierID,
	}

	result, err := h.quoteService.ListQuotes(c.Request.Context(), input)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "ERR_LIST_QUOTES", err.Error())
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetQuote handles GET /api/v1/quotes/:id
func (h *QuoteHandler) GetQuote(c *gin.Context) {
	userCtx := auth.GetUserContext(c)
	if userCtx == nil {
		RespondError(c, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "User not authenticated")
		return
	}

	id, ok := ParseUint64Param(c, "id")
	if !ok {
		RespondError(c, http.StatusBadRequest, "ERR_INVALID_ID", "Invalid quote ID")
		return
	}

	quote, err := h.quoteService.GetQuote(c.Request.Context(), id, userCtx.Role, userCtx.SupplierID)
	if err != nil {
		if err.Error() == "quote not found" {
			RespondError(c, http.StatusNotFound, "ERR_NOT_FOUND", "Quote not found")
			return
		}
		if err.Error() == "forbidden: cannot access other supplier's quotes" {
			RespondError(c, http.StatusForbidden, "ERR_FORBIDDEN", err.Error())
			return
		}
		RespondError(c, http.StatusInternalServerError, "ERR_GET_QUOTE", err.Error())
		return
	}

	c.JSON(http.StatusOK, quote)
}

// VoidQuote handles PUT /api/v1/quotes/:id/void
func (h *QuoteHandler) VoidQuote(c *gin.Context) {
	userCtx := auth.GetUserContext(c)
	if userCtx == nil {
		RespondError(c, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "User not authenticated")
		return
	}

	id, ok := ParseUint64Param(c, "id")
	if !ok {
		RespondError(c, http.StatusBadRequest, "ERR_INVALID_ID", "Invalid quote ID")
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "ERR_INVALID_REQUEST", err.Error())
		return
	}

	quote, err := h.quoteService.VoidQuote(c.Request.Context(), id, req.Reason, userCtx.UserID, userCtx.Role, userCtx.SupplierID)
	if err != nil {
		if err.Error() == "quote not found" {
			RespondError(c, http.StatusNotFound, "ERR_NOT_FOUND", "Quote not found")
			return
		}
		if err.Error() == "forbidden: cannot void other supplier's quotes" {
			RespondError(c, http.StatusForbidden, "ERR_FORBIDDEN", err.Error())
			return
		}
		if err.Error() == "quote is not active" {
			RespondError(c, http.StatusBadRequest, "ERR_INVALID_STATE", err.Error())
			return
		}
		RespondError(c, http.StatusInternalServerError, "ERR_VOID_QUOTE", err.Error())
		return
	}

	c.JSON(http.StatusOK, quote)
}
