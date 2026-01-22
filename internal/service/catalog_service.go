package service

import (
	"context"
	"errors"
	"fmt"

	"cruise-price-compare/internal/domain"
	"cruise-price-compare/internal/obs"
	"cruise-price-compare/internal/repo"
)

var (
	ErrCruiseLineNotFound    = errors.New("cruise line not found")
	ErrShipNotFound          = errors.New("ship not found")
	ErrCabinCategoryNotFound = errors.New("cabin category not found")
	ErrCabinTypeNotFound     = errors.New("cabin type not found")
	ErrSailingNotFound       = errors.New("sailing not found")
	ErrSupplierNotFound      = errors.New("supplier not found")
	ErrDuplicateName         = errors.New("duplicate name")
)

// CatalogService handles catalog operations
type CatalogService struct {
	cruiseLineRepo    *repo.CruiseLineRepository
	shipRepo          *repo.ShipRepository
	cabinCategoryRepo *repo.CabinCategoryRepository
	cabinTypeRepo     *repo.CabinTypeRepository
	sailingRepo       *repo.SailingRepository
	supplierRepo      *repo.SupplierRepository
	audit             *obs.AuditService
	logger            *obs.Logger
}

// NewCatalogService creates a new catalog service
func NewCatalogService(
	cruiseLineRepo *repo.CruiseLineRepository,
	shipRepo *repo.ShipRepository,
	cabinCategoryRepo *repo.CabinCategoryRepository,
	cabinTypeRepo *repo.CabinTypeRepository,
	sailingRepo *repo.SailingRepository,
	supplierRepo *repo.SupplierRepository,
	audit *obs.AuditService,
	logger *obs.Logger,
) *CatalogService {
	return &CatalogService{
		cruiseLineRepo:    cruiseLineRepo,
		shipRepo:          shipRepo,
		cabinCategoryRepo: cabinCategoryRepo,
		cabinTypeRepo:     cabinTypeRepo,
		sailingRepo:       sailingRepo,
		supplierRepo:      supplierRepo,
		audit:             audit,
		logger:            logger,
	}
}

// CruiseLine operations

func (s *CatalogService) GetCruiseLine(ctx context.Context, id uint64) (*domain.CruiseLine, error) {
	return s.cruiseLineRepo.GetByID(ctx, id)
}

func (s *CatalogService) ListCruiseLines(ctx context.Context, pagination repo.Pagination, status *domain.EntityStatus) (repo.PaginatedResult[domain.CruiseLine], error) {
	return s.cruiseLineRepo.List(ctx, pagination, status)
}

func (s *CatalogService) CreateCruiseLine(ctx context.Context, userID uint64, cl *domain.CruiseLine) error {
	exists, err := s.cruiseLineRepo.ExistsByName(ctx, cl.Name, nil)
	if err != nil {
		return fmt.Errorf("failed to check cruise line exists: %w", err)
	}
	if exists {
		return ErrDuplicateName
	}

	cl.Status = domain.EntityStatusActive
	createdBy := userID
	cl.CreatedBy = &createdBy

	if err := s.cruiseLineRepo.Create(ctx, cl); err != nil {
		return fmt.Errorf("failed to create cruise line: %w", err)
	}

	_ = s.audit.LogCreate(ctx, userID, nil, domain.EntityTypeCruiseLine, cl.ID, cl)
	return nil
}

func (s *CatalogService) UpdateCruiseLine(ctx context.Context, userID uint64, cl *domain.CruiseLine) error {
	old, err := s.cruiseLineRepo.GetByID(ctx, cl.ID)
	if err != nil {
		return fmt.Errorf("failed to get cruise line: %w", err)
	}
	if old == nil {
		return ErrCruiseLineNotFound
	}

	exists, err := s.cruiseLineRepo.ExistsByName(ctx, cl.Name, &cl.ID)
	if err != nil {
		return fmt.Errorf("failed to check cruise line exists: %w", err)
	}
	if exists {
		return ErrDuplicateName
	}

	if err := s.cruiseLineRepo.Update(ctx, cl); err != nil {
		return fmt.Errorf("failed to update cruise line: %w", err)
	}

	_ = s.audit.LogUpdate(ctx, userID, nil, domain.EntityTypeCruiseLine, cl.ID, old, cl)
	return nil
}

