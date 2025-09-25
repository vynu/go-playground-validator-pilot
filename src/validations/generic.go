// Package validations contains generic validation logic and business rules.
// This module implements custom validators and business logic for generic payloads.
package validations

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"goplayground-data-validator/models"
)

// GenericValidator provides generic validation functionality.
type GenericValidator struct {
	validator *validator.Validate
}

// NewGenericValidator creates a new generic validator instance.
func NewGenericValidator() *GenericValidator {
	v := validator.New()

	// Register generic custom validators
	v.RegisterValidation("priority_level", validatePriorityLevel)
	v.RegisterValidation("semver", validateSemVer)

	return &GenericValidator{validator: v}
}

// ValidatePayload validates a generic payload with comprehensive rules.
func (gv *GenericValidator) ValidatePayload(payload models.GenericPayload) models.ValidationResult {
	start := time.Now()

	result := models.ValidationResult{
		IsValid:   true,
		ModelType: "GenericPayload",
		Provider:  "generic_validator",
		Timestamp: time.Now(),
		Errors:    []models.ValidationError{},
		Warnings:  []models.ValidationWarning{},
	}

	// Perform struct validation
	if err := gv.validator.Struct(payload); err != nil {
		result.IsValid = false

		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, fieldError := range validationErrors {
				result.Errors = append(result.Errors, models.ValidationError{
					Field:      fieldError.Field(),
					Message:    formatGenericValidationError(fieldError),
					Code:       fieldError.Tag(),
					Value:      fieldError.Value(),
					Expected:   fieldError.Param(),
					Constraint: fieldError.Tag(),
					Path:       fieldError.Namespace(),
					Severity:   "error",
				})
			}
		}
	}

	// Perform business logic validation
	if result.IsValid {
		warnings := ValidateGenericBusinessLogic(payload)
		result.Warnings = append(result.Warnings, warnings...)
	}

	// Add performance metrics
	result.ProcessingDuration = time.Since(start)
	result.PerformanceMetrics = &models.PerformanceMetrics{
		ValidationDuration: time.Since(start),
		Provider:           "generic_validator",
		FieldCount:         countGenericStructFields(payload),
		RuleCount:          gv.getRuleCount(),
	}

	return result
}

// ValidateAPIModel validates an API model with comprehensive rules.
func (gv *GenericValidator) ValidateAPIModel(apiModel models.APIModel) models.ValidationResult {
	start := time.Now()

	result := models.ValidationResult{
		IsValid:   true,
		ModelType: "APIModel",
		Provider:  "generic_validator",
		Timestamp: time.Now(),
		Errors:    []models.ValidationError{},
		Warnings:  []models.ValidationWarning{},
	}

	// Perform struct validation
	if err := gv.validator.Struct(apiModel); err != nil {
		result.IsValid = false

		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, fieldError := range validationErrors {
				result.Errors = append(result.Errors, models.ValidationError{
					Field:      fieldError.Field(),
					Message:    formatGenericValidationError(fieldError),
					Code:       fieldError.Tag(),
					Value:      fieldError.Value(),
					Expected:   fieldError.Param(),
					Constraint: fieldError.Tag(),
					Path:       fieldError.Namespace(),
					Severity:   "error",
				})
			}
		}
	}

	// Perform business logic validation
	if result.IsValid {
		warnings := ValidateAPIModelBusinessLogic(apiModel)
		result.Warnings = append(result.Warnings, warnings...)
	}

	// Add performance metrics
	result.ProcessingDuration = time.Since(start)
	result.PerformanceMetrics = &models.PerformanceMetrics{
		ValidationDuration: time.Since(start),
		Provider:           "generic_validator",
		FieldCount:         countAPIModelFields(apiModel),
		RuleCount:          gv.getRuleCount(),
	}

	return result
}

