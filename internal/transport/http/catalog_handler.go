package http

import (
	"net/http"
	"strconv"

	"cruise-price-compare/internal/auth"
	"cruise-price-compare/internal/domain"
	"cruise-price-compare/internal/service"

	"github.com/gin-gonic/gin"
)

// CatalogHandler handles catalog-related HTTP requests
type CatalogHandler struct {
	catalogService *service.CatalogService
}

// NewCatalogHandler creates a new catalog handler
func NewCatalogHandler(catalogService *service.CatalogService) *CatalogHandler {
	return &CatalogHandler{catalogService: catalogService}
}

// CruiseLine handlers

// ListCruiseLines returns a paginated list of cruise lines
func (h *CatalogHandler) ListCruiseLines(c *gin.Context) {
	pagination := ParsePagination(c)
	statusParam := c.Query("status")
	var status *domain.EntityStatus
	if statusParam != "" {
		s := domain.EntityStatus(statusParam)
		status = &s
	}

	result, err := h.catalogService.ListCruiseLines(c.Request.Context(), pagination, status)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "ERR_LIST_CRUISE_LINES", err.Error())
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetCruiseLine returns a cruise line by ID
func (h *CatalogHandler) GetCruiseLine(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "ERR_INVALID_ID", "Invalid cruise line ID")
		return
	}

	cruiseLine, err := h.catalogService.GetCruiseLine(c.Request.Context(), id)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "ERR_GET_CRUISE_LINE", err.Error())
		return
	}
	if cruiseLine == nil {
		RespondError(c, http.StatusNotFound, "ERR_NOT_FOUND", "Cruise line not found")
		return
	}

	c.JSON(http.StatusOK, cruiseLine)
}

// CreateCruiseLine creates a new cruise line
func (h *CatalogHandler) CreateCruiseLine(c *gin.Context) {
	userCtx := auth.GetUserContext(c)
	if userCtx == nil {
		RespondError(c, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "User not authenticated")
		return
	}

	var req struct {
		Name    string   `json:"name" binding:"required"`
		LogoURL *string  `json:"logo_url"`
		Aliases []string `json:"aliases"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "ERR_INVALID_REQUEST", err.Error())
		return
	}

	cl := &domain.CruiseLine{
		Name:    req.Name,
		LogoURL: req.LogoURL,
		Aliases: req.Aliases,
	}

	if errs := domain.ValidateCruiseLine(cl); len(errs) > 0 {
		RespondValidationErrors(c, errs)
		return
	}

	if err := h.catalogService.CreateCruiseLine(c.Request.Context(), userCtx.UserID, cl); err != nil {
		if err == service.ErrDuplicateName {
			RespondError(c, http.StatusConflict, "ERR_DUPLICATE_NAME", "Cruise line with this name already exists")
			return
		}
		RespondError(c, http.StatusInternalServerError, "ERR_CREATE_CRUISE_LINE", err.Error())
		return
	}

	c.JSON(http.StatusCreated, cl)
}

// UpdateCruiseLine updates an existing cruise line
func (h *CatalogHandler) UpdateCruiseLine(c *gin.Context) {
	userCtx := auth.GetUserContext(c)
	if userCtx == nil {
		RespondError(c, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "User not authenticated")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "ERR_INVALID_ID", "Invalid cruise line ID")
		return
	}

	var req struct {
		Name    string              `json:"name" binding:"required"`
		LogoURL *string             `json:"logo_url"`
		Aliases []string            `json:"aliases"`
		Status  domain.EntityStatus `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "ERR_INVALID_REQUEST", err.Error())
		return
	}

	cl := &domain.CruiseLine{
		ID:      id,
		Name:    req.Name,
		LogoURL: req.LogoURL,
		Aliases: req.Aliases,
		Status:  req.Status,
	}

	if errs := domain.ValidateCruiseLine(cl); len(errs) > 0 {
		RespondValidationErrors(c, errs)
		return
	}

	if err := h.catalogService.UpdateCruiseLine(c.Request.Context(), userCtx.UserID, cl); err != nil {
		if err == service.ErrCruiseLineNotFound {
			RespondError(c, http.StatusNotFound, "ERR_NOT_FOUND", "Cruise line not found")
			return
		}
		if err == service.ErrDuplicateName {
			RespondError(c, http.StatusConflict, "ERR_DUPLICATE_NAME", "Cruise line with this name already exists")
			return
		}
		RespondError(c, http.StatusInternalServerError, "ERR_UPDATE_CRUISE_LINE", err.Error())
		return
	}

	c.JSON(http.StatusOK, cl)
}

