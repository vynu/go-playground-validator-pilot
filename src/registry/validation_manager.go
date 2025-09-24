// Package registry provides validation management and coordination capabilities.
// This module coordinates between different validators and provides unified validation interface.
package registry

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github-data-validator/models"
)

// ValidationProfile represents different validation strictness levels.
type ValidationProfile string

const (
	ProfileStrict     ValidationProfile = "strict"
	ProfilePermissive ValidationProfile = "permissive"
	ProfileMinimal    ValidationProfile = "minimal"
)

// ValidationOptions contains options for validation behavior.
type ValidationOptions struct {
	Profile           ValidationProfile      `json:"profile"`
	StopOnFirstError  bool                   `json:"stop_on_first_error"`
	IncludeWarnings   bool                   `json:"include_warnings"`
	IncludeMetrics    bool                   `json:"include_metrics"`
	MaxErrors         int                    `json:"max_errors"`
	MaxWarnings       int                    `json:"max_warnings"`
	CustomRules       []string               `json:"custom_rules"`
	IgnoredFields     []string               `json:"ignored_fields"`
	RequiredFields    []string               `json:"required_fields"`
	ValidationContext map[string]interface{} `json:"validation_context"`
}

// DefaultValidationOptions returns default validation options.
func DefaultValidationOptions() ValidationOptions {
	return ValidationOptions{
		Profile:           ProfileStrict,
		StopOnFirstError:  false,
		IncludeWarnings:   true,
		IncludeMetrics:    true,
		MaxErrors:         100,
		MaxWarnings:       50,
		CustomRules:       []string{},
		IgnoredFields:     []string{},
		RequiredFields:    []string{},
		ValidationContext: make(map[string]interface{}),
	}
}

// ValidationManager manages validation operations across different model types.
type ValidationManager struct {
	registry *ModelRegistry
	options  ValidationOptions
}

// NewValidationManager creates a new validation manager.
func NewValidationManager(registry *ModelRegistry) *ValidationManager {
	return &ValidationManager{
		registry: registry,
		options:  DefaultValidationOptions(),
	}
}

// SetOptions sets validation options.
func (vm *ValidationManager) SetOptions(options ValidationOptions) {
	vm.options = options
}

// GetOptions returns current validation options.
func (vm *ValidationManager) GetOptions() ValidationOptions {
	return vm.options
}

// ValidatePayload validates a payload with the specified model type and options.
func (vm *ValidationManager) ValidatePayload(modelType ModelType, payload interface{}, options *ValidationOptions) (models.ValidationResult, error) {
	start := time.Now()

	// Use provided options or fall back to manager defaults
	opts := vm.options
	if options != nil {
		opts = *options
	}

	// Get the appropriate validator
	validator, err := vm.registry.GetValidator(modelType)
	if err != nil {
		return models.ValidationResult{}, fmt.Errorf("failed to get validator for type %s: %v", modelType, err)
	}

	// Perform validation
	result := validator.ValidatePayload(payload)

	// Apply validation options
	result = vm.applyValidationOptions(result, opts)

	// Add validation manager metadata
	result.ValidationProfile = string(opts.Profile)
	result.RequestID = generateRequestID()

	if result.Context == nil {
		result.Context = make(map[string]interface{})
	}
	result.Context["validation_manager"] = true
	result.Context["model_type"] = string(modelType)
	result.Context["options"] = opts
	result.Context["total_processing_duration"] = time.Since(start)

	return result, nil
}

// ValidateMultiplePayloads validates multiple payloads in batch.
func (vm *ValidationManager) ValidateMultiplePayloads(payloads []PayloadWithType, options *ValidationOptions) ([]models.ValidationResult, error) {
	opts := vm.options
	if options != nil {
		opts = *options
	}

	results := make([]models.ValidationResult, len(payloads))

	for i, payloadWithType := range payloads {
		result, err := vm.ValidatePayload(payloadWithType.ModelType, payloadWithType.Payload, &opts)
		if err != nil {
			// Create error result
			result = models.ValidationResult{
				IsValid:   false,
				ModelType: string(payloadWithType.ModelType),
				Provider:  "validation_manager",
				Timestamp: time.Now(),
				Errors: []models.ValidationError{{
					Field:   "payload",
					Message: err.Error(),
					Code:    "VALIDATION_ERROR",
				}},
			}
		}

		results[i] = result

		// Stop on first error if configured
		if opts.StopOnFirstError && !result.IsValid {
			// Truncate results to only include up to the failed one
			return results[:i+1], nil
		}
	}

	return results, nil
}

