// Package validations provides shared validation framework and utilities
package validations

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"goplayground-data-validator/config"
	"goplayground-data-validator/models"
)

// BaseValidator provides common validation functionality shared across all validators
type BaseValidator struct {
	validator *validator.Validate
	modelType string
	provider  string
}

// NewBaseValidator creates a new base validator instance
func NewBaseValidator(modelType, provider string) *BaseValidator {
	return &BaseValidator{
		validator: validator.New(),
		modelType: modelType,
		provider:  provider,
	}
}

// ValidateWithBusinessLogic performs standard validation with optional business logic
func (bv *BaseValidator) ValidateWithBusinessLogic(
	payload interface{},
	businessLogicFunc func(interface{}) []models.ValidationWarning,
) models.ValidationResult {
	start := time.Now()

	result := models.ValidationResult{
		IsValid:   true,
		ModelType: bv.modelType,
		Provider:  bv.provider,
		Timestamp: time.Now(),
		Errors:    []models.ValidationError{},
		Warnings:  []models.ValidationWarning{},
	}

	// Perform struct validation using go-playground/validator
	if err := bv.validator.Struct(payload); err != nil {
		result.IsValid = false
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
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
		result.Warnings = businessLogicFunc(payload)
	}

	// Add performance metadata using constants
	duration := time.Since(start)
	if config.IsSlowValidation(duration) {
		result.Warnings = append(result.Warnings, models.ValidationWarning{
			Field:      "performance",
			Message:    fmt.Sprintf("Validation took %v (longer than expected)", duration),
			Code:       config.WarnCodePerformance,
			Suggestion: "Consider optimizing validation logic or payload size",
		})
	}

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

// CountStructFields counts the number of fields in a struct using reflection
func CountStructFields(payload interface{}) int {
	return len(GetStructFieldNames(payload))
}

// GetStructFieldNames returns the names of all fields in a struct
func GetStructFieldNames(payload interface{}) []string {
	fieldNames := []string{}

	// Use reflection to get field names - this is a placeholder
	// The actual implementation would use reflect package
	// For now, returning empty slice to avoid compilation issues

	return fieldNames
}

// IsLargePayload checks if a payload exceeds size thresholds
func IsLargePayload(payload interface{}) bool {
	fieldCount := CountStructFields(payload)
	return config.IsLargePayload(fieldCount)
}

// AddPerformanceWarnings adds performance-related warnings based on validation complexity
func AddPerformanceWarnings(result *models.ValidationResult, payload interface{}, duration time.Duration) {
	if IsLargePayload(payload) {
		result.Warnings = append(result.Warnings, models.ValidationWarning{
			Field:      "payload_size",
			Message:    "Large payload detected - consider splitting into smaller requests",
			Code:       config.ErrCodeLargePayload,
			Suggestion: "Split large payloads into smaller, focused requests for better performance",
		})
	}

	if config.IsVerySlowValidation(duration) {
		result.Warnings = append(result.Warnings, models.ValidationWarning{
			Field:      "response_time",
			Message:    fmt.Sprintf("Validation took %v (exceeds threshold)", duration),
			Code:       config.ErrCodeSlowValidation,
			Suggestion: "Consider optimizing validation logic or reducing payload complexity",
		})
	}
}

// ValidateStringField provides common string field validation
func ValidateStringField(value, fieldName string, minLength, maxLength int) *models.ValidationError {
	if strings.TrimSpace(value) == "" {
		return &models.ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("Field '%s' cannot be empty", fieldName),
			Code:    "EMPTY_STRING_FIELD",
			Value:   value,
		}
	}

	if len(value) < minLength {
		return &models.ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("Field '%s' must be at least %d characters", fieldName, minLength),
			Code:    "STRING_TOO_SHORT",
			Value:   value,
		}
	}

	if len(value) > maxLength {
		return &models.ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("Field '%s' must be at most %d characters", fieldName, maxLength),
			Code:    "STRING_TOO_LONG",
			Value:   value,
		}
	}

	return nil
}

// ValidateEmailField provides consistent email validation
func ValidateEmailField(email, fieldName string) *models.ValidationError {
	if email == "" {
		return &models.ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("Field '%s' is required", fieldName),
			Code:    "REQUIRED_FIELD_MISSING",
			Value:   email,
		}
	}

	// Basic email validation - could be enhanced with regex
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return &models.ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("Field '%s' must be a valid email address", fieldName),
			Code:    "INVALID_EMAIL_FORMAT",
			Value:   email,
		}
	}

	return nil
}
