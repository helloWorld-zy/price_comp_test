package http

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"cruise-price-compare/internal/auth"
	"cruise-price-compare/internal/domain"
	"cruise-price-compare/internal/service"

	"github.com/gin-gonic/gin"
)

// TemplateHandler 模板导入处理器
type TemplateHandler struct {
	templateService *service.TemplateImportService
}

// NewTemplateHandler 创建模板处理器
func NewTemplateHandler(templateService *service.TemplateImportService) *TemplateHandler {
	return &TemplateHandler{
		templateService: templateService,
	}
}

// DownloadSailingTemplate 下载航次模板
// GET /api/v1/template/sailing/download
func (h *TemplateHandler) DownloadSailingTemplate(c *gin.Context) {
	userCtx := auth.GetUserContext(c)
	if userCtx == nil {
		RespondError(c, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "User not authenticated")
		return
	}

	// 只有管理员可以下载模板
	if userCtx.Role != domain.UserRoleAdmin {
		RespondError(c, http.StatusForbidden, "ERR_FORBIDDEN", "Only admins can download templates")
		return
	}

	// 生成模板
	file, err := h.templateService.GenerateSailingTemplate(c.Request.Context())
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "ERR_GENERATE_TEMPLATE", err.Error())
		return
	}
	defer file.Close()

	// 设置响应头
	filename := fmt.Sprintf("sailing_template_%s.xlsx", time.Now().Format("20060102_150405"))
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")

	// 写入响应
	if err := file.Write(c.Writer); err != nil {
		RespondError(c, http.StatusInternalServerError, "ERR_WRITE_FILE", err.Error())
		return
	}
}

// DownloadCabinTypeTemplate 下载房型模板
// GET /api/v1/template/cabin-type/download
func (h *TemplateHandler) DownloadCabinTypeTemplate(c *gin.Context) {
	userCtx := auth.GetUserContext(c)
	if userCtx == nil {
		RespondError(c, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "User not authenticated")
		return
	}

	// 只有管理员可以下载模板
	if userCtx.Role != domain.UserRoleAdmin {
		RespondError(c, http.StatusForbidden, "ERR_FORBIDDEN", "Only admins can download templates")
		return
	}

	// 生成模板
	file, err := h.templateService.GenerateCabinTypeTemplate(c.Request.Context())
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "ERR_GENERATE_TEMPLATE", err.Error())
		return
	}
	defer file.Close()

	// 设置响应头
	filename := fmt.Sprintf("cabin_type_template_%s.xlsx", time.Now().Format("20060102_150405"))
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")

	// 写入响应
	if err := file.Write(c.Writer); err != nil {
		RespondError(c, http.StatusInternalServerError, "ERR_WRITE_FILE", err.Error())
		return
	}
}

// UploadSailingTemplate 上传并导入航次模板
// POST /api/v1/template/sailing/import
func (h *TemplateHandler) UploadSailingTemplate(c *gin.Context) {
	userCtx := auth.GetUserContext(c)
	if userCtx == nil {
		RespondError(c, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "User not authenticated")
		return
	}

	// 只有管理员可以导入模板
	if userCtx.Role != domain.UserRoleAdmin {
		RespondError(c, http.StatusForbidden, "ERR_FORBIDDEN", "Only admins can import templates")
		return
	}

	// 解析上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		RespondError(c, http.StatusBadRequest, "ERR_INVALID_FILE", "File is required")
		return
	}

	// 验证文件类型
	if filepath.Ext(file.Filename) != ".xlsx" {
		RespondError(c, http.StatusBadRequest, "ERR_INVALID_FILE_TYPE", "Only .xlsx files are supported")
		return
	}

	// 验证文件大小（最大 5MB）
	const maxFileSize = 5 * 1024 * 1024
	if file.Size > maxFileSize {
		RespondError(c, http.StatusBadRequest, "ERR_FILE_TOO_LARGE", "File size exceeds 5MB")
		return
	}

	// 保存临时文件
	tempDir := os.TempDir()
	tempFile := filepath.Join(tempDir, fmt.Sprintf("sailing_import_%d_%s", time.Now().Unix(), file.Filename))
	if err := c.SaveUploadedFile(file, tempFile); err != nil {
		RespondError(c, http.StatusInternalServerError, "ERR_SAVE_FILE", "Failed to save uploaded file")
		return
	}
	defer os.Remove(tempFile)

	// 导入模板
	result, err := h.templateService.ImportSailingTemplate(c.Request.Context(), tempFile, userCtx.UserID)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "ERR_IMPORT_FAILED", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": result,
	})
}

// UploadCabinTypeTemplate 上传并导入房型模板
// POST /api/v1/template/cabin-type/import
func (h *TemplateHandler) UploadCabinTypeTemplate(c *gin.Context) {
	userCtx := auth.GetUserContext(c)
	if userCtx == nil {
		RespondError(c, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "User not authenticated")
		return
	}

	// 只有管理员可以导入模板
	if userCtx.Role != domain.UserRoleAdmin {
		RespondError(c, http.StatusForbidden, "ERR_FORBIDDEN", "Only admins can import templates")
		return
	}

	// 解析上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		RespondError(c, http.StatusBadRequest, "ERR_INVALID_FILE", "File is required")
		return
	}

	// 验证文件类型
	if filepath.Ext(file.Filename) != ".xlsx" {
		RespondError(c, http.StatusBadRequest, "ERR_INVALID_FILE_TYPE", "Only .xlsx files are supported")
		return
	}

	// 验证文件大小（最大 5MB）
	const maxFileSize = 5 * 1024 * 1024
	if file.Size > maxFileSize {
		RespondError(c, http.StatusBadRequest, "ERR_FILE_TOO_LARGE", "File size exceeds 5MB")
		return
	}

	// 保存临时文件
	tempDir := os.TempDir()
	tempFile := filepath.Join(tempDir, fmt.Sprintf("cabin_type_import_%d_%s", time.Now().Unix(), file.Filename))
	if err := c.SaveUploadedFile(file, tempFile); err != nil {
		RespondError(c, http.StatusInternalServerError, "ERR_SAVE_FILE", "Failed to save uploaded file")
		return
	}
	defer os.Remove(tempFile)

	// 导入模板
	result, err := h.templateService.ImportCabinTypeTemplate(c.Request.Context(), tempFile, userCtx.UserID)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "ERR_IMPORT_FAILED", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": result,
	})
}