// PayloadWithType combines a payload with its model type.
type PayloadWithType struct {
	ModelType ModelType   `json:"model_type"`
	Payload   interface{} `json:"payload"`
}

// AutoDetectModelType attempts to automatically detect the model type from a payload.
func (vm *ValidationManager) AutoDetectModelType(payload interface{}) (ModelType, error) {
	// Try to determine model type based on payload structure
	payloadType := reflect.TypeOf(payload)

	// If it's a pointer, get the underlying type
	if payloadType.Kind() == reflect.Ptr {
		payloadType = payloadType.Elem()
	}

	// Check against registered model types
	for modelType, modelInfo := range vm.registry.GetAllModels() {
		if payloadType == modelInfo.ModelStruct {
			return modelType, nil
		}
	}

	// Try to detect based on JSON structure if payload is a map
	if payloadMap, ok := payload.(map[string]interface{}); ok {
		return vm.detectModelTypeFromMap(payloadMap)
	}

	return "", fmt.Errorf("unable to auto-detect model type for payload of type %s", payloadType)
}

// detectModelTypeFromMap detects model type based on map structure.
func (vm *ValidationManager) detectModelTypeFromMap(payloadMap map[string]interface{}) (ModelType, error) {
	// GitHub webhook detection
	if _, hasAction := payloadMap["action"]; hasAction {
		if _, hasPullRequest := payloadMap["pull_request"]; hasPullRequest {
			if _, hasRepository := payloadMap["repository"]; hasRepository {
				return ModelTypeGitHub, nil
			}
		}
	}

	// GitLab webhook detection
	if objectKind, ok := payloadMap["object_kind"].(string); ok {
		if objectKind == "merge_request" || objectKind == "push" {
			return ModelTypeGitLab, nil
		}
	}

	// Bitbucket webhook detection
	if _, hasRepository := payloadMap["repository"]; hasRepository {
		if _, hasActor := payloadMap["actor"]; hasActor {
			if _, hasPullRequest := payloadMap["pullrequest"]; hasPullRequest {
				return ModelTypeBitbucket, nil
			}
		}
	}

	// Slack message detection
	if _, hasText := payloadMap["text"]; hasText {
		if _, hasChannel := payloadMap["channel"]; hasChannel {
			return ModelTypeSlack, nil
		}
	}

	// API request/response detection
	if method, ok := payloadMap["method"].(string); ok {
		if _, hasURL := payloadMap["url"]; hasURL {
			httpMethods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
			for _, httpMethod := range httpMethods {
				if strings.ToUpper(method) == httpMethod {
					return ModelTypeAPI, nil
				}
			}
		}
	}

	// Database query detection
	if operation, ok := payloadMap["operation"].(string); ok {
		if _, hasQuery := payloadMap["query"]; hasQuery {
			dbOperations := []string{"SELECT", "INSERT", "UPDATE", "DELETE", "CREATE", "DROP", "ALTER"}
			for _, dbOperation := range dbOperations {
				if strings.ToUpper(operation) == dbOperation {
					return ModelTypeDatabase, nil
				}
			}
		}
	}

	// Generic payload detection (fallback)
	if _, hasType := payloadMap["type"]; hasType {
		if _, hasData := payloadMap["data"]; hasData {
			return ModelTypeGeneric, nil
		}
	}

	return "", fmt.Errorf("unable to detect model type from map structure")
}

// ValidateWithAutoDetection validates a payload with automatic model type detection.
func (vm *ValidationManager) ValidateWithAutoDetection(payload interface{}, options *ValidationOptions) (models.ValidationResult, error) {
	modelType, err := vm.AutoDetectModelType(payload)
	if err != nil {
		return models.ValidationResult{}, fmt.Errorf("auto-detection failed: %v", err)
	}

	return vm.ValidatePayload(modelType, payload, options)
}

