// Package models contains generic and common data structures used across multiple model types.
// This module provides base structures and common validation patterns.
package models

import (
	"time"
)

// GenericPayload represents a flexible payload structure for general JSON validation.
type GenericPayload struct {
	ID        string                 `json:"id" validate:"omitempty"`
	Type      string                 `json:"type" validate:"required,min=1,max=100"`
	Version   string                 `json:"version" validate:"omitempty,semver"`
	Timestamp time.Time              `json:"timestamp" validate:"required"`
	Source    string                 `json:"source" validate:"required,url"`
	Data      map[string]interface{} `json:"data" validate:"required"`
	Metadata  map[string]string      `json:"metadata" validate:"omitempty"`
	Tags      []string               `json:"tags" validate:"omitempty,dive,min=1,max=50"`
	Priority  string                 `json:"priority" validate:"omitempty,priority_level"`
	Status    string                 `json:"status" validate:"omitempty,oneof=pending processing completed failed"`
	Checksum  string                 `json:"checksum" validate:"omitempty,len=64,hexadecimal"`
}

// ValidationResult represents the result of any validation operation.
type ValidationResult struct {
	IsValid            bool                   `json:"is_valid"`
	ModelType          string                 `json:"model_type"`
	Provider           string                 `json:"provider,omitempty"`
	ValidationProfile  string                 `json:"validation_profile,omitempty"`
	Errors             []ValidationError      `json:"errors,omitempty"`
	Warnings           []ValidationWarning    `json:"warnings,omitempty"`
	Timestamp          time.Time              `json:"timestamp"`
	ProcessingDuration time.Duration          `json:"processing_duration,omitempty"`
	PerformanceMetrics *PerformanceMetrics    `json:"performance_metrics,omitempty"`
	RequestID          string                 `json:"request_id,omitempty"`
	Context            map[string]interface{} `json:"context,omitempty"`
}

// ValidationError represents a validation error with detailed information.
type ValidationError struct {
	Field      string                 `json:"field"`
	Message    string                 `json:"message"`
	Code       string                 `json:"code"`
	Value      interface{}            `json:"value,omitempty"`
	Expected   interface{}            `json:"expected,omitempty"`
	Constraint string                 `json:"constraint,omitempty"`
	Path       string                 `json:"path,omitempty"`
	Context    map[string]interface{} `json:"context,omitempty"`
	Severity   string                 `json:"severity,omitempty"` // error, warning, info
}

// ValidationWarning represents a validation warning that doesn't prevent processing.
type ValidationWarning struct {
	Field      string                 `json:"field"`
	Message    string                 `json:"message"`
	Code       string                 `json:"code"`
	Value      interface{}            `json:"value,omitempty"`
	Suggestion string                 `json:"suggestion,omitempty"`
	Path       string                 `json:"path,omitempty"`
	Context    map[string]interface{} `json:"context,omitempty"`
	Category   string                 `json:"category,omitempty"` // security, performance, compliance, etc.
}

// PerformanceMetrics contains detailed performance information for validation operations.
type PerformanceMetrics struct {
	ValidationDuration    time.Duration `json:"validation_duration"`
	CacheHits             int           `json:"cache_hits,omitempty"`
	CacheMisses           int           `json:"cache_misses,omitempty"`
	MemoryUsage           int64         `json:"memory_usage,omitempty"`
	Provider              string        `json:"provider,omitempty"`
	RuleCount             int           `json:"rule_count,omitempty"`
	FieldCount            int           `json:"field_count,omitempty"`
	ParsingDuration       time.Duration `json:"parsing_duration,omitempty"`
	BusinessLogicDuration time.Duration `json:"business_logic_duration,omitempty"`
}

// APIModel represents a generic API endpoint validation structure.
type APIModel struct {
	Method     string                 `json:"method" validate:"required,oneof=GET POST PUT PATCH DELETE HEAD OPTIONS"`
	URL        string                 `json:"url" validate:"required,url"`
	Headers    map[string]string      `json:"headers" validate:"omitempty"`
	Parameters map[string]interface{} `json:"parameters" validate:"omitempty"`
	Body       interface{}            `json:"body" validate:"omitempty"`
	StatusCode int                    `json:"status_code" validate:"omitempty,gte=100,lte=599"`
	Response   interface{}            `json:"response" validate:"omitempty"`
	Timestamp  time.Time              `json:"timestamp" validate:"required"`
	Duration   time.Duration          `json:"duration" validate:"omitempty"`
	UserAgent  string                 `json:"user_agent" validate:"omitempty,max=500"`
	RemoteIP   string                 `json:"remote_ip" validate:"omitempty,ip"`
	RequestID  string                 `json:"request_id" validate:"omitempty"`
	TraceID    string                 `json:"trace_id" validate:"omitempty"`
}

