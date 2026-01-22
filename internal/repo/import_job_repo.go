package repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"cruise-price-compare/internal/domain"
)

// ImportJobRepository handles import job data access
type ImportJobRepository struct {
	db *DB
}

// NewImportJobRepository creates a new import job repository
func NewImportJobRepository(db *DB) *ImportJobRepository {
	return &ImportJobRepository{db: db}
}

// GetByID retrieves an import job by ID
func (r *ImportJobRepository) GetByID(ctx context.Context, id uint64) (*domain.ImportJob, error) {
	var row importJobRow
	query := `SELECT id, type, status, file_name, file_hash, file_size, file_path, raw_text, 
              idempotency_key, model_version, prompt_version, result_summary, error_message, 
              started_at, completed_at, duration_ms, created_at, created_by 
              FROM import_job WHERE id = ?`

	if err := r.db.GetContext(ctx, &row, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get import job by id: %w", err)
	}

	return row.toDomain(), nil
}

// GetByIdempotencyKey retrieves an import job by idempotency key
func (r *ImportJobRepository) GetByIdempotencyKey(ctx context.Context, key string) (*domain.ImportJob, error) {
	var row importJobRow
	query := `SELECT id, type, status, file_name, file_hash, file_size, file_path, raw_text, 
              idempotency_key, model_version, prompt_version, result_summary, error_message, 
              started_at, completed_at, duration_ms, created_at, created_by 
              FROM import_job WHERE idempotency_key = ?`

	if err := r.db.GetContext(ctx, &row, query, key); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get import job by key: %w", err)
	}

	return row.toDomain(), nil
}

// List retrieves import jobs with pagination
func (r *ImportJobRepository) List(ctx context.Context, pagination Pagination, userID *uint64, status *domain.ImportJobStatus, jobType *domain.ImportJobType) (PaginatedResult[domain.ImportJob], error) {
	var rows []importJobRow
	var total int64

	countQuery := "SELECT COUNT(*) FROM import_job WHERE 1=1"
	selectQuery := `SELECT id, type, status, file_name, file_hash, file_size, file_path, raw_text, 
                    idempotency_key, model_version, prompt_version, result_summary, error_message, 
                    started_at, completed_at, duration_ms, created_at, created_by FROM import_job WHERE 1=1`
	var args []interface{}

	if userID != nil {
		countQuery += " AND created_by = ?"
		selectQuery += " AND created_by = ?"
		args = append(args, *userID)
	}

	if status != nil {
		countQuery += " AND status = ?"
		selectQuery += " AND status = ?"
		args = append(args, *status)
	}

	if jobType != nil {
		countQuery += " AND type = ?"
		selectQuery += " AND type = ?"
		args = append(args, *jobType)
	}

	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return PaginatedResult[domain.ImportJob]{}, fmt.Errorf("failed to count import jobs: %w", err)
	}

	selectQuery += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, pagination.Limit(), pagination.Offset())

	if err := r.db.SelectContext(ctx, &rows, selectQuery, args...); err != nil {
		return PaginatedResult[domain.ImportJob]{}, fmt.Errorf("failed to list import jobs: %w", err)
	}

	items := make([]domain.ImportJob, len(rows))
	for i, row := range rows {
		items[i] = *row.toDomain()
	}

	return NewPaginatedResult(items, total, pagination), nil
}

// Create creates a new import job
func (r *ImportJobRepository) Create(ctx context.Context, job *domain.ImportJob) error {
	var resultJSON []byte
	if job.ResultSummary != nil {
		var err error
		resultJSON, err = json.Marshal(job.ResultSummary)
		if err != nil {
			return fmt.Errorf("failed to marshal result summary: %w", err)
		}
	}

	query := `INSERT INTO import_job (type, status, file_name, file_hash, file_size, file_path, 
              raw_text, idempotency_key, model_version, prompt_version, result_summary, 
              error_message, started_at, completed_at, created_by) 
              VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db.ExecContext(ctx, query, job.Type, job.Status, job.FileName, job.FileHash,
		job.FileSize, job.FilePath, job.RawText, job.IdempotencyKey, job.ModelVersion,
		job.PromptVersion, resultJSON, job.ErrorMessage, job.StartedAt, job.CompletedAt, job.CreatedBy)
	if err != nil {
		return fmt.Errorf("failed to create import job: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	job.ID = uint64(id)

	return nil
}

// UpdateStatus updates an import job status
func (r *ImportJobRepository) UpdateStatus(ctx context.Context, id uint64, status domain.ImportJobStatus, errorMsg string) error {
	query := `UPDATE import_job SET status = ?, error_message = ? WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, status, errorMsg, id)
	if err != nil {
		return fmt.Errorf("failed to update import job status: %w", err)
	}

	return nil
}

