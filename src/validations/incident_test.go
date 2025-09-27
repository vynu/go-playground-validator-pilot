package validations

import (
	"strings"
	"testing"
	"time"

	"goplayground-data-validator/models"
)

func TestNewIncidentValidator(t *testing.T) {
	validator := NewIncidentValidator()

	if validator == nil {
		t.Fatal("NewIncidentValidator returned nil")
	}
	if validator.validator == nil {
		t.Error("validator instance should not be nil")
	}
}

func TestIncidentValidator_ValidatePayload(t *testing.T) {
	validator := NewIncidentValidator()

	tests := []struct {
		name            string
		payload         models.IncidentPayload
		expectValid     bool
		expectErrors    int
		expectWarnings  int
		checkErrorField string
	}{
		{
			name:         "valid incident",
			payload:      getValidIncidentPayload(),
			expectValid:  true,
			expectErrors: 0,
		},
		{
			name: "missing required fields",
			payload: models.IncidentPayload{
				Title: "Short title", // Too short
				// Missing other required fields
			},
			expectValid:  false,
			expectErrors: 1, // At least one error
		},
		{
			name: "invalid ID format",
			payload: func() models.IncidentPayload {
				p := getValidIncidentPayload()
				p.ID = "invalid-id-format"
				return p
			}(),
			expectValid:     false,
			expectErrors:    1,
			checkErrorField: "id",
		},
		{
			name: "invalid priority severity combination",
			payload: func() models.IncidentPayload {
				p := getValidIncidentPayload()
				p.Priority = 1          // Low priority
				p.Severity = "critical" // High severity - inconsistent
				return p
			}(),
			expectValid:     false,
			expectErrors:    1,
			checkErrorField: "priority",
		},
		{
			name: "critical incident without assignee",
			payload: func() models.IncidentPayload {
				p := getValidIncidentPayload()
				p.Severity = "critical"
				p.Priority = 5    // Consistent with critical severity
				p.AssignedTo = "" // No assignee for critical
				return p
			}(),
			expectValid:    true, // Basic validation passes
			expectWarnings: 1,    // Business logic warning
		},
		{
			name: "production incident with low priority",
			payload: func() models.IncidentPayload {
				p := getValidIncidentPayload()
				p.Environment = "production"
				p.Priority = 1     // Low priority for production
				p.Severity = "low" // Consistent with priority
				return p
			}(),
			expectValid:    true, // Basic validation passes
			expectWarnings: 1,    // Business logic warning
		},
		{
			name: "old incident still open",
			payload: func() models.IncidentPayload {
				p := getValidIncidentPayload()
				p.Status = "open"
				p.Environment = "staging"                      // Avoid production warning
				p.ReportedAt = time.Now().Add(-25 * time.Hour) // 25 hours ago
				return p
			}(),
			expectValid:    true, // Basic validation passes
			expectWarnings: 1,    // Business logic warning
		},
		{
			name: "security incident without tags",
			payload: func() models.IncidentPayload {
				p := getValidIncidentPayload()
				p.Category = "security"
				p.Environment = "staging" // Avoid production warning
				p.Tags = []string{}       // No tags for security incident
				return p
			}(),
			expectValid:    true, // Basic validation passes
			expectWarnings: 1,    // Business logic warning
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidatePayload(tt.payload)

			if result.IsValid != tt.expectValid {
				t.Errorf("IsValid = %v, want %v", result.IsValid, tt.expectValid)
			}

			if tt.expectErrors > 0 && len(result.Errors) < tt.expectErrors {
				t.Errorf("Expected at least %d errors, got %d", tt.expectErrors, len(result.Errors))
			}

			if tt.expectWarnings > 0 && len(result.Warnings) < tt.expectWarnings {
				t.Errorf("Expected at least %d warnings, got %d", tt.expectWarnings, len(result.Warnings))
			}

			if tt.checkErrorField != "" {
				found := false
				for _, err := range result.Errors {
					if err.Field == tt.checkErrorField {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error for field %s", tt.checkErrorField)
				}
			}

			// Check basic result structure
			if result.ModelType != "incident" {
				t.Errorf("ModelType = %s, want incident", result.ModelType)
			}
			if result.Provider != "go-playground" {
				t.Errorf("Provider = %s, want go-playground", result.Provider)
			}
		})
	}
}

func TestIncidentValidator_validateIncidentIDFormat(t *testing.T) {
	validator := NewIncidentValidator()

	tests := []struct {
		id      string
		wantErr bool
	}{
		{"INC-20240927-0001", false},
		{"INC-20240101-9999", false},
		{"INC-2024-001", true},
		{"INC-20240101-001", true},
		{"INCIDENT-2024-001", true},
		{"invalid", true},
		{"INC-INVALID-001", true},
		{"", true},
		{"INC-2024", true}, // Too short
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			err := validator.validateIncidentIDFormat(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateIncidentIDFormat(%s) error = %v, wantErr %v", tt.id, err, tt.wantErr)
			}
		})
	}
}

