// Package config provides configuration constants and validation thresholds
package config

import "time"

// Validation Thresholds
const (
	// Performance thresholds
	SlowValidationThreshold = 100 * time.Millisecond
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
)

// IsSlowValidation checks if validation time exceeds threshold
func IsSlowValidation(duration time.Duration) bool {
	return duration > SlowValidationThreshold
}
