package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cruise-price-compare/internal/domain"
	"cruise-price-compare/internal/obs"
	"cruise-price-compare/internal/parsers"
	"cruise-price-compare/internal/repo"

	"github.com/xuri/excelize/v2"
)

// TemplateImportService 模板导入服务
type TemplateImportService struct {
	cruiseLineRepo    *repo.CruiseLineRepository
	shipRepo          *repo.ShipRepository
	cabinCategoryRepo *repo.CabinCategoryRepository
	cabinTypeRepo     *repo.CabinTypeRepository
	sailingRepo       *repo.SailingRepository
	auditService      *obs.AuditService
	logger            obs.Logger
}

// NewTemplateImportService 创建模板导入服务
func NewTemplateImportService(
	cruiseLineRepo *repo.CruiseLineRepository,
	shipRepo *repo.ShipRepository,
	cabinCategoryRepo *repo.CabinCategoryRepository,
	cabinTypeRepo *repo.CabinTypeRepository,
	sailingRepo *repo.SailingRepository,
	auditService *obs.AuditService,
	logger obs.Logger,
) *TemplateImportService {
	return &TemplateImportService{
		cruiseLineRepo:    cruiseLineRepo,
		shipRepo:          shipRepo,
		cabinCategoryRepo: cabinCategoryRepo,
		cabinTypeRepo:     cabinTypeRepo,
		sailingRepo:       sailingRepo,
		auditService:      auditService,
		logger:            logger,
	}
}

// ImportResult 导入结果
type ImportResult struct {
	TotalRows   int              `json:"total_rows"`
	SuccessRows int              `json:"success_rows"`
	ErrorRows   int              `json:"error_rows"`
	Errors      []ImportRowError `json:"errors"`
	CreatedIDs  []uint64         `json:"created_ids"`
}

// ImportRowError 行错误
type ImportRowError struct {
	RowNumber int      `json:"row_number"`
	Errors    []string `json:"errors"`
}

// GenerateSailingTemplate 生成航次模板
func (s *TemplateImportService) GenerateSailingTemplate(ctx context.Context) (*excelize.File, error) {
	generator := parsers.NewExcelTemplateGenerator()
	return generator.GenerateSailingTemplate()
}

// GenerateCabinTypeTemplate 生成房型模板
func (s *TemplateImportService) GenerateCabinTypeTemplate(ctx context.Context) (*excelize.File, error) {
	generator := parsers.NewExcelTemplateGenerator()
	return generator.GenerateCabinTypeTemplate()
}

// ImportSailingTemplate 导入航次模板
func (s *TemplateImportService) ImportSailingTemplate(ctx context.Context, filePath string, userID uint64) (*ImportResult, error) {
	// 解析 Excel 文件
	rows, err := parsers.ParseSailingExcel(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse excel: %w", err)
	}

	result := &ImportResult{
		TotalRows:   len(rows),
		SuccessRows: 0,
		ErrorRows:   0,
		Errors:      []ImportRowError{},
		CreatedIDs:  []uint64{},
	}

	for _, row := range rows {
		// 验证行数据
		validationErrors := parsers.ValidateSailingRow(row)
		if len(validationErrors) > 0 {
			result.ErrorRows++
			result.Errors = append(result.Errors, ImportRowError{
				RowNumber: row.RowNumber,
				Errors:    validationErrors,
			})
			continue
		}

		// 创建航次
		sailingID, err := s.createSailing(ctx, row, userID)
		if err != nil {
			result.ErrorRows++
			result.Errors = append(result.Errors, ImportRowError{
				RowNumber: row.RowNumber,
				Errors:    []string{err.Error()},
			})
			continue
		}

		result.SuccessRows++
		result.CreatedIDs = append(result.CreatedIDs, sailingID)
	}

	return result, nil
}

