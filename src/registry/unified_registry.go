// Package registry provides a unified, fully automatic model registration system.
// This system combines auto-discovery, file monitoring, and HTTP endpoint management
// into a single, efficient implementation with zero configuration required.
package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"
	"unicode"

	"goplayground-data-validator/models"
	"goplayground-data-validator/validations"
)

// Removed FileSystemWatcher - keeping the system simple with startup-only registration

// UnifiedRegistry is the single, consolidated registry system that handles:
// - Automatic model discovery and registration
// - File system monitoring for dynamic updates
// - HTTP endpoint creation and management
// - Universal validation with any validator type
type UnifiedRegistry struct {
	models          map[ModelType]*ModelInfo
	modelsPath      string
	validationsPath string
	mux             *http.ServeMux
	mutex           sync.RWMutex
}

// NewUnifiedRegistry creates a new unified registry instance
func NewUnifiedRegistry(modelsPath, validationsPath string) *UnifiedRegistry {
	return &UnifiedRegistry{
		models:          make(map[ModelType]*ModelInfo),
		modelsPath:      modelsPath,
		validationsPath: validationsPath,
		mutex:           sync.RWMutex{},
	}
}

// StartAutoRegistration performs initial discovery and starts file system monitoring
func (ur *UnifiedRegistry) StartAutoRegistration(_ context.Context, mux *http.ServeMux) error {
	ur.mux = mux

	log.Println("🚀 Starting unified automatic model registration system...")

	// Phase 1: Initial discovery and registration
	if err := ur.discoverAndRegisterAll(); err != nil {
		log.Printf("⚠️ Initial discovery had issues: %v", err)
	}

	// Phase 2: Register HTTP endpoints for discovered models (only once)
	ur.registerAllHTTPEndpoints()

	// Phase 3: File system monitoring removed - keeping it simple with startup-only registration
	log.Println("✅ Pure auto-registration completed - models will be discovered on each startup")

	return nil
}

// discoverAndRegisterAll scans directories and registers all found models
func (ur *UnifiedRegistry) discoverAndRegisterAll() error {
	log.Println("🔍 Starting comprehensive model discovery...")

	modelFiles, err := filepath.Glob(filepath.Join(ur.modelsPath, "*.go"))
	if err != nil {
		return fmt.Errorf("scanning models directory: %w", err)
	}

	registered := 0
	var errors []string

	for _, modelFile := range modelFiles {
		baseName := strings.TrimSuffix(filepath.Base(modelFile), ".go")

		// Skip test files and package files
		if strings.HasSuffix(baseName, "_test") || baseName == "models" {
			continue
		}

		// Check if corresponding validator exists
		validatorFile := filepath.Join(ur.validationsPath, baseName+".go")
		if _, err := os.Stat(validatorFile); os.IsNotExist(err) {
			log.Printf("⚠️ No validator found for model: %s (skipping)", baseName)
			continue
		}

		// Attempt automatic registration
		if err := ur.registerModelAutomatically(baseName); err != nil {
			log.Printf("❌ Failed to register model %s: %v", baseName, err)
			errors = append(errors, fmt.Sprintf("%s: %v", baseName, err))
			continue
		}

		registered++
		log.Printf("✅ Auto-registered model: %s", baseName)
	}

	log.Printf("🎉 Discovery completed: %d models registered", registered)

	if len(errors) > 0 {
		log.Printf("⚠️ %d models had registration issues:", len(errors))
		for _, errMsg := range errors {
			log.Printf("   - %s", errMsg)
		}
	}

	return nil
}

// registerModelAutomatically discovers and registers a single model
func (ur *UnifiedRegistry) registerModelAutomatically(baseName string) error {
	log.Printf("🔍 Auto-registering model: %s", baseName)

	// Step 1: Find model struct type
	modelStruct, structName, err := ur.discoverModelStruct(baseName)
	if err != nil {
		return fmt.Errorf("discovering model struct: %w", err)
	}

	// Step 2: Create validator instance
	validatorInstance, err := ur.createValidatorInstance(baseName)
	if err != nil {
		return fmt.Errorf("creating validator: %w", err)
	}

	// Step 3: Create universal wrapper
	wrapper := &UniversalValidatorWrapper{
		modelType:         baseName,
		validatorInstance: validatorInstance,
		modelStructType:   modelStruct,
	}

	// Step 4: Create model info with metadata
	modelInfo := &ModelInfo{
		Type:        ModelType(baseName),
		Name:        ur.generateModelName(baseName, structName),
		Description: ur.generateModelDescription(baseName),
		ModelStruct: modelStruct,
		Validator:   wrapper,
		Version:     "1.0.0",
		CreatedAt:   time.Now().Format(time.RFC3339),
		Author:      "Unified Auto-Registry",
		Tags:        ur.generateModelTags(baseName),
	}

	// Step 5: Register the model
	return ur.RegisterModel(modelInfo)
}

