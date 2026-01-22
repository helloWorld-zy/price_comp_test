package service

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"cruise-price-compare/internal/domain"
	"cruise-price-compare/internal/llm"
	"cruise-price-compare/internal/llm/prompts"
	"cruise-price-compare/internal/obs"
	"cruise-price-compare/internal/repo"
)

// ImportJobService handles import job operations
type ImportJobService struct {
	jobRepo        *repo.ImportJobRepository
	fileStorage    *FileStorageService
	pdfExtractor   *llm.PDFExtractor
	wordExtractor  *llm.WordExtractor
	ollamaClient   *llm.OllamaClient
	responseParser *llm.ResponseParser
	dataMatcher    *DataMatcher
	quoteService   *QuoteService
	auditService   *obs.AuditService
}

// NewImportJobService creates a new import job service
func NewImportJobService(
	jobRepo *repo.ImportJobRepository,
	fileStorage *FileStorageService,
	ollamaClient *llm.OllamaClient,
	dataMatcher *DataMatcher,
	quoteService *QuoteService,
	auditService *obs.AuditService,
) *ImportJobService {
	return &ImportJobService{
		jobRepo:        jobRepo,
		fileStorage:    fileStorage,
		pdfExtractor:   llm.NewPDFExtractor(),
		wordExtractor:  llm.NewWordExtractor(),
		ollamaClient:   ollamaClient,
		responseParser: llm.NewResponseParser(),
		dataMatcher:    dataMatcher,
		quoteService:   quoteService,
		auditService:   auditService,
	}
}

// CreateImportJobInput represents input for creating an import job
type CreateImportJobInput struct {
	FileName       string
	FileContent    []byte
	UserID         uint64
	SupplierID     uint64
	IdempotencyKey string // Optional, for duplicate detection
}

// CreateImportJob creates a new import job from uploaded file
func (s *ImportJobService) CreateImportJob(ctx context.Context, input CreateImportJobInput) (*domain.ImportJob, error) {
	// Check for duplicate if idempotency key provided
	if input.IdempotencyKey != "" {
		existing, err := s.jobRepo.GetByIdempotencyKey(ctx, input.IdempotencyKey)
		if err != nil {
			return nil, fmt.Errorf("failed to check for duplicate: %w", err)
		}
		if existing != nil {
			return existing, nil // Return existing job
		}
	}

	// Store the file
	filePath, fileHash, fileSize, err := s.fileStorage.UploadFile(ctx, input.FileName, bytes.NewReader(input.FileContent))
	if err != nil {
		return nil, fmt.Errorf("failed to store file: %w", err)
	}

	// Determine job type from file extension
	ext := strings.ToLower(filepath.Ext(input.FileName))
	var jobType domain.ImportJobType
	switch ext {
	case ".pdf", ".docx", ".doc":
		jobType = domain.ImportJobTypeFileUpload
	default:
		return nil, fmt.Errorf("unsupported file type: %s", ext)
	}

	// Create the import job
	job := &domain.ImportJob{
		Type:           jobType,
		Status:         domain.ImportJobStatusPending,
		FileName:       input.FileName,
		FileHash:       fileHash,
		FileSize:       fileSize,
		FilePath:       filePath,
		IdempotencyKey: input.IdempotencyKey,
		CreatedBy:      input.UserID,
	}

	if err := s.jobRepo.Create(ctx, job); err != nil {
		return nil, fmt.Errorf("failed to create import job: %w", err)
	}

	// Audit log
	if s.auditService != nil {
		_ = s.auditService.LogCreate(ctx, input.UserID, &input.SupplierID, "import_job", job.ID, job)
	}

	return job, nil
}

// ProcessImportJob processes a single import job
// This is called by the worker
func (s *ImportJobService) ProcessImportJob(ctx context.Context, jobID uint64) error {
	// Get the job
	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}
	if job == nil {
		return fmt.Errorf("job not found: %d", jobID)
	}

	// Mark as running
	if err := s.jobRepo.UpdateStarted(ctx, jobID); err != nil {
		return fmt.Errorf("failed to mark job as started: %w", err)
	}

	// Process based on type
	var processErr error
	var summary *domain.ImportResultSummary

	// Determine file type from extension
	ext := strings.ToLower(filepath.Ext(job.FileName))
	if ext == ".pdf" {
		summary, processErr = s.processPDFJob(ctx, job)
	} else if ext == ".docx" || ext == ".doc" {
		summary, processErr = s.processWordJob(ctx, job)
	} else {
		processErr = fmt.Errorf("unsupported file type: %s", ext)
	}

	// Update job status
	var status domain.ImportJobStatus
	var errorMsg string

	if processErr != nil {
		status = domain.ImportJobStatusFailed
		errorMsg = processErr.Error()
	} else {
		status = domain.ImportJobStatusSucceeded
	}

	if err := s.jobRepo.UpdateCompleted(ctx, jobID, status, summary, errorMsg); err != nil {
		return fmt.Errorf("failed to update job completion: %w", err)
	}

	return processErr
}

