// Package models contains database-specific models with comprehensive validation rules.
// This module defines structures for validating database operations and related data.
package models

import (
	"time"
)

// DatabaseQuery represents a database query validation structure.
type DatabaseQuery struct {
	ID                string                 `json:"id" validate:"omitempty"`
	Operation         string                 `json:"operation" validate:"required,oneof=SELECT INSERT UPDATE DELETE CREATE DROP ALTER TRUNCATE GRANT REVOKE"`
	Table             string                 `json:"table" validate:"required_if=Operation SELECT,required_if=Operation INSERT,required_if=Operation UPDATE,required_if=Operation DELETE,min=1,max=255"`
	Database          string                 `json:"database" validate:"required,min=1,max=255"`
	Schema            string                 `json:"schema" validate:"omitempty,min=1,max=255"`
	Query             string                 `json:"query" validate:"required,min=1,max=100000"`
	Parameters        []interface{}          `json:"parameters" validate:"omitempty"`
	PreparedStatement bool                   `json:"prepared_statement"`
	Timestamp         time.Time              `json:"timestamp" validate:"required"`
	Duration          time.Duration          `json:"duration" validate:"omitempty,gte=0"`
	RowsAffected      int64                  `json:"rows_affected" validate:"omitempty,gte=0"`
	RowsReturned      int64                  `json:"rows_returned" validate:"omitempty,gte=0"`
	TransactionID     string                 `json:"transaction_id" validate:"omitempty"`
	ConnectionInfo    DatabaseConnectionInfo `json:"connection_info" validate:"required"`
	ExecutionPlan     *DatabaseExecutionPlan `json:"execution_plan,omitempty" validate:"omitempty"`
	Performance       *DatabasePerformance   `json:"performance,omitempty" validate:"omitempty"`
	Security          *DatabaseSecurity      `json:"security,omitempty" validate:"omitempty"`
	Context           map[string]interface{} `json:"context,omitempty" validate:"omitempty"`
}

// DatabaseConnectionInfo represents database connection information.
type DatabaseConnectionInfo struct {
	Host           string `json:"host" validate:"required,hostname_rfc1123"`
	Port           int    `json:"port" validate:"required,gte=1,lte=65535"`
	Database       string `json:"database" validate:"required,min=1,max=255"`
	Username       string `json:"username" validate:"required,min=1,max=255"`
	Driver         string `json:"driver" validate:"required,oneof=postgres mysql sqlite oracle sqlserver mongodb cassandra redis"`
	Version        string `json:"version" validate:"omitempty"`
	SSLMode        string `json:"ssl_mode" validate:"omitempty,oneof=disable allow prefer require verify-ca verify-full"`
	ConnectionPool string `json:"connection_pool" validate:"omitempty,min=1,max=255"`
	ConnectionID   string `json:"connection_id" validate:"omitempty"`
	ServerID       string `json:"server_id" validate:"omitempty"`
	ClusterID      string `json:"cluster_id" validate:"omitempty"`
	Timezone       string `json:"timezone" validate:"omitempty"`
	Charset        string `json:"charset" validate:"omitempty"`
	Collation      string `json:"collation" validate:"omitempty"`
}

// DatabaseExecutionPlan represents query execution plan information.
type DatabaseExecutionPlan struct {
	PlanID        string                  `json:"plan_id" validate:"omitempty"`
	QueryHash     string                  `json:"query_hash" validate:"omitempty,len=64,hexadecimal"`
	Operations    []DatabasePlanOperation `json:"operations" validate:"omitempty,dive"`
	EstimatedCost float64                 `json:"estimated_cost" validate:"omitempty,gte=0"`
	ActualCost    *float64                `json:"actual_cost,omitempty" validate:"omitempty,gte=0"`
	Indexes       []DatabaseIndexUsage    `json:"indexes" validate:"omitempty,dive"`
	Warnings      []string                `json:"warnings" validate:"omitempty,dive,min=1"`
	Suggestions   []string                `json:"suggestions" validate:"omitempty,dive,min=1"`
	CacheHit      bool                    `json:"cache_hit"`
	Parallel      bool                    `json:"parallel"`
}

// DatabasePlanOperation represents an operation in the execution plan.
type DatabasePlanOperation struct {
	Type     string                  `json:"type" validate:"required,min=1"`
	Table    string                  `json:"table" validate:"omitempty,min=1"`
	Index    string                  `json:"index" validate:"omitempty,min=1"`
	Rows     int64                   `json:"rows" validate:"omitempty,gte=0"`
	Cost     float64                 `json:"cost" validate:"omitempty,gte=0"`
	Duration time.Duration           `json:"duration" validate:"omitempty,gte=0"`
	Filter   string                  `json:"filter" validate:"omitempty"`
	Sort     []string                `json:"sort" validate:"omitempty,dive,min=1"`
	Join     *DatabaseJoinInfo       `json:"join,omitempty" validate:"omitempty"`
	Children []DatabasePlanOperation `json:"children" validate:"omitempty,dive"`
}