func (s *CatalogService) DeleteCruiseLine(ctx context.Context, userID uint64, id uint64) error {
	old, err := s.cruiseLineRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get cruise line: %w", err)
	}
	if old == nil {
		return ErrCruiseLineNotFound
	}

	if err := s.cruiseLineRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete cruise line: %w", err)
	}

	_ = s.audit.LogDelete(ctx, userID, nil, domain.EntityTypeCruiseLine, id, old)
	return nil
}

// Ship operations

func (s *CatalogService) GetShip(ctx context.Context, id uint64) (*domain.Ship, error) {
	return s.shipRepo.GetByID(ctx, id)
}

func (s *CatalogService) ListShips(ctx context.Context, pagination repo.Pagination, cruiseLineID *uint64, status *domain.EntityStatus) (repo.PaginatedResult[domain.Ship], error) {
	return s.shipRepo.List(ctx, pagination, cruiseLineID, status)
}

func (s *CatalogService) ListShipsByCruiseLine(ctx context.Context, cruiseLineID uint64) ([]domain.Ship, error) {
	return s.shipRepo.ListByCruiseLine(ctx, cruiseLineID)
}

func (s *CatalogService) CreateShip(ctx context.Context, userID uint64, ship *domain.Ship) error {
	exists, err := s.shipRepo.ExistsByName(ctx, ship.CruiseLineID, ship.Name, nil)
	if err != nil {
		return fmt.Errorf("failed to check ship exists: %w", err)
	}
	if exists {
		return ErrDuplicateName
	}

	ship.Status = domain.EntityStatusActive
	createdBy := userID
	ship.CreatedBy = &createdBy

	if err := s.shipRepo.Create(ctx, ship); err != nil {
		return fmt.Errorf("failed to create ship: %w", err)
	}

	_ = s.audit.LogCreate(ctx, userID, nil, domain.EntityTypeShip, ship.ID, ship)
	return nil
}

func (s *CatalogService) UpdateShip(ctx context.Context, userID uint64, ship *domain.Ship) error {
	old, err := s.shipRepo.GetByID(ctx, ship.ID)
	if err != nil {
		return fmt.Errorf("failed to get ship: %w", err)
	}
	if old == nil {
		return ErrShipNotFound
	}

	exists, err := s.shipRepo.ExistsByName(ctx, ship.CruiseLineID, ship.Name, &ship.ID)
	if err != nil {
		return fmt.Errorf("failed to check ship exists: %w", err)
	}
	if exists {
		return ErrDuplicateName
	}

	if err := s.shipRepo.Update(ctx, ship); err != nil {
		return fmt.Errorf("failed to update ship: %w", err)
	}

	_ = s.audit.LogUpdate(ctx, userID, nil, domain.EntityTypeShip, ship.ID, old, ship)
	return nil
}

func (s *CatalogService) DeleteShip(ctx context.Context, userID uint64, id uint64) error {
	old, err := s.shipRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get ship: %w", err)
	}
	if old == nil {
		return ErrShipNotFound
	}

	if err := s.shipRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete ship: %w", err)
	}

	_ = s.audit.LogDelete(ctx, userID, nil, domain.EntityTypeShip, id, old)
	return nil
}

// CabinCategory operations

func (s *CatalogService) GetCabinCategory(ctx context.Context, id uint64) (*domain.CabinCategory, error) {
	return s.cabinCategoryRepo.GetByID(ctx, id)
}

func (s *CatalogService) ListCabinCategories(ctx context.Context) ([]domain.CabinCategory, error) {
	return s.cabinCategoryRepo.List(ctx)
}

func (s *CatalogService) CreateCabinCategory(ctx context.Context, userID uint64, cc *domain.CabinCategory) error {
	exists, err := s.cabinCategoryRepo.ExistsByName(ctx, cc.Name, nil)
	if err != nil {
		return fmt.Errorf("failed to check cabin category exists: %w", err)
	}
	if exists {
		return ErrDuplicateName
	}

	if err := s.cabinCategoryRepo.Create(ctx, cc); err != nil {
		return fmt.Errorf("failed to create cabin category: %w", err)
	}

	_ = s.audit.LogCreate(ctx, userID, nil, domain.EntityTypeCabinCategory, cc.ID, cc)
	return nil
}