// DeleteCruiseLine deletes a cruise line
func (h *CatalogHandler) DeleteCruiseLine(c *gin.Context) {
	userCtx := auth.GetUserContext(c)
	if userCtx == nil {
		RespondError(c, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "User not authenticated")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "ERR_INVALID_ID", "Invalid cruise line ID")
		return
	}

	if err := h.catalogService.DeleteCruiseLine(c.Request.Context(), userCtx.UserID, id); err != nil {
		if err == service.ErrCruiseLineNotFound {
			RespondError(c, http.StatusNotFound, "ERR_NOT_FOUND", "Cruise line not found")
			return
		}
		RespondError(c, http.StatusInternalServerError, "ERR_DELETE_CRUISE_LINE", err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

// Ship handlers

// ListShips returns a paginated list of ships
func (h *CatalogHandler) ListShips(c *gin.Context) {
	pagination := ParsePagination(c)
	var cruiseLineID *uint64
	if id := c.Query("cruise_line_id"); id != "" {
		if parsed, err := strconv.ParseUint(id, 10, 64); err == nil {
			cruiseLineID = &parsed
		}
	}
	statusParam := c.Query("status")
	var status *domain.EntityStatus
	if statusParam != "" {
		s := domain.EntityStatus(statusParam)
		status = &s
	}

	result, err := h.catalogService.ListShips(c.Request.Context(), pagination, cruiseLineID, status)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "ERR_LIST_SHIPS", err.Error())
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetShip returns a ship by ID
func (h *CatalogHandler) GetShip(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "ERR_INVALID_ID", "Invalid ship ID")
		return
	}

	ship, err := h.catalogService.GetShip(c.Request.Context(), id)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "ERR_GET_SHIP", err.Error())
		return
	}
	if ship == nil {
		RespondError(c, http.StatusNotFound, "ERR_NOT_FOUND", "Ship not found")
		return
	}

	c.JSON(http.StatusOK, ship)
}

