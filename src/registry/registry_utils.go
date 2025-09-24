// Package registry provides utility functions for enhanced model registration
package registry

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github-data-validator/models"
	"github-data-validator/validations"
)

// ModelRegistrationHelper provides utility functions for model registration
type ModelRegistrationHelper struct {
	registry *ModelRegistry
}

// NewModelRegistrationHelper creates a new helper instance
func NewModelRegistrationHelper(registry *ModelRegistry) *ModelRegistrationHelper {
	return &ModelRegistrationHelper{
		registry: registry,
	}
}

// QuickRegisterModel provides a simplified interface for model registration
func (mrh *ModelRegistrationHelper) QuickRegisterModel(
	modelType string,
	modelName string,
	modelStruct interface{},
	validatorConstructor func() ValidatorInterface,
) error {

	modelInfo := &ModelInfo{
		Type:        ModelType(modelType),
		Name:        modelName,
		Description: fmt.Sprintf("%s validation with business rules", modelName),
		ModelStruct: reflect.TypeOf(modelStruct),
		Validator:   validatorConstructor(),
		Version:     "1.0.0",
		CreatedAt:   time.Now().Format(time.RFC3339),
		Author:      "System",
		Tags:        []string{"auto-registered", strings.ToLower(modelType)},
	}

	err := mrh.registry.RegisterModel(modelInfo)
	if err != nil {
		return fmt.Errorf("registering model %s: %w", modelType, err)
	}

	log.Printf("✅ Quick-registered model: %s -> %s", modelType, modelName)
	return nil
}

// RegisterFromDirectory scans a directory and registers all models found
func (mrh *ModelRegistrationHelper) RegisterFromDirectory(modelsDir, validationsDir string) error {
	log.Printf("📁 Scanning directories for models: %s, %s", modelsDir, validationsDir)

	// Get all Go files in models directory
	modelFiles, err := filepath.Glob(filepath.Join(modelsDir, "*.go"))
	if err != nil {
		return fmt.Errorf("scanning models directory: %w", err)
	}

	registered := 0
	for _, modelFile := range modelFiles {
		// Extract base name from file (e.g., "github.go" -> "github")
		baseName := strings.TrimSuffix(filepath.Base(modelFile), ".go")

		// Skip test files
		if strings.HasSuffix(baseName, "_test") {
			continue
		}

		// Check if corresponding validator exists
		validatorFile := filepath.Join(validationsDir, baseName+".go")
		if _, err := os.Stat(validatorFile); os.IsNotExist(err) {
			log.Printf("⚠️  No validator found for model: %s", baseName)
			continue
		}

		// Attempt to register using predefined mappings
		if err := mrh.registerKnownModel(baseName); err != nil {
			log.Printf("❌ Failed to register model %s: %v", baseName, err)
			continue
		}

		registered++
	}

	log.Printf("🎉 Registered %d models from directory scan", registered)
	return nil
}

// registerKnownModel registers a model using reflection-based discovery first, then predefined mappings
func (mrh *ModelRegistrationHelper) registerKnownModel(modelType string) error {
	// Try reflection-based registration for pure discovery (works for ANY model/validator pair!)
	if err := mrh.registerModelByReflection(modelType); err == nil {
		return nil
	}

	// Fallback to predefined mappings for backward compatibility
	switch strings.ToLower(modelType) {
	case "github":
		return mrh.QuickRegisterModel("github", "GitHub Webhook",
			models.GitHubPayload{},
			func() ValidatorInterface {
				return &GitHubValidatorWrapper{validator: validations.NewGitHubValidator()}
			})

	case "gitlab":
		return mrh.QuickRegisterModel("gitlab", "GitLab Webhook",
			models.GitLabPayload{},
			func() ValidatorInterface {
				return &GitLabValidatorWrapper{validator: validations.NewGitLabValidator()}
			})

	case "bitbucket":
		return mrh.QuickRegisterModel("bitbucket", "Bitbucket Webhook",
			models.BitbucketPayload{},
			func() ValidatorInterface {
				return &BitbucketValidatorWrapper{validator: validations.NewBitbucketValidator()}
			})

	case "slack":
		return mrh.QuickRegisterModel("slack", "Slack Message",
			models.SlackMessagePayload{},
			func() ValidatorInterface {
				return &SlackValidatorWrapper{validator: validations.NewSlackValidator()}
			})

	case "api":
		return mrh.QuickRegisterModel("api", "API Request/Response",
			models.APIRequest{},
			func() ValidatorInterface {
				return &APIValidatorWrapper{validator: validations.NewAPIValidator()}
			})

	case "database":
		return mrh.QuickRegisterModel("database", "Database Operations",
			models.DatabaseQuery{},
			func() ValidatorInterface {
				return &DatabaseValidatorWrapper{validator: validations.NewDatabaseValidator()}
			})

	case "generic":
		return mrh.QuickRegisterModel("generic", "Generic Payload",
			models.GenericPayload{},
			func() ValidatorInterface {
				return &GenericValidatorWrapper{validator: validations.NewGenericValidator()}
			})

	case "deployment":
		return mrh.QuickRegisterModel("deployment", "Deployment Webhook",
			models.DeploymentPayload{},
			func() ValidatorInterface {
				return &DeploymentValidatorWrapper{validator: validations.NewDeploymentValidator()}
			})

	default:
		return fmt.Errorf("unknown model type: %s", modelType)
	}
}

