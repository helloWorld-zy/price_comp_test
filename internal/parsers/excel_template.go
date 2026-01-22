package parsers

import (
	"fmt"
	"time"

	"github.com/xuri/excelize/v2"
)

// ExcelTemplateGenerator 生成 Excel 模板
type ExcelTemplateGenerator struct{}

// NewExcelTemplateGenerator 创建 Excel 模板生成器
func NewExcelTemplateGenerator() *ExcelTemplateGenerator {
	return &ExcelTemplateGenerator{}
}

// GenerateSailingTemplate 生成航次导入模板
func (g *ExcelTemplateGenerator) GenerateSailingTemplate() (*excelize.File, error) {
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			// Log error but don't fail
		}
	}()

	sheetName := "航次数据"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to create sheet: %w", err)
	}
	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1")

	// 设置表头
	headers := []string{
		"邮轮公司",
		"邮轮名称",
		"航次编号",
		"出发日期",
		"返回日期",
		"航线描述",
		"停靠港口",
		"备注",
	}

	// 设置表头样式
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   12,
			Color:  "FFFFFF",
			Family: "Arial",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"4472C4"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create header style: %w", err)
	}

	// 写入表头
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		if err := f.SetCellValue(sheetName, cell, header); err != nil {
			return nil, fmt.Errorf("failed to set header: %w", err)
		}
		if err := f.SetCellStyle(sheetName, cell, cell, headerStyle); err != nil {
			return nil, fmt.Errorf("failed to set header style: %w", err)
		}
	}

	// 设置列宽
	columnWidths := []struct {
		col   string
		width float64
	}{
		{"A", 15}, // 邮轮公司
		{"B", 15}, // 邮轮名称
		{"C", 15}, // 航次编号
		{"D", 12}, // 出发日期
		{"E", 12}, // 返回日期
		{"F", 30}, // 航线描述
		{"G", 30}, // 停靠港口
		{"H", 20}, // 备注
	}
	for _, cw := range columnWidths {
		if err := f.SetColWidth(sheetName, cw.col, cw.col, cw.width); err != nil {
			return nil, fmt.Errorf("failed to set column width: %w", err)
		}
	}

	// 添加示例数据
	exampleData := [][]interface{}{
		{"皇家加勒比", "海洋量子号", "QN20260515", "2026-05-15", "2026-05-20", "日本航线", "东京,大阪,福冈", ""},
		{"歌诗达", "大西洋号", "AT20260601", "2026-06-01", "2026-06-08", "地中海航线", "巴塞罗那,那不勒斯,雅典", "含岸上观光"},
	}

	exampleStyle, err := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"FFF2CC"},
			Pattern: 1,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create example style: %w", err)
	}

	for rowIdx, rowData := range exampleData {
		for colIdx, value := range rowData {
			cell := fmt.Sprintf("%c%d", 'A'+colIdx, rowIdx+2)
			if err := f.SetCellValue(sheetName, cell, value); err != nil {
				return nil, fmt.Errorf("failed to set example data: %w", err)
			}
			if err := f.SetCellStyle(sheetName, cell, cell, exampleStyle); err != nil {
				return nil, fmt.Errorf("failed to set example style: %w", err)
			}
		}
	}

	// 添加说明页
	instructionSheet := "填写说明"
	instructionIndex, err := f.NewSheet(instructionSheet)
	if err != nil {
		return nil, fmt.Errorf("failed to create instruction sheet: %w", err)
	}
	_ = instructionIndex

	instructions := []string{
		"航次数据导入说明",
		"",
		"1. 必填字段：",
		"   - 邮轮公司：必须是系统中已存在的邮轮公司名称",
		"   - 邮轮名称：必须是系统中已存在的邮轮名称",
		"   - 出发日期：格式为 YYYY-MM-DD（如 2026-05-15）",
		"   - 返回日期：格式为 YYYY-MM-DD，必须晚于出发日期",
		"   - 航线描述：航线名称或简要说明",
		"",
		"2. 可选字段：",
		"   - 航次编号：邮轮公司的航次编号（如有）",
		"   - 停靠港口：多个港口用逗号分隔",
		"   - 备注：其他补充信息",
		"",
		"3. 注意事项：",
		"   - 黄色背景行是示例数据，请删除后填入真实数据",
		"   - 不要修改表头（第一行）",
		"   - 出发日期必须是未来的日期",
		"   - 同一邮轮不能有重复的航次（相同日期）",
		"",
		"4. 导入流程：",
		"   - 填写完成后保存文件",
		"   - 返回系统上传此文件",
		"   - 系统将自动校验数据",
		"   - 若有错误会给出详细说明",
		"   - 修正错误后可重新上传",
	}

	for i, instruction := range instructions {
		cell := fmt.Sprintf("A%d", i+1)
		if err := f.SetCellValue(instructionSheet, cell, instruction); err != nil {
			return nil, fmt.Errorf("failed to set instruction: %w", err)
		}
	}

	if err := f.SetColWidth(instructionSheet, "A", "A", 60); err != nil {
		return nil, fmt.Errorf("failed to set instruction column width: %w", err)
	}

	return f, nil
}

