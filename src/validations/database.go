// Package validations contains database-specific validation logic and business rules.
// This module implements custom validators and business logic for database operations.
package validations

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github-data-validator/models"
	"github.com/go-playground/validator/v10"
)

// DatabaseValidator provides database-specific validation functionality.
type DatabaseValidator struct {
	validator *validator.Validate
}

// NewDatabaseValidator creates a new database validator instance.
func NewDatabaseValidator() *DatabaseValidator {
	v := validator.New()

	// Register database-specific custom validators
	v.RegisterValidation("hostname_rfc1123", validateHostnameRFC1123)
	v.RegisterValidation("semver", validateSemVer)

	return &DatabaseValidator{validator: v}
}

// ValidateQuery validates a database query with comprehensive rules.
func (dv *DatabaseValidator) ValidateQuery(query models.DatabaseQuery) models.ValidationResult {
	start := time.Now()

	result := models.ValidationResult{
		IsValid:   true,
		ModelType: "DatabaseQuery",
		Provider:  "database_validator",
		Timestamp: time.Now(),
		Errors:    []models.ValidationError{},
		Warnings:  []models.ValidationWarning{},
	}

	// Perform struct validation
	if err := dv.validator.Struct(query); err != nil {
		result.IsValid = false

		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, fieldError := range validationErrors {
				result.Errors = append(result.Errors, models.ValidationError{
					Field:      fieldError.Field(),
					Message:    formatDatabaseValidationError(fieldError),
					Code:       fieldError.Tag(),
					Value:      fieldError.Value(),
					Expected:   fieldError.Param(),
					Constraint: fieldError.Tag(),
					Path:       fieldError.Namespace(),
					Severity:   "error",
				})
			}
		}
	}

	// Perform business logic validation
	if result.IsValid {
		warnings := ValidateDatabaseQueryBusinessLogic(query)
		result.Warnings = append(result.Warnings, warnings...)
	}

	// Add performance metrics
	result.ProcessingDuration = time.Since(start)
	result.PerformanceMetrics = &models.PerformanceMetrics{
		ValidationDuration: time.Since(start),
		Provider:           "database_validator",
		FieldCount:         countDatabaseQueryFields(query),
		RuleCount:          dv.getRuleCount(),
	}

	return result
}

// ValidateTransaction validates a database transaction with comprehensive rules.
func (dv *DatabaseValidator) ValidateTransaction(transaction models.DatabaseTransaction) models.ValidationResult {
	start := time.Now()

	result := models.ValidationResult{
		IsValid:   true,
		ModelType: "DatabaseTransaction",
		Provider:  "database_validator",
		Timestamp: time.Now(),
		Errors:    []models.ValidationError{},
		Warnings:  []models.ValidationWarning{},
	}

	// Perform struct validation
	if err := dv.validator.Struct(transaction); err != nil {
		result.IsValid = false

		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, fieldError := range validationErrors {
				result.Errors = append(result.Errors, models.ValidationError{
					Field:      fieldError.Field(),
					Message:    formatDatabaseValidationError(fieldError),
					Code:       fieldError.Tag(),
					Value:      fieldError.Value(),
					Expected:   fieldError.Param(),
					Constraint: fieldError.Tag(),
					Path:       fieldError.Namespace(),
					Severity:   "error",
				})
			}
		}
	}

	// Perform business logic validation
	if result.IsValid {
		warnings := ValidateDatabaseTransactionBusinessLogic(transaction)
		result.Warnings = append(result.Warnings, warnings...)
	}

	// Add performance metrics
	result.ProcessingDuration = time.Since(start)
	result.PerformanceMetrics = &models.PerformanceMetrics{
		ValidationDuration: time.Since(start),
		Provider:           "database_validator",
		FieldCount:         countDatabaseTransactionFields(transaction),
		RuleCount:          dv.getRuleCount(),
	}

	return result
}