// CreateShip creates a new ship
func (h *CatalogHandler) CreateShip(c *gin.Context) {
	userCtx := auth.GetUserContext(c)
	if userCtx == nil {
		RespondError(c, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "User not authenticated")
		return
	}

	var req struct {
		CruiseLineID uint64   `json:"cruise_line_id" binding:"required"`
		Name         string   `json:"name" binding:"required"`
		IMO          *string  `json:"imo"`
		Aliases      []string `json:"aliases"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "ERR_INVALID_REQUEST", err.Error())
		return
	}

	ship := &domain.Ship{
		CruiseLineID: req.CruiseLineID,
		Name:         req.Name,
		IMO:          req.IMO,
		Aliases:      req.Aliases,
	}

	if errs := domain.ValidateShip(ship); len(errs) > 0 {
		RespondValidationErrors(c, errs)
		return
	}

	if err := h.catalogService.CreateShip(c.Request.Context(), userCtx.UserID, ship); err != nil {
		if err == service.ErrDuplicateName {
			RespondError(c, http.StatusConflict, "ERR_DUPLICATE_NAME", "Ship with this name already exists for this cruise line")
			return
		}
		RespondError(c, http.StatusInternalServerError, "ERR_CREATE_SHIP", err.Error())
		return
	}

	c.JSON(http.StatusCreated, ship)
}

// UpdateShip updates an existing ship
func (h *CatalogHandler) UpdateShip(c *gin.Context) {
	userCtx := auth.GetUserContext(c)
	if userCtx == nil {
		RespondError(c, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "User not authenticated")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "ERR_INVALID_ID", "Invalid ship ID")
		return
	}

	var req struct {
		CruiseLineID uint64              `json:"cruise_line_id" binding:"required"`
		Name         string              `json:"name" binding:"required"`
		IMO          *string             `json:"imo"`
		Aliases      []string            `json:"aliases"`
		Status       domain.EntityStatus `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "ERR_INVALID_REQUEST", err.Error())
		return
	}

	ship := &domain.Ship{
		ID:           id,
		CruiseLineID: req.CruiseLineID,
		Name:         req.Name,
		IMO:          req.IMO,
		Aliases:      req.Aliases,
		Status:       req.Status,
	}

	if errs := domain.ValidateShip(ship); len(errs) > 0 {
		RespondValidationErrors(c, errs)
		return
	}

	if err := h.catalogService.UpdateShip(c.Request.Context(), userCtx.UserID, ship); err != nil {
		if err == service.ErrShipNotFound {
			RespondError(c, http.StatusNotFound, "ERR_NOT_FOUND", "Ship not found")
			return
		}
		if err == service.ErrDuplicateName {
			RespondError(c, http.StatusConflict, "ERR_DUPLICATE_NAME", "Ship with this name already exists for this cruise line")
			return
		}
		RespondError(c, http.StatusInternalServerError, "ERR_UPDATE_SHIP", err.Error())
		return
	}

	c.JSON(http.StatusOK, ship)
}

// DeleteShip deletes a ship
func (h *CatalogHandler) DeleteShip(c *gin.Context) {
	userCtx := auth.GetUserContext(c)
	if userCtx == nil {
		RespondError(c, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "User not authenticated")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "ERR_INVALID_ID", "Invalid ship ID")
		return
	}

	if err := h.catalogService.DeleteShip(c.Request.Context(), userCtx.UserID, id); err != nil {
		if err == service.ErrShipNotFound {
			RespondError(c, http.StatusNotFound, "ERR_NOT_FOUND", "Ship not found")
			return
		}
		RespondError(c, http.StatusInternalServerError, "ERR_DELETE_SHIP", err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

// ListCabinTypesByShip returns cabin types for a specific ship
func (h *CatalogHandler) ListCabinTypesByShip(c *gin.Context) {
	shipID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "ERR_INVALID_ID", "Invalid ship ID")
		return
	}

	cabinTypes, err := h.catalogService.ListCabinTypesByShip(c.Request.Context(), shipID)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "ERR_LIST_CABIN_TYPES", err.Error())
		return
	}

	c.JSON(http.StatusOK, cabinTypes)
}

// CabinCategory handlers

// ListCabinCategories returns all cabin categories
func (h *CatalogHandler) ListCabinCategories(c *gin.Context) {
	categories, err := h.catalogService.ListCabinCategories(c.Request.Context())
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "ERR_LIST_CABIN_CATEGORIES", err.Error())
		return
	}

	c.JSON(http.StatusOK, categories)
}

