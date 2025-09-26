// Package config provides configuration constants and validation thresholds
package config

import "time"

// Validation Thresholds
const (
	// Performance thresholds
	SlowValidationThreshold     = 100 * time.Millisecond
	VerySlowValidationThreshold = 5 * time.Second

	// Size thresholds
	LargePayloadFieldCount  = 50
	LargeChangesetThreshold = 1000
	MaxFileChangeThreshold  = 50
	LargeResponseThreshold  = 10 * 1024 * 1024 // 10MB

	// Business logic thresholds
	MaxSeverityThreshold     = 5000
	MinimumDescriptionLength = 20
	MaximumTitleLength       = 200
	MaximumDescriptionLength = 1000
)

// String validation constants
const (
	MinEmailLength    = 5
	MaxEmailLength    = 254
	MinUsernameLength = 3
	MaxUsernameLength = 50
	MinPasswordLength = 8
	MaxPasswordLength = 128
	MinTitleLength    = 10
	MaxTitleLength    = 200
	MinURLLength      = 10
	MaxURLLength      = 2048
)

// HTTP and API constants
const (
	DefaultTimeout      = 30 * time.Second
	MaxRequestSize      = 100 * 1024 * 1024 // 100MB
	DefaultPort         = 8080
	HealthCheckInterval = 30 * time.Second
)

// Registry and discovery constants
const (
	MaxModelsToDiscover        = 100
	ModelDiscoveryTimeout      = 10 * time.Second
	ValidatorConstructorPrefix = "New"
	ValidatorConstructorSuffix = "Validator"
	ModelStructSuffix          = "Payload"
)

// Security constants
const (
	MinRandomStringLength = 16
	MaxRandomStringLength = 64
	RandomStringCharset   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	SessionTokenLength    = 32
	APIKeyLength          = 64
)

// Logging and monitoring constants
const (
	LogTimestampFormat        = "2006-01-02 15:04:05"
	MetricsCollectionInterval = 1 * time.Minute
	MaxLogFileSize            = 100 * 1024 * 1024 // 100MB
	MaxLogFiles               = 5
)

// Default validation messages
const (
	DefaultRequiredMessage  = "This field is required"
	DefaultMinLengthMessage = "This field is too short"
	DefaultMaxLengthMessage = "This field is too long"
	DefaultEmailMessage     = "Invalid email format"
	DefaultURLMessage       = "Invalid URL format"
	DefaultNumericMessage   = "This field must be numeric"
)

// Model-specific constants
const (
	// GitHub webhook constants
	GitHubMaxCommitCount      = 20
	GitHubMaxBranchNameLength = 255
	GitHubMinRepoNameLength   = 1
	GitHubMaxRepoNameLength   = 100

	// API request constants
	APIMaxEndpointLength = 255
	APIMaxHeaderCount    = 50
	APIMaxParameterCount = 100

	// Database query constants
	DBMaxQueryLength     = 10000
	DBMaxTableNameLength = 64
	DBMaxColumnCount     = 1000

	// Incident report constants
	IncidentMinTitleLength = 10
	IncidentMaxTitleLength = 200
	IncidentMinDescLength  = 20
	IncidentMaxDescLength  = 1000
	IncidentMaxTagCount    = 10
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

// Warning codes for business logic warnings
const (
	WarnCodePerformance     = "PERFORMANCE_WARNING"
	WarnCodeBusinessRule    = "BUSINESS_RULE_WARNING"
	WarnCodeSecurityConcern = "SECURITY_CONCERN"
	WarnCodeBestPractice    = "BEST_PRACTICE_VIOLATION"
	WarnCodeDeprecated      = "DEPRECATED_USAGE"
)

// Feature flags for enabling/disabling validation features
var (
	EnablePerformanceWarnings = true
	EnableBusinessLogic       = true
	EnableSecurityValidation  = true
	EnableDetailedLogging     = false
	EnableMetricsCollection   = false
)

// GetValidationTimeout returns the appropriate timeout based on payload complexity
func GetValidationTimeout(fieldCount int) time.Duration {
	if fieldCount > LargePayloadFieldCount {
		return VerySlowValidationThreshold * 2
	}
	return VerySlowValidationThreshold
}

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