// DatabaseModel represents database operation validation structures.
type DatabaseModel struct {
	Operation      string                 `json:"operation" validate:"required,oneof=SELECT INSERT UPDATE DELETE CREATE DROP ALTER"`
	Table          string                 `json:"table" validate:"required,min=1,max=255"`
	Database       string                 `json:"database" validate:"required,min=1,max=255"`
	Schema         string                 `json:"schema" validate:"omitempty,min=1,max=255"`
	Query          string                 `json:"query" validate:"omitempty,max=10000"`
	Parameters     []interface{}          `json:"parameters" validate:"omitempty"`
	Timestamp      time.Time              `json:"timestamp" validate:"required"`
	Duration       time.Duration          `json:"duration" validate:"omitempty"`
	RowsAffected   int64                  `json:"rows_affected" validate:"omitempty,gte=0"`
	TransactionID  string                 `json:"transaction_id" validate:"omitempty"`
	ConnectionInfo DatabaseConnection     `json:"connection_info" validate:"required"`
	RecordData     map[string]interface{} `json:"record_data" validate:"omitempty"`
	Constraints    DatabaseConstraints    `json:"constraints" validate:"omitempty"`
	Indexes        []DatabaseIndex        `json:"indexes" validate:"omitempty,dive"`
	AuditTrail     DatabaseAuditTrail     `json:"audit_trail" validate:"omitempty"`
}

// DatabaseConnection represents database connection information.
type DatabaseConnection struct {
	Host           string `json:"host" validate:"required,hostname"`
	Port           int    `json:"port" validate:"required,gte=1,lte=65535"`
	Database       string `json:"database" validate:"required,min=1,max=255"`
	Username       string `json:"username" validate:"required,min=1,max=255"`
	SSLMode        string `json:"ssl_mode" validate:"omitempty,oneof=disable allow prefer require"`
	ConnectionPool string `json:"connection_pool" validate:"omitempty,min=1,max=255"`
	ConnectionID   string `json:"connection_id" validate:"omitempty"`
	Driver         string `json:"driver" validate:"omitempty,oneof=postgres mysql sqlite oracle sqlserver"`
	Version        string `json:"version" validate:"omitempty"`
}

// DatabaseConstraints represents database constraint information.
type DatabaseConstraints struct {
	PrimaryKey        DatabasePrimaryKey         `json:"primary_key" validate:"omitempty"`
	UniqueConstraints []DatabaseUniqueConstraint `json:"unique_constraints" validate:"omitempty,dive"`
	ForeignKeys       []DatabaseForeignKey       `json:"foreign_keys" validate:"omitempty,dive"`
	CheckConstraints  []DatabaseCheckConstraint  `json:"check_constraints" validate:"omitempty,dive"`
	NotNullFields     []string                   `json:"not_null_fields" validate:"omitempty,dive,min=1"`
	DataTypes         map[string]string          `json:"data_types" validate:"omitempty"`
}

// DatabasePrimaryKey represents primary key constraint information.
type DatabasePrimaryKey struct {
	Field   string `json:"field" validate:"required,min=1"`
	Type    string `json:"type" validate:"required,min=1"`
	Unique  bool   `json:"unique"`
	NotNull bool   `json:"not_null"`
}

// DatabaseUniqueConstraint represents unique constraint information.
type DatabaseUniqueConstraint struct {
	Fields []string `json:"fields" validate:"required,dive,min=1"`
	Name   string   `json:"name" validate:"required,min=1"`
}

// DatabaseForeignKey represents foreign key constraint information.
type DatabaseForeignKey struct {
	Field      string                      `json:"field" validate:"required,min=1"`
	References DatabaseForeignKeyReference `json:"references" validate:"required"`
	OnDelete   string                      `json:"on_delete" validate:"omitempty,oneof=CASCADE RESTRICT SET NULL SET DEFAULT NO ACTION"`
	OnUpdate   string                      `json:"on_update" validate:"omitempty,oneof=CASCADE RESTRICT SET NULL SET DEFAULT NO ACTION"`
}

// DatabaseForeignKeyReference represents the target of a foreign key.
type DatabaseForeignKeyReference struct {
	Table string `json:"table" validate:"required,min=1"`
	Field string `json:"field" validate:"required,min=1"`
}

// DatabaseCheckConstraint represents check constraint information.
type DatabaseCheckConstraint struct {
	Name      string `json:"name" validate:"required,min=1"`
	Condition string `json:"condition" validate:"required,min=1"`
}

// DatabaseIndex represents database index information.
type DatabaseIndex struct {
	Name   string   `json:"name" validate:"required,min=1"`
	Fields []string `json:"fields" validate:"required,dive,min=1"`
	Type   string   `json:"type" validate:"required,oneof=btree hash gin gist spgist brin"`
	Unique bool     `json:"unique"`
}