// processPDFJob processes a PDF import job
func (s *ImportJobService) processPDFJob(ctx context.Context, job *domain.ImportJob) (*domain.ImportResultSummary, error) {
	// Step 1: Extract text from PDF
	text, err := s.pdfExtractor.ExtractText(job.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to extract PDF text: %w", err)
	}

	// Step 2: Send to LLM for parsing
	prompt := prompts.QuoteParsePrompt(text)
	llmResponse, err := s.ollamaClient.Generate(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate LLM response: %w", err)
	}

	// Step 3: Parse LLM response
	parseResult, err := s.responseParser.ParseQuoteResponse(llmResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	// Step 4: Match sailing and cabin types
	summary, err := s.matchAndCreateQuotes(ctx, job, parseResult)
	if err != nil {
		return nil, fmt.Errorf("failed to create quotes: %w", err)
	}

	return summary, nil
}

// processWordJob processes a Word document import job
func (s *ImportJobService) processWordJob(ctx context.Context, job *domain.ImportJob) (*domain.ImportResultSummary, error) {
	// Step 1: Extract text from Word document
	text, err := s.wordExtractor.ExtractText(job.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to extract Word text: %w", err)
	}

	// Step 2: Send to LLM for parsing
	prompt := prompts.QuoteParsePrompt(text)
	llmResponse, err := s.ollamaClient.Generate(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate LLM response: %w", err)
	}

	// Step 3: Parse LLM response
	parseResult, err := s.responseParser.ParseQuoteResponse(llmResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	// Step 4: Match sailing and cabin types
	summary, err := s.matchAndCreateQuotes(ctx, job, parseResult)
	if err != nil {
		return nil, fmt.Errorf("failed to create quotes: %w", err)
	}

	return summary, nil
}

// matchAndCreateQuotes matches parsed data and creates quotes
func (s *ImportJobService) matchAndCreateQuotes(ctx context.Context, job *domain.ImportJob, parseResult *llm.QuoteParseResult) (*domain.ImportResultSummary, error) {
	summary := &domain.ImportResultSummary{
		TotalRows:     len(parseResult.Quotes),
		CreatedQuotes: 0,
		SkippedRows:   0,
		Warnings:      []string{},
	}

	// Parse departure date
	departureDate, err := time.Parse("2006-01-02", parseResult.DepartureDate)
	if err != nil {
		summary.Warnings = append(summary.Warnings, fmt.Sprintf("Invalid departure date: %s", parseResult.DepartureDate))
		return summary, fmt.Errorf("invalid departure date: %w", err)
	}

	// Match sailing
	matchResult, err := s.dataMatcher.MatchSailingData(ctx, parseResult.SailingCode, parseResult.ShipName, departureDate, parseResult.Nights)
	if err != nil {
		summary.Warnings = append(summary.Warnings, fmt.Sprintf("Sailing match error: %v", err))
		return summary, fmt.Errorf("sailing match failed: %w", err)
	}

	if matchResult.Sailing == nil {
		summary.Warnings = append(summary.Warnings, "Sailing not found in database")
		return summary, fmt.Errorf("sailing not found")
	}

	// Process each quote
	for _, parsedQuote := range parseResult.Quotes {
		// Match cabin type
		cabinType, confidence, err := s.dataMatcher.MatchCabinType(ctx, matchResult.Sailing.ShipID, parsedQuote.CabinTypeName, parsedQuote.CabinCategory)
		if err != nil || confidence < 0.6 {
			summary.SkippedRows++
			summary.Warnings = append(summary.Warnings, fmt.Sprintf("Cabin type '%s' not matched (confidence: %.2f)", parsedQuote.CabinTypeName, confidence))
			continue
		}

		// Create quote
		quoteInput := CreateQuoteInput{
			SailingID:   matchResult.Sailing.ShipID,
			CabinTypeID: cabinType.ID,
			Price:       fmt.Sprintf("%.2f", parsedQuote.Price),
			Currency:    parsedQuote.Currency,
			PricingUnit: s.responseParser.ConvertPricingUnit(parsedQuote.PricingUnit),
			Conditions:  parsedQuote.Conditions,
			Promotion:   parsedQuote.Promotion,
			Notes:       parsedQuote.Notes,
			UserID:      job.CreatedBy,
		}

		_, err = s.quoteService.CreateQuote(ctx, quoteInput)
		if err != nil {
			summary.Warnings = append(summary.Warnings, fmt.Sprintf("Failed to create quote for cabin '%s': %v", parsedQuote.CabinTypeName, err))
			summary.SkippedRows++
		} else {
			summary.SuccessRows++
			summary.CreatedQuotes++
		}
	}

	return summary, nil
}

// GetJob retrieves an import job by ID
func (s *ImportJobService) GetJob(ctx context.Context, id uint64, userID uint64, userRole domain.UserRole, supplierID uint64) (*domain.ImportJob, error) {
	job, err := s.jobRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}
	if job == nil {
		return nil, nil
	}

	// Check permissions
	if userRole == domain.UserRoleVendor && job.CreatedBy != userID {
		return nil, fmt.Errorf("permission denied")
	}

	return job, nil
}

// ListJobs lists import jobs with pagination
func (s *ImportJobService) ListJobs(ctx context.Context, pagination repo.Pagination, filterUserID *uint64, status *domain.ImportJobStatus, jobType *domain.ImportJobType, userRole domain.UserRole, actualUserID uint64) (repo.PaginatedResult[domain.ImportJob], error) {
	// If vendor role, force filter by their user ID
	userIDToUse := filterUserID
	if userRole == domain.UserRoleVendor {
		userIDToUse = &actualUserID
	}

	return s.jobRepo.List(ctx, pagination, userIDToUse, status, jobType)
}

// GetNextPendingJob gets the next pending job for processing
func (s *ImportJobService) GetNextPendingJob(ctx context.Context) (*domain.ImportJob, error) {
	jobs, err := s.jobRepo.ListPending(ctx, 1)
	if err != nil {
		return nil, err
	}
	if len(jobs) == 0 {
		return nil, nil
	}
	return &jobs[0], nil
}
