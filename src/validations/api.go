// Package validations contains API-specific validation logic and business rules.
// This module implements custom validators and business logic for API requests and responses.
package validations

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github-data-validator/models"
	"github.com/go-playground/validator/v10"
)

// APIValidator provides API-specific validation functionality.
type APIValidator struct {
	validator *validator.Validate
}

// NewAPIValidator creates a new API validator instance.
func NewAPIValidator() *APIValidator {
	v := validator.New()

	// Register API-specific custom validators
	v.RegisterValidation("api_content_type", validateAPIContentType)
	v.RegisterValidation("api_version", validateAPIVersion)

	return &APIValidator{validator: v}
}

// ValidateRequest validates an API request with comprehensive rules.
func (av *APIValidator) ValidateRequest(request models.APIRequest) models.ValidationResult {
	start := time.Now()

	result := models.ValidationResult{
		IsValid:   true,
		ModelType: "APIRequest",
		Provider:  "api_validator",
		Timestamp: time.Now(),
		Errors:    []models.ValidationError{},
		Warnings:  []models.ValidationWarning{},
	}

	// Perform struct validation
	if err := av.validator.Struct(request); err != nil {
		result.IsValid = false

		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, fieldError := range validationErrors {
				result.Errors = append(result.Errors, models.ValidationError{
					Field:      fieldError.Field(),
					Message:    formatAPIValidationError(fieldError),
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
		warnings := ValidateAPIRequestBusinessLogic(request)
		result.Warnings = append(result.Warnings, warnings...)
	}

	// Add performance metrics
	result.ProcessingDuration = time.Since(start)
	result.PerformanceMetrics = &models.PerformanceMetrics{
		ValidationDuration: time.Since(start),
		Provider:           "api_validator",
		FieldCount:         countAPIRequestFields(request),
		RuleCount:          av.getRuleCount(),
	}

	return result
}

// ValidateResponse validates an API response with comprehensive rules.
func (av *APIValidator) ValidateResponse(response models.APIResponse) models.ValidationResult {
	start := time.Now()

	result := models.ValidationResult{
		IsValid:   true,
		ModelType: "APIResponse",
		Provider:  "api_validator",
		Timestamp: time.Now(),
		Errors:    []models.ValidationError{},
		Warnings:  []models.ValidationWarning{},
	}

	// Perform struct validation
	if err := av.validator.Struct(response); err != nil {
		result.IsValid = false

		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, fieldError := range validationErrors {
				result.Errors = append(result.Errors, models.ValidationError{
					Field:      fieldError.Field(),
					Message:    formatAPIValidationError(fieldError),
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
		warnings := ValidateAPIResponseBusinessLogic(response)
		result.Warnings = append(result.Warnings, warnings...)
	}

	// Add performance metrics
	result.ProcessingDuration = time.Since(start)
	result.PerformanceMetrics = &models.PerformanceMetrics{
		ValidationDuration: time.Since(start),
		Provider:           "api_validator",
		FieldCount:         countAPIResponseFields(response),
		RuleCount:          av.getRuleCount(),
	}

	return result
}

// validateAPIContentType validates API content type format.
func validateAPIContentType(fl validator.FieldLevel) bool {
	contentType := fl.Field().String()

	if contentType == "" {
		return true // Allow empty for optional fields
	}

	// Common API content types
	validTypes := []string{
		"application/json",
		"application/xml",
		"application/x-www-form-urlencoded",
		"multipart/form-data",
		"text/plain",
		"text/html",
		"text/csv",
		"application/pdf",
		"application/octet-stream",
		"image/jpeg",
		"image/png",
		"image/gif",
		"video/mp4",
		"audio/mpeg",
	}

	// Check for exact matches first
	for _, validType := range validTypes {
		if contentType == validType {
			return true
		}
	}

	// Check for content type with charset or other parameters
	if strings.Contains(contentType, ";") {
		mainType := strings.Split(contentType, ";")[0]
		mainType = strings.TrimSpace(mainType)
		for _, validType := range validTypes {
			if mainType == validType {
				return true
			}
		}
	}

	// Validate format using regex for custom types
	matched, _ := regexp.MatchString(`^[a-zA-Z][a-zA-Z0-9][a-zA-Z0-9\!\#\$\&\-\^]*\/[a-zA-Z0-9][a-zA-Z0-9\!\#\$\&\-\^]*$`, contentType)
	return matched
}

// validateAPIVersion validates API version format.
func validateAPIVersion(fl validator.FieldLevel) bool {
	version := fl.Field().String()

	if version == "" {
		return true // Allow empty for optional fields
	}

	// Support various version formats:
	// - Semantic versioning: v1.2.3, 1.2.3
	// - Simple versioning: v1, v2, 1, 2
	// - Date-based: 2023-01-01, 20230101
	patterns := []string{
		`^v?[0-9]+$`,                                  // v1, 1
		`^v?[0-9]+\.[0-9]+$`,                          // v1.2, 1.2
		`^v?[0-9]+\.[0-9]+\.[0-9]+$`,                  // v1.2.3, 1.2.3
		`^v?[0-9]+\.[0-9]+\.[0-9]+-[a-zA-Z0-9\-\.]+$`, // v1.2.3-alpha.1
		`^[0-9]{4}-[0-9]{2}-[0-9]{2}$`,                // 2023-01-01
		`^[0-9]{8}$`,                                  // 20230101
	}

	for _, pattern := range patterns {
		matched, _ := regexp.MatchString(pattern, version)
		if matched {
			return true
		}
	}

	return false
}

// ValidateAPIRequestBusinessLogic performs API request-specific business logic validation.
func ValidateAPIRequestBusinessLogic(request models.APIRequest) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	// Check for security concerns
	warnings = append(warnings, checkAPIRequestSecurity(request)...)

	// Check for performance concerns
	warnings = append(warnings, checkAPIRequestPerformance(request)...)

	// Check for best practices
	warnings = append(warnings, checkAPIRequestBestPractices(request)...)

	// Check for rate limiting concerns
	warnings = append(warnings, checkAPIRequestRateLimit(request)...)

	return warnings
}

// ValidateAPIResponseBusinessLogic performs API response-specific business logic validation.
func ValidateAPIResponseBusinessLogic(response models.APIResponse) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	// Check for status code patterns
	warnings = append(warnings, checkAPIResponseStatus(response)...)

	// Check for performance concerns
	warnings = append(warnings, checkAPIResponsePerformance(response)...)

	// Check for security headers
	warnings = append(warnings, checkAPIResponseSecurity(response)...)

	// Check for content patterns
	warnings = append(warnings, checkAPIResponseContent(response)...)

	return warnings
}

// checkAPIRequestSecurity checks for security-related concerns in API requests.
func checkAPIRequestSecurity(request models.APIRequest) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	// Check for missing authorization
	if request.Authorization == nil && request.Source != "test" {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "Authorization",
			Message:    "API request missing authorization header",
			Code:       "MISSING_AUTHORIZATION",
			Suggestion: "Include appropriate authorization for authenticated endpoints",
			Category:   "security",
		})
	}

	// Check for insecure HTTP in production
	if strings.HasPrefix(strings.ToLower(request.URL), "http://") && request.Source != "test" {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "URL",
			Message:    "API request using insecure HTTP protocol",
			Code:       "INSECURE_PROTOCOL",
			Value:      request.URL,
			Suggestion: "Use HTTPS for production API requests",
			Category:   "security",
		})
	}

	// Check for sensitive data in query parameters
	for param := range request.QueryParams {
		paramLower := strings.ToLower(param)
		if strings.Contains(paramLower, "password") ||
			strings.Contains(paramLower, "secret") ||
			strings.Contains(paramLower, "token") ||
			strings.Contains(paramLower, "key") {
			warnings = append(warnings, models.ValidationWarning{
				Field:      "QueryParams",
				Message:    fmt.Sprintf("Potentially sensitive parameter in query string: %s", param),
				Code:       "SENSITIVE_QUERY_PARAM",
				Value:      param,
				Suggestion: "Move sensitive data to request body or headers",
				Category:   "security",
			})
		}
	}

	// Check user agent for bot patterns
	userAgent := strings.ToLower(request.UserAgent)
	botPatterns := []string{"bot", "crawler", "spider", "scraper", "scanner"}
	for _, pattern := range botPatterns {
		if strings.Contains(userAgent, pattern) {
			warnings = append(warnings, models.ValidationWarning{
				Field:      "UserAgent",
				Message:    fmt.Sprintf("Bot-like user agent detected: %s", pattern),
				Code:       "BOT_USER_AGENT",
				Value:      request.UserAgent,
				Suggestion: "Implement appropriate bot handling and rate limiting",
				Category:   "security",
			})
			break
		}
	}

	return warnings
}