// validatePriorityLevel validates priority level values.
func validatePriorityLevel(fl validator.FieldLevel) bool {
	priority := fl.Field().String()

	if priority == "" {
		return true // Allow empty for optional fields
	}

	validPriorities := []string{"low", "normal", "high", "urgent", "critical"}
	for _, validPriority := range validPriorities {
		if strings.ToLower(priority) == validPriority {
			return true
		}
	}

	// Also allow numeric priorities (1-5)
	matched, _ := regexp.MatchString(`^[1-5]$`, priority)
	return matched
}

// ValidateGenericBusinessLogic performs generic payload business logic validation.
func ValidateGenericBusinessLogic(payload models.GenericPayload) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	// Check timestamp consistency
	warnings = append(warnings, checkTimestampConsistency(payload)...)

	// Check data integrity
	warnings = append(warnings, checkDataIntegrity(payload)...)

	// Check metadata patterns
	warnings = append(warnings, checkMetadataPatterns(payload)...)

	// Check priority and status consistency
	warnings = append(warnings, checkPriorityStatusConsistency(payload)...)

	// Check tag patterns
	warnings = append(warnings, checkTagPatterns(payload)...)

	return warnings
}

// ValidateAPIModelBusinessLogic performs API model business logic validation.
func ValidateAPIModelBusinessLogic(apiModel models.APIModel) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	// Check HTTP method and URL consistency
	warnings = append(warnings, checkHTTPMethodURLConsistency(apiModel)...)

	// Check status code patterns
	warnings = append(warnings, checkStatusCodePatterns(apiModel)...)

	// Check security patterns
	warnings = append(warnings, checkAPISecurityPatterns(apiModel)...)

	// Check performance patterns
	warnings = append(warnings, checkAPIPerformancePatterns(apiModel)...)

	return warnings
}

// checkTimestampConsistency checks for timestamp-related inconsistencies.
func checkTimestampConsistency(payload models.GenericPayload) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	// Check if timestamp is in the future
	if payload.Timestamp.After(time.Now().Add(5 * time.Minute)) {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "Timestamp",
			Message:    fmt.Sprintf("Timestamp is in the future: %v", payload.Timestamp),
			Code:       "FUTURE_TIMESTAMP",
			Value:      payload.Timestamp,
			Suggestion: "Verify system clock synchronization",
			Category:   "temporal",
		})
	}

	// Check if timestamp is very old
	if time.Since(payload.Timestamp).Hours() > 24*365 { // More than a year old
		warnings = append(warnings, models.ValidationWarning{
			Field:      "Timestamp",
			Message:    fmt.Sprintf("Very old timestamp: %v", payload.Timestamp),
			Code:       "OLD_TIMESTAMP",
			Value:      payload.Timestamp,
			Suggestion: "Verify if this is historical data or a timestamp error",
			Category:   "temporal",
		})
	}

	return warnings
}

// checkDataIntegrity checks for data integrity issues.
func checkDataIntegrity(payload models.GenericPayload) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	// Check if checksum is provided but data is empty
	if payload.Checksum != "" && len(payload.Data) == 0 {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "Checksum",
			Message:    "Checksum provided but data is empty",
			Code:       "CHECKSUM_WITHOUT_DATA",
			Suggestion: "Remove checksum or add data content",
			Category:   "integrity",
		})
	}

	// Check for very large data payloads
	if len(payload.Data) > 100 {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "Data",
			Message:    fmt.Sprintf("Large data payload: %d fields", len(payload.Data)),
			Code:       "LARGE_DATA_PAYLOAD",
			Value:      len(payload.Data),
			Suggestion: "Consider pagination or data chunking for large payloads",
			Category:   "performance",
		})
	}

	// Check for empty required-looking data
	if len(payload.Data) == 0 && payload.Type != "heartbeat" && payload.Type != "ping" {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "Data",
			Message:    "Data field is empty for non-heartbeat payload",
			Code:       "EMPTY_DATA",
			Suggestion: "Verify if payload should contain data",
			Category:   "completeness",
		})
	}

	return warnings
}