// CreateCabinCategory creates a new cabin category
func (h *CatalogHandler) CreateCabinCategory(c *gin.Context) {
	userCtx := auth.GetUserContext(c)
	if userCtx == nil {
		RespondError(c, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "User not authenticated")
		return
	}

	var req struct {
		Name      string `json:"name" binding:"required"`
		SortOrder int    `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "ERR_INVALID_REQUEST", err.Error())
		return
	}

	cc := &domain.CabinCategory{
		Name:      req.Name,
		SortOrder: req.SortOrder,
	}

	if err := h.catalogService.CreateCabinCategory(c.Request.Context(), userCtx.UserID, cc); err != nil {
		if err == service.ErrDuplicateName {
			RespondError(c, http.StatusConflict, "ERR_DUPLICATE_NAME", "Cabin category with this name already exists")
			return
		}
		RespondError(c, http.StatusInternalServerError, "ERR_CREATE_CABIN_CATEGORY", err.Error())
		return
	}

	c.JSON(http.StatusCreated, cc)
}

// UpdateCabinCategory updates an existing cabin category
func (h *CatalogHandler) UpdateCabinCategory(c *gin.Context) {
	userCtx := auth.GetUserContext(c)
	if userCtx == nil {
		RespondError(c, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "User not authenticated")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "ERR_INVALID_ID", "Invalid cabin category ID")
		return
	}

	var req struct {
		Name      string `json:"name" binding:"required"`
		SortOrder int    `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "ERR_INVALID_REQUEST", err.Error())
		return
	}

	cc := &domain.CabinCategory{
		ID:        id,
		Name:      req.Name,
		SortOrder: req.SortOrder,
	}

	if err := h.catalogService.UpdateCabinCategory(c.Request.Context(), userCtx.UserID, cc); err != nil {
		if err == service.ErrCabinCategoryNotFound {
			RespondError(c, http.StatusNotFound, "ERR_NOT_FOUND", "Cabin category not found")
			return
		}
		RespondError(c, http.StatusInternalServerError, "ERR_UPDATE_CABIN_CATEGORY", err.Error())
		return
	}

	c.JSON(http.StatusOK, cc)
}

// DeleteCabinCategory deletes a cabin category
func (h *CatalogHandler) DeleteCabinCategory(c *gin.Context) {
	userCtx := auth.GetUserContext(c)
	if userCtx == nil {
		RespondError(c, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "User not authenticated")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "ERR_INVALID_ID", "Invalid cabin category ID")
		return
	}

	if err := h.catalogService.DeleteCabinCategory(c.Request.Context(), userCtx.UserID, id); err != nil {
		if err == service.ErrCabinCategoryNotFound {
			RespondError(c, http.StatusNotFound, "ERR_NOT_FOUND", "Cabin category not found")
			return
		}
		RespondError(c, http.StatusInternalServerError, "ERR_DELETE_CABIN_CATEGORY", err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

// CabinType handlers

// ListCabinTypes returns a paginated list of cabin types
func (h *CatalogHandler) ListCabinTypes(c *gin.Context) {
	pagination := ParsePagination(c)
	var shipID, categoryID *uint64
	if id := c.Query("ship_id"); id != "" {
		if parsed, err := strconv.ParseUint(id, 10, 64); err == nil {
			shipID = &parsed
		}
	}
	if id := c.Query("category_id"); id != "" {
		if parsed, err := strconv.ParseUint(id, 10, 64); err == nil {
			categoryID = &parsed
		}
	}
	enabledOnly := c.Query("enabled_only") == "true"

	result, err := h.catalogService.ListCabinTypes(c.Request.Context(), pagination, shipID, categoryID, enabledOnly)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "ERR_LIST_CABIN_TYPES", err.Error())
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetCabinType returns a cabin type by ID
func (h *CatalogHandler) GetCabinType(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "ERR_INVALID_ID", "Invalid cabin type ID")
		return
	}

	cabinType, err := h.catalogService.GetCabinType(c.Request.Context(), id)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "ERR_GET_CABIN_TYPE", err.Error())
		return
	}
	if cabinType == nil {
		RespondError(c, http.StatusNotFound, "ERR_NOT_FOUND", "Cabin type not found")
		return
	}

	c.JSON(http.StatusOK, cabinType)
}