// discoverModelStruct finds the correct struct type for a model
func (ur *UnifiedRegistry) discoverModelStruct(baseName string) (reflect.Type, string, error) {
	// Parse the Go file to find struct names
	modelFile := filepath.Join(ur.modelsPath, baseName+".go")
	discoveredStructs, err := ur.parseGoFileForStructs(modelFile)
	if err != nil {
		log.Printf("⚠️ Could not parse %s: %v", modelFile, err)
	}

	// Try different naming patterns
	possibleNames := discoveredStructs
	titleCase := toTitleCase(baseName)
	possibleNames = append(possibleNames,
		titleCase+"Payload",
		titleCase+"Model",
		titleCase+"Request",
		titleCase+"Data",
		titleCase,
	)

	// Get known model types
	knownTypes := ur.getKnownModelTypes()

	for _, name := range possibleNames {
		if structType, exists := knownTypes[name]; exists {
			log.Printf("✅ Found model struct: %s -> %s", baseName, name)
			return structType, name, nil
		}
	}

	return nil, "", fmt.Errorf("no model struct found for %s (tried: %v)", baseName, possibleNames)
}

// parseGoFileForStructs extracts struct names from a Go file
func (ur *UnifiedRegistry) parseGoFileForStructs(filename string) ([]string, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var structNames []string
	ast.Inspect(node, func(n ast.Node) bool {
		if typeSpec, ok := n.(*ast.TypeSpec); ok {
			if _, ok := typeSpec.Type.(*ast.StructType); ok {
				structNames = append(structNames, typeSpec.Name.Name)
			}
		}
		return true
	})

	return structNames, nil
}

// createValidatorInstance creates validator using naming conventions
func (ur *UnifiedRegistry) createValidatorInstance(baseName string) (interface{}, error) {
	// Try multiple naming patterns
	// Handle special cases first
	specialCases := map[string]string{
		"github": "GitHub",
		"api":    "API",
		"db":     "DB",
		"http":   "HTTP",
		"json":   "JSON",
		"xml":    "XML",
		"url":    "URL",
	}

	var titleCase string
	if special, exists := specialCases[baseName]; exists {
		titleCase = special
	} else {
		titleCase = toTitleCase(baseName)
	}

	possibleNames := []string{
		"New" + titleCase + "Validator",                 // NewGitHubValidator, NewAPIValidator
		"New" + toTitleCase(baseName) + "Validator",     // NewGithubValidator
		"New" + strings.ToUpper(baseName) + "Validator", // NewGITHUBValidator
	}

	knownValidators := ur.getKnownValidatorConstructors()

	for _, constructorName := range possibleNames {
		log.Printf("🔍 Looking for: %s", constructorName)
		if constructor, exists := knownValidators[constructorName]; exists {
			log.Printf("✅ Found validator constructor: %s", constructorName)
			return constructor(), nil
		}
	}

	return nil, fmt.Errorf("no validator constructor found (tried: %v)", possibleNames)
}

// getKnownModelTypes returns available model types via reflection
func (ur *UnifiedRegistry) getKnownModelTypes() map[string]reflect.Type {
	return map[string]reflect.Type{
		"IncidentPayload":   reflect.TypeOf(models.IncidentPayload{}),
		"GitHubPayload":     reflect.TypeOf(models.GitHubPayload{}),
		"APIRequest":        reflect.TypeOf(models.APIRequest{}),
		"DatabaseQuery":     reflect.TypeOf(models.DatabaseQuery{}),
		"GenericPayload":    reflect.TypeOf(models.GenericPayload{}),
		"DeploymentPayload": reflect.TypeOf(models.DeploymentPayload{}),
	}
}