// DatabaseJoinInfo represents join operation information.
type DatabaseJoinInfo struct {
	Type      string `json:"type" validate:"required,oneof=INNER LEFT RIGHT FULL CROSS"`
	Condition string `json:"condition" validate:"required,min=1"`
	Algorithm string `json:"algorithm" validate:"omitempty,oneof=NESTED_LOOP HASH MERGE SORT"`
}

// DatabaseIndexUsage represents index usage information.
type DatabaseIndexUsage struct {
	Name        string     `json:"name" validate:"required,min=1"`
	Table       string     `json:"table" validate:"required,min=1"`
	Type        string     `json:"type" validate:"required,oneof=btree hash gin gist spgist brin unique clustered"`
	Used        bool       `json:"used"`
	Selectivity float64    `json:"selectivity" validate:"omitempty,gte=0,lte=1"`
	Rows        int64      `json:"rows" validate:"omitempty,gte=0"`
	Size        int64      `json:"size" validate:"omitempty,gte=0"`
	LastUsed    *time.Time `json:"last_used,omitempty" validate:"omitempty"`
	HitRatio    float64    `json:"hit_ratio" validate:"omitempty,gte=0,lte=1"`
	Suggestion  string     `json:"suggestion" validate:"omitempty"`
}

// DatabasePerformance represents performance metrics.
type DatabasePerformance struct {
	QueryTime      time.Duration `json:"query_time" validate:"required,gte=0"`
	ParseTime      time.Duration `json:"parse_time" validate:"omitempty,gte=0"`
	PlanTime       time.Duration `json:"plan_time" validate:"omitempty,gte=0"`
	ExecuteTime    time.Duration `json:"execute_time" validate:"omitempty,gte=0"`
	NetworkTime    time.Duration `json:"network_time" validate:"omitempty,gte=0"`
	LockWaitTime   time.Duration `json:"lock_wait_time" validate:"omitempty,gte=0"`
	IOWaitTime     time.Duration `json:"io_wait_time" validate:"omitempty,gte=0"`
	CPUTime        time.Duration `json:"cpu_time" validate:"omitempty,gte=0"`
	MemoryUsed     int64         `json:"memory_used" validate:"omitempty,gte=0"`
	TempSpaceUsed  int64         `json:"temp_space_used" validate:"omitempty,gte=0"`
	BufferHits     int64         `json:"buffer_hits" validate:"omitempty,gte=0"`
	BufferMisses   int64         `json:"buffer_misses" validate:"omitempty,gte=0"`
	PhysicalReads  int64         `json:"physical_reads" validate:"omitempty,gte=0"`
	LogicalReads   int64         `json:"logical_reads" validate:"omitempty,gte=0"`
	Writes         int64         `json:"writes" validate:"omitempty,gte=0"`
	NetworkBytes   int64         `json:"network_bytes" validate:"omitempty,gte=0"`
	ConnectionTime time.Duration `json:"connection_time" validate:"omitempty,gte=0"`
	LockCount      int           `json:"lock_count" validate:"omitempty,gte=0"`
	DeadlockCount  int           `json:"deadlock_count" validate:"omitempty,gte=0"`
}

// DatabaseSecurity represents security-related information.
type DatabaseSecurity struct {
	UserRole           string                `json:"user_role" validate:"omitempty,min=1"`
	Permissions        []string              `json:"permissions" validate:"omitempty,dive,min=1"`
	AccessLevel        string                `json:"access_level" validate:"omitempty,oneof=read write admin"`
	EncryptionUsed     bool                  `json:"encryption_used"`
	SSLUsed            bool                  `json:"ssl_used"`
	AuditLogged        bool                  `json:"audit_logged"`
	Sensitive          bool                  `json:"sensitive"`
	DataClassification string                `json:"data_classification" validate:"omitempty,oneof=public internal confidential restricted"`
	ComplianceFlags    []string              `json:"compliance_flags" validate:"omitempty,dive,oneof=GDPR HIPAA SOX PCI-DSS"`
	AccessPattern      DatabaseAccessPattern `json:"access_pattern" validate:"omitempty"`
	Anomaly            *DatabaseAnomaly      `json:"anomaly,omitempty" validate:"omitempty"`
}