// checkAPIRequestPerformance checks for performance-related concerns.
func checkAPIRequestPerformance(request models.APIRequest) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	// Check for excessive timeout
	if request.Timeout != nil && *request.Timeout > 60 {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "Timeout",
			Message:    fmt.Sprintf("Very high timeout value: %d seconds", *request.Timeout),
			Code:       "HIGH_TIMEOUT",
			Value:      *request.Timeout,
			Suggestion: "Consider optimizing request or using asynchronous processing",
			Category:   "performance",
		})
	}

	// Check for excessive retry count
	if request.RetryCount > 5 {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "RetryCount",
			Message:    fmt.Sprintf("High retry count: %d", request.RetryCount),
			Code:       "HIGH_RETRY_COUNT",
			Value:      request.RetryCount,
			Suggestion: "Consider exponential backoff and circuit breaker patterns",
			Category:   "performance",
		})
	}

	// Check for large query parameter sets
	if len(request.QueryParams) > 20 {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "QueryParams",
			Message:    fmt.Sprintf("Large number of query parameters: %d", len(request.QueryParams)),
			Code:       "LARGE_QUERY_PARAMS",
			Value:      len(request.QueryParams),
			Suggestion: "Consider using request body for complex data",
			Category:   "performance",
		})
	}

	return warnings
}

