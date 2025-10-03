// Package models provides array validation result structures
package models

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// ArrayValidationResult represents the result of validating an array of records
type ArrayValidationResult struct {
	BatchID        string                `json:"batch_id"`           // Universal tracking
	Status         string                `json:"status"`             // "completed"
	TotalRecords   int                   `json:"total_records"`      // Total number of records
	ValidRecords   int                   `json:"valid_records"`      // Number of valid records
	InvalidRecords int                   `json:"invalid_records"`    // Number of invalid records
	ProcessingTime int64                 `json:"processing_time_ms"` // Processing time in milliseconds
	CompletedAt    time.Time             `json:"completed_at"`       // Completion timestamp
	Summary        ValidationSummary     `json:"summary"`            // Summary of validation
	Results        []RowValidationResult `json:"results"`            // Individual row results
}

// RowValidationResult represents the validation result for a single row
type RowValidationResult struct {
	RowIndex         int                 `json:"row_index"`          // Index of the row
	RecordIdentifier string              `json:"record_identifier"`  // Auto-detected ID
	IsValid          bool                `json:"is_valid"`           // Whether the row is valid
	ValidationTime   int64               `json:"validation_time_ms"` // Validation time in milliseconds
	TestName         string              `json:"test_name"`          // Name of the validation test applied (e.g., "IncidentValidator:IDFormat")
	Errors           []ValidationError   `json:"errors,omitempty"`   // Validation errors
	Warnings         []ValidationWarning `json:"warnings,omitempty"` // Validation warnings
}

// ValidationSummary provides aggregated statistics about the validation
type ValidationSummary struct {
	SuccessRate           float64  `json:"success_rate"`            // Percentage of valid records
	ValidationErrors      int      `json:"validation_errors"`       // Total number of errors
	ValidationWarnings    int      `json:"validation_warnings"`     // Total number of warnings
	TotalRecordsProcessed int      `json:"total_records_processed"` // Total records processed
	TotalTestsRan         int      `json:"total_tests_ran"`         // Total validation tests ran
	SuccessfulTestNames   []string `json:"successful_test_names"`   // Names of successful tests
	FailedTestNames       []string `json:"failed_test_names"`       // Names of failed tests
}

// GenerateBatchID generates a unique batch ID with a given prefix
func GenerateBatchID(prefix string) string {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID if random generation fails
		return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
	}
	return fmt.Sprintf("%s_%s", prefix, hex.EncodeToString(bytes))
}

// DetectRecordIdentifier attempts to extract a unique identifier from a record
func DetectRecordIdentifier(record map[string]interface{}, rowIndex int) string {
	// Common ID field patterns to check
	idPatterns := []string{"id", "ID", "_id", "uuid", "UUID", "identifier", "recordId", "record_id"}

	for _, pattern := range idPatterns {
		if val, ok := record[pattern]; ok && val != nil {
			return fmt.Sprintf("%v", val)
		}
	}

	// Fallback: use row index
	return fmt.Sprintf("row_%d", rowIndex)
}

// BuildSummary builds a ValidationSummary from an array of RowValidationResults
func BuildSummary(results []RowValidationResult) ValidationSummary {
	// Pre-allocate slices with estimated capacity to avoid repeated allocations
	totalRecords := len(results)
	summary := ValidationSummary{
		TotalRecordsProcessed: totalRecords,
		SuccessfulTestNames:   make([]string, 0, totalRecords),
		FailedTestNames:       make([]string, 0, totalRecords),
	}

	validCount := 0
	totalErrors := 0
	totalWarnings := 0

	// Use maps to track unique test names
	successfulTests := make(map[string]bool)
	failedTests := make(map[string]bool)

	for _, result := range results {
		if result.IsValid {
			validCount++
			// Add to successful test names set if test name is provided
			if result.TestName != "" {
				successfulTests[result.TestName] = true
			}
		} else {
			// Add to failed test names set if test name is provided
			if result.TestName != "" {
				failedTests[result.TestName] = true
			}
		}
		totalErrors += len(result.Errors)
		totalWarnings += len(result.Warnings)
	}

	// Convert sets to slices for unique test names
	for testName := range successfulTests {
		summary.SuccessfulTestNames = append(summary.SuccessfulTestNames, testName)
	}
	for testName := range failedTests {
		summary.FailedTestNames = append(summary.FailedTestNames, testName)
	}

	summary.ValidationErrors = totalErrors
	summary.ValidationWarnings = totalWarnings

	if totalRecords > 0 {
		summary.SuccessRate = float64(validCount) / float64(totalRecords) * 100
	}

	// Total tests ran is the number of unique test types that were executed
	// Count all unique test names (merge both successful and failed)
	allTests := make(map[string]bool)
	for testName := range successfulTests {
		allTests[testName] = true
	}
	for testName := range failedTests {
		allTests[testName] = true
	}
	summary.TotalTestsRan = len(allTests)

	return summary
}
