package validations

import (
	"testing"
	"time"

	"goplayground-data-validator/models"
)

func TestNewBaseValidator(t *testing.T) {
	tests := []struct {
		name      string
		modelType string
		provider  string
	}{
		{"incident validator", "incident", "test-provider"},
		{"github validator", "github", "test-provider"},
		{"api validator", "api", "test-provider"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bv := NewBaseValidator(tt.modelType, tt.provider)

			if bv == nil {
				t.Fatal("NewBaseValidator returned nil")
			}
			if bv.modelType != tt.modelType {
				t.Errorf("modelType = %s, want %s", bv.modelType, tt.modelType)
			}
			if bv.provider != tt.provider {
				t.Errorf("provider = %s, want %s", bv.provider, tt.provider)
			}
			if bv.validator == nil {
				t.Error("validator instance is nil")
			}
		})
	}
}

func TestBaseValidator_CreateValidationResult(t *testing.T) {
	bv := NewBaseValidator("test", "test-provider")

	result := bv.CreateValidationResult()

	if !result.IsValid {
		t.Error("Expected IsValid to be true by default")
	}
	if result.ModelType != "test" {
		t.Errorf("ModelType = %s, want test", result.ModelType)
	}
	if result.Provider != "test-provider" {
		t.Errorf("Provider = %s, want test-provider", result.Provider)
	}
	if result.Errors == nil {
		t.Error("Errors slice should be initialized")
	}
	if result.Warnings == nil {
		t.Error("Warnings slice should be initialized")
	}
	if len(result.Errors) != 0 {
		t.Errorf("Expected empty errors slice, got length %d", len(result.Errors))
	}
	if len(result.Warnings) != 0 {
		t.Errorf("Expected empty warnings slice, got length %d", len(result.Warnings))
	}
}

func TestBaseValidator_AddPerformanceMetrics(t *testing.T) {
	bv := NewBaseValidator("incident", "test-provider")
	result := bv.CreateValidationResult()

	start := time.Now().Add(-time.Millisecond * 50)
	bv.AddPerformanceMetrics(&result, start)

	if result.ProcessingDuration == 0 {
		t.Error("ProcessingDuration should be set")
	}
	if result.PerformanceMetrics == nil {
		t.Fatal("PerformanceMetrics should be set")
	}
	if result.PerformanceMetrics.ValidationDuration == 0 {
		t.Error("ValidationDuration should be set")
	}
	if result.PerformanceMetrics.FieldCount == 0 {
		t.Error("FieldCount should be greater than 0")
	}
	if result.PerformanceMetrics.RuleCount == 0 {
		t.Error("RuleCount should be greater than 0")
	}
}

func TestBaseValidator_AddPerformanceMetrics_SlowValidation(t *testing.T) {
	bv := NewBaseValidator("incident", "test-provider")
	result := bv.CreateValidationResult()

	// Simulate slow validation (> 1 second ago)
	start := time.Now().Add(-time.Second * 2)
	bv.AddPerformanceMetrics(&result, start)

	// Should add a performance warning
	found := false
	for _, warning := range result.Warnings {
		if warning.Field == "performance" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected performance warning for slow validation")
	}
}

func TestBaseValidator_countStructFields(t *testing.T) {
	tests := []struct {
		modelType string
		expected  int
	}{
		{"github", 25},
		{"incident", 10},
		{"api", 15},
		{"database", 12},
		{"generic", 8},
		{"deployment", 18},
		{"unknown", 10}, // Default
	}

	bv := NewBaseValidator("test", "test-provider")

	for _, tt := range tests {
		t.Run(tt.modelType, func(t *testing.T) {
			count := bv.countStructFields(tt.modelType)
			if count != tt.expected {
				t.Errorf("countStructFields(%s) = %d, want %d", tt.modelType, count, tt.expected)
			}
		})
	}
}

func TestBaseValidator_getRuleCount(t *testing.T) {
	tests := []struct {
		modelType string
		expected  int
	}{
		{"github", 50},
		{"incident", 25},
		{"api", 30},
		{"database", 28},
		{"generic", 15},
		{"deployment", 35},
		{"unknown", 20}, // Default
	}

	for _, tt := range tests {
		t.Run(tt.modelType, func(t *testing.T) {
			bv := NewBaseValidator(tt.modelType, "test-provider")
			count := bv.getRuleCount()
			if count != tt.expected {
				t.Errorf("getRuleCount() for %s = %d, want %d", tt.modelType, count, tt.expected)
			}
		})
	}
}

func TestBaseValidator_ValidateWithBusinessLogic(t *testing.T) {
	bv := NewBaseValidator("test", "test-provider")

	// Test payload
	payload := struct {
		Name string `validate:"required"`
	}{
		Name: "test",
	}

	// Business logic function that adds a warning
	businessLogic := func(interface{}) []models.ValidationWarning {
		return []models.ValidationWarning{
			{
				Field:   "business",
				Message: "Business logic warning",
				Code:    "BUSINESS_WARNING",
			},
		}
	}

	result := bv.ValidateWithBusinessLogic(payload, businessLogic)

	if !result.IsValid {
		t.Error("Expected validation to pass")
	}
	if len(result.Warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(result.Warnings))
	}
	if result.Warnings[0].Field != "business" {
		t.Errorf("Expected business warning, got %s", result.Warnings[0].Field)
	}
	if result.ProcessingDuration == 0 {
		t.Error("ProcessingDuration should be set")
	}
}

func TestBaseValidator_ValidateWithBusinessLogic_ValidationFailed(t *testing.T) {
	bv := NewBaseValidator("test", "test-provider")

	// Test payload with validation error
	payload := struct {
		Name string `validate:"required"`
	}{
		Name: "", // Empty name should fail
	}

	// Business logic function
	businessLogic := func(interface{}) []models.ValidationWarning {
		return []models.ValidationWarning{}
	}

	result := bv.ValidateWithBusinessLogic(payload, businessLogic)

	if result.IsValid {
		t.Error("Expected validation to fail")
	}
	if len(result.Errors) == 0 {
		t.Error("Expected validation errors")
	}
}

func BenchmarkBaseValidator_CreateValidationResult(b *testing.B) {
	bv := NewBaseValidator("test", "test-provider")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bv.CreateValidationResult()
	}
}

func BenchmarkBaseValidator_AddPerformanceMetrics(b *testing.B) {
	bv := NewBaseValidator("incident", "test-provider")
	start := time.Now()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := bv.CreateValidationResult()
		bv.AddPerformanceMetrics(&result, start)
	}
}
