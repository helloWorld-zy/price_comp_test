package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cruise-price-compare/internal/llm"
	"cruise-price-compare/internal/obs"
	"cruise-price-compare/internal/repo"
	"cruise-price-compare/internal/service"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// Load configuration from environment
	dbDSN := os.Getenv("DB_DSN")
	if dbDSN == "" {
		dbDSN = "root:password@tcp(localhost:3306)/cruise_price_compare?parseTime=true"
	}

	ollamaURL := os.Getenv("OLLAMA_URL")
	if ollamaURL == "" {
		ollamaURL = "http://localhost:11434"
	}

	ollamaModel := os.Getenv("OLLAMA_MODEL")
	if ollamaModel == "" {
		ollamaModel = "llama2"
	}

	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "./uploads"
	}

	pollInterval := 5 * time.Second
	maxConcurrent := 1 // Process one job at a time

	// Initialize database
	db, err := repo.NewDB(repo.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     3306,
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Database: os.Getenv("DB_NAME"),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize logger
	logger := obs.NewLogger(obs.LogConfig{
		Level:  obs.LogLevelInfo,
		Format: "json",
	})

	// Initialize repositories
	jobRepo := repo.NewImportJobRepository(db)
	quoteRepo := repo.NewPriceQuoteRepository(db)
	sailingRepo := repo.NewSailingRepository(db)
	cabinTypeRepo := repo.NewCabinTypeRepository(db)
	shipRepo := repo.NewShipRepository(db)
	cruiseLineRepo := repo.NewCruiseLineRepository(db)
	supplierRepo := repo.NewSupplierRepository(db)
	auditRepo := repo.NewAuditLogRepository(db)

	// Initialize services
	fileStorage := service.NewFileStorageService(uploadDir)
	ollamaClient := llm.NewOllamaClient(ollamaURL, ollamaModel)
	auditService := obs.NewAuditService(auditRepo, logger)

	dataMatcher := service.NewDataMatcher(
		shipRepo,
		sailingRepo,
		cabinTypeRepo,
		cruiseLineRepo,
	)

	quoteService := service.NewQuoteService(
		quoteRepo,
		sailingRepo,
		cabinTypeRepo,
		supplierRepo,
		auditService,
	)

	importJobService := service.NewImportJobService(
		jobRepo,
		fileStorage,
		ollamaClient,
		dataMatcher,
		quoteService,
		auditService,
	)

	// Create worker
	worker := NewWorker(importJobService, logger, pollInterval, maxConcurrent)

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Received shutdown signal, gracefully stopping worker...")
		cancel()
	}()

	// Start worker
	logger.Info("Starting import job worker...")
	logger.Info(fmt.Sprintf("Poll interval: %v, Max concurrent: %d", pollInterval, maxConcurrent))

	if err := worker.Run(ctx); err != nil {
		logger.WithError(err).Error("Worker stopped with error")
		os.Exit(1)
	}

	logger.Info("Worker stopped gracefully")
}

// Worker processes import jobs
type Worker struct {
	service       *service.ImportJobService
	logger        *obs.Logger
	pollInterval  time.Duration
	maxConcurrent int
	jobChan       chan uint64
}

// NewWorker creates a new worker
func NewWorker(service *service.ImportJobService, logger *obs.Logger, pollInterval time.Duration, maxConcurrent int) *Worker {
	return &Worker{
		service:       service,
		logger:        logger,
		pollInterval:  pollInterval,
		maxConcurrent: maxConcurrent,
		jobChan:       make(chan uint64, maxConcurrent),
	}
}

// Run starts the worker loop
func (w *Worker) Run(ctx context.Context) error {
	// Start job processors
	for i := 0; i < w.maxConcurrent; i++ {
		go w.processJobs(ctx, i+1)
	}

	// Poll for pending jobs
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Worker context cancelled, stopping...")
			close(w.jobChan)
			return nil

		case <-ticker.C:
			// Check for pending jobs
			job, err := w.service.GetNextPendingJob(ctx)
			if err != nil {
				w.logger.WithError(err).Error("Failed to get next pending job")
				continue
			}

			if job != nil {
				w.logger.WithField("job_id", job.ID).Info("Found pending job")

				// Try to send to job channel (non-blocking)
				select {
				case w.jobChan <- job.ID:
					w.logger.WithField("job_id", job.ID).Info("Job queued for processing")
				default:
					w.logger.WithField("job_id", job.ID).Warn("Job channel full, will retry later")
				}
			}
		}
	}
}

// processJobs processes jobs from the channel
func (w *Worker) processJobs(ctx context.Context, workerID int) {
	logger := w.logger.WithField("worker_id", workerID)
	logger.Info("Job processor started")

	for {
		select {
		case <-ctx.Done():
			logger.Info("Job processor stopping...")
			return

		case jobID, ok := <-w.jobChan:
			if !ok {
				logger.Info("Job channel closed, processor stopping...")
				return
			}

			logger.WithField("job_id", jobID).Info("Processing job")
			startTime := time.Now()

			// Process the job
			err := w.service.ProcessImportJob(ctx, jobID)

			duration := time.Since(startTime)

			if err != nil {
				logger.WithField("job_id", jobID).
					WithField("duration_ms", duration.Milliseconds()).
					WithError(err).
					Error("Job processing failed")
			} else {
				logger.WithField("job_id", jobID).
					WithField("duration_ms", duration.Milliseconds()).
					Info("Job processing completed successfully")
			}
		}
	}
}