func (s *CatalogService) UpdateCabinCategory(ctx context.Context, userID uint64, cc *domain.CabinCategory) error {
	old, err := s.cabinCategoryRepo.GetByID(ctx, cc.ID)
	if err != nil {
		return fmt.Errorf("failed to get cabin category: %w", err)
	}
	if old == nil {
		return ErrCabinCategoryNotFound
	}

	if err := s.cabinCategoryRepo.Update(ctx, cc); err != nil {
		return fmt.Errorf("failed to update cabin category: %w", err)
	}

	_ = s.audit.LogUpdate(ctx, userID, nil, domain.EntityTypeCabinCategory, cc.ID, old, cc)
	return nil
}

func (s *CatalogService) DeleteCabinCategory(ctx context.Context, userID uint64, id uint64) error {
	old, err := s.cabinCategoryRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get cabin category: %w", err)
	}
	if old == nil {
		return ErrCabinCategoryNotFound
	}

	if err := s.cabinCategoryRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete cabin category: %w", err)
	}

	_ = s.audit.LogDelete(ctx, userID, nil, domain.EntityTypeCabinCategory, id, old)
	return nil
}

// CabinType operations

func (s *CatalogService) GetCabinType(ctx context.Context, id uint64) (*domain.CabinType, error) {
	return s.cabinTypeRepo.GetByID(ctx, id)
}

func (s *CatalogService) ListCabinTypes(ctx context.Context, pagination repo.Pagination, shipID, categoryID *uint64, enabledOnly bool) (repo.PaginatedResult[domain.CabinType], error) {
	return s.cabinTypeRepo.List(ctx, pagination, shipID, categoryID, enabledOnly)
}

func (s *CatalogService) ListCabinTypesByShip(ctx context.Context, shipID uint64) ([]domain.CabinType, error) {
	return s.cabinTypeRepo.ListByShip(ctx, shipID)
}

func (s *CatalogService) CreateCabinType(ctx context.Context, userID uint64, ct *domain.CabinType) error {
	ct.IsEnabled = true

	if err := s.cabinTypeRepo.Create(ctx, ct); err != nil {
		return fmt.Errorf("failed to create cabin type: %w", err)
	}

	_ = s.audit.LogCreate(ctx, userID, nil, domain.EntityTypeCabinType, ct.ID, ct)
	return nil
}

func (s *CatalogService) UpdateCabinType(ctx context.Context, userID uint64, ct *domain.CabinType) error {
	old, err := s.cabinTypeRepo.GetByID(ctx, ct.ID)
	if err != nil {
		return fmt.Errorf("failed to get cabin type: %w", err)
	}
	if old == nil {
		return ErrCabinTypeNotFound
	}

	if err := s.cabinTypeRepo.Update(ctx, ct); err != nil {
		return fmt.Errorf("failed to update cabin type: %w", err)
	}

	_ = s.audit.LogUpdate(ctx, userID, nil, domain.EntityTypeCabinType, ct.ID, old, ct)
	return nil
}

func (s *CatalogService) DeleteCabinType(ctx context.Context, userID uint64, id uint64) error {
	old, err := s.cabinTypeRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get cabin type: %w", err)
	}
	if old == nil {
		return ErrCabinTypeNotFound
	}

	if err := s.cabinTypeRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete cabin type: %w", err)
	}

	_ = s.audit.LogDelete(ctx, userID, nil, domain.EntityTypeCabinType, id, old)
	return nil
}

// Sailing operations

func (s *CatalogService) GetSailing(ctx context.Context, id uint64) (*domain.Sailing, error) {
	return s.sailingRepo.GetByID(ctx, id)
}

func (s *CatalogService) ListSailings(ctx context.Context, pagination repo.Pagination, shipID *uint64, status *domain.SailingStatus) (repo.PaginatedResult[domain.Sailing], error) {
	return s.sailingRepo.List(ctx, pagination, shipID, status, nil, nil)
}