// validateHostnameRFC1123 validates hostname according to RFC 1123.
func validateHostnameRFC1123(fl validator.FieldLevel) bool {
	hostname := fl.Field().String()

	if len(hostname) == 0 || len(hostname) > 253 {
		return false
	}

	// RFC 1123 allows letters, digits, and hyphens
	// Cannot start or end with hyphen
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9]([a-zA-Z0-9\-]*[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]*[a-zA-Z0-9])?)*$`, hostname)
	return matched
}

// validateSemVer validates semantic versioning format.
func validateSemVer(fl validator.FieldLevel) bool {
	version := fl.Field().String()

	if version == "" {
		return true // Allow empty for optional fields
	}

	// Semantic versioning pattern: MAJOR.MINOR.PATCH[-PRERELEASE][+BUILD]
	matched, _ := regexp.MatchString(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`, version)
	return matched
}

// ValidateDatabaseQueryBusinessLogic performs database query-specific business logic validation.
func ValidateDatabaseQueryBusinessLogic(query models.DatabaseQuery) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	// Check for SQL injection patterns
	warnings = append(warnings, checkSQLInjectionPatterns(query)...)

	// Check for performance concerns
	warnings = append(warnings, checkDatabasePerformanceConcerns(query)...)

	// Check for security concerns
	warnings = append(warnings, checkDatabaseSecurityConcerns(query)...)

	// Check for best practices
	warnings = append(warnings, checkDatabaseBestPractices(query)...)

	// Check execution plan concerns
	if query.ExecutionPlan != nil {
		warnings = append(warnings, checkExecutionPlanConcerns(*query.ExecutionPlan)...)
	}

	return warnings
}

// ValidateDatabaseTransactionBusinessLogic performs database transaction-specific business logic validation.
func ValidateDatabaseTransactionBusinessLogic(transaction models.DatabaseTransaction) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	// Check for transaction patterns
	warnings = append(warnings, checkTransactionPatterns(transaction)...)

	// Check for lock concerns
	warnings = append(warnings, checkTransactionLockConcerns(transaction)...)

	// Check for isolation level concerns
	warnings = append(warnings, checkIsolationLevelConcerns(transaction)...)

	// Check for long-running transactions
	warnings = append(warnings, checkLongRunningTransaction(transaction)...)

	return warnings
}

// checkSQLInjectionPatterns checks for potential SQL injection patterns.
func checkSQLInjectionPatterns(query models.DatabaseQuery) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	queryLower := strings.ToLower(query.Query)

	// Check for suspicious patterns
	suspiciousPatterns := []string{
		"' or '1'='1",
		"' or 1=1",
		"union select",
		"drop table",
		"delete from",
		"truncate table",
		"alter table",
		"create table",
		"exec ",
		"execute ",
		"sp_",
		"xp_",
		"--",
		"/*",
		"*/",
	}

	foundPatterns := []string{}
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(queryLower, pattern) {
			foundPatterns = append(foundPatterns, pattern)
		}
	}

	if len(foundPatterns) > 0 {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "Query",
			Message:    fmt.Sprintf("Potentially dangerous SQL patterns detected: %s", strings.Join(foundPatterns, ", ")),
			Code:       "SUSPICIOUS_SQL_PATTERNS",
			Suggestion: "Use parameterized queries and validate all inputs",
			Category:   "security",
		})
	}

	// Check for unprepared statements with user input indicators
	if !query.PreparedStatement && len(query.Parameters) == 0 {
		inputIndicators := []string{"@", "$", "?", ":"}
		hasInputIndicators := false
		for _, indicator := range inputIndicators {
			if strings.Contains(query.Query, indicator) {
				hasInputIndicators = true
				break
			}
		}

		if !hasInputIndicators && (strings.Contains(queryLower, "where") || strings.Contains(queryLower, "having")) {
			warnings = append(warnings, models.ValidationWarning{
				Field:      "PreparedStatement",
				Message:    "Query with conditions not using prepared statements",
				Code:       "UNPREPARED_CONDITIONAL_QUERY",
				Suggestion: "Use prepared statements to prevent SQL injection",
				Category:   "security",
			})
		}
	}

	return warnings
}