// DatabaseAuditTrail represents audit trail information for database operations.
type DatabaseAuditTrail struct {
	OperationType     string                 `json:"operation_type" validate:"required"`
	PerformedBy       string                 `json:"performed_by" validate:"required,min=1"`
	PerformedAt       time.Time              `json:"performed_at" validate:"required"`
	SourceApplication string                 `json:"source_application" validate:"omitempty"`
	SourceIP          string                 `json:"source_ip" validate:"omitempty,ip"`
	AffectedRows      int64                  `json:"affected_rows" validate:"gte=0"`
	RollbackInfo      DatabaseRollbackInfo   `json:"rollback_info" validate:"omitempty"`
	Context           map[string]interface{} `json:"context" validate:"omitempty"`
}

// DatabaseRollbackInfo represents rollback information for database operations.
type DatabaseRollbackInfo struct {
	Available     bool   `json:"available"`
	RetentionDays int    `json:"retention_days" validate:"omitempty,gt=0"`
	BackupID      string `json:"backup_id" validate:"omitempty"`
}

// BaseModel represents a common base structure for all model types.
type BaseModel struct {
	ID        string                 `json:"id" validate:"omitempty"`
	Type      string                 `json:"type" validate:"required,min=1"`
	Version   string                 `json:"version" validate:"omitempty,semver"`
	CreatedAt time.Time              `json:"created_at" validate:"required"`
	UpdatedAt time.Time              `json:"updated_at" validate:"omitempty,gtefield=CreatedAt"`
	Metadata  map[string]interface{} `json:"metadata" validate:"omitempty"`
	Tags      []string               `json:"tags" validate:"omitempty,dive,min=1,max=50"`
	Status    string                 `json:"status" validate:"omitempty,oneof=active inactive pending archived"`
}

// ModelRegistry represents the structure for dynamic model registration.
type ModelRegistry struct {
	Name        string                 `json:"name" validate:"required,min=1,max=100"`
	Type        string                 `json:"type" validate:"required"`
	Version     string                 `json:"version" validate:"required,semver"`
	Description string                 `json:"description" validate:"omitempty,max=500"`
	Schema      map[string]interface{} `json:"schema" validate:"required"`
	Validators  []string               `json:"validators" validate:"omitempty,dive,min=1"`
	Examples    []interface{}          `json:"examples" validate:"omitempty"`
	CreatedAt   time.Time              `json:"created_at" validate:"required"`
	UpdatedAt   time.Time              `json:"updated_at" validate:"omitempty,gtefield=CreatedAt"`
	CreatedBy   string                 `json:"created_by" validate:"required,min=1"`
	IsActive    bool                   `json:"is_active"`
}

// BatchValidationRequest represents a request to validate multiple payloads.
type BatchValidationRequest struct {
	RequestID string        `json:"request_id" validate:"omitempty"`
	ModelType string        `json:"model_type" validate:"required,min=1"`
	Profile   string        `json:"profile" validate:"omitempty"`
	Payloads  []interface{} `json:"payloads" validate:"required,min=1,max=100,dive"`
	Options   BatchOptions  `json:"options" validate:"omitempty"`
}

// BatchValidationResponse represents the response from batch validation.
type BatchValidationResponse struct {
	RequestID    string             `json:"request_id,omitempty"`
	TotalCount   int                `json:"total_count"`
	ValidCount   int                `json:"valid_count"`
	InvalidCount int                `json:"invalid_count"`
	Results      []ValidationResult `json:"results"`
	Summary      BatchSummary       `json:"summary"`
	Timestamp    time.Time          `json:"timestamp"`
	Duration     time.Duration      `json:"duration"`
}

// BatchOptions represents options for batch validation.
type BatchOptions struct {
	StopOnFirstError bool          `json:"stop_on_first_error"`
	MaxConcurrency   int           `json:"max_concurrency" validate:"omitempty,gte=1,lte=100"`
	Timeout          time.Duration `json:"timeout" validate:"omitempty"`
	IncludeMetrics   bool          `json:"include_metrics"`
	FailFast         bool          `json:"fail_fast"`
}

// BatchSummary represents a summary of batch validation results.
type BatchSummary struct {
	SuccessRate           float64            `json:"success_rate"`
	AverageProcessingTime time.Duration      `json:"average_processing_time"`
	ErrorsByCode          map[string]int     `json:"errors_by_code"`
	WarningsByCode        map[string]int     `json:"warnings_by_code"`
	PerformanceStats      PerformanceMetrics `json:"performance_stats"`
	CommonErrors          []string           `json:"common_errors"`
	Recommendations       []string           `json:"recommendations"`
}