// ImportCabinTypeTemplate 导入房型模板
func (s *TemplateImportService) ImportCabinTypeTemplate(ctx context.Context, filePath string, userID uint64) (*ImportResult, error) {
	// 解析 Excel 文件
	rows, err := parsers.ParseCabinTypeExcel(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse excel: %w", err)
	}

	result := &ImportResult{
		TotalRows:   len(rows),
		SuccessRows: 0,
		ErrorRows:   0,
		Errors:      []ImportRowError{},
		CreatedIDs:  []uint64{},
	}

	for _, row := range rows {
		// 验证行数据
		validationErrors := parsers.ValidateCabinTypeRow(row)
		if len(validationErrors) > 0 {
			result.ErrorRows++
			result.Errors = append(result.Errors, ImportRowError{
				RowNumber: row.RowNumber,
				Errors:    validationErrors,
			})
			continue
		}

		// 创建房型
		cabinTypeID, err := s.createCabinType(ctx, row, userID)
		if err != nil {
			result.ErrorRows++
			result.Errors = append(result.Errors, ImportRowError{
				RowNumber: row.RowNumber,
				Errors:    []string{err.Error()},
			})
			continue
		}

		result.SuccessRows++
		result.CreatedIDs = append(result.CreatedIDs, cabinTypeID)
	}

	return result, nil
}

// createSailing 创建航次
func (s *TemplateImportService) createSailing(ctx context.Context, row parsers.SailingRowData, userID uint64) (uint64, error) {
	// 查找邮轮公司
	pagination := repo.Pagination{Page: 1, PageSize: 100}
	activeStatus := domain.EntityStatusActive
	cruiseLineResult, err := s.cruiseLineRepo.List(ctx, pagination, &activeStatus)
	if err != nil {
		return 0, fmt.Errorf("failed to list cruise lines: %w", err)
	}

	var cruiseLineID uint64
	for _, cl := range cruiseLineResult.Items {
		if cl.Name == row.CruiseLineName {
			cruiseLineID = cl.ID
			break
		}
	}
	if cruiseLineID == 0 {
		return 0, fmt.Errorf("cruise line '%s' not found", row.CruiseLineName)
	}

	// 查找邮轮
	shipResult, err := s.shipRepo.List(ctx, pagination, &cruiseLineID, &activeStatus)
	if err != nil {
		return 0, fmt.Errorf("failed to list ships: %w", err)
	}

	var shipID uint64
	for _, ship := range shipResult.Items {
		if ship.Name == row.ShipName {
			shipID = ship.ID
			break
		}
	}
	if shipID == 0 {
		return 0, fmt.Errorf("ship '%s' not found in cruise line '%s'", row.ShipName, row.CruiseLineName)
	}

	// 解析日期
	departureDate, err := time.Parse("2006-01-02", row.DepartureDate)
	if err != nil {
		return 0, fmt.Errorf("invalid departure date: %w", err)
	}

	returnDate, err := time.Parse("2006-01-02", row.ReturnDate)
	if err != nil {
		return 0, fmt.Errorf("invalid return date: %w", err)
	}

	// 计算晚数
	nights := int(returnDate.Sub(departureDate).Hours() / 24)

	// 解析停靠港口
	var ports []string
	if row.Ports != "" {
		ports = strings.Split(row.Ports, ",")
		for i := range ports {
			ports[i] = strings.TrimSpace(ports[i])
		}
	}

	// 创建航次
	createdByPtr := &userID
	sailing := &domain.Sailing{
		ShipID:        shipID,
		SailingCode:   row.SailingCode,
		DepartureDate: departureDate,
		ReturnDate:    returnDate,
		Nights:        nights,
		Route:         row.Route,
		Ports:         ports,
		Description:   row.Notes,
		Status:        domain.SailingStatusActive,
		CreatedBy:     createdByPtr,
	}

	// 检查是否已存在
	existingSailings, err := s.sailingRepo.ListByShip(ctx, shipID)
	if err != nil {
		return 0, fmt.Errorf("failed to check existing sailings: %w", err)
	}

	for _, existing := range existingSailings {
		if existing.DepartureDate.Format("2006-01-02") == departureDate.Format("2006-01-02") {
			return 0, fmt.Errorf("sailing already exists for ship '%s' on %s", row.ShipName, row.DepartureDate)
		}
	}

	// 创建航次
	err = s.sailingRepo.Create(ctx, sailing)
	if err != nil {
		return 0, fmt.Errorf("failed to create sailing: %w", err)
	}
	sailingID := sailing.ID

	// 记录审计日志
	_ = s.auditService.LogCreate(ctx, userID, nil, "sailing", sailingID, sailing)

	return sailingID, nil
}