// DatabaseAccessPattern represents access pattern analysis.
type DatabaseAccessPattern struct {
	Frequency       string    `json:"frequency" validate:"omitempty,oneof=low medium high"`
	TimePattern     string    `json:"time_pattern" validate:"omitempty,oneof=business_hours off_hours mixed"`
	LocationPattern string    `json:"location_pattern" validate:"omitempty,oneof=internal external mixed"`
	UserPattern     string    `json:"user_pattern" validate:"omitempty,oneof=single multiple automated"`
	DataVolume      string    `json:"data_volume" validate:"omitempty,oneof=small medium large bulk"`
	LastAccess      time.Time `json:"last_access" validate:"omitempty"`
	AccessCount     int64     `json:"access_count" validate:"omitempty,gte=0"`
	UniqueUsers     int       `json:"unique_users" validate:"omitempty,gte=0"`
}

// DatabaseAnomaly represents detected anomalies.
type DatabaseAnomaly struct {
	Type           string    `json:"type" validate:"required,oneof=unusual_access suspicious_query high_volume unusual_time unauthorized_access"`
	Severity       string    `json:"severity" validate:"required,oneof=low medium high critical"`
	Description    string    `json:"description" validate:"required,min=1,max=1000"`
	Confidence     float64   `json:"confidence" validate:"required,gte=0,lte=1"`
	DetectedAt     time.Time `json:"detected_at" validate:"required"`
	Evidence       []string  `json:"evidence" validate:"omitempty,dive,min=1"`
	Recommendation string    `json:"recommendation" validate:"omitempty,max=1000"`
}

// DatabaseTransaction represents a database transaction.
type DatabaseTransaction struct {
	ID             string                 `json:"id" validate:"required,min=1"`
	Status         string                 `json:"status" validate:"required,oneof=active committed rolled_back preparing prepared"`
	IsolationLevel string                 `json:"isolation_level" validate:"required,oneof=read_uncommitted read_committed repeatable_read serializable"`
	StartTime      time.Time              `json:"start_time" validate:"required"`
	EndTime        *time.Time             `json:"end_time,omitempty" validate:"omitempty,gtefield=StartTime"`
	Duration       time.Duration          `json:"duration" validate:"omitempty,gte=0"`
	Queries        []DatabaseQuery        `json:"queries" validate:"omitempty,dive"`
	ReadOnly       bool                   `json:"read_only"`
	AutoCommit     bool                   `json:"auto_commit"`
	LockTimeout    time.Duration          `json:"lock_timeout" validate:"omitempty,gte=0"`
	Savepoints     []DatabaseSavepoint    `json:"savepoints" validate:"omitempty,dive"`
	Locks          []DatabaseLock         `json:"locks" validate:"omitempty,dive"`
	Performance    *DatabasePerformance   `json:"performance,omitempty" validate:"omitempty"`
	ConnectionInfo DatabaseConnectionInfo `json:"connection_info" validate:"required"`
	Context        map[string]interface{} `json:"context,omitempty" validate:"omitempty"`
}

// DatabaseSavepoint represents a transaction savepoint.
type DatabaseSavepoint struct {
	Name       string    `json:"name" validate:"required,min=1,max=255"`
	CreatedAt  time.Time `json:"created_at" validate:"required"`
	Released   bool      `json:"released"`
	RolledBack bool      `json:"rolled_back"`
}

// DatabaseLock represents a database lock.
type DatabaseLock struct {
	Type       string        `json:"type" validate:"required,oneof=shared exclusive update intention"`
	Resource   string        `json:"resource" validate:"required,min=1"`
	Mode       string        `json:"mode" validate:"required,oneof=table row page extent database"`
	AcquiredAt time.Time     `json:"acquired_at" validate:"required"`
	Duration   time.Duration `json:"duration" validate:"omitempty,gte=0"`
	Granted    bool          `json:"granted"`
	Waiting    bool          `json:"waiting"`
	Blocking   []string      `json:"blocking" validate:"omitempty,dive,min=1"`
	BlockedBy  []string      `json:"blocked_by" validate:"omitempty,dive,min=1"`
}

