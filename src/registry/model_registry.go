// Package registry provides dynamic model registration and loading capabilities.
// This module allows for extensible model and validation system.
package registry

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"sync"

	"github-data-validator/models"
	"github-data-validator/validations"
)

// ModelType represents different types of models that can be registered.
type ModelType string

const (
	ModelTypeGitHub     ModelType = "github"
	ModelTypeGitLab     ModelType = "gitlab"
	ModelTypeBitbucket  ModelType = "bitbucket"
	ModelTypeSlack      ModelType = "slack"
	ModelTypeAPI        ModelType = "api"
	ModelTypeDatabase   ModelType = "database"
	ModelTypeGeneric    ModelType = "generic"
	ModelTypeDeployment ModelType = "deployment"
)

// ValidatorInterface represents the interface that all validators must implement.
type ValidatorInterface interface {
	ValidatePayload(payload interface{}) models.ValidationResult
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

// ModelRegistry manages the registration and lookup of models and validators.
type ModelRegistry struct {
	models map[ModelType]*ModelInfo
	mutex  sync.RWMutex
}

// NewModelRegistry creates a new model registry instance.
func NewModelRegistry() *ModelRegistry {
	registry := &ModelRegistry{
		models: make(map[ModelType]*ModelInfo),
	}

	// Register built-in models
	registry.registerBuiltInModels()

	return registry
}

// RegisterModel registers a new model with its validator.
func (mr *ModelRegistry) RegisterModel(info *ModelInfo) error {
	mr.mutex.Lock()
	defer mr.mutex.Unlock()

	if info.Type == "" {
		return fmt.Errorf("model type cannot be empty")
	}

	if info.Name == "" {
		return fmt.Errorf("model name cannot be empty")
	}

	if info.ModelStruct == nil {
		return fmt.Errorf("model struct cannot be nil")
	}

	if info.Validator == nil {
		return fmt.Errorf("validator cannot be nil")
	}

	mr.models[info.Type] = info
	return nil
}

// GetModel retrieves model information by type.
func (mr *ModelRegistry) GetModel(modelType ModelType) (*ModelInfo, error) {
	mr.mutex.RLock()
	defer mr.mutex.RUnlock()

	model, exists := mr.models[modelType]
	if !exists {
		return nil, fmt.Errorf("model type '%s' not found", modelType)
	}

	return model, nil
}

// GetValidator retrieves a validator by model type.
func (mr *ModelRegistry) GetValidator(modelType ModelType) (ValidatorInterface, error) {
	model, err := mr.GetModel(modelType)
	if err != nil {
		return nil, err
	}

	return model.Validator, nil
}

// ListModels returns a list of all registered model types.
func (mr *ModelRegistry) ListModels() []ModelType {
	mr.mutex.RLock()
	defer mr.mutex.RUnlock()

	var types []ModelType
	for modelType := range mr.models {
		types = append(types, modelType)
	}

	return types
}

// GetAllModels returns all registered model information.
func (mr *ModelRegistry) GetAllModels() map[ModelType]*ModelInfo {
	mr.mutex.RLock()
	defer mr.mutex.RUnlock()

	result := make(map[ModelType]*ModelInfo)
	for k, v := range mr.models {
		result[k] = v
	}

	return result
}

// ValidatePayload validates a payload using the appropriate validator.
func (mr *ModelRegistry) ValidatePayload(modelType ModelType, payload interface{}) (models.ValidationResult, error) {
	validator, err := mr.GetValidator(modelType)
	if err != nil {
		return models.ValidationResult{}, err
	}

	return validator.ValidatePayload(payload), nil
}

// CreateModelInstance creates a new instance of the specified model type.
func (mr *ModelRegistry) CreateModelInstance(modelType ModelType) (interface{}, error) {
	model, err := mr.GetModel(modelType)
	if err != nil {
		return nil, err
	}

	// Create a new instance of the model struct
	modelValue := reflect.New(model.ModelStruct).Interface()
	return modelValue, nil
}

// IsRegistered checks if a model type is registered.
func (mr *ModelRegistry) IsRegistered(modelType ModelType) bool {
	mr.mutex.RLock()
	defer mr.mutex.RUnlock()

	_, exists := mr.models[modelType]
	return exists
}

// UnregisterModel removes a model from the registry.
func (mr *ModelRegistry) UnregisterModel(modelType ModelType) error {
	mr.mutex.Lock()
	defer mr.mutex.Unlock()

	if _, exists := mr.models[modelType]; !exists {
		return fmt.Errorf("model type '%s' not found", modelType)
	}

	delete(mr.models, modelType)
	return nil
}

// GetModelStats returns statistics about registered models.
func (mr *ModelRegistry) GetModelStats() map[string]interface{} {
	mr.mutex.RLock()
	defer mr.mutex.RUnlock()

	stats := map[string]interface{}{
		"total_models": len(mr.models),
		"model_types":  make([]string, 0, len(mr.models)),
	}

	for modelType := range mr.models {
		stats["model_types"] = append(stats["model_types"].([]string), string(modelType))
	}

	return stats
}

// registerBuiltInModels registers all the built-in model types.
// AUTOMATED: This function now uses the automatic directory scanning system
func (mr *ModelRegistry) registerBuiltInModels() {
	log.Println("🚀 Using automatic directory scanning model registration system...")

	// Use the directory scanning registration helper
	helper := NewModelRegistrationHelper(mr)

	// First try directory scanning for automatic discovery
	log.Println("🔍 Scanning directories for models and validators...")
	if err := helper.RegisterFromDirectory("models", "validations"); err != nil {
		log.Printf("⚠️  Directory scanning failed: %v", err)
	}

	// Also register known models (for backward compatibility)
	if err := helper.AutoRegisterAllKnownModels(); err != nil {
		log.Printf("⚠️  Error in auto-registration, falling back to manual: %v", err)
		mr.fallbackToManualRegistration()
		return
	}

	log.Println("✨ Automatic model registration completed successfully")
}

// fallbackToManualRegistration provides manual registration as a fallback
func (mr *ModelRegistry) fallbackToManualRegistration() {
	log.Println("🔄 Falling back to manual model registration...")

	// Register GitHub model manually
	mr.models[ModelTypeGitHub] = &ModelInfo{
		Type:        ModelTypeGitHub,
		Name:        "GitHub Webhook",
		Description: "GitHub webhook payload validation with comprehensive business rules",
		ModelStruct: reflect.TypeOf(models.GitHubPayload{}),
		Validator:   &GitHubValidatorWrapper{validator: validations.NewGitHubValidator()},
		Version:     "1.0.0",
		Author:      "System",
		Tags:        []string{"webhook", "github", "git", "collaboration"},
	}

	// Register GitLab model
	mr.models[ModelTypeGitLab] = &ModelInfo{
		Type:        ModelTypeGitLab,
		Name:        "GitLab Webhook",
		Description: "GitLab webhook payload validation with comprehensive business rules",
		ModelStruct: reflect.TypeOf(models.GitLabPayload{}),
		Validator:   &GitLabValidatorWrapper{validator: validations.NewGitLabValidator()},
		Version:     "1.0.0",
		Author:      "System",
		Tags:        []string{"webhook", "gitlab", "git", "collaboration"},
	}

	// Register Bitbucket model
	mr.models[ModelTypeBitbucket] = &ModelInfo{
		Type:        ModelTypeBitbucket,
		Name:        "Bitbucket Webhook",
		Description: "Bitbucket webhook payload validation with comprehensive business rules",
		ModelStruct: reflect.TypeOf(models.BitbucketPayload{}),
		Validator:   &BitbucketValidatorWrapper{validator: validations.NewBitbucketValidator()},
		Version:     "1.0.0",
		Author:      "System",
		Tags:        []string{"webhook", "bitbucket", "git", "collaboration"},
	}

	// Register Slack model
	mr.models[ModelTypeSlack] = &ModelInfo{
		Type:        ModelTypeSlack,
		Name:        "Slack Message",
		Description: "Slack message payload validation with comprehensive business rules",
		ModelStruct: reflect.TypeOf(models.SlackMessagePayload{}),
		Validator:   &SlackValidatorWrapper{validator: validations.NewSlackValidator()},
		Version:     "1.0.0",
		Author:      "System",
		Tags:        []string{"messaging", "slack", "communication"},
	}

	// Register API model
	mr.models[ModelTypeAPI] = &ModelInfo{
		Type:        ModelTypeAPI,
		Name:        "API Request/Response",
		Description: "API request and response validation with comprehensive business rules",
		ModelStruct: reflect.TypeOf(models.APIRequest{}),
		Validator:   &APIValidatorWrapper{validator: validations.NewAPIValidator()},
		Version:     "1.0.0",
		Author:      "System",
		Tags:        []string{"api", "http", "rest", "web"},
	}

	// Register Database model
	mr.models[ModelTypeDatabase] = &ModelInfo{
		Type:        ModelTypeDatabase,
		Name:        "Database Operations",
		Description: "Database query and transaction validation with comprehensive business rules",
		ModelStruct: reflect.TypeOf(models.DatabaseQuery{}),
		Validator:   &DatabaseValidatorWrapper{validator: validations.NewDatabaseValidator()},
		Version:     "1.0.0",
		Author:      "System",
		Tags:        []string{"database", "sql", "transaction", "query"},
	}

	// Register Generic model
	mr.models[ModelTypeGeneric] = &ModelInfo{
		Type:        ModelTypeGeneric,
		Name:        "Generic Payload",
		Description: "Generic payload validation with flexible business rules",
		ModelStruct: reflect.TypeOf(models.GenericPayload{}),
		Validator:   &GenericValidatorWrapper{validator: validations.NewGenericValidator()},
		Version:     "1.0.0",
		Author:      "System",
		Tags:        []string{"generic", "flexible", "json", "general"},
	}

	// Register Deployment model
	mr.models[ModelTypeDeployment] = &ModelInfo{
		Type:        ModelTypeDeployment,
		Name:        "Deployment Webhook",
		Description: "Deployment webhook payload validation with semantic versioning and business rules",
		ModelStruct: reflect.TypeOf(models.DeploymentPayload{}),
		Validator:   &DeploymentValidatorWrapper{validator: validations.NewDeploymentValidator()},
		Version:     "1.0.0",
		Author:      "System",
		Tags:        []string{"deployment", "webhook", "devops", "ci/cd"},
	}
}

// Validator wrapper implementations to adapt specific validators to the common interface

// GitHubValidatorWrapper wraps the GitHub validator.
type GitHubValidatorWrapper struct {
	validator *validations.GitHubValidator
}

func (gvw *GitHubValidatorWrapper) ValidatePayload(payload interface{}) models.ValidationResult {
	githubPayload, ok := payload.(models.GitHubPayload)
	if !ok {
		return models.ValidationResult{
			IsValid: false,
			Errors: []models.ValidationError{{
				Field:   "payload",
				Message: "payload is not a GitHub payload",
				Code:    "TYPE_MISMATCH",
			}},
		}
	}
	return gvw.validator.ValidatePayload(githubPayload)
}

// GitLabValidatorWrapper wraps the GitLab validator.
type GitLabValidatorWrapper struct {
	validator *validations.GitLabValidator
}

func (gvw *GitLabValidatorWrapper) ValidatePayload(payload interface{}) models.ValidationResult {
	gitlabPayload, ok := payload.(models.GitLabPayload)
	if !ok {
		return models.ValidationResult{
			IsValid: false,
			Errors: []models.ValidationError{{
				Field:   "payload",
				Message: "payload is not a GitLab payload",
				Code:    "TYPE_MISMATCH",
			}},
		}
	}
	return gvw.validator.ValidatePayload(gitlabPayload)
}

// BitbucketValidatorWrapper wraps the Bitbucket validator.
type BitbucketValidatorWrapper struct {
	validator *validations.BitbucketValidator
}

func (bvw *BitbucketValidatorWrapper) ValidatePayload(payload interface{}) models.ValidationResult {
	bitbucketPayload, ok := payload.(models.BitbucketPayload)
	if !ok {
		return models.ValidationResult{
			IsValid: false,
			Errors: []models.ValidationError{{
				Field:   "payload",
				Message: "payload is not a Bitbucket payload",
				Code:    "TYPE_MISMATCH",
			}},
		}
	}
	return bvw.validator.ValidatePayload(bitbucketPayload)
}

// SlackValidatorWrapper wraps the Slack validator.
type SlackValidatorWrapper struct {
	validator *validations.SlackValidator
}

func (svw *SlackValidatorWrapper) ValidatePayload(payload interface{}) models.ValidationResult {
	slackPayload, ok := payload.(models.SlackMessagePayload)
	if !ok {
		return models.ValidationResult{
			IsValid: false,
			Errors: []models.ValidationError{{
				Field:   "payload",
				Message: "payload is not a Slack payload",
				Code:    "TYPE_MISMATCH",
			}},
		}
	}
	return svw.validator.ValidatePayload(slackPayload)
}

// APIValidatorWrapper wraps the API validator.
type APIValidatorWrapper struct {
	validator *validations.APIValidator
}

func (avw *APIValidatorWrapper) ValidatePayload(payload interface{}) models.ValidationResult {
	// API validator can handle both requests and responses
	if apiRequest, ok := payload.(models.APIRequest); ok {
		return avw.validator.ValidateRequest(apiRequest)
	}
	if apiResponse, ok := payload.(models.APIResponse); ok {
		return avw.validator.ValidateResponse(apiResponse)
	}
	return models.ValidationResult{
		IsValid: false,
		Errors: []models.ValidationError{{
			Field:   "payload",
			Message: "payload is not an API request or response",
			Code:    "TYPE_MISMATCH",
		}},
	}
}

// DatabaseValidatorWrapper wraps the Database validator.
type DatabaseValidatorWrapper struct {
	validator *validations.DatabaseValidator
}

func (dvw *DatabaseValidatorWrapper) ValidatePayload(payload interface{}) models.ValidationResult {
	// Database validator can handle both queries and transactions
	if dbQuery, ok := payload.(models.DatabaseQuery); ok {
		return dvw.validator.ValidateQuery(dbQuery)
	}
	if dbTransaction, ok := payload.(models.DatabaseTransaction); ok {
		return dvw.validator.ValidateTransaction(dbTransaction)
	}
	return models.ValidationResult{
		IsValid: false,
		Errors: []models.ValidationError{{
			Field:   "payload",
			Message: "payload is not a database query or transaction",
			Code:    "TYPE_MISMATCH",
		}},
	}
}

// GenericValidatorWrapper wraps the Generic validator.
type GenericValidatorWrapper struct {
	validator *validations.GenericValidator
}

func (gvw *GenericValidatorWrapper) ValidatePayload(payload interface{}) models.ValidationResult {
	// Generic validator can handle multiple payload types
	if genericPayload, ok := payload.(models.GenericPayload); ok {
		return gvw.validator.ValidatePayload(genericPayload)
	}
	if apiModel, ok := payload.(models.APIModel); ok {
		return gvw.validator.ValidateAPIModel(apiModel)
	}
	return models.ValidationResult{
		IsValid: false,
		Errors: []models.ValidationError{{
			Field:   "payload",
			Message: "payload is not a supported generic type",
			Code:    "TYPE_MISMATCH",
		}},
	}
}

// Global registry instance
var globalRegistry *ModelRegistry

// GetGlobalRegistry returns the global model registry instance.
func GetGlobalRegistry() *ModelRegistry {
	if globalRegistry == nil {
		globalRegistry = NewModelRegistry()
	}
	return globalRegistry
}

// Helper functions for common operations

// ValidateWithRegistry validates a payload using the global registry.
func ValidateWithRegistry(modelType ModelType, payload interface{}) (models.ValidationResult, error) {
	return GetGlobalRegistry().ValidatePayload(modelType, payload)
}

// RegisterCustomModel registers a custom model with the global registry.
func RegisterCustomModel(info *ModelInfo) error {
	return GetGlobalRegistry().RegisterModel(info)
}

// GetRegisteredModels returns all registered models from the global registry.
func GetRegisteredModels() map[ModelType]*ModelInfo {
	return GetGlobalRegistry().GetAllModels()
}

// HTTP endpoint management for automatic registration

// RegisterHTTPEndpoints automatically registers HTTP endpoints for all registered models
func (mr *ModelRegistry) RegisterHTTPEndpoints(mux *http.ServeMux) {
	mr.mutex.RLock()
	defer mr.mutex.RUnlock()

	log.Println("🔄 Registering dynamic HTTP endpoints for all models...")

	for modelType, modelInfo := range mr.models {
		endpointPath := "/validate/" + string(modelType)

		// Create a closure to capture the current modelType and modelInfo
		func(mt ModelType, mi *ModelInfo) {
			mux.HandleFunc("POST "+endpointPath, mr.createDynamicHandler(mt, mi))
			log.Printf("✅ Registered endpoint: POST %s -> %s", endpointPath, mi.Name)
		}(modelType, modelInfo)
	}

	log.Printf("🎉 Successfully registered %d dynamic validation endpoints", len(mr.models))
}

// createDynamicHandler creates a dynamic HTTP handler for a specific model type
func (mr *ModelRegistry) createDynamicHandler(modelType ModelType, modelInfo *ModelInfo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set content type
		w.Header().Set("Content-Type", "application/json")

		// Create new instance of the model struct
		modelInstance := reflect.New(modelInfo.ModelStruct).Interface()

		// Parse JSON payload into the model struct
		if err := json.NewDecoder(r.Body).Decode(modelInstance); err != nil {
			sendJSONError(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		// Dereference the pointer to get the actual struct value
		modelValue := reflect.ValueOf(modelInstance).Elem().Interface()

		// Validate using the registry
		result, err := mr.ValidatePayload(modelType, modelValue)
		if err != nil {
			sendJSONError(w, "Validation failed", http.StatusInternalServerError)
			return
		}

		// Set appropriate status code
		if !result.IsValid {
			w.WriteHeader(http.StatusUnprocessableEntity)
		}

		// Encode and send response
		json.NewEncoder(w).Encode(result)
	}
}

// RegisterModelWithEndpoint registers a model and optionally creates its HTTP endpoint
func (mr *ModelRegistry) RegisterModelWithEndpoint(info *ModelInfo, mux *http.ServeMux) error {
	// Register the model first
	if err := mr.RegisterModel(info); err != nil {
		return err
	}

	// Create HTTP endpoint if mux is provided
	if mux != nil {
		endpointPath := "/validate/" + string(info.Type)
		mux.HandleFunc("POST "+endpointPath, mr.createDynamicHandler(info.Type, info))
		log.Printf("✅ Auto-registered endpoint: POST %s -> %s", endpointPath, info.Name)
	}

	return nil
}

// Helper function to send JSON error responses
func sendJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":     message,
		"status":    statusCode,
		"timestamp": "2023-01-01T00:00:00Z", // You might want to use actual timestamp
	})
}

// DeploymentValidatorWrapper wraps the Deployment validator.
type DeploymentValidatorWrapper struct {
	validator *validations.DeploymentValidator
}

func (dvw *DeploymentValidatorWrapper) ValidatePayload(payload interface{}) models.ValidationResult {
	deploymentPayload, ok := payload.(models.DeploymentPayload)
	if !ok {
		return models.ValidationResult{
			IsValid: false,
			Errors: []models.ValidationError{{
				Field:   "payload",
				Message: "payload is not a Deployment payload",
				Code:    "TYPE_MISMATCH",
			}},
		}
	}
	return dvw.validator.ValidatePayload(deploymentPayload)
}

// RegisterCustomModelWithEndpoint registers a custom model with automatic HTTP endpoint
func RegisterCustomModelWithEndpoint(info *ModelInfo, mux *http.ServeMux) error {
	return GetGlobalRegistry().RegisterModelWithEndpoint(info, mux)
}
