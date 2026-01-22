package domain

import (
	"encoding/json"
	"time"
)

// ImportJobType represents the type of import job
type ImportJobType string

const (
	ImportJobTypeFileUpload       ImportJobType = "FILE_UPLOAD"
	ImportJobTypeTextInput        ImportJobType = "TEXT_INPUT"
	ImportJobTypeTemplateImport   ImportJobType = "TEMPLATE_IMPORT"
	ImportJobTypeAdminLLMGenerate ImportJobType = "ADMIN_LLM_GENERATE"
)

// ImportJobStatus represents the status of an import job
type ImportJobStatus string

const (
	ImportJobStatusPending           ImportJobStatus = "PENDING"
	ImportJobStatusRunning           ImportJobStatus = "RUNNING"
	ImportJobStatusNeedsConfirmation ImportJobStatus = "NEEDS_CONFIRMATION"
	ImportJobStatusSucceeded         ImportJobStatus = "SUCCEEDED"
	ImportJobStatusFailed            ImportJobStatus = "FAILED"
)

// ImportResultSummary represents the summary of import results
type ImportResultSummary struct {
	TotalRows     int      `json:"total_rows"`
	SuccessRows   int      `json:"success_rows"`
	FailedRows    int      `json:"failed_rows"`
	SkippedRows   int      `json:"skipped_rows"`
	Warnings      []string `json:"warnings,omitempty"`
	CreatedQuotes int      `json:"created_quotes,omitempty"`
}

// ImportJob represents an import task
type ImportJob struct {
	ID             uint64               `json:"id" db:"id"`
	Type           ImportJobType        `json:"type" db:"type"`
	Status         ImportJobStatus      `json:"status" db:"status"`
	FileName       string               `json:"file_name,omitempty" db:"file_name"`
	FileHash       string               `json:"file_hash,omitempty" db:"file_hash"`
	FileSize       int64                `json:"file_size,omitempty" db:"file_size"`
	FilePath       string               `json:"file_path,omitempty" db:"file_path"`
	RawText        string               `json:"raw_text,omitempty" db:"raw_text"`
	IdempotencyKey string               `json:"idempotency_key,omitempty" db:"idempotency_key"`
	ModelVersion   string               `json:"model_version,omitempty" db:"model_version"`
	PromptVersion  string               `json:"prompt_version,omitempty" db:"prompt_version"`
	ResultSummary  *ImportResultSummary `json:"result_summary,omitempty" db:"-"`
	ResultJSON     json.RawMessage      `json:"-" db:"result_summary"`
	ErrorMessage   string               `json:"error_message,omitempty" db:"error_message"`
	StartedAt      *time.Time           `json:"started_at,omitempty" db:"started_at"`
	CompletedAt    *time.Time           `json:"completed_at,omitempty" db:"completed_at"`
	DurationMs     *int64               `json:"duration_ms,omitempty" db:"duration_ms"`
	CreatedAt      time.Time            `json:"created_at" db:"created_at"`
	CreatedBy      uint64               `json:"created_by" db:"created_by"`

	// Loaded relations
	ParseJobs   []ParseJob   `json:"parse_jobs,omitempty" db:"-"`
	PriceQuotes []PriceQuote `json:"price_quotes,omitempty" db:"-"`
}

// IsPending checks if job is pending
func (ij *ImportJob) IsPending() bool {
	return ij.Status == ImportJobStatusPending
}

// IsRunning checks if job is running
func (ij *ImportJob) IsRunning() bool {
	return ij.Status == ImportJobStatusRunning
}

// IsCompleted checks if job is completed (success or failure)
func (ij *ImportJob) IsCompleted() bool {
	return ij.Status == ImportJobStatusSucceeded || ij.Status == ImportJobStatusFailed
}

// NeedsConfirmation checks if job needs user confirmation
func (ij *ImportJob) NeedsConfirmation() bool {
	return ij.Status == ImportJobStatusNeedsConfirmation
}

// ParseJobStatus represents the status of a parse job
type ParseJobStatus string

const (
	ParseJobStatusPending   ParseJobStatus = "PENDING"
	ParseJobStatusRunning   ParseJobStatus = "RUNNING"
	ParseJobStatusSucceeded ParseJobStatus = "SUCCEEDED"
	ParseJobStatusFailed    ParseJobStatus = "FAILED"
)

// ParsedDataItem represents a single parsed quote item from LLM
type ParsedDataItem struct {
	SailingCode   string  `json:"sailing_code,omitempty"`
	ShipName      string  `json:"ship_name,omitempty"`
	CruiseLine    string  `json:"cruise_line,omitempty"`
	DepartureDate string  `json:"departure_date,omitempty"`
	ReturnDate    string  `json:"return_date,omitempty"`
	Route         string  `json:"route,omitempty"`
	CabinType     string  `json:"cabin_type,omitempty"`
	CabinCategory string  `json:"cabin_category,omitempty"`
	Price         float64 `json:"price,omitempty"`
	Currency      string  `json:"currency,omitempty"`
	PricingUnit   string  `json:"pricing_unit,omitempty"`
	GuestCount    int     `json:"guest_count,omitempty"`
	Conditions    string  `json:"conditions,omitempty"`
	Promotion     string  `json:"promotion,omitempty"`
	ValidUntil    string  `json:"valid_until,omitempty"`
	Notes         string  `json:"notes,omitempty"`
}

// ParseJob represents an LLM parsing task
type ParseJob struct {
	ID           uint64           `json:"id" db:"id"`
	ImportJobID  uint64           `json:"import_job_id" db:"import_job_id"`
	Status       ParseJobStatus   `json:"status" db:"status"`
	ParsedData   []ParsedDataItem `json:"parsed_data,omitempty" db:"-"`
	ParsedJSON   json.RawMessage  `json:"-" db:"parsed_data"`
	Confidence   *float64         `json:"confidence,omitempty" db:"confidence"`
	Warnings     []string         `json:"warnings,omitempty" db:"-"`
	WarningsJSON json.RawMessage  `json:"-" db:"warnings"`
	PageInfo     json.RawMessage  `json:"page_info,omitempty" db:"page_info"`
	ErrorMessage string           `json:"error_message,omitempty" db:"error_message"`
	StartedAt    *time.Time       `json:"started_at,omitempty" db:"started_at"`
	CompletedAt  *time.Time       `json:"completed_at,omitempty" db:"completed_at"`
	CreatedAt    time.Time        `json:"created_at" db:"created_at"`

	// Loaded relations
	ImportJob *ImportJob `json:"import_job,omitempty" db:"-"`
}

// IsSucceeded checks if parse job succeeded
func (pj *ParseJob) IsSucceeded() bool {
	return pj.Status == ParseJobStatusSucceeded
}

// IsFailed checks if parse job failed
func (pj *ParseJob) IsFailed() bool {
	return pj.Status == ParseJobStatusFailed
}
