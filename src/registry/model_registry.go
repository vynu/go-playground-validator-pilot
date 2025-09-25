// Package registry provides core model registration types and interfaces.
// The main functionality has been moved to unified_registry.go to eliminate duplication.
package registry

import (
	"reflect"
)

// ModelType represents different types of models that can be registered.
type ModelType string

// ValidatorInterface represents the interface that all validators must implement.
type ValidatorInterface interface {
	ValidatePayload(payload interface{}) interface{} // Return type is now interface{} for flexibility
}

// ModelInfo contains information about a registered model.
type ModelInfo struct {
	Type        ModelType
	Name        string
	Description string
	ModelStruct reflect.Type
	Validator   ValidatorInterface
	Examples    []interface{}
	Version     string
	CreatedAt   string
	Author      string
	Tags        []string
}

// UniversalValidatorWrapper - A universal wrapper that works with any validator using reflection
type UniversalValidatorWrapper struct {
	modelType         string
	validatorInstance interface{}
	modelStructType   reflect.Type
}

// ValidatePayload implements ValidatorInterface using reflection to call any validator
func (uvw *UniversalValidatorWrapper) ValidatePayload(payload interface{}) interface{} {
	// Use reflection to call the validator's ValidatePayload method
	validatorValue := reflect.ValueOf(uvw.validatorInstance)

	// Try different method names that validators might use
	methodNames := []string{"ValidatePayload", "Validate", "ValidateRequest", "ValidateModel"}

	for _, methodName := range methodNames {
		validateMethod := validatorValue.MethodByName(methodName)
		if !validateMethod.IsValid() {
			continue
		}

		// Call the method with the payload
		results := validateMethod.Call([]reflect.Value{reflect.ValueOf(payload)})

		// Return the first result (should be ValidationResult)
		if len(results) > 0 {
			result := results[0].Interface()

			// Ensure model type is set for validation results
			if validationResult, ok := result.(map[string]interface{}); ok {
				if validationResult["model_type"] == "" || validationResult["model_type"] == nil {
					validationResult["model_type"] = uvw.modelType
				}
				if validationResult["provider"] == "" || validationResult["provider"] == nil {
					validationResult["provider"] = "universal-wrapper"
				}
			}

			return result
		}
	}

	// Fallback: create a basic validation result
	return map[string]interface{}{
		"is_valid":   false,
		"model_type": uvw.modelType,
		"provider":   "universal-wrapper-fallback",
		"errors": []map[string]interface{}{{
			"field":   "validator",
			"message": "No suitable validation method found for " + uvw.modelType,
			"code":    "METHOD_NOT_FOUND",
		}},
	}
}