func (s *CatalogService) CreateSailing(ctx context.Context, userID uint64, sailing *domain.Sailing) error {
	sailing.Status = domain.SailingStatusActive
	createdBy := userID
	sailing.CreatedBy = &createdBy

	if err := s.sailingRepo.Create(ctx, sailing); err != nil {
		return fmt.Errorf("failed to create sailing: %w", err)
	}

	_ = s.audit.LogCreate(ctx, userID, nil, domain.EntityTypeSailing, sailing.ID, sailing)
	return nil
}

func (s *CatalogService) UpdateSailing(ctx context.Context, userID uint64, sailing *domain.Sailing) error {
	old, err := s.sailingRepo.GetByID(ctx, sailing.ID)
	if err != nil {
		return fmt.Errorf("failed to get sailing: %w", err)
	}
	if old == nil {
		return ErrSailingNotFound
	}

	if err := s.sailingRepo.Update(ctx, sailing); err != nil {
		return fmt.Errorf("failed to update sailing: %w", err)
	}

	_ = s.audit.LogUpdate(ctx, userID, nil, domain.EntityTypeSailing, sailing.ID, old, sailing)
	return nil
}

func (s *CatalogService) DeleteSailing(ctx context.Context, userID uint64, id uint64) error {
	old, err := s.sailingRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get sailing: %w", err)
	}
	if old == nil {
		return ErrSailingNotFound
	}

	if err := s.sailingRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete sailing: %w", err)
	}

	_ = s.audit.LogDelete(ctx, userID, nil, domain.EntityTypeSailing, id, old)
	return nil
}

// Supplier operations

func (s *CatalogService) GetSupplier(ctx context.Context, id uint64) (*domain.Supplier, error) {
	return s.supplierRepo.GetByID(ctx, id)
}

func (s *CatalogService) ListSuppliers(ctx context.Context, pagination repo.Pagination, status *domain.EntityStatus) (repo.PaginatedResult[domain.Supplier], error) {
	return s.supplierRepo.List(ctx, pagination, status)
}

func (s *CatalogService) CreateSupplier(ctx context.Context, userID uint64, supplier *domain.Supplier) error {
	exists, err := s.supplierRepo.ExistsByName(ctx, supplier.Name, nil)
	if err != nil {
		return fmt.Errorf("failed to check supplier exists: %w", err)
	}
	if exists {
		return ErrDuplicateName
	}

	supplier.Status = domain.EntityStatusActive
	createdBy := userID
	supplier.CreatedBy = &createdBy

	if err := s.supplierRepo.Create(ctx, supplier); err != nil {
		return fmt.Errorf("failed to create supplier: %w", err)
	}

	_ = s.audit.LogCreate(ctx, userID, nil, domain.EntityTypeSupplier, supplier.ID, supplier)
	return nil
}

func (s *CatalogService) UpdateSupplier(ctx context.Context, userID uint64, supplier *domain.Supplier) error {
	old, err := s.supplierRepo.GetByID(ctx, supplier.ID)
	if err != nil {
		return fmt.Errorf("failed to get supplier: %w", err)
	}
	if old == nil {
		return ErrSupplierNotFound
	}

	exists, err := s.supplierRepo.ExistsByName(ctx, supplier.Name, &supplier.ID)
	if err != nil {
		return fmt.Errorf("failed to check supplier exists: %w", err)
	}
	if exists {
		return ErrDuplicateName
	}

	if err := s.supplierRepo.Update(ctx, supplier); err != nil {
		return fmt.Errorf("failed to update supplier: %w", err)
	}

	_ = s.audit.LogUpdate(ctx, userID, nil, domain.EntityTypeSupplier, supplier.ID, old, supplier)
	return nil
}

func (s *CatalogService) DeleteSupplier(ctx context.Context, userID uint64, id uint64) error {
	old, err := s.supplierRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get supplier: %w", err)
	}
	if old == nil {
		return ErrSupplierNotFound
	}

	if err := s.supplierRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete supplier: %w", err)
	}

	_ = s.audit.LogDelete(ctx, userID, nil, domain.EntityTypeSupplier, id, old)
	return nil
}