// GenerateCabinTypeTemplate 生成房型导入模板
func (g *ExcelTemplateGenerator) GenerateCabinTypeTemplate() (*excelize.File, error) {
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			// Log error but don't fail
		}
	}()

	sheetName := "房型数据"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to create sheet: %w", err)
	}
	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1")

	// 设置表头
	headers := []string{
		"邮轮公司",
		"邮轮名称",
		"房型大类",
		"房型名称",
		"房型代码",
		"房型描述",
		"排序",
	}

	// 设置表头样式
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   12,
			Color:  "FFFFFF",
			Family: "Arial",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"4472C4"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create header style: %w", err)
	}

	// 写入表头
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		if err := f.SetCellValue(sheetName, cell, header); err != nil {
			return nil, fmt.Errorf("failed to set header: %w", err)
		}
		if err := f.SetCellStyle(sheetName, cell, cell, headerStyle); err != nil {
			return nil, fmt.Errorf("failed to set header style: %w", err)
		}
	}

	// 设置列宽
	columnWidths := []struct {
		col   string
		width float64
	}{
		{"A", 15}, // 邮轮公司
		{"B", 15}, // 邮轮名称
		{"C", 12}, // 房型大类
		{"D", 20}, // 房型名称
		{"E", 12}, // 房型代码
		{"F", 30}, // 房型描述
		{"G", 8},  // 排序
	}
	for _, cw := range columnWidths {
		if err := f.SetColWidth(sheetName, cw.col, cw.col, cw.width); err != nil {
			return nil, fmt.Errorf("failed to set column width: %w", err)
		}
	}

	// 添加示例数据
	exampleData := [][]interface{}{
		{"皇家加勒比", "海洋量子号", "内舱", "内舱房", "IN", "标准内舱房", 1},
		{"皇家加勒比", "海洋量子号", "海景", "海景房", "OV", "带窗户海景房", 2},
		{"皇家加勒比", "海洋量子号", "阳台", "豪华阳台房", "BA", "带私人阳台", 3},
		{"皇家加勒比", "海洋量子号", "套房", "初级套房", "JS", "小型套房", 4},
	}

	exampleStyle, err := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"FFF2CC"},
			Pattern: 1,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create example style: %w", err)
	}

	for rowIdx, rowData := range exampleData {
		for colIdx, value := range rowData {
			cell := fmt.Sprintf("%c%d", 'A'+colIdx, rowIdx+2)
			if err := f.SetCellValue(sheetName, cell, value); err != nil {
				return nil, fmt.Errorf("failed to set example data: %w", err)
			}
			if err := f.SetCellStyle(sheetName, cell, cell, exampleStyle); err != nil {
				return nil, fmt.Errorf("failed to set example style: %w", err)
			}
		}
	}

	// 添加说明页
	instructionSheet := "填写说明"
	instructionIndex, err := f.NewSheet(instructionSheet)
	if err != nil {
		return nil, fmt.Errorf("failed to create instruction sheet: %w", err)
	}
	_ = instructionIndex

	instructions := []string{
		"房型数据导入说明",
		"",
		"1. 必填字段：",
		"   - 邮轮公司：必须是系统中已存在的邮轮公司名称",
		"   - 邮轮名称：必须是系统中已存在的邮轮名称",
		"   - 房型大类：必须是以下之一：内舱、海景、阳台、套房",
		"   - 房型名称：房型的具体名称",
		"",
		"2. 可选字段：",
		"   - 房型代码：邮轮公司的房型代码（如 BA、JS 等）",
		"   - 房型描述：房型的详细描述",
		"   - 排序：数字，用于控制显示顺序（默认为 0）",
		"",
		"3. 注意事项：",
		"   - 黄色背景行是示例数据，请删除后填入真实数据",
		"   - 不要修改表头（第一行）",
		"   - 同一邮轮的房型名称不能重复",
		"   - 房型大类必须严格匹配（区分大小写）",
		"",
		"4. 导入流程：",
		"   - 填写完成后保存文件",
		"   - 返回系统上传此文件",
		"   - 系统将自动校验数据",
		"   - 若有错误会给出详细说明",
		"   - 修正错误后可重新上传",
	}

	for i, instruction := range instructions {
		cell := fmt.Sprintf("A%d", i+1)
		if err := f.SetCellValue(instructionSheet, cell, instruction); err != nil {
			return nil, fmt.Errorf("failed to set instruction: %w", err)
		}
	}

	if err := f.SetColWidth(instructionSheet, "A", "A", 60); err != nil {
		return nil, fmt.Errorf("failed to set instruction column width: %w", err)
	}

	return f, nil
}

// SailingRowData 航次行数据
type SailingRowData struct {
	RowNumber      int
	CruiseLineName string
	ShipName       string
	SailingCode    string
	DepartureDate  string
	ReturnDate     string
	Route          string
	Ports          string
	Notes          string
}