// applyValidationOptions applies validation options to filter and modify results.
func (vm *ValidationManager) applyValidationOptions(result models.ValidationResult, options ValidationOptions) models.ValidationResult {
	// Initialize Context if nil
	if result.Context == nil {
		result.Context = make(map[string]interface{})
	}

	// Apply error limits
	if len(result.Errors) > options.MaxErrors {
		result.Errors = result.Errors[:options.MaxErrors]
		result.Context["errors_truncated"] = true
		result.Context["original_error_count"] = len(result.Errors)
	}

	// Apply warning limits and inclusion
	if !options.IncludeWarnings {
		result.Warnings = []models.ValidationWarning{}
	} else if len(result.Warnings) > options.MaxWarnings {
		result.Warnings = result.Warnings[:options.MaxWarnings]
		result.Context["warnings_truncated"] = true
		result.Context["original_warning_count"] = len(result.Warnings)
	}

	// Apply metrics inclusion
	if !options.IncludeMetrics {
		result.PerformanceMetrics = nil
	}

	// Apply profile-specific filtering
	switch options.Profile {
	case ProfileMinimal:
		// Keep only critical errors
		filteredErrors := []models.ValidationError{}
		for _, err := range result.Errors {
			if err.Severity == "error" || err.Severity == "critical" {
				filteredErrors = append(filteredErrors, err)
			}
		}
		result.Errors = filteredErrors
		result.Warnings = []models.ValidationWarning{} // Remove all warnings in minimal mode

	case ProfilePermissive:
		// Convert some errors to warnings
		permissiveErrors := []models.ValidationError{}
		for _, err := range result.Errors {
			if err.Severity == "critical" || err.Code == "required" {
				permissiveErrors = append(permissiveErrors, err)
			} else {
				// Convert to warning
				warning := models.ValidationWarning{
					Field:      err.Field,
					Message:    err.Message,
					Code:       err.Code,
					Value:      err.Value,
					Suggestion: "Consider addressing this validation issue",
					Category:   "permissive-mode",
				}
				result.Warnings = append(result.Warnings, warning)
			}
		}
		result.Errors = permissiveErrors

		// Update validity if only non-critical errors remain
		if len(permissiveErrors) == 0 {
			result.IsValid = true
		}
	}

	// Apply field filtering
	if len(options.IgnoredFields) > 0 {
		result.Errors = vm.filterErrorsByFields(result.Errors, options.IgnoredFields, true)
		result.Warnings = vm.filterWarningsByFields(result.Warnings, options.IgnoredFields, true)
	}

	return result
}

// filterErrorsByFields filters errors based on field names.
func (vm *ValidationManager) filterErrorsByFields(errors []models.ValidationError, fields []string, exclude bool) []models.ValidationError {
	filtered := []models.ValidationError{}

	for _, err := range errors {
		shouldInclude := true

		for _, field := range fields {
			if strings.Contains(err.Field, field) || strings.Contains(err.Path, field) {
				shouldInclude = !exclude // If exclude=true and field matches, don't include
				break
			}
		}

		if shouldInclude {
			filtered = append(filtered, err)
		}
	}

	return filtered
}

// filterWarningsByFields filters warnings based on field names.
func (vm *ValidationManager) filterWarningsByFields(warnings []models.ValidationWarning, fields []string, exclude bool) []models.ValidationWarning {
	filtered := []models.ValidationWarning{}

	for _, warning := range warnings {
		shouldInclude := true

		for _, field := range fields {
			if strings.Contains(warning.Field, field) || strings.Contains(warning.Path, field) {
				shouldInclude = !exclude // If exclude=true and field matches, don't include
				break
			}
		}

		if shouldInclude {
			filtered = append(filtered, warning)
		}
	}

	return filtered
}

// GetSupportedModelTypes returns a list of all supported model types.
func (vm *ValidationManager) GetSupportedModelTypes() []ModelType {
	return vm.registry.ListModels()
}

// GetModelInfo returns information about a specific model type.
func (vm *ValidationManager) GetModelInfo(modelType ModelType) (*ModelInfo, error) {
	return vm.registry.GetModel(modelType)
}

// GetValidationStats returns statistics about validation operations.
func (vm *ValidationManager) GetValidationStats() map[string]interface{} {
	stats := vm.registry.GetModelStats()
	stats["validation_profiles"] = []string{string(ProfileStrict), string(ProfilePermissive), string(ProfileMinimal)}
	stats["current_profile"] = string(vm.options.Profile)
	return stats
}

