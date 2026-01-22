package http

import (
	"net/http"
	"strconv"

	"cruise-price-compare/internal/auth"
	"cruise-price-compare/internal/domain"
	"cruise-price-compare/internal/repo"
	"cruise-price-compare/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ImportHandler handles import-related HTTP requests
type ImportHandler struct {
	importService *service.ImportJobService
}

// NewImportHandler creates a new import handler
func NewImportHandler(importService *service.ImportJobService) *ImportHandler {
	return &ImportHandler{
		importService: importService,
	}
}

// UploadFileRequest represents the file upload request
type UploadFileRequest struct {
	File *gin.Context `form:"file" binding:"required"`
}

// UploadFile handles file upload for quote import
// POST /api/v1/import/upload
func (h *ImportHandler) UploadFile(c *gin.Context) {
	userCtx := auth.GetUserContext(c)
	if userCtx == nil {
		RespondError(c, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "User not authenticated")
		return
	}

	// Only vendors can upload files
	if userCtx.Role != domain.UserRoleVendor {
		RespondError(c, http.StatusForbidden, "ERR_FORBIDDEN", "Only vendors can upload files")
		return
	}

	// Parse multipart form
	file, err := c.FormFile("file")
	if err != nil {
		RespondError(c, http.StatusBadRequest, "ERR_INVALID_FILE", "File is required")
		return
	}

	// Validate file size (max 10MB)
	const maxFileSize = 10 * 1024 * 1024
	if file.Size > maxFileSize {
		RespondError(c, http.StatusBadRequest, "ERR_FILE_TOO_LARGE", "File size exceeds 10MB")
		return
	}

	// Validate file type
	ext := file.Filename[len(file.Filename)-5:]
	if ext != ".pdf" && ext != ".docx" && ext[len(ext)-4:] != ".doc" {
		RespondError(c, http.StatusBadRequest, "ERR_INVALID_FILE_TYPE", "Only PDF and Word documents are supported")
		return
	}

	// Read file content
	fileContent, err := file.Open()
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "ERR_FILE_READ", "Failed to read file")
		return
	}
	defer fileContent.Close()

	// Read into memory
	fileBytes := make([]byte, file.Size)
	_, err = fileContent.Read(fileBytes)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "ERR_FILE_READ", "Failed to read file content")
		return
	}

	// Create idempotency key
	idempotencyKey := uuid.New().String()

	// Create import job
	job, err := h.importService.CreateImportJob(c.Request.Context(), service.CreateImportJobInput{
		FileName:       file.Filename,
		FileContent:    fileBytes,
		UserID:         userCtx.UserID,
		SupplierID:     userCtx.SupplierID,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "ERR_CREATE_JOB", err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": job,
	})
}

// ListJobs lists import jobs
// GET /api/v1/import/jobs
func (h *ImportHandler) ListJobs(c *gin.Context) {
	userCtx := auth.GetUserContext(c)
	if userCtx == nil {
		RespondError(c, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "User not authenticated")
		return
	}

	// Parse pagination
	page, pageSize := ParsePagination(c)
	pagination := repo.Pagination{Page: page, PageSize: pageSize}

	// Parse filters
	var status *domain.ImportJobStatus
	if statusStr := c.Query("status"); statusStr != "" {
		s := domain.ImportJobStatus(statusStr)
		status = &s
	}

	var jobType *domain.ImportJobType
	if typeStr := c.Query("type"); typeStr != "" {
		t := domain.ImportJobType(typeStr)
		jobType = &t
	}

	// List jobs
	result, err := h.importService.ListJobs(
		c.Request.Context(),
		pagination,
		nil,
		status,
		jobType,
		userCtx.Role,
		userCtx.UserID,
	)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "ERR_LIST_JOBS", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": result.Items,
		"pagination": gin.H{
			"page":       result.Page,
			"page_size":  result.PageSize,
			"total":      result.Total,
			"total_page": result.TotalPage,
		},
	})
}

// GetJob retrieves a single import job
// GET /api/v1/import/jobs/:id
func (h *ImportHandler) GetJob(c *gin.Context) {
	userCtx := auth.GetUserContext(c)
	if userCtx == nil {
		RespondError(c, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "User not authenticated")
		return
	}

	// Parse job ID
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "ERR_INVALID_ID", "Invalid job ID")
		return
	}

	// Get job
	job, err := h.importService.GetJob(c.Request.Context(), id, userCtx.UserID, userCtx.Role, userCtx.SupplierID)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "ERR_GET_JOB", err.Error())
		return
	}

	if job == nil {
		RespondError(c, http.StatusNotFound, "ERR_NOT_FOUND", "Job not found")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": job,
	})
}

// RetryJob retries a failed import job
// POST /api/v1/import/jobs/:id/retry
func (h *ImportHandler) RetryJob(c *gin.Context) {
	userCtx := auth.GetUserContext(c)
	if userCtx == nil {
		RespondError(c, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "User not authenticated")
		return
	}

	// Parse job ID
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		RespondError(c, http.StatusBadRequest, "ERR_INVALID_ID", "Invalid job ID")
		return
	}

	// Get job first to verify ownership
	job, err := h.importService.GetJob(c.Request.Context(), id, userCtx.UserID, userCtx.Role, userCtx.SupplierID)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "ERR_GET_JOB", err.Error())
		return
	}

	if job == nil {
		RespondError(c, http.StatusNotFound, "ERR_NOT_FOUND", "Job not found")
		return
	}

	// Only retry failed jobs
	if job.Status != domain.ImportJobStatusFailed {
		RespondError(c, http.StatusBadRequest, "ERR_INVALID_STATUS", "Only failed jobs can be retried")
		return
	}

	// Process the job immediately (or reset to pending)
	err = h.importService.ProcessImportJob(c.Request.Context(), id)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "ERR_RETRY_FAILED", err.Error())
		return
	}

	// Reload job
	job, _ = h.importService.GetJob(c.Request.Context(), id, userCtx.UserID, userCtx.Role, userCtx.SupplierID)

	c.JSON(http.StatusOK, gin.H{
		"data": job,
	})
}