// CabinTypeRowData 房型行数据
type CabinTypeRowData struct {
	RowNumber      int
	CruiseLineName string
	ShipName       string
	CategoryName   string
	CabinTypeName  string
	CabinTypeCode  string
	Description    string
	SortOrder      int
}

// ParseSailingExcel 解析航次 Excel 文件
func ParseSailingExcel(filePath string) ([]SailingRowData, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	sheetName := "航次数据"
	sheets := f.GetSheetList()
	found := false
	for _, s := range sheets {
		if s == sheetName {
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("sheet '航次数据' not found")
	}

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to get rows: %w", err)
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("no data rows found")
	}

	var result []SailingRowData
	for i, row := range rows {
		if i == 0 {
			// Skip header
			continue
		}

		// Skip empty rows
		if len(row) == 0 || row[0] == "" {
			continue
		}

		data := SailingRowData{
			RowNumber: i + 1,
		}

		if len(row) > 0 {
			data.CruiseLineName = row[0]
		}
		if len(row) > 1 {
			data.ShipName = row[1]
		}
		if len(row) > 2 {
			data.SailingCode = row[2]
		}
		if len(row) > 3 {
			data.DepartureDate = row[3]
		}
		if len(row) > 4 {
			data.ReturnDate = row[4]
		}
		if len(row) > 5 {
			data.Route = row[5]
		}
		if len(row) > 6 {
			data.Ports = row[6]
		}
		if len(row) > 7 {
			data.Notes = row[7]
		}

		result = append(result, data)
	}

	return result, nil
}

// ParseCabinTypeExcel 解析房型 Excel 文件
func ParseCabinTypeExcel(filePath string) ([]CabinTypeRowData, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	sheetName := "房型数据"
	sheets := f.GetSheetList()
	found := false
	for _, s := range sheets {
		if s == sheetName {
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("sheet '房型数据' not found")
	}

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to get rows: %w", err)
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("no data rows found")
	}

	var result []CabinTypeRowData
	for i, row := range rows {
		if i == 0 {
			// Skip header
			continue
		}

		// Skip empty rows
		if len(row) == 0 || row[0] == "" {
			continue
		}

		data := CabinTypeRowData{
			RowNumber: i + 1,
		}

		if len(row) > 0 {
			data.CruiseLineName = row[0]
		}
		if len(row) > 1 {
			data.ShipName = row[1]
		}
		if len(row) > 2 {
			data.CategoryName = row[2]
		}
		if len(row) > 3 {
			data.CabinTypeName = row[3]
		}
		if len(row) > 4 {
			data.CabinTypeCode = row[4]
		}
		if len(row) > 5 {
			data.Description = row[5]
		}
		if len(row) > 6 {
			sortOrder := 0
			fmt.Sscanf(row[6], "%d", &sortOrder)
			data.SortOrder = sortOrder
		}

		result = append(result, data)
	}

	return result, nil
}

// ValidateSailingRow 验证航次行数据
func ValidateSailingRow(row SailingRowData) []string {
	var errors []string

	if row.CruiseLineName == "" {
		errors = append(errors, "邮轮公司不能为空")
	}
	if row.ShipName == "" {
		errors = append(errors, "邮轮名称不能为空")
	}
	if row.DepartureDate == "" {
		errors = append(errors, "出发日期不能为空")
	} else {
		if _, err := time.Parse("2006-01-02", row.DepartureDate); err != nil {
			errors = append(errors, "出发日期格式错误，应为 YYYY-MM-DD")
		}
	}
	if row.ReturnDate == "" {
		errors = append(errors, "返回日期不能为空")
	} else {
		if _, err := time.Parse("2006-01-02", row.ReturnDate); err != nil {
			errors = append(errors, "返回日期格式错误，应为 YYYY-MM-DD")
		}
	}
	if row.Route == "" {
		errors = append(errors, "航线描述不能为空")
	}

	// Validate date order
	if row.DepartureDate != "" && row.ReturnDate != "" {
		dept, err1 := time.Parse("2006-01-02", row.DepartureDate)
		ret, err2 := time.Parse("2006-01-02", row.ReturnDate)
		if err1 == nil && err2 == nil {
			if !ret.After(dept) {
				errors = append(errors, "返回日期必须晚于出发日期")
			}
		}
	}

	return errors
}

// ValidateCabinTypeRow 验证房型行数据
func ValidateCabinTypeRow(row CabinTypeRowData) []string {
	var errors []string

	if row.CruiseLineName == "" {
		errors = append(errors, "邮轮公司不能为空")
	}
	if row.ShipName == "" {
		errors = append(errors, "邮轮名称不能为空")
	}
	if row.CategoryName == "" {
		errors = append(errors, "房型大类不能为空")
	} else {
		validCategories := map[string]bool{
			"内舱": true,
			"海景": true,
			"阳台": true,
			"套房": true,
		}
		if !validCategories[row.CategoryName] {
			errors = append(errors, "房型大类必须是：内舱、海景、阳台、套房之一")
		}
	}
	if row.CabinTypeName == "" {
		errors = append(errors, "房型名称不能为空")
	}

	return errors
}