// getKnownValidatorConstructors returns available validator constructors
func (ur *UnifiedRegistry) getKnownValidatorConstructors() map[string]func() interface{} {
	return map[string]func() interface{}{
		"NewIncidentValidator":   func() interface{} { return validations.NewIncidentValidator() },
		"NewGitHubValidator":     func() interface{} { return validations.NewGitHubValidator() },
		"NewAPIValidator":        func() interface{} { return validations.NewAPIValidator() },
		"NewDatabaseValidator":   func() interface{} { return validations.NewDatabaseValidator() },
		"NewGenericValidator":    func() interface{} { return validations.NewGenericValidator() },
		"NewDeploymentValidator": func() interface{} { return validations.NewDeploymentValidator() },
	}
}

// generateModelName creates human-readable model names
func (ur *UnifiedRegistry) generateModelName(baseName, structName string) string {
	contextualNames := map[string]string{
		"github":     "GitHub Webhook",
		"api":        "API Request/Response",
		"database":   "Database Operations",
		"generic":    "Generic Payload",
		"deployment": "Deployment Webhook",
		"incident":   "Incident Report",
	}

	if name, exists := contextualNames[strings.ToLower(baseName)]; exists {
		return name
	}

	if strings.Contains(strings.ToLower(structName), "payload") {
		return toTitleCase(baseName) + " Payload"
	}

	return toTitleCase(baseName) + " Data"
}

// generateModelDescription creates model descriptions
func (ur *UnifiedRegistry) generateModelDescription(baseName string) string {
	contextualDescriptions := map[string]string{
		"github":     "GitHub webhook payload validation with comprehensive business rules",
		"api":        "API request and response validation with comprehensive business rules",
		"database":   "Database query and transaction validation with comprehensive business rules",
		"generic":    "Generic payload validation with flexible business rules",
		"deployment": "Deployment webhook payload validation with semantic versioning and business rules",
		"incident":   "Incident report validation with operational context and business rules",
	}

	if desc, exists := contextualDescriptions[strings.ToLower(baseName)]; exists {
		return desc
	}

	return fmt.Sprintf("Automatically discovered %s validation with comprehensive business rules", toTitleCase(baseName))
}

// generateModelTags creates appropriate tags
func (ur *UnifiedRegistry) generateModelTags(baseName string) []string {
	baseTags := []string{"auto-discovered", strings.ToLower(baseName)}

	contextualTags := map[string][]string{
		"github":     {"webhook", "github", "git", "collaboration"},
		"api":        {"api", "http", "rest", "web"},
		"database":   {"database", "sql", "transaction", "query"},
		"generic":    {"generic", "flexible", "json", "general"},
		"deployment": {"deployment", "webhook", "devops", "ci/cd"},
		"incident":   {"incident", "monitoring", "alert", "operations"},
	}

	if tags, exists := contextualTags[strings.ToLower(baseName)]; exists {
		return append(baseTags, tags...)
	}

	return append(baseTags, "custom", "flexible")
}

// RegisterModel registers a model with the unified registry
func (ur *UnifiedRegistry) RegisterModel(info *ModelInfo) error {
	ur.mutex.Lock()
	defer ur.mutex.Unlock()

	if info.Type == "" {
		return fmt.Errorf("model type cannot be empty")
	}

	ur.models[info.Type] = info
	log.Printf("✅ Registered model: %s -> %s", info.Type, info.Name)
	return nil
}

// UnregisterModel removes a model from the registry
func (ur *UnifiedRegistry) UnregisterModel(modelType ModelType) error {
	ur.mutex.Lock()
	defer ur.mutex.Unlock()

	if _, exists := ur.models[modelType]; !exists {
		return fmt.Errorf("model type '%s' not found", modelType)
	}

	delete(ur.models, modelType)
	log.Printf("🗑️ Unregistered model: %s", modelType)
	return nil
}

// GetModel retrieves model information
func (ur *UnifiedRegistry) GetModel(modelType ModelType) (*ModelInfo, error) {
	ur.mutex.RLock()
	defer ur.mutex.RUnlock()

	model, exists := ur.models[modelType]
	if !exists {
		return nil, fmt.Errorf("model type '%s' not found", modelType)
	}

	return model, nil
}

// GetAllModels returns all registered models
func (ur *UnifiedRegistry) GetAllModels() map[ModelType]*ModelInfo {
	ur.mutex.RLock()
	defer ur.mutex.RUnlock()

	result := make(map[ModelType]*ModelInfo)
	for k, v := range ur.models {
		result[k] = v
	}
	return result
}

