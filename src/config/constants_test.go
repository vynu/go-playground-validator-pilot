package config

import (
	"testing"
	"time"
)

func TestIsSlowValidation(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected bool
	}{
		{
			name:     "fast validation - 50ms",
			duration: 50 * time.Millisecond,
			expected: false,
		},
		{
			name:     "exact threshold - 100ms",
			duration: 100 * time.Millisecond,
			expected: false,
		},
		{
			name:     "slow validation - 150ms",
			duration: 150 * time.Millisecond,
			expected: true,
		},
		{
			name:     "very slow validation - 1s",
			duration: 1 * time.Second,
			expected: true,
		},
		{
			name:     "instant validation - 0ms",
			duration: 0,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsSlowValidation(tt.duration)
			if result != tt.expected {
				t.Errorf("IsSlowValidation(%v) = %v, expected %v", tt.duration, result, tt.expected)
			}
		})
	}
}

func TestConstants(t *testing.T) {
	t.Run("validation threshold constant", func(t *testing.T) {
		if SlowValidationThreshold != 100*time.Millisecond {
			t.Errorf("Expected SlowValidationThreshold to be 100ms, got %v", SlowValidationThreshold)
		}
	})

	t.Run("error code constants", func(t *testing.T) {
		errorCodes := map[string]string{
			"VALIDATION_FAILED":      ErrCodeValidationFailed,
			"REQUIRED_FIELD_MISSING": ErrCodeRequiredMissing,
			"VALUE_TOO_SHORT":        ErrCodeValueTooShort,
			"VALUE_TOO_LONG":         ErrCodeValueTooLong,
			"INVALID_FORMAT":         ErrCodeInvalidFormat,
			"INVALID_EMAIL_FORMAT":   ErrCodeInvalidEmail,
			"INVALID_URL_FORMAT":     ErrCodeInvalidURL,
			"INVALID_ENUM_VALUE":     ErrCodeInvalidEnum,
		}

		for expected, actual := range errorCodes {
			if expected != actual {
				t.Errorf("Expected error code %s, got %s", expected, actual)
			}
		}
	})
}