// checkDatabasePerformanceConcerns checks for performance-related concerns.
func checkDatabasePerformanceConcerns(query models.DatabaseQuery) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	queryLower := strings.ToLower(query.Query)

	// Check for SELECT * queries
	if strings.Contains(queryLower, "select *") {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "Query",
			Message:    "SELECT * query detected",
			Code:       "SELECT_ALL_COLUMNS",
			Suggestion: "Specify only required columns for better performance",
			Category:   "performance",
		})
	}

	// Check for queries without WHERE clause on large operations
	if (strings.Contains(queryLower, "delete from") ||
		strings.Contains(queryLower, "update ")) &&
		!strings.Contains(queryLower, "where") {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "Query",
			Message:    "DELETE/UPDATE query without WHERE clause",
			Code:       "MISSING_WHERE_CLAUSE",
			Suggestion: "Add WHERE clause to limit affected rows",
			Category:   "performance",
		})
	}

	// Check for LIKE queries with leading wildcards
	if strings.Contains(queryLower, "like '%") {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "Query",
			Message:    "LIKE query with leading wildcard detected",
			Code:       "LEADING_WILDCARD_LIKE",
			Suggestion: "Leading wildcards prevent index usage - consider full-text search",
			Category:   "performance",
		})
	}

	// Check for subqueries that could be JOINs
	subqueryCount := strings.Count(queryLower, "select")
	if subqueryCount > 1 {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "Query",
			Message:    fmt.Sprintf("Multiple SELECT statements detected: %d", subqueryCount),
			Code:       "MULTIPLE_SUBQUERIES",
			Suggestion: "Consider using JOINs instead of subqueries for better performance",
			Category:   "performance",
		})
	}

	// Check for long query execution time
	if query.Duration > 5*time.Second {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "Duration",
			Message:    fmt.Sprintf("Slow query execution: %v", query.Duration),
			Code:       "SLOW_QUERY",
			Value:      query.Duration,
			Suggestion: "Optimize query performance with indexes or query restructuring",
			Category:   "performance",
		})
	}

	// Check for high row count operations
	if query.RowsAffected > 100000 {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "RowsAffected",
			Message:    fmt.Sprintf("High number of rows affected: %d", query.RowsAffected),
			Code:       "HIGH_ROW_COUNT",
			Value:      query.RowsAffected,
			Suggestion: "Consider batch processing for large data operations",
			Category:   "performance",
		})
	}

	return warnings
}