// CreateCabinType creates a new cabin type
func (h *CatalogHandler) CreateCabinType(c *gin.Context) {
	userCtx := auth.GetUserContext(c)
	if userCtx == nil {
		RespondError(c, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "User not authenticated")
		return
	}

	var req struct {
		ShipID     uint64  `json:"ship_id" binding:"required"`
		CategoryID uint64  `json:"category_id" binding:"required"`
		Name       string  `json:"name" binding:"required"`
		Code       *string `json:"code"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "ERR_INVALID_REQUEST", err.Error())
		return
	}

	ct := &domain.CabinType{
		ShipID:     req.ShipID,
		CategoryID: req.CategoryID,
		Name:       req.Name,
		Code:       req.Code,
	}

	if errs := domain.ValidateCabinType(ct); len(errs) > 0 {
		RespondValidationErrors(c, errs)
		return
	}

	if err := h.catalogService.CreateCabinType(c.Request.Context(), userCtx.UserID, ct); err != nil {
		RespondError(c, http.StatusInternalServerError, "ERR_CREATE_CABIN_TYPE", err.Error())
		return
	}

	c.JSON(http.StatusCreated, ct)
}

// UpdateCabinType updates an existing cabin type
func (h *CatalogHandler) UpdateCabinType(c *gin.Context) {
	userCtx := auth.GetUserContext(c)
	if userCtx == nil {
		RespondError(c, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "User not authenticated")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "ERR_INVALID_ID", "Invalid cabin type ID")
		return
	}

	var req struct {
		ShipID     uint64  `json:"ship_id" binding:"required"`
		CategoryID uint64  `json:"category_id" binding:"required"`
		Name       string  `json:"name" binding:"required"`
		Code       *string `json:"code"`
		IsEnabled  bool    `json:"is_enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "ERR_INVALID_REQUEST", err.Error())
		return
	}

	ct := &domain.CabinType{
		ID:         id,
		ShipID:     req.ShipID,
		CategoryID: req.CategoryID,
		Name:       req.Name,
		Code:       req.Code,
		IsEnabled:  req.IsEnabled,
	}

	if errs := domain.ValidateCabinType(ct); len(errs) > 0 {
		RespondValidationErrors(c, errs)
		return
	}

	if err := h.catalogService.UpdateCabinType(c.Request.Context(), userCtx.UserID, ct); err != nil {
		if err == service.ErrCabinTypeNotFound {
			RespondError(c, http.StatusNotFound, "ERR_NOT_FOUND", "Cabin type not found")
			return
		}
		RespondError(c, http.StatusInternalServerError, "ERR_UPDATE_CABIN_TYPE", err.Error())
		return
	}

	c.JSON(http.StatusOK, ct)
}

// DeleteCabinType deletes a cabin type
func (h *CatalogHandler) DeleteCabinType(c *gin.Context) {
	userCtx := auth.GetUserContext(c)
	if userCtx == nil {
		RespondError(c, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "User not authenticated")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "ERR_INVALID_ID", "Invalid cabin type ID")
		return
	}

	if err := h.catalogService.DeleteCabinType(c.Request.Context(), userCtx.UserID, id); err != nil {
		if err == service.ErrCabinTypeNotFound {
			RespondError(c, http.StatusNotFound, "ERR_NOT_FOUND", "Cabin type not found")
			return
		}
		RespondError(c, http.StatusInternalServerError, "ERR_DELETE_CABIN_TYPE", err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

// Sailing handlers

// ListSailings returns a paginated list of sailings
func (h *CatalogHandler) ListSailings(c *gin.Context) {
	pagination := ParsePagination(c)
	var shipID *uint64
	if id := c.Query("ship_id"); id != "" {
		if parsed, err := strconv.ParseUint(id, 10, 64); err == nil {
			shipID = &parsed
		}
	}
	statusParam := c.Query("status")
	var status *domain.SailingStatus
	if statusParam != "" {
		s := domain.SailingStatus(statusParam)
		status = &s
	}

	result, err := h.catalogService.ListSailings(c.Request.Context(), pagination, shipID, status)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "ERR_LIST_SAILINGS", err.Error())
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetSailing returns a sailing by ID
func (h *CatalogHandler) GetSailing(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "ERR_INVALID_ID", "Invalid sailing ID")
		return
	}

	sailing, err := h.catalogService.GetSailing(c.Request.Context(), id)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "ERR_GET_SAILING", err.Error())
		return
	}
	if sailing == nil {
		RespondError(c, http.StatusNotFound, "ERR_NOT_FOUND", "Sailing not found")
		return
	}

	c.JSON(http.StatusOK, sailing)
}