// checkAPIRequestBestPractices checks for API best practice violations.
func checkAPIRequestBestPractices(request models.APIRequest) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	// Check for missing request ID
	if request.RequestID == "" {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "RequestID",
			Message:    "Missing request ID for tracing",
			Code:       "MISSING_REQUEST_ID",
			Suggestion: "Include unique request ID for debugging and tracing",
			Category:   "observability",
		})
	}

	// Check for missing content type on POST/PUT requests
	if (request.Method == "POST" || request.Method == "PUT" || request.Method == "PATCH") &&
		request.ContentType == "" {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "ContentType",
			Message:    "Missing content type for request with body",
			Code:       "MISSING_CONTENT_TYPE",
			Suggestion: "Specify appropriate content type for requests with body",
			Category:   "api-design",
		})
	}

	// Check for non-idempotent methods without appropriate headers
	if request.Method == "PUT" || request.Method == "DELETE" {
		if _, hasIfMatch := request.Headers["If-Match"]; !hasIfMatch {
			if _, hasIfUnmodified := request.Headers["If-Unmodified-Since"]; !hasIfUnmodified {
				warnings = append(warnings, models.ValidationWarning{
					Field:      "Headers",
					Message:    "Idempotent operation missing conditional headers",
					Code:       "MISSING_CONDITIONAL_HEADERS",
					Suggestion: "Include If-Match or If-Unmodified-Since for safe idempotent operations",
					Category:   "api-design",
				})
			}
		}
	}

	return warnings
}