// checkMetadataPatterns checks for metadata-related patterns.
func checkMetadataPatterns(payload models.GenericPayload) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	// Check for sensitive information in metadata
	for key, value := range payload.Metadata {
		keyLower := strings.ToLower(key)
		valueLower := strings.ToLower(value)

		sensitiveKeys := []string{"password", "secret", "token", "key", "credential"}
		for _, sensitiveKey := range sensitiveKeys {
			if strings.Contains(keyLower, sensitiveKey) || strings.Contains(valueLower, sensitiveKey) {
				warnings = append(warnings, models.ValidationWarning{
					Field:      "Metadata",
					Message:    fmt.Sprintf("Potentially sensitive information in metadata: %s", key),
					Code:       "SENSITIVE_METADATA",
					Value:      key,
					Suggestion: "Avoid storing sensitive information in metadata",
					Category:   "security",
				})
				break
			}
		}
	}

	// Check for excessive metadata
	if len(payload.Metadata) > 50 {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "Metadata",
			Message:    fmt.Sprintf("Large number of metadata fields: %d", len(payload.Metadata)),
			Code:       "EXCESSIVE_METADATA",
			Value:      len(payload.Metadata),
			Suggestion: "Consider consolidating or restructuring metadata",
			Category:   "maintainability",
		})
	}

	return warnings
}

// checkPriorityStatusConsistency checks for priority and status consistency.
func checkPriorityStatusConsistency(payload models.GenericPayload) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	// Check priority-status combinations
	if payload.Priority == "critical" && payload.Status == "pending" {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "Priority",
			Message:    "Critical priority item still pending",
			Code:       "CRITICAL_PENDING",
			Suggestion: "Critical items should be processed immediately",
			Category:   "workflow",
		})
	}

	if payload.Priority == "low" && payload.Status == "failed" {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "Status",
			Message:    "Low priority item has failed status",
			Code:       "LOW_PRIORITY_FAILED",
			Suggestion: "Review if failed low priority items need attention",
			Category:   "workflow",
		})
	}

	// Check for completed items with high priority
	if payload.Status == "completed" && (payload.Priority == "urgent" || payload.Priority == "critical") {
		// This is actually good, but we can note it for metrics
		warnings = append(warnings, models.ValidationWarning{
			Field:      "Status",
			Message:    "High priority item completed successfully",
			Code:       "HIGH_PRIORITY_COMPLETED",
			Suggestion: "Good - high priority items should be completed quickly",
			Category:   "workflow",
		})
	}

	return warnings
}

// checkTagPatterns checks for tag-related patterns.
func checkTagPatterns(payload models.GenericPayload) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	// Check for duplicate tags
	tagMap := make(map[string]bool)
	duplicates := []string{}
	for _, tag := range payload.Tags {
		if tagMap[tag] {
			duplicates = append(duplicates, tag)
		}
		tagMap[tag] = true
	}

	if len(duplicates) > 0 {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "Tags",
			Message:    fmt.Sprintf("Duplicate tags found: %s", strings.Join(duplicates, ", ")),
			Code:       "DUPLICATE_TAGS",
			Value:      duplicates,
			Suggestion: "Remove duplicate tags",
			Category:   "data-quality",
		})
	}

	// Check for excessive number of tags
	if len(payload.Tags) > 20 {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "Tags",
			Message:    fmt.Sprintf("Large number of tags: %d", len(payload.Tags)),
			Code:       "EXCESSIVE_TAGS",
			Value:      len(payload.Tags),
			Suggestion: "Consider using fewer, more specific tags",
			Category:   "maintainability",
		})
	}

	// Check for tag naming patterns
	for _, tag := range payload.Tags {
		if strings.Contains(tag, " ") {
			warnings = append(warnings, models.ValidationWarning{
				Field:      "Tags",
				Message:    fmt.Sprintf("Tag contains spaces: '%s'", tag),
				Code:       "SPACED_TAG",
				Value:      tag,
				Suggestion: "Use hyphens or underscores instead of spaces in tags",
				Category:   "data-quality",
			})
		}

		if len(tag) > 30 {
			warnings = append(warnings, models.ValidationWarning{
				Field:      "Tags",
				Message:    fmt.Sprintf("Very long tag: '%s'", tag),
				Code:       "LONG_TAG",
				Value:      tag,
				Suggestion: "Use shorter, more concise tags",
				Category:   "maintainability",
			})
		}
	}

	return warnings
}

