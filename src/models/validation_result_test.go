package models

import (
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
			Errors:           []ValidationError{},
			Warnings:         []ValidationWarning{},
		},
		{
			RowIndex:         1,
			RecordIdentifier: "rec-2",
			IsValid:          false,
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

	// Test total tests ran
	if summary.TotalTestsRan != 3 {
		t.Errorf("Expected TotalTestsRan = 3, got %d", summary.TotalTestsRan)
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