// DatabaseMigration represents a database migration.
type DatabaseMigration struct {
	ID           string                     `json:"id" validate:"required,min=1"`
	Version      string                     `json:"version" validate:"required,semver"`
	Name         string                     `json:"name" validate:"required,min=1,max=255"`
	Description  string                     `json:"description" validate:"omitempty,max=1000"`
	Type         string                     `json:"type" validate:"required,oneof=schema data both"`
	Direction    string                     `json:"direction" validate:"required,oneof=up down"`
	SQL          string                     `json:"sql" validate:"required,min=1"`
	Checksum     string                     `json:"checksum" validate:"required,len=64,hexadecimal"`
	ExecutedAt   *time.Time                 `json:"executed_at,omitempty" validate:"omitempty"`
	Duration     time.Duration              `json:"duration" validate:"omitempty,gte=0"`
	Success      bool                       `json:"success"`
	Error        string                     `json:"error" validate:"omitempty"`
	Rollback     *DatabaseMigrationRollback `json:"rollback,omitempty" validate:"omitempty"`
	Dependencies []string                   `json:"dependencies" validate:"omitempty,dive,min=1"`
	Tags         []string                   `json:"tags" validate:"omitempty,dive,min=1,max=50"`
	Environment  string                     `json:"environment" validate:"omitempty,oneof=development staging production"`
	Author       string                     `json:"author" validate:"omitempty,min=1,max=255"`
	Context      map[string]interface{}     `json:"context,omitempty" validate:"omitempty"`
}

// DatabaseMigrationRollback represents rollback information for a migration.
type DatabaseMigrationRollback struct {
	Available  bool       `json:"available"`
	SQL        string     `json:"sql" validate:"omitempty,min=1"`
	Automatic  bool       `json:"automatic"`
	ExecutedAt *time.Time `json:"executed_at,omitempty" validate:"omitempty"`
	Success    bool       `json:"success"`
	Error      string     `json:"error" validate:"omitempty"`
}

// DatabaseBackup represents a database backup operation.
type DatabaseBackup struct {
	ID              string                     `json:"id" validate:"required,min=1"`
	Type            string                     `json:"type" validate:"required,oneof=full incremental differential transaction_log"`
	Status          string                     `json:"status" validate:"required,oneof=pending running completed failed canceled"`
	StartTime       time.Time                  `json:"start_time" validate:"required"`
	EndTime         *time.Time                 `json:"end_time,omitempty" validate:"omitempty,gtefield=StartTime"`
	Duration        time.Duration              `json:"duration" validate:"omitempty,gte=0"`
	Size            int64                      `json:"size" validate:"omitempty,gte=0"`
	CompressedSize  int64                      `json:"compressed_size" validate:"omitempty,gte=0"`
	Location        string                     `json:"location" validate:"required,min=1"`
	Checksum        string                     `json:"checksum" validate:"omitempty,len=64,hexadecimal"`
	Encryption      *DatabaseBackupEncryption  `json:"encryption,omitempty" validate:"omitempty"`
	Compression     *DatabaseBackupCompression `json:"compression,omitempty" validate:"omitempty"`
	RetentionPolicy *DatabaseRetentionPolicy   `json:"retention_policy,omitempty" validate:"omitempty"`
	Tables          []string                   `json:"tables" validate:"omitempty,dive,min=1"`
	ExcludedTables  []string                   `json:"excluded_tables" validate:"omitempty,dive,min=1"`
	Error           string                     `json:"error" validate:"omitempty"`
	Metadata        map[string]interface{}     `json:"metadata,omitempty" validate:"omitempty"`
	ConnectionInfo  DatabaseConnectionInfo     `json:"connection_info" validate:"required"`
}

// DatabaseBackupEncryption represents backup encryption settings.
type DatabaseBackupEncryption struct {
	Enabled   bool   `json:"enabled"`
	Algorithm string `json:"algorithm" validate:"omitempty,oneof=AES256 AES128"`
	KeyID     string `json:"key_id" validate:"omitempty,min=1"`
}

// DatabaseBackupCompression represents backup compression settings.
type DatabaseBackupCompression struct {
	Enabled   bool    `json:"enabled"`
	Algorithm string  `json:"algorithm" validate:"omitempty,oneof=gzip lz4 zstd"`
	Level     int     `json:"level" validate:"omitempty,gte=1,lte=9"`
	Ratio     float64 `json:"ratio" validate:"omitempty,gte=0,lte=1"`
}

// DatabaseRetentionPolicy represents backup retention policy.
type DatabaseRetentionPolicy struct {
	Days       int    `json:"days" validate:"required,gt=0"`
	Weeks      int    `json:"weeks" validate:"omitempty,gt=0"`
	Months     int    `json:"months" validate:"omitempty,gt=0"`
	Years      int    `json:"years" validate:"omitempty,gt=0"`
	MaxBackups int    `json:"max_backups" validate:"omitempty,gt=0"`
	Policy     string `json:"policy" validate:"omitempty,oneof=time_based count_based size_based"`
	AutoDelete bool   `json:"auto_delete"`
}