// checkAPIRequestRateLimit checks for rate limiting concerns.
func checkAPIRequestRateLimit(request models.APIRequest) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	// Check if rate limit info is available
	if request.RateLimit != nil {
		rl := *request.RateLimit

		// Check if approaching rate limit
		if float64(rl.Remaining)/float64(rl.Limit) < 0.1 {
			warnings = append(warnings, models.ValidationWarning{
				Field:      "RateLimit.Remaining",
				Message:    fmt.Sprintf("Approaching rate limit: %d/%d remaining", rl.Remaining, rl.Limit),
				Code:       "RATE_LIMIT_WARNING",
				Value:      rl.Remaining,
				Suggestion: "Implement rate limiting handling and consider request throttling",
				Category:   "rate-limiting",
			})
		}

		// Check for very short reset window with high usage
		resetDuration := time.Until(rl.Reset)
		if resetDuration < time.Minute && float64(rl.Remaining)/float64(rl.Limit) < 0.5 {
			warnings = append(warnings, models.ValidationWarning{
				Field:      "RateLimit.Reset",
				Message:    "High rate limit usage with short reset window",
				Code:       "HIGH_RATE_USAGE",
				Suggestion: "Consider implementing request queuing or backoff strategies",
				Category:   "rate-limiting",
			})
		}
	}

	return warnings
}

// checkAPIResponseStatus checks for status code patterns.
func checkAPIResponseStatus(response models.APIResponse) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	// Check for client error responses
	if response.StatusCode >= 400 && response.StatusCode < 500 {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "StatusCode",
			Message:    fmt.Sprintf("Client error response: %d", response.StatusCode),
			Code:       "CLIENT_ERROR",
			Value:      response.StatusCode,
			Suggestion: "Review request parameters and authentication",
			Category:   "http-status",
		})
	}

	// Check for server error responses
	if response.StatusCode >= 500 {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "StatusCode",
			Message:    fmt.Sprintf("Server error response: %d", response.StatusCode),
			Code:       "SERVER_ERROR",
			Value:      response.StatusCode,
			Suggestion: "Check server health and implement retry logic",
			Category:   "http-status",
		})
	}

	// Check for redirect responses without location
	if response.StatusCode >= 300 && response.StatusCode < 400 {
		if _, hasLocation := response.Headers["Location"]; !hasLocation {
			warnings = append(warnings, models.ValidationWarning{
				Field:      "Headers",
				Message:    "Redirect response missing Location header",
				Code:       "MISSING_LOCATION",
				Value:      response.StatusCode,
				Suggestion: "Include Location header for redirect responses",
				Category:   "http-headers",
			})
		}
	}

	return warnings
}

// checkAPIResponsePerformance checks for performance-related concerns in responses.
func checkAPIResponsePerformance(response models.APIResponse) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	// Check for slow response times
	if response.Duration > 5*time.Second {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "Duration",
			Message:    fmt.Sprintf("Slow response time: %v", response.Duration),
			Code:       "SLOW_RESPONSE",
			Value:      response.Duration,
			Suggestion: "Optimize API performance or consider caching",
			Category:   "performance",
		})
	}

	// Check for large response bodies
	if response.ContentLength != nil && *response.ContentLength > 10*1024*1024 { // 10MB
		warnings = append(warnings, models.ValidationWarning{
			Field:      "ContentLength",
			Message:    fmt.Sprintf("Large response body: %d bytes", *response.ContentLength),
			Code:       "LARGE_RESPONSE",
			Value:      *response.ContentLength,
			Suggestion: "Consider pagination or response compression",
			Category:   "performance",
		})
	}

	// Check for missing cache headers on cacheable responses
	if response.StatusCode == 200 {
		if _, hasCacheControl := response.Headers["Cache-Control"]; !hasCacheControl {
			if _, hasETag := response.Headers["ETag"]; !hasETag {
				warnings = append(warnings, models.ValidationWarning{
					Field:      "Headers",
					Message:    "Cacheable response missing cache headers",
					Code:       "MISSING_CACHE_HEADERS",
					Suggestion: "Include Cache-Control, ETag, or Last-Modified headers",
					Category:   "performance",
				})
			}
		}
	}

	return warnings
}