// checkDatabaseSecurityConcerns checks for security-related concerns.
func checkDatabaseSecurityConcerns(query models.DatabaseQuery) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	queryLower := strings.ToLower(query.Query)

	// Check for administrative operations
	adminOperations := []string{
		"create user",
		"drop user",
		"grant ",
		"revoke ",
		"alter user",
		"create role",
		"drop role",
		"create database",
		"drop database",
	}

	for _, operation := range adminOperations {
		if strings.Contains(queryLower, operation) {
			warnings = append(warnings, models.ValidationWarning{
				Field:      "Query",
				Message:    fmt.Sprintf("Administrative operation detected: %s", operation),
				Code:       "ADMIN_OPERATION",
				Suggestion: "Ensure proper authorization for administrative operations",
				Category:   "security",
			})
			break
		}
	}

	// Check for schema modifications
	schemaOperations := []string{
		"create table",
		"drop table",
		"alter table",
		"create index",
		"drop index",
		"create view",
		"drop view",
	}

	for _, operation := range schemaOperations {
		if strings.Contains(queryLower, operation) {
			warnings = append(warnings, models.ValidationWarning{
				Field:      "Query",
				Message:    fmt.Sprintf("Schema modification detected: %s", operation),
				Code:       "SCHEMA_MODIFICATION",
				Suggestion: "Schema changes should be managed through migration scripts",
				Category:   "security",
			})
			break
		}
	}

	// Check for potential data exposure
	if strings.Contains(queryLower, "password") ||
		strings.Contains(queryLower, "secret") ||
		strings.Contains(queryLower, "token") ||
		strings.Contains(queryLower, "ssn") ||
		strings.Contains(queryLower, "credit_card") {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "Query",
			Message:    "Query accesses potentially sensitive data",
			Code:       "SENSITIVE_DATA_ACCESS",
			Suggestion: "Ensure proper authorization and audit logging for sensitive data",
			Category:   "security",
		})
	}

	// Check security context
	if query.Security != nil {
		security := *query.Security

		if !security.AuditLogged {
			warnings = append(warnings, models.ValidationWarning{
				Field:      "Security.AuditLogged",
				Message:    "Query execution not logged for audit",
				Code:       "NO_AUDIT_LOG",
				Suggestion: "Enable audit logging for compliance and security",
				Category:   "security",
			})
		}

		if security.Sensitive && !security.EncryptionUsed {
			warnings = append(warnings, models.ValidationWarning{
				Field:      "Security.EncryptionUsed",
				Message:    "Sensitive data accessed without encryption",
				Code:       "UNENCRYPTED_SENSITIVE_ACCESS",
				Suggestion: "Use encryption for sensitive data access",
				Category:   "security",
			})
		}
	}

	return warnings
}

// checkDatabaseBestPractices checks for database best practice violations.
func checkDatabaseBestPractices(query models.DatabaseQuery) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	// Check for missing transaction context on modifications
	if (query.Operation == "INSERT" || query.Operation == "UPDATE" || query.Operation == "DELETE") &&
		query.TransactionID == "" {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "TransactionID",
			Message:    "Data modification query not within a transaction",
			Code:       "NO_TRANSACTION_CONTEXT",
			Suggestion: "Use transactions for data consistency",
			Category:   "best-practices",
		})
	}

	// Check for very long query strings
	if len(query.Query) > 10000 {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "Query",
			Message:    fmt.Sprintf("Very long query string: %d characters", len(query.Query)),
			Code:       "LONG_QUERY_STRING",
			Value:      len(query.Query),
			Suggestion: "Consider breaking down complex queries or using stored procedures",
			Category:   "maintainability",
		})
	}

	// Check for missing query comments on complex operations
	queryLower := strings.ToLower(query.Query)
	if !strings.Contains(queryLower, "--") && !strings.Contains(queryLower, "/*") {
		complexityIndicators := []string{"join", "union", "case when", "exists", "group by", "having"}
		complexityCount := 0
		for _, indicator := range complexityIndicators {
			if strings.Contains(queryLower, indicator) {
				complexityCount++
			}
		}

		if complexityCount >= 2 {
			warnings = append(warnings, models.ValidationWarning{
				Field:      "Query",
				Message:    "Complex query without documentation comments",
				Code:       "UNDOCUMENTED_COMPLEX_QUERY",
				Suggestion: "Add comments to explain complex query logic",
				Category:   "maintainability",
			})
		}
	}

	return warnings
}

