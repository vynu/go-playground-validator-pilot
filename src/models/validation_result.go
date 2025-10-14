// Package models provides array validation result structures
package models

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// ArrayValidationResult represents the result of validating an array of records
type ArrayValidationResult struct {
	BatchID        string                `json:"batch_id"`            // Universal tracking
	Status         string                `json:"status"`              // "success" or "failed" based on threshold
	TotalRecords   int                   `json:"total_records"`       // Total number of records
	ValidRecords   int                   `json:"valid_records"`       // Number of valid records
	InvalidRecords int                   `json:"invalid_records"`     // Number of invalid records
	WarningRecords int                   `json:"warning_records"`     // Number of records with warnings only
	Threshold      *float64              `json:"threshold,omitempty"` // Optional threshold percentage (e.g., 20.0 for 20%)
	ProcessingTime int64                 `json:"processing_time_ms"`  // Processing time in milliseconds
	CompletedAt    time.Time             `json:"completed_at"`        // Completion timestamp
	Summary        ValidationSummary     `json:"summary"`             // Summary of validation
	Results        []RowValidationResult `json:"results"`             // Individual row results (only invalid/warning rows)
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

// BatchSession tracks validation across multiple requests
type BatchSession struct {
	BatchID        string    `json:"batch_id"`
	TotalRecords   int       `json:"total_records"`
	ValidRecords   int       `json:"valid_records"`
	InvalidRecords int       `json:"invalid_records"`
	WarningRecords int       `json:"warning_records"`
	Threshold      *float64  `json:"threshold,omitempty"`
	StartedAt      time.Time `json:"started_at"`
	LastUpdated    time.Time `json:"last_updated"`
	IsFinal        bool      `json:"is_final"` // Set to true when client sends final batch
	mutex          sync.RWMutex
}

// BatchSessionManager manages batch sessions across multiple requests
type BatchSessionManager struct {
	sessions map[string]*BatchSession
	mutex    sync.RWMutex
}

var (
	globalBatchManager *BatchSessionManager
	batchManagerOnce   sync.Once
)

// GetBatchSessionManager returns the global batch session manager
func GetBatchSessionManager() *BatchSessionManager {
	batchManagerOnce.Do(func() {
		globalBatchManager = &BatchSessionManager{
			sessions: make(map[string]*BatchSession),
		}
	})
	return globalBatchManager
}

// CreateBatchSession creates a new batch session
func (bsm *BatchSessionManager) CreateBatchSession(batchID string, threshold *float64) *BatchSession {
	bsm.mutex.Lock()
	defer bsm.mutex.Unlock()

	session := &BatchSession{
		BatchID:     batchID,
		Threshold:   threshold,
		StartedAt:   time.Now(),
		LastUpdated: time.Now(),
		IsFinal:     false,
	}
	bsm.sessions[batchID] = session
	return session
}

// GetBatchSession retrieves a batch session by ID
func (bsm *BatchSessionManager) GetBatchSession(batchID string) (*BatchSession, bool) {
	bsm.mutex.RLock()
	defer bsm.mutex.RUnlock()

	session, exists := bsm.sessions[batchID]
	return session, exists
}

// UpdateBatchSession adds validation results to existing batch session
func (bsm *BatchSessionManager) UpdateBatchSession(batchID string, validCount, invalidCount, warningCount int) error {
	bsm.mutex.Lock()
	defer bsm.mutex.Unlock()

	session, exists := bsm.sessions[batchID]
	if !exists {
		return fmt.Errorf("batch session %s not found", batchID)
	}

	session.mutex.Lock()
	defer session.mutex.Unlock()

	session.TotalRecords += validCount + invalidCount
	session.ValidRecords += validCount
	session.InvalidRecords += invalidCount
	session.WarningRecords += warningCount
	session.LastUpdated = time.Now()

	return nil
}

// FinalizeBatchSession marks the batch as complete and returns final status
func (bsm *BatchSessionManager) FinalizeBatchSession(batchID string) (string, error) {
	bsm.mutex.Lock()
	defer bsm.mutex.Unlock()

	session, exists := bsm.sessions[batchID]
	if !exists {
		return "", fmt.Errorf("batch session %s not found", batchID)
	}

	session.mutex.Lock()
	defer session.mutex.Unlock()

	session.IsFinal = true
	session.LastUpdated = time.Now()

	// Calculate final status based on threshold
	status := "success"
	if session.Threshold != nil && session.TotalRecords > 0 {
		successRate := (float64(session.ValidRecords) / float64(session.TotalRecords)) * 100.0
		if successRate < *session.Threshold {
			status = "failed"
		}
	} else if session.TotalRecords == 1 && session.InvalidRecords > 0 {
		status = "failed"
	}

	return status, nil
}

// DeleteBatchSession removes a batch session
func (bsm *BatchSessionManager) DeleteBatchSession(batchID string) {
	bsm.mutex.Lock()
	defer bsm.mutex.Unlock()

	delete(bsm.sessions, batchID)
}

// CleanupExpiredBatches removes batch sessions older than 30 minutes
func (bsm *BatchSessionManager) CleanupExpiredBatches() {
	bsm.mutex.Lock()
	defer bsm.mutex.Unlock()

	now := time.Now()
	expirationDuration := 30 * time.Minute

	for batchID, session := range bsm.sessions {
		session.mutex.RLock()
		age := now.Sub(session.LastUpdated)
		session.mutex.RUnlock()

		if age > expirationDuration {
			delete(bsm.sessions, batchID)
		}
	}
}

// StartCleanupRoutine starts a background goroutine to cleanup expired batches
func (bsm *BatchSessionManager) StartCleanupRoutine() {
	go func() {
		ticker := time.NewTicker(5 * time.Minute) // Check every 5 minutes
		defer ticker.Stop()

		for range ticker.C {
			bsm.CleanupExpiredBatches()
		}
	}()
}

// GetBatchStatus returns current status of batch session
func (bs *BatchSession) GetStatus() map[string]interface{} {
	bs.mutex.RLock()
	defer bs.mutex.RUnlock()

	successRate := 0.0
	if bs.TotalRecords > 0 {
		successRate = (float64(bs.ValidRecords) / float64(bs.TotalRecords)) * 100.0
	}

	status := "in_progress"
	if bs.IsFinal {
		status = "success"
		if bs.Threshold != nil && successRate < *bs.Threshold {
			status = "failed"
		} else if bs.TotalRecords == 1 && bs.InvalidRecords > 0 {
			status = "failed"
		}
	}

	return map[string]interface{}{
		"batch_id":        bs.BatchID,
		"status":          status,
		"total_records":   bs.TotalRecords,
		"valid_records":   bs.ValidRecords,
		"invalid_records": bs.InvalidRecords,
		"warning_records": bs.WarningRecords,
		"success_rate":    successRate,
		"threshold":       bs.Threshold,
		"started_at":      bs.StartedAt,
		"last_updated":    bs.LastUpdated,
		"is_final":        bs.IsFinal,
	}
}