// UpdateStarted marks job as started
func (r *ImportJobRepository) UpdateStarted(ctx context.Context, id uint64) error {
	now := time.Now()
	query := `UPDATE import_job SET status = 'RUNNING', started_at = ? WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, now, id)
	if err != nil {
		return fmt.Errorf("failed to update import job started: %w", err)
	}

	return nil
}

// UpdateCompleted marks job as completed
func (r *ImportJobRepository) UpdateCompleted(ctx context.Context, id uint64, status domain.ImportJobStatus, summary *domain.ImportResultSummary, errorMsg string) error {
	now := time.Now()
	var resultJSON []byte
	if summary != nil {
		var err error
		resultJSON, err = json.Marshal(summary)
		if err != nil {
			return fmt.Errorf("failed to marshal result summary: %w", err)
		}
	}

	query := `UPDATE import_job SET status = ?, result_summary = ?, error_message = ?, completed_at = ? WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, status, resultJSON, errorMsg, now, id)
	if err != nil {
		return fmt.Errorf("failed to update import job completed: %w", err)
	}

	return nil
}

// ListPending retrieves pending import jobs
func (r *ImportJobRepository) ListPending(ctx context.Context, limit int) ([]domain.ImportJob, error) {
	var rows []importJobRow
	query := `SELECT id, type, status, file_name, file_hash, file_size, file_path, raw_text, 
              idempotency_key, model_version, prompt_version, result_summary, error_message, 
              started_at, completed_at, duration_ms, created_at, created_by 
              FROM import_job WHERE status = 'PENDING' ORDER BY created_at LIMIT ?`

	if err := r.db.SelectContext(ctx, &rows, query, limit); err != nil {
		return nil, fmt.Errorf("failed to list pending import jobs: %w", err)
	}

	items := make([]domain.ImportJob, len(rows))
	for i, row := range rows {
		items[i] = *row.toDomain()
	}

	return items, nil
}

type importJobRow struct {
	ID             uint64         `db:"id"`
	Type           string         `db:"type"`
	Status         string         `db:"status"`
	FileName       sql.NullString `db:"file_name"`
	FileHash       sql.NullString `db:"file_hash"`
	FileSize       sql.NullInt64  `db:"file_size"`
	FilePath       sql.NullString `db:"file_path"`
	RawText        sql.NullString `db:"raw_text"`
	IdempotencyKey sql.NullString `db:"idempotency_key"`
	ModelVersion   sql.NullString `db:"model_version"`
	PromptVersion  sql.NullString `db:"prompt_version"`
	ResultSummary  []byte         `db:"result_summary"`
	ErrorMessage   sql.NullString `db:"error_message"`
	StartedAt      sql.NullTime   `db:"started_at"`
	CompletedAt    sql.NullTime   `db:"completed_at"`
	DurationMs     sql.NullInt64  `db:"duration_ms"`
	CreatedAt      sql.NullTime   `db:"created_at"`
	CreatedBy      uint64         `db:"created_by"`
}

func (r *importJobRow) toDomain() *domain.ImportJob {
	job := &domain.ImportJob{
		ID:        r.ID,
		Type:      domain.ImportJobType(r.Type),
		Status:    domain.ImportJobStatus(r.Status),
		CreatedBy: r.CreatedBy,
	}

	if r.FileName.Valid {
		job.FileName = r.FileName.String
	}
	if r.FileHash.Valid {
		job.FileHash = r.FileHash.String
	}
	if r.FileSize.Valid {
		job.FileSize = r.FileSize.Int64
	}
	if r.FilePath.Valid {
		job.FilePath = r.FilePath.String
	}
	if r.RawText.Valid {
		job.RawText = r.RawText.String
	}
	if r.IdempotencyKey.Valid {
		job.IdempotencyKey = r.IdempotencyKey.String
	}
	if r.ModelVersion.Valid {
		job.ModelVersion = r.ModelVersion.String
	}
	if r.PromptVersion.Valid {
		job.PromptVersion = r.PromptVersion.String
	}
	if r.ResultSummary != nil {
		var summary domain.ImportResultSummary
		if json.Unmarshal(r.ResultSummary, &summary) == nil {
			job.ResultSummary = &summary
		}
	}
	if r.ErrorMessage.Valid {
		job.ErrorMessage = r.ErrorMessage.String
	}
	if r.StartedAt.Valid {
		job.StartedAt = &r.StartedAt.Time
	}
	if r.CompletedAt.Valid {
		job.CompletedAt = &r.CompletedAt.Time
	}
	if r.DurationMs.Valid {
		d := r.DurationMs.Int64
		job.DurationMs = &d
	}
	if r.CreatedAt.Valid {
		job.CreatedAt = r.CreatedAt.Time
	}

	return job
}