// checkHTTPMethodURLConsistency checks for HTTP method and URL consistency.
func checkHTTPMethodURLConsistency(apiModel models.APIModel) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	// Check for GET requests with body
	if apiModel.Method == "GET" && apiModel.Body != nil {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "Body",
			Message:    "GET request contains body",
			Code:       "GET_WITH_BODY",
			Suggestion: "GET requests should not contain request body",
			Category:   "http-semantics",
		})
	}

	// Check for POST/PUT requests without body
	if (apiModel.Method == "POST" || apiModel.Method == "PUT") && apiModel.Body == nil {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "Body",
			Message:    fmt.Sprintf("%s request missing body", apiModel.Method),
			Code:       "MISSING_BODY",
			Suggestion: "POST/PUT requests typically require a request body",
			Category:   "http-semantics",
		})
	}

	// Check for DELETE with body
	if apiModel.Method == "DELETE" && apiModel.Body != nil {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "Body",
			Message:    "DELETE request contains body",
			Code:       "DELETE_WITH_BODY",
			Suggestion: "DELETE requests typically should not contain body",
			Category:   "http-semantics",
		})
	}

	return warnings
}

// checkStatusCodePatterns checks for status code patterns.
func checkStatusCodePatterns(apiModel models.APIModel) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	// Check for method-status code consistency
	switch apiModel.Method {
	case "POST":
		if apiModel.StatusCode == 200 {
			warnings = append(warnings, models.ValidationWarning{
				Field:      "StatusCode",
				Message:    "POST request returned 200 instead of 201",
				Code:       "POST_WRONG_STATUS",
				Value:      apiModel.StatusCode,
				Suggestion: "POST requests should typically return 201 for created resources",
				Category:   "http-semantics",
			})
		}
	case "PUT":
		if apiModel.StatusCode == 200 {
			// This could be either update (200) or create (201), so just note it
			warnings = append(warnings, models.ValidationWarning{
				Field:      "StatusCode",
				Message:    "PUT request returned 200 (update) vs 201 (create)",
				Code:       "PUT_STATUS_AMBIGUOUS",
				Value:      apiModel.StatusCode,
				Suggestion: "Consider using 201 for creation, 200 for updates",
				Category:   "http-semantics",
			})
		}
	case "DELETE":
		if apiModel.StatusCode == 200 && apiModel.Response != nil {
			warnings = append(warnings, models.ValidationWarning{
				Field:      "Response",
				Message:    "DELETE request with response body",
				Code:       "DELETE_WITH_RESPONSE",
				Suggestion: "Consider using 204 No Content for DELETE operations",
				Category:   "http-semantics",
			})
		}
	}

	return warnings
}

// checkAPISecurityPatterns checks for API security patterns.
func checkAPISecurityPatterns(apiModel models.APIModel) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	// Check for insecure HTTP
	if strings.HasPrefix(strings.ToLower(apiModel.URL), "http://") {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "URL",
			Message:    "API request using insecure HTTP protocol",
			Code:       "INSECURE_HTTP",
			Value:      apiModel.URL,
			Suggestion: "Use HTTPS for API requests",
			Category:   "security",
		})
	}

	// Check for sensitive data in query parameters
	for paramName := range apiModel.Parameters {
		paramLower := strings.ToLower(paramName)
		if strings.Contains(paramLower, "password") ||
			strings.Contains(paramLower, "secret") ||
			strings.Contains(paramLower, "token") {
			warnings = append(warnings, models.ValidationWarning{
				Field:      "Parameters",
				Message:    fmt.Sprintf("Potentially sensitive parameter: %s", paramName),
				Code:       "SENSITIVE_QUERY_PARAM",
				Value:      paramName,
				Suggestion: "Move sensitive data to request body or headers",
				Category:   "security",
			})
		}
	}

	return warnings
}