// checkAPIResponseSecurity checks for security-related concerns in responses.
func checkAPIResponseSecurity(response models.APIResponse) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	// Check for missing security headers
	securityHeaders := map[string]string{
		"X-Content-Type-Options":    "nosniff",
		"X-Frame-Options":           "DENY or SAMEORIGIN",
		"X-XSS-Protection":          "1; mode=block",
		"Strict-Transport-Security": "max-age=31536000",
	}

	missingHeaders := []string{}
	for header := range securityHeaders {
		if _, exists := response.Headers[header]; !exists {
			missingHeaders = append(missingHeaders, header)
		}
	}

	if len(missingHeaders) > 0 {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "Headers",
			Message:    fmt.Sprintf("Missing security headers: %s", strings.Join(missingHeaders, ", ")),
			Code:       "MISSING_SECURITY_HEADERS",
			Suggestion: "Include security headers to protect against common attacks",
			Category:   "security",
		})
	}

	// Check for sensitive data exposure in error responses
	if response.StatusCode >= 400 && len(response.Errors) > 0 {
		for _, apiError := range response.Errors {
			errorDetail := strings.ToLower(apiError.Detail)
			if strings.Contains(errorDetail, "password") ||
				strings.Contains(errorDetail, "token") ||
				strings.Contains(errorDetail, "secret") ||
				strings.Contains(errorDetail, "key") {
				warnings = append(warnings, models.ValidationWarning{
					Field:      "Errors.Detail",
					Message:    "Error message may contain sensitive information",
					Code:       "SENSITIVE_ERROR_DETAIL",
					Suggestion: "Sanitize error messages to avoid exposing sensitive data",
					Category:   "security",
				})
				break
			}
		}
	}

	return warnings
}

// checkAPIResponseContent checks for content-related patterns.
func checkAPIResponseContent(response models.APIResponse) []models.ValidationWarning {
	var warnings []models.ValidationWarning

	// Check for missing content type
	if response.ContentType == "" && response.StatusCode == 200 {
		warnings = append(warnings, models.ValidationWarning{
			Field:      "ContentType",
			Message:    "Response missing content type",
			Code:       "MISSING_CONTENT_TYPE",
			Suggestion: "Include appropriate content type header",
			Category:   "http-headers",
		})
	}

	// Check pagination completeness
	if response.Pagination != nil {
		pagination := *response.Pagination

		// Check for missing pagination links
		if pagination.HasNext && (pagination.Links == nil || pagination.Links.Next == nil) {
			warnings = append(warnings, models.ValidationWarning{
				Field:      "Pagination.Links.Next",
				Message:    "Pagination indicates next page but missing next link",
				Code:       "MISSING_PAGINATION_LINK",
				Suggestion: "Include pagination links for better API navigation",
				Category:   "api-design",
			})
		}

		// Check for very large page sizes
		if pagination.PerPage > 1000 {
			warnings = append(warnings, models.ValidationWarning{
				Field:      "Pagination.PerPage",
				Message:    fmt.Sprintf("Very large page size: %d", pagination.PerPage),
				Code:       "LARGE_PAGE_SIZE",
				Value:      pagination.PerPage,
				Suggestion: "Consider smaller page sizes for better performance",
				Category:   "performance",
			})
		}
	}

	return warnings
}

// formatAPIValidationError formats validation errors with API-specific context.
func formatAPIValidationError(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("Field '%s' is required for API validation", fe.Field())
	case "api_content_type":
		return fmt.Sprintf("Field '%s' must be a valid API content type", fe.Field())
	case "api_version":
		return fmt.Sprintf("Field '%s' must be a valid API version format", fe.Field())
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
	default:
		return fmt.Sprintf("Field '%s' failed validation: %s", fe.Field(), fe.Tag())
	}
}

// countAPIRequestFields counts the number of fields in an API request for metrics.
func countAPIRequestFields(request models.APIRequest) int {
	count := 15 // Base fields
	count += len(request.Headers)
	count += len(request.QueryParams)
	count += len(request.PathParams)
	return count
}

// countAPIResponseFields counts the number of fields in an API response for metrics.
func countAPIResponseFields(response models.APIResponse) int {
	count := 15 // Base fields
	count += len(response.Headers)
	count += len(response.Errors)
	count += len(response.Warnings)
	count += len(response.Links)
	return count
}

// getRuleCount returns the number of validation rules applied.
func (av *APIValidator) getRuleCount() int {
	// Return approximate number of validation rules
	return 40
}
