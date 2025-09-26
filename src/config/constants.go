// Package config provides configuration constants and validation thresholds
package config

import "time"

// Validation Thresholds
const (
	// Performance thresholds
	SlowValidationThreshold     = 100 * time.Millisecond
	VerySlowValidationThreshold = 5 * time.Second

	// Size thresholds
	LargePayloadFieldCount = 50
)

// Error codes for standardized error handling
const (
	ErrCodeValidationFailed = "VALIDATION_FAILED"
	ErrCodeRequiredMissing  = "REQUIRED_FIELD_MISSING"
	ErrCodeValueTooShort    = "VALUE_TOO_SHORT"
	ErrCodeValueTooLong     = "VALUE_TOO_LONG"
	ErrCodeInvalidFormat    = "INVALID_FORMAT"
	ErrCodeInvalidEmail     = "INVALID_EMAIL_FORMAT"
	ErrCodeInvalidURL       = "INVALID_URL_FORMAT"
	ErrCodeInvalidEnum      = "INVALID_ENUM_VALUE"
	ErrCodeSlowValidation   = "SLOW_VALIDATION"
	ErrCodeLargePayload     = "LARGE_PAYLOAD"
)

// IsLargePayload checks if a payload exceeds the size threshold
func IsLargePayload(fieldCount int) bool {
	return fieldCount > LargePayloadFieldCount
}

// IsSlowValidation checks if validation time exceeds threshold
func IsSlowValidation(duration time.Duration) bool {
	return duration > SlowValidationThreshold
}

// IsVerySlowValidation checks if validation time exceeds severe threshold
func IsVerySlowValidation(duration time.Duration) bool {
	return duration > VerySlowValidationThreshold
}