// CreateValidationReport creates a comprehensive validation report.
func (vm *ValidationManager) CreateValidationReport(results []models.ValidationResult) map[string]interface{} {
	report := map[string]interface{}{
		"summary": map[string]interface{}{
			"total_validations": len(results),
			"successful":        0,
			"failed":            0,
			"total_errors":      0,
			"total_warnings":    0,
		},
		"by_model_type":      make(map[string]interface{}),
		"error_categories":   make(map[string]int),
		"warning_categories": make(map[string]int),
		"performance_stats": map[string]interface{}{
			"average_duration": time.Duration(0),
			"min_duration":     time.Duration(0),
			"max_duration":     time.Duration(0),
		},
	}

	var totalDuration time.Duration
	var minDuration, maxDuration time.Duration
	modelTypeStats := make(map[string]map[string]int)
	errorCategories := make(map[string]int)
	warningCategories := make(map[string]int)

	for i, result := range results {
		// Update summary
		if result.IsValid {
			report["summary"].(map[string]interface{})["successful"] = report["summary"].(map[string]interface{})["successful"].(int) + 1
		} else {
			report["summary"].(map[string]interface{})["failed"] = report["summary"].(map[string]interface{})["failed"].(int) + 1
		}

		report["summary"].(map[string]interface{})["total_errors"] = report["summary"].(map[string]interface{})["total_errors"].(int) + len(result.Errors)
		report["summary"].(map[string]interface{})["total_warnings"] = report["summary"].(map[string]interface{})["total_warnings"].(int) + len(result.Warnings)

		// Update model type stats
		if _, exists := modelTypeStats[result.ModelType]; !exists {
			modelTypeStats[result.ModelType] = map[string]int{
				"total": 0, "successful": 0, "failed": 0, "errors": 0, "warnings": 0,
			}
		}
		modelTypeStats[result.ModelType]["total"]++
		if result.IsValid {
			modelTypeStats[result.ModelType]["successful"]++
		} else {
			modelTypeStats[result.ModelType]["failed"]++
		}
		modelTypeStats[result.ModelType]["errors"] += len(result.Errors)
		modelTypeStats[result.ModelType]["warnings"] += len(result.Warnings)

		// Update error categories
		for _, err := range result.Errors {
			if err.Code != "" {
				errorCategories[err.Code]++
			}
		}

		// Update warning categories
		for _, warning := range result.Warnings {
			if warning.Category != "" {
				warningCategories[warning.Category]++
			}
		}

		// Update performance stats
		if result.ProcessingDuration > 0 {
			totalDuration += result.ProcessingDuration
			if i == 0 || result.ProcessingDuration < minDuration {
				minDuration = result.ProcessingDuration
			}
			if result.ProcessingDuration > maxDuration {
				maxDuration = result.ProcessingDuration
			}
		}
	}

	// Calculate average duration
	if len(results) > 0 {
		report["performance_stats"].(map[string]interface{})["average_duration"] = totalDuration / time.Duration(len(results))
		report["performance_stats"].(map[string]interface{})["min_duration"] = minDuration
		report["performance_stats"].(map[string]interface{})["max_duration"] = maxDuration
	}

	report["by_model_type"] = modelTypeStats
	report["error_categories"] = errorCategories
	report["warning_categories"] = warningCategories

	return report
}

// generateRequestID generates a unique request ID for tracking.
func generateRequestID() string {
	return fmt.Sprintf("val_%d", time.Now().UnixNano())
}

// Global validation manager instance
var globalValidationManager *ValidationManager

// GetGlobalValidationManager returns the global validation manager instance.
func GetGlobalValidationManager() *ValidationManager {
	if globalValidationManager == nil {
		globalValidationManager = NewValidationManager(GetGlobalRegistry())
	}
	return globalValidationManager
}

// Helper functions for common operations

// ValidateWithManager validates a payload using the global validation manager.
func ValidateWithManager(modelType ModelType, payload interface{}, options *ValidationOptions) (models.ValidationResult, error) {
	return GetGlobalValidationManager().ValidatePayload(modelType, payload, options)
}

// AutoValidate validates a payload with automatic model type detection using the global manager.
func AutoValidate(payload interface{}, options *ValidationOptions) (models.ValidationResult, error) {
	return GetGlobalValidationManager().ValidateWithAutoDetection(payload, options)
}