// ListModels returns all model types
func (ur *UnifiedRegistry) ListModels() []ModelType {
	ur.mutex.RLock()
	defer ur.mutex.RUnlock()

	var types []ModelType
	for modelType := range ur.models {
		types = append(types, modelType)
	}
	return types
}

// IsRegistered checks if a model type is registered
func (ur *UnifiedRegistry) IsRegistered(modelType ModelType) bool {
	ur.mutex.RLock()
	defer ur.mutex.RUnlock()

	_, exists := ur.models[modelType]
	return exists
}

// GetValidator retrieves validator for a model type
func (ur *UnifiedRegistry) GetValidator(modelType ModelType) (ValidatorInterface, error) {
	model, err := ur.GetModel(modelType)
	if err != nil {
		return nil, err
	}
	return model.Validator, nil
}

// ValidatePayload validates payload using appropriate validator
func (ur *UnifiedRegistry) ValidatePayload(modelType ModelType, payload interface{}) (interface{}, error) {
	validator, err := ur.GetValidator(modelType)
	if err != nil {
		return nil, err
	}
	return validator.ValidatePayload(payload), nil
}

// ValidateArray validates an array of records and returns structured results
// Only invalid rows are included in the results array (successful validations return empty results)
// Status is determined by threshold: if no threshold provided, status is "success" for single records
// For multiple records with threshold, status is "success" if success_rate >= threshold, otherwise "failed"
func (ur *UnifiedRegistry) ValidateArray(modelType ModelType, records []map[string]interface{}, threshold *float64) (*models.ArrayValidationResult, error) {
	// Generate batch_id for tracking
	batchID := models.GenerateBatchID("auto")
	startTime := time.Now()

	allResults := make([]models.RowValidationResult, len(records))
	validCount := 0
	invalidCount := 0
	warningCount := 0

	// Get model info for struct creation
	modelInfo, err := ur.GetModel(modelType)
	if err != nil {
		return nil, fmt.Errorf("model type not found: %w", err)
	}

	// Sequential validation (can be optimized later with worker pool)
	for i, record := range records {
		rowResult := ur.validateSingleRow(modelType, modelInfo, record, i)
		allResults[i] = rowResult

		if rowResult.IsValid {
			validCount++
			// Check if it has warnings only (valid but with warnings)
			if len(rowResult.Warnings) > 0 {
				warningCount++
			}
		} else {
			invalidCount++
		}
	}

	// Filter results: only include invalid rows (failed validations)
	// Successful validations should NOT return any results
	filteredResults := make([]models.RowValidationResult, 0)
	for _, result := range allResults {
		// Only include if invalid (IsValid == false)
		if !result.IsValid {
			filteredResults = append(filteredResults, result)
		}
	}

	// Calculate success rate
	totalRecords := len(records)
	successRate := 0.0
	if totalRecords > 0 {
		successRate = (float64(validCount) / float64(totalRecords)) * 100.0
	}

	// Determine status based on threshold logic
	status := "success" // default status

	if threshold != nil {
		// Threshold is provided - apply strict comparison
		// success_rate >= threshold means success, otherwise failed
		if successRate < *threshold {
			status = "failed"
		}
	} else {
		// No threshold provided
		// For single record: "success" if valid, "failed" if invalid
		// For multiple records: "success" (no threshold means we don't fail the batch)
		if totalRecords == 1 && invalidCount > 0 {
			status = "failed"
		}
		// For multiple records without threshold, status remains "success"
	}

	arrayResult := &models.ArrayValidationResult{
		BatchID:        batchID,
		Status:         status,
		TotalRecords:   totalRecords,
		ValidRecords:   validCount,
		InvalidRecords: invalidCount,
		WarningRecords: warningCount,
		Threshold:      threshold,
		ProcessingTime: time.Since(startTime).Milliseconds(),
		CompletedAt:    time.Now(),
		Results:        filteredResults,                 // Only invalid rows (successful validations excluded)
		Summary:        models.BuildSummary(allResults), // Summary includes all rows
	}

	return arrayResult, nil
}