// checkExecutionPlanConcerns checks execution plan for potential issues.
func checkExecutionPlanConcerns(plan models.DatabaseExecutionPlan) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	// Check for high estimated cost
	if plan.EstimatedCost > 1000 {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "ExecutionPlan.EstimatedCost",
			Message:    fmt.Sprintf("High estimated query cost: %.2f", plan.EstimatedCost),
			Code:       "HIGH_QUERY_COST",
			Value:      plan.EstimatedCost,
			Suggestion: "Consider query optimization or index creation",
			Category:   "performance",
		})
	}

	// Check for table scans
	for _, operation := range plan.Operations {
		if strings.ToLower(operation.Type) == "table scan" || strings.ToLower(operation.Type) == "seq scan" {
			warnings = append(warnings, models.ValidationWarning{
				Field:      "ExecutionPlan.Operations",
				Message:    fmt.Sprintf("Table scan detected on table: %s", operation.Table),
				Code:       "TABLE_SCAN",
				Value:      operation.Table,
				Suggestion: "Consider adding appropriate indexes",
				Category:   "performance",
			})
		}
	}

	// Check for unused indexes
	for _, index := range plan.Indexes {
		if !index.Used {
			warnings = append(warnings, models.ValidationWarning{
				Field:      "ExecutionPlan.Indexes",
				Message:    fmt.Sprintf("Index not used: %s", index.Name),
				Code:       "UNUSED_INDEX",
				Value:      index.Name,
				Suggestion: "Review index necessity or query structure",
				Category:   "performance",
			})
		}
	}

	return warnings
}

// checkTransactionPatterns checks for transaction-specific patterns.
func checkTransactionPatterns(transaction models.DatabaseTransaction) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	// Check for autocommit transactions with multiple operations
	if transaction.AutoCommit && len(transaction.Queries) > 1 {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "AutoCommit",
			Message:    "Multiple operations in autocommit transaction",
			Code:       "AUTOCOMMIT_MULTIPLE_OPS",
			Suggestion: "Use explicit transactions for multiple related operations",
			Category:   "best-practices",
		})
	}

	// Check for read-only transactions with write operations
	if transaction.ReadOnly {
		for _, query := range transaction.Queries {
			if query.Operation == "INSERT" || query.Operation == "UPDATE" || query.Operation == "DELETE" {
				warnings = append(warnings, models.ValidationWarning{
					Field:      "ReadOnly",
					Message:    "Write operation in read-only transaction",
					Code:       "WRITE_IN_READONLY",
					Value:      query.Operation,
					Suggestion: "Remove read-only flag or change to read operations",
					Category:   "consistency",
				})
				break
			}
		}
	}

	// Check for excessive savepoints
	if len(transaction.Savepoints) > 10 {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "Savepoints",
			Message:    fmt.Sprintf("High number of savepoints: %d", len(transaction.Savepoints)),
			Code:       "EXCESSIVE_SAVEPOINTS",
			Value:      len(transaction.Savepoints),
			Suggestion: "Consider simplifying transaction logic",
			Category:   "performance",
		})
	}

	return warnings
}

// checkTransactionLockConcerns checks for lock-related concerns.
func checkTransactionLockConcerns(transaction models.DatabaseTransaction) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	// Check for blocking locks
	blockingLocks := 0
	for _, lock := range transaction.Locks {
		if len(lock.Blocking) > 0 {
			blockingLocks++
		}
	}

	if blockingLocks > 0 {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "Locks",
			Message:    fmt.Sprintf("Transaction has %d blocking locks", blockingLocks),
			Code:       "BLOCKING_LOCKS",
			Value:      blockingLocks,
			Suggestion: "Consider reducing lock duration or changing isolation level",
			Category:   "concurrency",
		})
	}

	// Check for long-held locks
	for _, lock := range transaction.Locks {
		if lock.Duration > 30*time.Second {
			warnings = append(warnings, models.ValidationWarning{
				Field:      "Locks.Duration",
				Message:    fmt.Sprintf("Long-held lock: %v on %s", lock.Duration, lock.Resource),
				Code:       "LONG_HELD_LOCK",
				Value:      lock.Duration,
				Suggestion: "Optimize transaction to reduce lock duration",
				Category:   "concurrency",
			})
		}
	}

	return warnings
}