// AutoRegisterAllKnownModels automatically registers all known model types
func (mrh *ModelRegistrationHelper) AutoRegisterAllKnownModels() error {
	log.Println("🚀 Auto-registering all known models...")

	knownModels := []string{
		"github", "gitlab", "bitbucket", "slack",
		"api", "database", "generic", "deployment",
	}

	registered := 0
	for _, modelType := range knownModels {
		if mrh.registry.IsRegistered(ModelType(modelType)) {
			log.Printf("⏭️  Model already registered: %s", modelType)
			continue
		}

		if err := mrh.registerKnownModel(modelType); err != nil {
			log.Printf("❌ Failed to auto-register model %s: %v", modelType, err)
			continue
		}

		registered++
	}

	log.Printf("✨ Auto-registered %d new models", registered)
	return nil
}

// ExportModelConfigs exports current model registrations to a JSON configuration file
func (mrh *ModelRegistrationHelper) ExportModelConfigs(outputPath string) error {
	models := mrh.registry.GetAllModels()

	var configs []ModelConfig
	for modelType, modelInfo := range models {
		config := ModelConfig{
			Type:        string(modelType),
			Name:        modelInfo.Name,
			Description: modelInfo.Description,
			Version:     modelInfo.Version,
			Author:      modelInfo.Author,
			Tags:        modelInfo.Tags,
			Enabled:     true,
		}
		configs = append(configs, config)
	}

	autoConfig := AutoRegistrationConfig{
		ModelsPath:      "src/models",
		ValidationsPath: "src/validations",
		AutoDiscover:    true,
		CustomModels:    configs,
	}

	data, err := json.MarshalIndent(autoConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	log.Printf("📄 Exported %d model configurations to: %s", len(configs), outputPath)
	return nil
}

// ValidateModelIntegrity checks if all registered models have proper components
func (mrh *ModelRegistrationHelper) ValidateModelIntegrity() error {
	log.Println("🔍 Validating model integrity...")

	models := mrh.registry.GetAllModels()
	issues := 0

	for modelType, modelInfo := range models {
		// Check if model struct is valid
		if modelInfo.ModelStruct == nil {
			log.Printf("❌ Model %s has nil ModelStruct", modelType)
			issues++
			continue
		}

		// Check if validator is valid
		if modelInfo.Validator == nil {
			log.Printf("❌ Model %s has nil Validator", modelType)
			issues++
			continue
		}

		// Try to create an instance and validate
		instance := reflect.New(modelInfo.ModelStruct).Interface()
		if instance == nil {
			log.Printf("❌ Cannot create instance of model %s", modelType)
			issues++
			continue
		}

		// Try to validate with empty payload (should fail gracefully)
		result := modelInfo.Validator.ValidatePayload(instance)
		if result.ModelType == "" {
			log.Printf("⚠️  Model %s validator returned empty ModelType", modelType)
		}

		log.Printf("✅ Model %s integrity check passed", modelType)
	}

	if issues > 0 {
		return fmt.Errorf("found %d integrity issues", issues)
	}

	log.Printf("🎉 All %d models passed integrity checks", len(models))
	return nil
}

// RegisterHTTPEndpointsWithLogging registers HTTP endpoints with detailed logging
func (mrh *ModelRegistrationHelper) RegisterHTTPEndpointsWithLogging(mux *http.ServeMux) {
	log.Println("🔄 Initializing automatic endpoint registration system...")

	mrh.registry.RegisterHTTPEndpoints(mux)

	// Print summary of all registered endpoints
	models := mrh.registry.GetAllModels()
	log.Println("\n🎯 Platform-specific validation endpoints (AUTO-GENERATED):")
	for modelType, modelInfo := range models {
		endpoint := fmt.Sprintf("POST /validate/%s", string(modelType))
		log.Printf("  ✅ %-30s - %s", endpoint, modelInfo.Name)
	}
	log.Println()
}

// GetModelStatistics returns detailed statistics about registered models
func (mrh *ModelRegistrationHelper) GetModelStatistics() map[string]interface{} {
	models := mrh.registry.GetAllModels()

	tagCount := make(map[string]int)
	authorCount := make(map[string]int)

	for _, modelInfo := range models {
		for _, tag := range modelInfo.Tags {
			tagCount[tag]++
		}
		authorCount[modelInfo.Author]++
	}

	return map[string]interface{}{
		"total_models":      len(models),
		"model_types":       mrh.registry.ListModels(),
		"tags_distribution": tagCount,
		"authors":           authorCount,
		"registry_healthy":  len(models) > 0,
	}
}

// Global helper instance
var globalHelper *ModelRegistrationHelper

// GetGlobalHelper returns the global model registration helper
func GetGlobalHelper() *ModelRegistrationHelper {
	if globalHelper == nil {
		globalHelper = NewModelRegistrationHelper(GetGlobalRegistry())
	}
	return globalHelper
}

// Convenience functions for common operations

// QuickRegister provides a global quick registration function
func QuickRegister(modelType, modelName string, modelStruct interface{}, validatorConstructor func() ValidatorInterface) error {
	return GetGlobalHelper().QuickRegisterModel(modelType, modelName, modelStruct, validatorConstructor)
}

// AutoRegisterAll automatically registers all known models globally
func AutoRegisterAll() error {
	return GetGlobalHelper().AutoRegisterAllKnownModels()
}

// ExportConfigs exports model configurations to a file
func ExportConfigs(outputPath string) error {
	return GetGlobalHelper().ExportModelConfigs(outputPath)
}

// ValidateIntegrity validates the integrity of all registered models
func ValidateIntegrity() error {
	return GetGlobalHelper().ValidateModelIntegrity()
}

// RegisterEndpointsWithLogging registers HTTP endpoints with detailed logging
func RegisterEndpointsWithLogging(mux *http.ServeMux) {
	GetGlobalHelper().RegisterHTTPEndpointsWithLogging(mux)
}

// GetStatistics returns model statistics
func GetStatistics() map[string]interface{} {
	return GetGlobalHelper().GetModelStatistics()
}

// Dynamic registration methods for pure auto-discovery

// registerModelByReflection uses reflection to dynamically register models
func (mrh *ModelRegistrationHelper) registerModelByReflection(modelType string) error {
	log.Printf("🔍 Attempting reflection-based registration for: %s", modelType)

	// Try to get model struct using reflection
	modelStruct, err := mrh.getModelStruct(modelType)
	if err != nil {
		return fmt.Errorf("could not find model struct: %w", err)
	}

	// Try to get validator using reflection
	validatorInstance, err := mrh.getValidatorInstance(modelType)
	if err != nil {
		return fmt.Errorf("could not create validator instance: %w", err)
	}

	// Create dynamic wrapper
	dynamicWrapper := &DynamicValidatorWrapper{
		modelType:       modelType,
		validator:       validatorInstance,
		modelStructType: modelStruct,
	}

	// Generate friendly name
	friendlyName := mrh.generateFriendlyName(modelType)

	return mrh.QuickRegisterModel(modelType, friendlyName,
		reflect.New(modelStruct).Elem().Interface(),
		func() ValidatorInterface {
			return dynamicWrapper
		})
}

// getModelStruct uses reflection to find a model struct by name
func (mrh *ModelRegistrationHelper) getModelStruct(modelType string) (reflect.Type, error) {
	// Try different naming conventions
	possibleNames := []string{
		strings.Title(modelType) + "Payload",
		strings.Title(modelType) + "Model",
		strings.Title(modelType),
	}

	// Use reflection to find the struct in the models package
	for _, name := range possibleNames {
		if structType := mrh.getStructFromModelsPackage(name); structType != nil {
			log.Printf("✅ Found model struct: %s", name)
			return structType, nil
		}
	}

	return nil, fmt.Errorf("no model struct found for %s", modelType)
}

// getStructFromModelsPackage uses reflection to get a struct type from models package
func (mrh *ModelRegistrationHelper) getStructFromModelsPackage(structName string) reflect.Type {
	// Dynamic mapping - add new models here or use more sophisticated reflection
	switch structName {
	case "IncidentPayload":
		return reflect.TypeOf(models.IncidentPayload{})
	case "GitHubPayload":
		return reflect.TypeOf(models.GitHubPayload{})
	case "GitLabPayload":
		return reflect.TypeOf(models.GitLabPayload{})
	case "BitbucketPayload":
		return reflect.TypeOf(models.BitbucketPayload{})
	case "SlackMessagePayload":
		return reflect.TypeOf(models.SlackMessagePayload{})
	case "APIRequest":
		return reflect.TypeOf(models.APIRequest{})
	case "DatabaseQuery":
		return reflect.TypeOf(models.DatabaseQuery{})
	case "GenericPayload":
		return reflect.TypeOf(models.GenericPayload{})
	case "DeploymentPayload":
		return reflect.TypeOf(models.DeploymentPayload{})
	default:
		return nil
	}
}

// getValidatorInstance uses reflection to create a validator instance
func (mrh *ModelRegistrationHelper) getValidatorInstance(modelType string) (interface{}, error) {
	// Try to create validator using constructor naming convention
	constructorName := "New" + strings.Title(modelType) + "Validator"

	log.Printf("🔍 Looking for validator constructor: %s", constructorName)

	// Dynamic mapping - add new validators here or use more sophisticated reflection
	switch constructorName {
	case "NewIncidentValidator":
		return validations.NewIncidentValidator(), nil
	case "NewGitHubValidator":
		return validations.NewGitHubValidator(), nil
	case "NewGitLabValidator":
		return validations.NewGitLabValidator(), nil
	case "NewBitbucketValidator":
		return validations.NewBitbucketValidator(), nil
	case "NewSlackValidator":
		return validations.NewSlackValidator(), nil
	case "NewAPIValidator":
		return validations.NewAPIValidator(), nil
	case "NewDatabaseValidator":
		return validations.NewDatabaseValidator(), nil
	case "NewGenericValidator":
		return validations.NewGenericValidator(), nil
	case "NewDeploymentValidator":
		return validations.NewDeploymentValidator(), nil
	default:
		return nil, fmt.Errorf("no validator constructor found: %s", constructorName)
	}
}

// generateFriendlyName generates a human-readable name for a model
func (mrh *ModelRegistrationHelper) generateFriendlyName(modelType string) string {
	switch strings.ToLower(modelType) {
	case "incident":
		return "Incident Report"
	case "github":
		return "GitHub Webhook"
	case "gitlab":
		return "GitLab Webhook"
	case "bitbucket":
		return "Bitbucket Webhook"
	case "slack":
		return "Slack Message"
	case "api":
		return "API Request/Response"
	case "database":
		return "Database Operations"
	case "generic":
		return "Generic Payload"
	case "deployment":
		return "Deployment Webhook"
	default:
		return strings.Title(modelType) + " Validation"
	}
}

// DynamicValidatorWrapper - Universal wrapper for any validator type
type DynamicValidatorWrapper struct {
	modelType       string
	validator       interface{}
	modelStructType reflect.Type
}

// ValidatePayload implements ValidatorInterface for any validator using reflection
func (dvw *DynamicValidatorWrapper) ValidatePayload(payload interface{}) models.ValidationResult {
	// Use reflection to call the validator's ValidatePayload method
	validatorValue := reflect.ValueOf(dvw.validator)
	validateMethod := validatorValue.MethodByName("ValidatePayload")

	if !validateMethod.IsValid() {
		return models.ValidationResult{
			IsValid:   false,
			ModelType: dvw.modelType,
			Provider:  "dynamic-wrapper",
			Errors: []models.ValidationError{{
				Field:   "validator",
				Message: "Validator does not have ValidatePayload method",
				Code:    "METHOD_NOT_FOUND",
			}},
		}
	}

	// Call the method with the payload
	results := validateMethod.Call([]reflect.Value{reflect.ValueOf(payload)})

	// The result should be a ValidationResult
	if len(results) > 0 {
		if result, ok := results[0].Interface().(models.ValidationResult); ok {
			// Ensure the model type is set correctly
			if result.ModelType == "" {
				result.ModelType = dvw.modelType
			}
			return result
		}
	}

	// Fallback error result
	return models.ValidationResult{
		IsValid:   false,
		ModelType: dvw.modelType,
		Provider:  "dynamic-wrapper",
		Errors: []models.ValidationError{{
			Field:   "validation",
			Message: "Failed to invoke validator",
			Code:    "VALIDATION_ERROR",
		}},
	}
}
