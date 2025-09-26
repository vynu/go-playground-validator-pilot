// Package validations provides shared validation framework and utilities
package validations

import (
	"fmt"
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
			Code:       config.ErrCodeValidationFailed,
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