// CreateSailing creates a new sailing
func (h *CatalogHandler) CreateSailing(c *gin.Context) {
	userCtx := auth.GetUserContext(c)
	if userCtx == nil {
		RespondError(c, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "User not authenticated")
		return
	}

	var req domain.Sailing
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "ERR_INVALID_REQUEST", err.Error())
		return
	}

	if errs := domain.ValidateSailing(&req); len(errs) > 0 {
		RespondValidationErrors(c, errs)
		return
	}

	if err := h.catalogService.CreateSailing(c.Request.Context(), userCtx.UserID, &req); err != nil {
		RespondError(c, http.StatusInternalServerError, "ERR_CREATE_SAILING", err.Error())
		return
	}

	c.JSON(http.StatusCreated, req)
}

// UpdateSailing updates an existing sailing
func (h *CatalogHandler) UpdateSailing(c *gin.Context) {
	userCtx := auth.GetUserContext(c)
	if userCtx == nil {
		RespondError(c, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "User not authenticated")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "ERR_INVALID_ID", "Invalid sailing ID")
		return
	}

	var req domain.Sailing
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "ERR_INVALID_REQUEST", err.Error())
		return
	}
	req.ID = id

	if errs := domain.ValidateSailing(&req); len(errs) > 0 {
		RespondValidationErrors(c, errs)
		return
	}

	if err := h.catalogService.UpdateSailing(c.Request.Context(), userCtx.UserID, &req); err != nil {
		if err == service.ErrSailingNotFound {
			RespondError(c, http.StatusNotFound, "ERR_NOT_FOUND", "Sailing not found")
			return
		}
		RespondError(c, http.StatusInternalServerError, "ERR_UPDATE_SAILING", err.Error())
		return
	}

	c.JSON(http.StatusOK, req)
}

// DeleteSailing deletes a sailing
func (h *CatalogHandler) DeleteSailing(c *gin.Context) {
	userCtx := auth.GetUserContext(c)
	if userCtx == nil {
		RespondError(c, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "User not authenticated")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "ERR_INVALID_ID", "Invalid sailing ID")
		return
	}

	if err := h.catalogService.DeleteSailing(c.Request.Context(), userCtx.UserID, id); err != nil {
		if err == service.ErrSailingNotFound {
			RespondError(c, http.StatusNotFound, "ERR_NOT_FOUND", "Sailing not found")
			return
		}
		RespondError(c, http.StatusInternalServerError, "ERR_DELETE_SAILING", err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

// Supplier handlers

// ListSuppliers returns a paginated list of suppliers
func (h *CatalogHandler) ListSuppliers(c *gin.Context) {
	pagination := ParsePagination(c)
	statusParam := c.Query("status")
	var status *domain.EntityStatus
	if statusParam != "" {
		s := domain.EntityStatus(statusParam)
		status = &s
	}

	result, err := h.catalogService.ListSuppliers(c.Request.Context(), pagination, status)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "ERR_LIST_SUPPLIERS", err.Error())
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetSupplier returns a supplier by ID
func (h *CatalogHandler) GetSupplier(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "ERR_INVALID_ID", "Invalid supplier ID")
		return
	}

	supplier, err := h.catalogService.GetSupplier(c.Request.Context(), id)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "ERR_GET_SUPPLIER", err.Error())
		return
	}
	if supplier == nil {
		RespondError(c, http.StatusNotFound, "ERR_NOT_FOUND", "Supplier not found")
		return
	}

	c.JSON(http.StatusOK, supplier)
}

