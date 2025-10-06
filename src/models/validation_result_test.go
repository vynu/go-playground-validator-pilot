package models

import (
	"fmt"
	"testing"
)

func TestGenerateBatchID(t *testing.T) {
	prefix := "test"
	batchID := GenerateBatchID(prefix)

	// Check that batch ID starts with the prefix
	if batchID[:len(prefix)] != prefix {
		t.Errorf("Expected batch ID to start with %s, got %s", prefix, batchID)
	}

	// Check that batch ID contains an underscore separator
	if len(batchID) <= len(prefix)+1 {
		t.Errorf("Expected batch ID to contain more than just the prefix, got %s", batchID)
	}

	// Generate another batch ID and ensure they're different
	batchID2 := GenerateBatchID(prefix)
	if batchID == batchID2 {
		t.Errorf("Expected batch IDs to be unique, got duplicate: %s", batchID)
	}
}

func TestDetectRecordIdentifier(t *testing.T) {
	tests := []struct {
		name     string
		record   map[string]interface{}
		rowIndex int
		expected string
	}{
		{
			name:     "ID field present",
			record:   map[string]interface{}{"id": "test-123"},
			rowIndex: 0,
			expected: "test-123",
		},
		{
			name:     "Uppercase ID field",
			record:   map[string]interface{}{"ID": "TEST-456"},
			rowIndex: 1,
			expected: "TEST-456",
		},
		{
			name:     "_id field present",
			record:   map[string]interface{}{"_id": "mongo-789"},
			rowIndex: 2,
			expected: "mongo-789",
		},
		{
			name:     "uuid field present",
			record:   map[string]interface{}{"uuid": "uuid-abc"},
			rowIndex: 3,
			expected: "uuid-abc",
		},
		{
			name:     "No ID field - fallback to row index",
			record:   map[string]interface{}{"name": "test"},
			rowIndex: 5,
			expected: "row_5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectRecordIdentifier(tt.record, tt.rowIndex)
			if result != tt.expected {
				t.Errorf("DetectRecordIdentifier() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestBuildSummary(t *testing.T) {
	results := []RowValidationResult{
		{
			RowIndex:         0,
			RecordIdentifier: "rec-1",
			IsValid:          true,
			TestName:         "UserValidator",
			Errors:           []ValidationError{},
			Warnings:         []ValidationWarning{},
		},
		{
			RowIndex:         1,
			RecordIdentifier: "rec-2",
			IsValid:          false,
			TestName:         "UserValidator",
			Errors: []ValidationError{
				{Field: "name", Message: "required", Code: "REQ"},
				{Field: "email", Message: "invalid", Code: "INV"},
			},
			Warnings: []ValidationWarning{
				{Field: "age", Message: "suspicious", Code: "WARN"},
			},
		},
		{
			RowIndex:         2,
			RecordIdentifier: "rec-3",
			IsValid:          true,
			TestName:         "UserValidator",
			Errors:           []ValidationError{},
			Warnings:         []ValidationWarning{},
		},
	}

	summary := BuildSummary(results)

	// Test total records processed
	if summary.TotalRecordsProcessed != 3 {
		t.Errorf("Expected TotalRecordsProcessed = 3, got %d", summary.TotalRecordsProcessed)
	}

	// Test success rate (2 out of 3 valid = 66.67%)
	expectedSuccessRate := float64(2) / float64(3) * 100
	if summary.SuccessRate != expectedSuccessRate {
		t.Errorf("Expected SuccessRate = %.2f, got %.2f", expectedSuccessRate, summary.SuccessRate)
	}

	// Test total errors
	if summary.ValidationErrors != 2 {
		t.Errorf("Expected ValidationErrors = 2, got %d", summary.ValidationErrors)
	}

	// Test total warnings
	if summary.ValidationWarnings != 1 {
		t.Errorf("Expected ValidationWarnings = 1, got %d", summary.ValidationWarnings)
	}

	// Test total tests ran (unique tests: 1 successful + 1 failed = 1 unique test "UserValidator")
	if summary.TotalTestsRan != 1 {
		t.Errorf("Expected TotalTestsRan = 1 (unique tests), got %d", summary.TotalTestsRan)
	}

	// Test successful test names (should have exactly 1 unique test name)
	if len(summary.SuccessfulTestNames) != 1 {
		t.Errorf("Expected 1 unique successful test name, got %d", len(summary.SuccessfulTestNames))
	}
	if len(summary.SuccessfulTestNames) > 0 && summary.SuccessfulTestNames[0] != "UserValidator" {
		t.Errorf("Expected successful test name 'UserValidator', got '%s'", summary.SuccessfulTestNames[0])
	}

	// Test failed test names (should have exactly 1 unique test name)
	if len(summary.FailedTestNames) != 1 {
		t.Errorf("Expected 1 unique failed test name, got %d", len(summary.FailedTestNames))
	}
	if len(summary.FailedTestNames) > 0 && summary.FailedTestNames[0] != "UserValidator" {
		t.Errorf("Expected failed test name 'UserValidator', got '%s'", summary.FailedTestNames[0])
	}
}

func TestBuildSummary_EmptyResults(t *testing.T) {
	results := []RowValidationResult{}
	summary := BuildSummary(results)

	if summary.TotalRecordsProcessed != 0 {
		t.Errorf("Expected TotalRecordsProcessed = 0, got %d", summary.TotalRecordsProcessed)
	}

	if summary.SuccessRate != 0 {
		t.Errorf("Expected SuccessRate = 0, got %.2f", summary.SuccessRate)
	}

	if summary.ValidationErrors != 0 {
		t.Errorf("Expected ValidationErrors = 0, got %d", summary.ValidationErrors)
	}

	if summary.TotalTestsRan != 0 {
		t.Errorf("Expected TotalTestsRan = 0, got %d", summary.TotalTestsRan)
	}
}

func TestBuildSummary_AllInvalid(t *testing.T) {
	results := []RowValidationResult{
		{
			RowIndex:         0,
			RecordIdentifier: "rec-1",
			IsValid:          false,
			Errors: []ValidationError{
				{Field: "field1", Message: "error1", Code: "E1"},
			},
			Warnings: []ValidationWarning{},
		},
		{
			RowIndex:         1,
			RecordIdentifier: "rec-2",
			IsValid:          false,
			Errors: []ValidationError{
				{Field: "field2", Message: "error2", Code: "E2"},
			},
			Warnings: []ValidationWarning{},
		},
	}

	summary := BuildSummary(results)

	// Success rate should be 0% when all records are invalid
	if summary.SuccessRate != 0 {
		t.Errorf("Expected SuccessRate = 0, got %.2f", summary.SuccessRate)
	}

	if summary.ValidationErrors != 2 {
		t.Errorf("Expected ValidationErrors = 2, got %d", summary.ValidationErrors)
	}
}

// Test Batch Session Manager
func TestBatchSessionManager_CreateSession(t *testing.T) {
	manager := GetBatchSessionManager()
	threshold := 20.0

	session := manager.CreateBatchSession("batch-001", &threshold)

	if session.BatchID != "batch-001" {
		t.Errorf("Expected BatchID = batch-001, got %s", session.BatchID)
	}

	if session.Threshold == nil || *session.Threshold != 20.0 {
		t.Errorf("Expected Threshold = 20.0, got %v", session.Threshold)
	}

	if session.TotalRecords != 0 {
		t.Errorf("Expected TotalRecords = 0, got %d", session.TotalRecords)
	}

	// Clean up
	manager.DeleteBatchSession("batch-001")
}

func TestBatchSessionManager_UpdateSession(t *testing.T) {
	manager := GetBatchSessionManager()
	threshold := 50.0

	_ = manager.CreateBatchSession("batch-002", &threshold)

	// Update with validation results
	err := manager.UpdateBatchSession("batch-002", 80, 20, 5)
	if err != nil {
		t.Errorf("UpdateBatchSession failed: %v", err)
	}

	// Verify updated values
	retrievedSession, exists := manager.GetBatchSession("batch-002")
	if !exists {
		t.Fatal("Session not found after update")
	}

	if retrievedSession.ValidRecords != 80 {
		t.Errorf("Expected ValidRecords = 80, got %d", retrievedSession.ValidRecords)
	}

	if retrievedSession.InvalidRecords != 20 {
		t.Errorf("Expected InvalidRecords = 20, got %d", retrievedSession.InvalidRecords)
	}

	if retrievedSession.TotalRecords != 100 {
		t.Errorf("Expected TotalRecords = 100, got %d", retrievedSession.TotalRecords)
	}

	if retrievedSession.WarningRecords != 5 {
		t.Errorf("Expected WarningRecords = 5, got %d", retrievedSession.WarningRecords)
	}

	// Clean up
	manager.DeleteBatchSession("batch-002")
}

func TestBatchSessionManager_FinalizeSession(t *testing.T) {
	manager := GetBatchSessionManager()

	tests := []struct {
		name           string
		threshold      *float64
		validCount     int
		invalidCount   int
		expectedStatus string
	}{
		{
			name:           "Success with threshold - exactly at threshold",
			threshold:      floatPtr(20.0),
			validCount:     20,
			invalidCount:   80,
			expectedStatus: "success",
		},
		{
			name:           "Success with threshold - above threshold",
			threshold:      floatPtr(20.0),
			validCount:     21,
			invalidCount:   79,
			expectedStatus: "success",
		},
		{
			name:           "Failed with threshold - below threshold",
			threshold:      floatPtr(20.0),
			validCount:     19,
			invalidCount:   81,
			expectedStatus: "failed",
		},
		{
			name:           "Success without threshold",
			threshold:      nil,
			validCount:     50,
			invalidCount:   50,
			expectedStatus: "success",
		},
		{
			name:           "Single record valid - no threshold",
			threshold:      nil,
			validCount:     1,
			invalidCount:   0,
			expectedStatus: "success",
		},
		{
			name:           "Single record invalid - no threshold",
			threshold:      nil,
			validCount:     0,
			invalidCount:   1,
			expectedStatus: "failed",
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			batchID := fmt.Sprintf("batch-finalize-%d", i)
			_ = manager.CreateBatchSession(batchID, tt.threshold)

			// Update session with counts
			manager.UpdateBatchSession(batchID, tt.validCount, tt.invalidCount, 0)

			// Finalize session
			status, err := manager.FinalizeBatchSession(batchID)
			if err != nil {
				t.Errorf("FinalizeBatchSession failed: %v", err)
			}

			if status != tt.expectedStatus {
				t.Errorf("Expected status = %s, got %s", tt.expectedStatus, status)
			}

			// Verify IsFinal flag
			retrievedSession, _ := manager.GetBatchSession(batchID)
			if !retrievedSession.IsFinal {
				t.Error("Expected IsFinal = true after finalization")
			}

			// Clean up
			manager.DeleteBatchSession(batchID)
		})
	}
}

func TestBatchSession_GetStatus(t *testing.T) {
	manager := GetBatchSessionManager()
	threshold := 30.0

	session := manager.CreateBatchSession("batch-status", &threshold)
	manager.UpdateBatchSession("batch-status", 40, 60, 5)

	status := session.GetStatus()

	// Verify status fields
	if status["batch_id"] != "batch-status" {
		t.Errorf("Expected batch_id = batch-status, got %v", status["batch_id"])
	}

	if status["total_records"] != 100 {
		t.Errorf("Expected total_records = 100, got %v", status["total_records"])
	}

	if status["valid_records"] != 40 {
		t.Errorf("Expected valid_records = 40, got %v", status["valid_records"])
	}

	successRate := status["success_rate"].(float64)
	if successRate != 40.0 {
		t.Errorf("Expected success_rate = 40.0, got %.2f", successRate)
	}

	// Before finalization, status should be "in_progress"
	if status["status"] != "in_progress" {
		t.Errorf("Expected status = in_progress, got %v", status["status"])
	}

	// After finalization with 40% success rate and 30% threshold, should be "success"
	manager.FinalizeBatchSession("batch-status")
	status = session.GetStatus()

	if status["status"] != "success" {
		t.Errorf("Expected status = success (40%% >= 30%%), got %v", status["status"])
	}

	// Clean up
	manager.DeleteBatchSession("batch-status")
}

func TestBatchSession_ThresholdEdgeCases(t *testing.T) {
	manager := GetBatchSessionManager()

	// Test exact threshold match (20.0% with 20% threshold)
	threshold := 20.0
	_ = manager.CreateBatchSession("batch-edge-1", &threshold)
	manager.UpdateBatchSession("batch-edge-1", 20, 80, 0)
	status, _ := manager.FinalizeBatchSession("batch-edge-1")

	if status != "success" {
		t.Errorf("Expected success with exact threshold match (20.0%% == 20%%), got %s", status)
	}
	manager.DeleteBatchSession("batch-edge-1")

	// Test just above threshold (20.0001% with 20% threshold)
	_ = manager.CreateBatchSession("batch-edge-2", &threshold)
	manager.UpdateBatchSession("batch-edge-2", 20001, 79999, 0)
	status2, _ := manager.FinalizeBatchSession("batch-edge-2")

	if status2 != "success" {
		t.Errorf("Expected success with 20.001%% > 20%%, got %s", status2)
	}
	manager.DeleteBatchSession("batch-edge-2")

	// Test just below threshold (19.9999% with 20% threshold)
	_ = manager.CreateBatchSession("batch-edge-3", &threshold)
	manager.UpdateBatchSession("batch-edge-3", 19999, 80001, 0)
	status3, _ := manager.FinalizeBatchSession("batch-edge-3")

	if status3 != "failed" {
		t.Errorf("Expected failed with 19.999%% < 20%%, got %s", status3)
	}
	manager.DeleteBatchSession("batch-edge-3")
}

// Helper function to create float64 pointer
func floatPtr(f float64) *float64 {
	return &f
}