// validateSingleRow validates a single row from an array
func (ur *UnifiedRegistry) validateSingleRow(modelType ModelType, modelInfo *ModelInfo, record map[string]interface{}, rowIndex int) models.RowValidationResult {
	rowStartTime := time.Now()
	recordID := models.DetectRecordIdentifier(record, rowIndex)

	// Generate test name from model type (e.g., "incident" -> "IncidentValidator")
	testName := fmt.Sprintf("%sValidator", toTitleCase(string(modelType)))

	// Helper to create error result
	createErrorResult := func(code, message string) models.RowValidationResult {
		return models.RowValidationResult{
			RowIndex:         rowIndex,
			RecordIdentifier: recordID,
			IsValid:          false,
			ValidationTime:   time.Since(rowStartTime).Milliseconds(),
			TestName:         testName,
			Errors: []models.ValidationError{{
				Field:   "record",
				Message: message,
				Code:    code,
			}},
			Warnings: []models.ValidationWarning{},
		}
	}

	// Create and populate model instance
	modelInstance := reflect.New(modelInfo.ModelStruct).Interface()

	jsonBytes, err := json.Marshal(record)
	if err != nil {
		return createErrorResult("JSON_MARSHAL_ERROR", fmt.Sprintf("Failed to marshal record: %v", err))
	}

	if err := json.Unmarshal(jsonBytes, modelInstance); err != nil {
		return createErrorResult("JSON_UNMARSHAL_ERROR", fmt.Sprintf("Failed to unmarshal record: %v", err))
	}

	// Validate using existing validator
	modelValue := reflect.ValueOf(modelInstance).Elem().Interface()
	result, err := ur.ValidatePayload(modelType, modelValue)
	if err != nil {
		return createErrorResult("VALIDATION_ERROR", fmt.Sprintf("Validation failed: %v", err))
	}

	// Convert validation result to row result
	rowResult := models.RowValidationResult{
		RowIndex:         rowIndex,
		RecordIdentifier: recordID,
		ValidationTime:   time.Since(rowStartTime).Milliseconds(),
		TestName:         testName,
		Errors:           []models.ValidationError{},
		Warnings:         []models.ValidationWarning{},
	}

	// Extract validation result fields
	if validationResult, ok := result.(models.ValidationResult); ok {
		rowResult.IsValid = validationResult.IsValid
		rowResult.Errors = validationResult.Errors
		rowResult.Warnings = validationResult.Warnings
	} else if resultMap, ok := result.(map[string]interface{}); ok {
		// Handle map-based validation result
		if isValid, exists := resultMap["is_valid"]; exists {
			if valid, ok := isValid.(bool); ok {
				rowResult.IsValid = valid
			}
		}

		// Extract errors
		if errors, exists := resultMap["errors"]; exists {
			if errSlice, ok := errors.([]models.ValidationError); ok {
				rowResult.Errors = errSlice
			}
		}

		// Extract warnings
		if warnings, exists := resultMap["warnings"]; exists {
			if warnSlice, ok := warnings.([]models.ValidationWarning); ok {
				rowResult.Warnings = warnSlice
			}
		}
	}

	// Add sub-test categorization based on error/warning codes
	if !rowResult.IsValid && len(rowResult.Errors) > 0 {
		// Determine sub-test from primary error code
		primaryErrorCode := rowResult.Errors[0].Code
		rowResult.TestName = fmt.Sprintf("%s:%s", testName, primaryErrorCode)
	} else if len(rowResult.Warnings) > 0 {
		// If valid but has warnings, categorize by warning code
		primaryWarningCode := rowResult.Warnings[0].Code
		rowResult.TestName = fmt.Sprintf("%s:%s", testName, primaryWarningCode)
	}

	return rowResult
}

// CreateModelInstance creates new instance of model struct
func (ur *UnifiedRegistry) CreateModelInstance(modelType ModelType) (interface{}, error) {
	model, err := ur.GetModel(modelType)
	if err != nil {
		return nil, err
	}

	modelValue := reflect.New(model.ModelStruct).Interface()
	return modelValue, nil
}