// CreateSupplier creates a new supplier
func (h *CatalogHandler) CreateSupplier(c *gin.Context) {
	userCtx := auth.GetUserContext(c)
	if userCtx == nil {
		RespondError(c, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "User not authenticated")
		return
	}

	var req struct {
		Name    string   `json:"name" binding:"required"`
		Contact *string  `json:"contact"`
		Aliases []string `json:"aliases"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "ERR_INVALID_REQUEST", err.Error())
		return
	}

	supplier := &domain.Supplier{
		Name:    req.Name,
		Contact: req.Contact,
		Aliases: req.Aliases,
	}

	if errs := domain.ValidateSupplier(supplier); len(errs) > 0 {
		RespondValidationErrors(c, errs)
		return
	}

	if err := h.catalogService.CreateSupplier(c.Request.Context(), userCtx.UserID, supplier); err != nil {
		if err == service.ErrDuplicateName {
			RespondError(c, http.StatusConflict, "ERR_DUPLICATE_NAME", "Supplier with this name already exists")
			return
		}
		RespondError(c, http.StatusInternalServerError, "ERR_CREATE_SUPPLIER", err.Error())
		return
	}

	c.JSON(http.StatusCreated, supplier)
}

// UpdateSupplier updates an existing supplier
func (h *CatalogHandler) UpdateSupplier(c *gin.Context) {
	userCtx := auth.GetUserContext(c)
	if userCtx == nil {
		RespondError(c, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "User not authenticated")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "ERR_INVALID_ID", "Invalid supplier ID")
		return
	}

	var req struct {
		Name    string              `json:"name" binding:"required"`
		Contact *string             `json:"contact"`
		Aliases []string            `json:"aliases"`
		Status  domain.EntityStatus `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "ERR_INVALID_REQUEST", err.Error())
		return
	}

	supplier := &domain.Supplier{
		ID:      id,
		Name:    req.Name,
		Contact: req.Contact,
		Aliases: req.Aliases,
		Status:  req.Status,
	}

	if errs := domain.ValidateSupplier(supplier); len(errs) > 0 {
		RespondValidationErrors(c, errs)
		return
	}

	if err := h.catalogService.UpdateSupplier(c.Request.Context(), userCtx.UserID, supplier); err != nil {
		if err == service.ErrSupplierNotFound {
			RespondError(c, http.StatusNotFound, "ERR_NOT_FOUND", "Supplier not found")
			return
		}
		if err == service.ErrDuplicateName {
			RespondError(c, http.StatusConflict, "ERR_DUPLICATE_NAME", "Supplier with this name already exists")
			return
		}
		RespondError(c, http.StatusInternalServerError, "ERR_UPDATE_SUPPLIER", err.Error())
		return
	}

	c.JSON(http.StatusOK, supplier)
}

// DeleteSupplier deletes a supplier
func (h *CatalogHandler) DeleteSupplier(c *gin.Context) {
	userCtx := auth.GetUserContext(c)
	if userCtx == nil {
		RespondError(c, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "User not authenticated")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "ERR_INVALID_ID", "Invalid supplier ID")
		return
	}

	if err := h.catalogService.DeleteSupplier(c.Request.Context(), userCtx.UserID, id); err != nil {
		if err == service.ErrSupplierNotFound {
			RespondError(c, http.StatusNotFound, "ERR_NOT_FOUND", "Supplier not found")
			return
		}
		RespondError(c, http.StatusInternalServerError, "ERR_DELETE_SUPPLIER", err.Error())
		return
	}

	c.Status(http.StatusNoContent)
}

// Helper function for validation errors
func RespondValidationErrors(c *gin.Context, errs domain.ValidationErrors) {
	c.JSON(http.StatusBadRequest, gin.H{
		"error":      "ERR_VALIDATION",
		"message":    "Validation failed",
		"validation": errs,
	})
}