// createCabinType 创建房型
func (s *TemplateImportService) createCabinType(ctx context.Context, row parsers.CabinTypeRowData, userID uint64) (uint64, error) {
	// 查找邮轮公司
	pagination := repo.Pagination{Page: 1, PageSize: 100}
	activeStatus := domain.EntityStatusActive
	cruiseLineResult, err := s.cruiseLineRepo.List(ctx, pagination, &activeStatus)
	if err != nil {
		return 0, fmt.Errorf("failed to list cruise lines: %w", err)
	}

	var cruiseLineID uint64
	for _, cl := range cruiseLineResult.Items {
		if cl.Name == row.CruiseLineName {
			cruiseLineID = cl.ID
			break
		}
	}
	if cruiseLineID == 0 {
		return 0, fmt.Errorf("cruise line '%s' not found", row.CruiseLineName)
	}

	// 查找邮轮
	shipResult, err := s.shipRepo.List(ctx, pagination, &cruiseLineID, &activeStatus)
	if err != nil {
		return 0, fmt.Errorf("failed to list ships: %w", err)
	}

	var shipID uint64
	for _, ship := range shipResult.Items {
		if ship.Name == row.ShipName {
			shipID = ship.ID
			break
		}
	}
	if shipID == 0 {
		return 0, fmt.Errorf("ship '%s' not found in cruise line '%s'", row.ShipName, row.CruiseLineName)
	}

	// 查找房型大类
	categories, err := s.cabinCategoryRepo.List(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to list categories: %w", err)
	}

	var categoryID uint64
	for _, cat := range categories {
		if cat.Name == row.CategoryName {
			categoryID = cat.ID
			break
		}
	}
	if categoryID == 0 {
		return 0, fmt.Errorf("cabin category '%s' not found", row.CategoryName)
	}

	// 检查房型是否已存在
	existingCabinTypes, err := s.cabinTypeRepo.ListByShip(ctx, shipID)
	if err != nil {
		return 0, fmt.Errorf("failed to check existing cabin types: %w", err)
	}

	for _, existing := range existingCabinTypes {
		if existing.Name == row.CabinTypeName {
			return 0, fmt.Errorf("cabin type '%s' already exists for ship '%s'", row.CabinTypeName, row.ShipName)
		}
		if row.CabinTypeCode != "" && existing.Code != "" && existing.Code == row.CabinTypeCode {
			return 0, fmt.Errorf("cabin type code '%s' already exists for ship '%s'", row.CabinTypeCode, row.ShipName)
		}
	}

	// 创建房型
	cabinType := &domain.CabinType{
		ShipID:      shipID,
		CategoryID:  categoryID,
		Name:        row.CabinTypeName,
		Code:        row.CabinTypeCode,
		Description: row.Description,
		SortOrder:   row.SortOrder,
		IsEnabled:   true,
	}

	err = s.cabinTypeRepo.Create(ctx, cabinType)
	if err != nil {
		return 0, fmt.Errorf("failed to create cabin type: %w", err)
	}
	cabinTypeID := cabinType.ID

	// 记录审计日志
	_ = s.auditService.LogCreate(ctx, userID, nil, "cabin_type", cabinTypeID, cabinType)

	return cabinTypeID, nil
}