// registerAllHTTPEndpoints creates HTTP endpoints for all registered models
func (ur *UnifiedRegistry) registerAllHTTPEndpoints() {
	if ur.mux == nil {
		log.Println("⚠️ No HTTP mux provided, skipping endpoint registration")
		return
	}

	ur.mutex.RLock()
	defer ur.mutex.RUnlock()

	log.Println("🔄 Registering HTTP endpoints for all models...")

	for modelType, modelInfo := range ur.models {
		endpointPath := "/validate/" + string(modelType)

		// Create closure to capture variables properly
		func(mt ModelType, mi *ModelInfo, path string) {
			ur.mux.HandleFunc("POST "+path, ur.createDynamicHandler(mt, mi))
			log.Printf("✅ Registered endpoint: POST %s -> %s", path, mi.Name)
		}(modelType, modelInfo, endpointPath)
	}

	log.Printf("🎉 Successfully registered %d HTTP endpoints", len(ur.models))
}

// createDynamicHandler creates HTTP handler for a specific model
func (ur *UnifiedRegistry) createDynamicHandler(modelType ModelType, modelInfo *ModelInfo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Ensure request body is closed and cleaned up
		defer r.Body.Close()

		w.Header().Set("Content-Type", "application/json")

		// Create model instance
		modelInstance := reflect.New(modelInfo.ModelStruct).Interface()

		// Parse JSON into model
		if err := json.NewDecoder(r.Body).Decode(modelInstance); err != nil {
			ur.sendJSONError(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		// Get actual struct value
		modelValue := reflect.ValueOf(modelInstance).Elem().Interface()

		// Validate
		result, err := ur.ValidatePayload(modelType, modelValue)
		if err != nil {
			ur.sendJSONError(w, "Validation failed", http.StatusInternalServerError)
			return
		}

		// Send response
		if err := json.NewEncoder(w).Encode(result); err != nil {
			log.Printf("Error encoding response: %v", err)
		}
	}
}

// sendJSONError sends standardized JSON error
func (ur *UnifiedRegistry) sendJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"error":     message,
		"status":    statusCode,
		"timestamp": time.Now().Format(time.RFC3339),
	}); err != nil {
		log.Printf("Error encoding error response: %v", err)
	}
}

// GetModelStats returns registry statistics
func (ur *UnifiedRegistry) GetModelStats() map[string]interface{} {
	ur.mutex.RLock()
	defer ur.mutex.RUnlock()

	modelTypes := make([]string, 0, len(ur.models))
	for modelType := range ur.models {
		modelTypes = append(modelTypes, string(modelType))
	}

	return map[string]interface{}{
		"total_models": len(ur.models),
		"model_types":  modelTypes,
		"monitoring":   false,
	}
}

// GetRegisteredModelsWithDetails returns detailed model information
func (ur *UnifiedRegistry) GetRegisteredModelsWithDetails() map[string]interface{} {
	ur.mutex.RLock()
	defer ur.mutex.RUnlock()

	details := make(map[string]interface{})
	for modelType, modelInfo := range ur.models {
		details[string(modelType)] = map[string]interface{}{
			"name":        modelInfo.Name,
			"description": modelInfo.Description,
			"version":     modelInfo.Version,
			"author":      modelInfo.Author,
			"tags":        modelInfo.Tags,
			"created_at":  modelInfo.CreatedAt,
			"endpoint":    "/validate/" + string(modelType),
		}
	}

	return map[string]interface{}{
		"models": details,
		"count":  len(details),
		"source": "unified-registry",
	}
}

// Global unified registry instance
var globalUnifiedRegistry *UnifiedRegistry

// GetGlobalRegistry returns the global unified registry
func GetGlobalRegistry() *UnifiedRegistry {
	if globalUnifiedRegistry == nil {
		globalUnifiedRegistry = NewUnifiedRegistry("src/models", "src/validations")
	}
	return globalUnifiedRegistry
}

// toTitleCase converts a string to title case (replacement for deprecated strings.Title)
func toTitleCase(s string) string {
	if s == "" {
		return ""
	}

	runes := []rune(s)
	result := make([]rune, len(runes))

	makeUpper := true
	for i, r := range runes {
		if unicode.IsSpace(r) || r == '_' || r == '-' {
			result[i] = r
			makeUpper = true
		} else if makeUpper {
			result[i] = unicode.ToUpper(r)
			makeUpper = false
		} else {
			result[i] = unicode.ToLower(r)
		}
	}

	return string(result)
}

// FileSystemWatcher methods removed - using simple startup-only registration

// StartRegistration starts the unified registration system
func StartRegistration(ctx context.Context, mux *http.ServeMux) error {
	return GetGlobalRegistry().StartAutoRegistration(ctx, mux)
}