func TestIncidentValidator_validatePrioritySeverityConsistency(t *testing.T) {
	validator := NewIncidentValidator()

	tests := []struct {
		priority int
		severity string
		wantErr  bool
	}{
		{5, "critical", false}, // High priority, critical severity
		{4, "high", false},     // High priority, high severity
		{3, "medium", false},   // Medium priority, medium severity
		{2, "low", false},      // Low priority, low severity
		{1, "critical", true},  // Low priority, critical severity - inconsistent
		{5, "low", true},       // High priority, low severity - inconsistent
		{1, "high", true},      // Low priority, high severity - inconsistent
	}

	for _, tt := range tests {
		t.Run(tt.severity, func(t *testing.T) {
			err := validator.validatePrioritySeverityConsistency(tt.priority, tt.severity)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePrioritySeverityConsistency(%d, %s) error = %v, wantErr %v",
					tt.priority, tt.severity, err, tt.wantErr)
			}
		})
	}
}

func TestIncidentValidator_validateBusinessLogic(t *testing.T) {
	validator := NewIncidentValidator()

	tests := []struct {
		name            string
		payload         models.IncidentPayload
		expectWarnings  int
		warningContains string
	}{
		{
			name:           "no warnings",
			payload:        getValidIncidentPayload(),
			expectWarnings: 0,
		},
		{
			name: "critical without assignee",
			payload: func() models.IncidentPayload {
				p := getValidIncidentPayload()
				p.Severity = "critical"
				p.Priority = 5 // Consistent with critical
				p.AssignedTo = ""
				p.Environment = "staging" // Avoid production warning
				return p
			}(),
			expectWarnings:  1,
			warningContains: "assigned",
		},
		{
			name: "production low priority",
			payload: func() models.IncidentPayload {
				p := getValidIncidentPayload()
				p.Environment = "production"
				p.Priority = 1
				p.Severity = "low" // Consistent with priority
				return p
			}(),
			expectWarnings:  1,
			warningContains: "priority",
		},
		{
			name: "old open incident",
			payload: func() models.IncidentPayload {
				p := getValidIncidentPayload()
				p.Status = "investigating"
				p.Environment = "staging" // Avoid production warning
				p.ReportedAt = time.Now().Add(-25 * time.Hour)
				return p
			}(),
			expectWarnings:  1,
			warningContains: "hours",
		},
		{
			name: "security without tags",
			payload: func() models.IncidentPayload {
				p := getValidIncidentPayload()
				p.Category = "security"
				p.Environment = "staging" // Avoid production warning
				p.Tags = []string{}
				return p
			}(),
			expectWarnings:  1,
			warningContains: "security",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warnings := validator.validateBusinessLogic(tt.payload)

			if len(warnings) != tt.expectWarnings {
				t.Errorf("Expected %d warnings, got %d", tt.expectWarnings, len(warnings))
			}

			if tt.warningContains != "" && len(warnings) > 0 {
				found := false
				for _, warning := range warnings {
					if strings.Contains(strings.ToLower(warning.Message), strings.ToLower(tt.warningContains)) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected warning containing '%s'", tt.warningContains)
				}
			}
		})
	}
}

func TestIncidentValidator_getCustomErrorMessage(t *testing.T) {
	validator := NewIncidentValidator()

	// Test custom error message formatting
	// This would require creating mock validation errors
	// For now, just test that the method exists and doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("getCustomErrorMessage panicked: %v", r)
		}
	}()

	// Create a mock payload to trigger validation errors
	payload := models.IncidentPayload{
		ID: "", // This will trigger required validation
	}

	result := validator.ValidatePayload(payload)
	if len(result.Errors) > 0 {
		// Check that error messages are formatted properly
		for _, err := range result.Errors {
			if err.Message == "" {
				t.Error("Error message should not be empty")
			}
			if err.Field == "" {
				t.Error("Error field should not be empty")
			}
		}
	}
}

func BenchmarkIncidentValidator_ValidatePayload(b *testing.B) {
	validator := NewIncidentValidator()
	payload := getValidIncidentPayload()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.ValidatePayload(payload)
	}
}

func BenchmarkIncidentValidator_BusinessLogic(b *testing.B) {
	validator := NewIncidentValidator()
	payload := getValidIncidentPayload()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.validateBusinessLogic(payload)
	}
}

// Helper functions

func getValidIncidentPayload() models.IncidentPayload {
	return models.IncidentPayload{
		ID:          "INC-20240927-0001",
		Title:       "Database connection timeout affecting user authentication and session management",
		Description: "Users are experiencing timeout errors when connecting to the main database server",
		Severity:    "medium",
		Status:      "resolved",
		Priority:    3,
		Category:    "performance",
		Environment: "staging",
		ReportedBy:  "john.doe@example.com",
		AssignedTo:  "jane.smith@example.com",
		ReportedAt:  time.Now().Add(-time.Hour), // 1 hour ago
		UpdatedAt:   time.Now(),
		Tags:        []string{"database", "timeout"},
		Impact:      "medium",
	}
}

func containsIgnoreCase(str, substr string) bool {
	return len(str) >= len(substr) &&
		len(substr) > 0 &&
		(str == substr ||
			(len(str) > len(substr) &&
				(str[:len(substr)] == substr ||
					str[len(str)-len(substr):] == substr ||
					containsSubstring(str, substr))))
}

func containsSubstring(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