// checkAPIPerformancePatterns checks for API performance patterns.
func checkAPIPerformancePatterns(apiModel models.APIModel) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	// Check for slow API responses
	if apiModel.Duration > 5*time.Second {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "Duration",
			Message:    fmt.Sprintf("Slow API response: %v", apiModel.Duration),
			Code:       "SLOW_API_RESPONSE",
			Value:      apiModel.Duration,
			Suggestion: "Optimize API performance or implement caching",
			Category:   "performance",
		})
	}

	// Check for very large parameter sets
	if len(apiModel.Parameters) > 20 {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "Parameters",
			Message:    fmt.Sprintf("Large number of parameters: %d", len(apiModel.Parameters)),
			Code:       "LARGE_PARAMETER_SET",
			Value:      len(apiModel.Parameters),
			Suggestion: "Consider using request body for complex data",
			Category:   "performance",
		})
	}

	return warnings
}

// formatGenericValidationError formats validation errors with generic context.
func formatGenericValidationError(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("Field '%s' is required", fe.Field())
	case "priority_level":
		return fmt.Sprintf("Field '%s' must be a valid priority level (low, normal, high, urgent, critical, or 1-5)", fe.Field())
	case "semver":
		return fmt.Sprintf("Field '%s' must be a valid semantic version", fe.Field())
	case "url":
		return fmt.Sprintf("Field '%s' must be a valid URL format", fe.Field())
	case "email":
		return fmt.Sprintf("Field '%s' must be a valid email address", fe.Field())
	case "ip":
		return fmt.Sprintf("Field '%s' must be a valid IP address", fe.Field())
	case "oneof":
		return fmt.Sprintf("Field '%s' must be one of: %s", fe.Field(), fe.Param())
	case "min":
		return fmt.Sprintf("Field '%s' must be at least %s characters long", fe.Field(), fe.Param())
	case "max":
		return fmt.Sprintf("Field '%s' must be at most %s characters long", fe.Field(), fe.Param())
	case "gt":
		return fmt.Sprintf("Field '%s' must be greater than %s", fe.Field(), fe.Param())
	case "gte":
		return fmt.Sprintf("Field '%s' must be greater than or equal to %s", fe.Field(), fe.Param())
	case "lte":
		return fmt.Sprintf("Field '%s' must be less than or equal to %s", fe.Field(), fe.Param())
	case "len":
		return fmt.Sprintf("Field '%s' must be exactly %s characters long", fe.Field(), fe.Param())
	case "hexadecimal":
		return fmt.Sprintf("Field '%s' must contain only hexadecimal characters", fe.Field())
	default:
		return fmt.Sprintf("Field '%s' failed validation: %s", fe.Field(), fe.Tag())
	}
}

// countGenericStructFields counts the number of fields in a generic struct for metrics.
func countGenericStructFields(payload models.GenericPayload) int {
	count := 10 // Base fields
	count += len(payload.Data)
	count += len(payload.Metadata)
	count += len(payload.Tags)
	return count
}

// countAPIModelFields counts the number of fields in an API model for metrics.
func countAPIModelFields(apiModel models.APIModel) int {
	count := 12 // Base fields
	count += len(apiModel.Headers)
	count += len(apiModel.Parameters)
	return count
}

// getRuleCount returns the number of validation rules applied.
func (gv *GenericValidator) getRuleCount() int {
	// Return approximate number of validation rules
	return 20
}