// checkIsolationLevelConcerns checks isolation level appropriateness.
func checkIsolationLevelConcerns(transaction models.DatabaseTransaction) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	// Check for serializable isolation with high concurrency needs
	if transaction.IsolationLevel == "serializable" && len(transaction.Locks) > 5 {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "IsolationLevel",
			Message:    "Serializable isolation with high lock count may cause contention",
			Code:       "HIGH_ISOLATION_CONTENTION",
			Suggestion: "Consider lower isolation level if data consistency allows",
			Category:   "concurrency",
		})
	}

	// Check for read uncommitted with sensitive operations
	if transaction.IsolationLevel == "read_uncommitted" {
		for _, query := range transaction.Queries {
			if query.Security != nil && query.Security.Sensitive {
				warnings = append(warnings, models.ValidationWarning{
					Field:      "IsolationLevel",
					Message:    "Read uncommitted isolation with sensitive data",
					Code:       "UNSAFE_ISOLATION_SENSITIVE",
					Suggestion: "Use higher isolation level for sensitive data operations",
					Category:   "security",
				})
				break
			}
		}
	}

	return warnings
}

// checkLongRunningTransaction checks for long-running transaction concerns.
func checkLongRunningTransaction(transaction models.DatabaseTransaction) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	if transaction.EndTime != nil {
		duration := transaction.EndTime.Sub(transaction.StartTime)
		if duration > 5*time.Minute {
			warnings = append(warnings, models.ValidationWarning{
				Field:      "Duration",
				Message:    fmt.Sprintf("Long-running transaction: %v", duration),
				Code:       "LONG_RUNNING_TRANSACTION",
				Value:      duration,
				Suggestion: "Break down into smaller transactions to reduce lock time",
				Category:   "performance",
			})
		}
	} else if time.Since(transaction.StartTime) > 10*time.Minute {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "StartTime",
			Message:    fmt.Sprintf("Transaction running for: %v", time.Since(transaction.StartTime)),
			Code:       "STALE_TRANSACTION",
			Value:      time.Since(transaction.StartTime),
			Suggestion: "Review transaction state and consider timeout mechanisms",
			Category:   "reliability",
		})
	}

	return warnings
}

// formatDatabaseValidationError formats validation errors with database-specific context.
func formatDatabaseValidationError(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("Field '%s' is required for database validation", fe.Field())
	case "hostname_rfc1123":
		return fmt.Sprintf("Field '%s' must be a valid hostname according to RFC 1123", fe.Field())
	case "semver":
		return fmt.Sprintf("Field '%s' must be a valid semantic version", fe.Field())
	case "oneof":
		return fmt.Sprintf("Field '%s' must be one of: %s", fe.Field(), fe.Param())
	case "min":
		return fmt.Sprintf("Field '%s' must be at least %s characters long", fe.Field(), fe.Param())
	case "max":
		return fmt.Sprintf("Field '%s' must be at most %s characters long", fe.Field(), fe.Param())
	case "gt":
		return fmt.Sprintf("Field '%s' must be greater than %s", fe.Field(), fe.Param())
	case "gte":
		return fmt.Sprintf("Field '%s' must be greater than or equal to %s", fe.Field(), fe.Param())
	case "lte":
		return fmt.Sprintf("Field '%s' must be less than or equal to %s", fe.Field(), fe.Param())
	default:
		return fmt.Sprintf("Field '%s' failed validation: %s", fe.Field(), fe.Tag())
	}
}

// countDatabaseQueryFields counts the number of fields in a database query for metrics.
func countDatabaseQueryFields(query models.DatabaseQuery) int {
	count := 15 // Base fields
	count += len(query.Parameters)
	if query.ExecutionPlan != nil {
		count += len(query.ExecutionPlan.Operations)
		count += len(query.ExecutionPlan.Indexes)
	}
	return count
}

// countDatabaseTransactionFields counts the number of fields in a database transaction for metrics.
func countDatabaseTransactionFields(transaction models.DatabaseTransaction) int {
	count := 15 // Base fields
	count += len(transaction.Queries)
	count += len(transaction.Savepoints)
	count += len(transaction.Locks)
	return count
}

// getRuleCount returns the number of validation rules applied.
func (dv *DatabaseValidator) getRuleCount() int {
	// Return approximate number of validation rules
	return 45
}
