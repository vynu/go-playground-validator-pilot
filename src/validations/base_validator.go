// Package validations provides shared validation framework and utilities
package validations

import (
	"fmt"
	"strings"
	"time"

	"goplayground-data-validator/config"
	"goplayground-data-validator/models"

	"github.com/go-playground/validator/v10"
)

// BaseValidator provides common validation functionality shared across all validators
type BaseValidator struct {
	validator *validator.Validate
	modelType string
	provider  string
}

// NewBaseValidator creates a new BaseValidator with optimized configuration
func NewBaseValidator(modelType, provider string) *BaseValidator {
	return &BaseValidator{
		validator: validator.New(),
		modelType: modelType,
		provider:  provider,
	}
}

// CreateValidationResult creates a standardized validation result with proper initialization
func (bv *BaseValidator) CreateValidationResult() models.ValidationResult {
	return models.ValidationResult{
		IsValid:   true,
		ModelType: bv.modelType,
		Provider:  bv.provider,
		Timestamp: time.Now(),
		Errors:    make([]models.ValidationError, 0, 5),   // Pre-allocate with capacity
		Warnings:  make([]models.ValidationWarning, 0, 3), // Pre-allocate with capacity
	}
}

// AddPerformanceMetrics adds standardized performance metrics to validation result
func (bv *BaseValidator) AddPerformanceMetrics(result *models.ValidationResult, start time.Time) {
	duration := time.Since(start)
	result.ProcessingDuration = duration

	result.PerformanceMetrics = &models.PerformanceMetrics{
		ValidationDuration: duration,
		FieldCount:         bv.countStructFields(result.ModelType),
		RuleCount:          bv.getRuleCount(),
		MemoryUsage:        getApproximateMemoryUsage(),
	}

	// Add performance warning if validation is slow
	if config.IsSlowValidation(duration) {
		result.Warnings = append(result.Warnings, models.ValidationWarning{
			Field:      "performance",
			Message:    fmt.Sprintf("Validation took %v (longer than expected)", duration),
			Code:       config.ErrCodeValidationFailed,
			Suggestion: "Consider optimizing validation logic or payload size",
		})
	}
}

// countStructFields returns the approximate number of fields in the model
func (bv *BaseValidator) countStructFields(modelType string) int {
	// Optimized field counting based on model type
	fieldCounts := map[string]int{
		"github":     25,
		"incident":   10,
		"api":        15,
		"database":   12,
		"generic":    8,
		"deployment": 18,
	}

	if count, exists := fieldCounts[strings.ToLower(modelType)]; exists {
		return count
	}
	return 10 // Default estimate
}

// getRuleCount returns the number of validation rules applied
func (bv *BaseValidator) getRuleCount() int {
	// Optimized rule counting based on model type
	ruleCounts := map[string]int{
		"github":     50,
		"incident":   25,
		"api":        30,
		"database":   28,
		"generic":    15,
		"deployment": 35,
	}

	if count, exists := ruleCounts[strings.ToLower(bv.modelType)]; exists {
		return count
	}
	return 20 // Default estimate
}

// getApproximateMemoryUsage returns estimated memory usage in bytes
func getApproximateMemoryUsage() int64 {
	// Simplified memory estimation - in production this could use runtime.ReadMemStats
	return 1024 * 64 // 64KB estimate
}

// ValidateWithBusinessLogic performs standard validation with optional business logic (optimized)
func (bv *BaseValidator) ValidateWithBusinessLogic(
	payload interface{},
	businessLogicFunc func(interface{}) []models.ValidationWarning,
) models.ValidationResult {
	start := time.Now()

	// Use optimized result creation
	result := bv.CreateValidationResult()

	// Perform struct validation using go-playground/validator
	if err := bv.validator.Struct(payload); err != nil {
		result.IsValid = false
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			// Pre-allocate slice if we know the error count
			if len(validationErrors) > cap(result.Errors) {
				result.Errors = make([]models.ValidationError, 0, len(validationErrors))
			}

			for _, ve := range validationErrors {
				result.Errors = append(result.Errors, models.ValidationError{
					Field:   ve.Field(),
					Message: FormatValidationError(ve, bv.modelType),
					Code:    GetErrorCode(ve.Tag()),
					Value:   fmt.Sprintf("%v", ve.Value()),
				})
			}
		}
	}

	// Apply business logic if validation passed basic checks
	if result.IsValid && businessLogicFunc != nil {
		businessWarnings := businessLogicFunc(payload)
		if len(businessWarnings) > 0 {
			// Pre-allocate if needed
			if len(businessWarnings) > cap(result.Warnings) {
				newWarnings := make([]models.ValidationWarning, len(result.Warnings), len(result.Warnings)+len(businessWarnings))
				copy(newWarnings, result.Warnings)
				result.Warnings = newWarnings
			}
			result.Warnings = append(result.Warnings, businessWarnings...)
		}
	}

	// Add optimized performance metadata
	bv.AddPerformanceMetrics(&result, start)

	return result
}

// FormatValidationError provides consistent error message formatting
func FormatValidationError(fe validator.FieldError, context string) string {
	field := fe.Field()

	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("Field '%s' is required", field)
	case "min":
		if fe.Kind().String() == "string" {
			return fmt.Sprintf("Field '%s' must be at least %s characters long", field, fe.Param())
		}
		return fmt.Sprintf("Field '%s' must be at least %s", field, fe.Param())
	case "max":
		if fe.Kind().String() == "string" {
			return fmt.Sprintf("Field '%s' must be at most %s characters long", field, fe.Param())
		}
		return fmt.Sprintf("Field '%s' must be at most %s", field, fe.Param())
	case "oneof":
		return fmt.Sprintf("Field '%s' must be one of: %s", field, fe.Param())
	case "email":
		return fmt.Sprintf("Field '%s' must be a valid email address", field)
	case "url":
		return fmt.Sprintf("Field '%s' must be a valid URL", field)
	case "uuid":
		return fmt.Sprintf("Field '%s' must be a valid UUID", field)
	case "numeric":
		return fmt.Sprintf("Field '%s' must be numeric", field)
	case "alpha":
		return fmt.Sprintf("Field '%s' must contain only letters", field)
	case "alphanum":
		return fmt.Sprintf("Field '%s' must contain only letters and numbers", field)
	default:
		return fmt.Sprintf("Field '%s' failed validation for %s context: %s", field, context, fe.Error())
	}
}

// GetErrorCode returns a standardized error code for validation tags
func GetErrorCode(tag string) string {
	switch tag {
	case "required":
		return config.ErrCodeRequiredMissing
	case "min":
		return config.ErrCodeValueTooShort
	case "max":
		return config.ErrCodeValueTooLong
	case "oneof":
		return config.ErrCodeInvalidEnum
	case "email":
		return config.ErrCodeInvalidEmail
	case "url":
		return config.ErrCodeInvalidURL
	case "uuid":
		return config.ErrCodeInvalidFormat
	case "numeric":
		return config.ErrCodeInvalidFormat
	case "alpha":
		return config.ErrCodeInvalidFormat
	case "alphanum":
		return config.ErrCodeInvalidFormat
	default:
		return config.ErrCodeValidationFailed
	}
}
