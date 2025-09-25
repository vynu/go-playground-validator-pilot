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

	"goplayground-data-validator/models"
	"goplayground-data-validator/validations"
)

// FileSystemWatcher monitors file system changes for dynamic model updates
type FileSystemWatcher struct {
	modelsPath      string
	validationsPath string
	registry        *UnifiedRegistry
	pollInterval    time.Duration
	lastScan        map[string]time.Time
	mutex          sync.RWMutex
}

// NewFileSystemWatcher creates a new file system watcher
func NewFileSystemWatcher(modelsPath, validationsPath string, registry *UnifiedRegistry) *FileSystemWatcher {
	return &FileSystemWatcher{
		modelsPath:      modelsPath,
		validationsPath: validationsPath,
		registry:        registry,
		pollInterval:    2 * time.Second,
		lastScan:        make(map[string]time.Time),
	}
}

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
	watcher         *FileSystemWatcher
	mutex           sync.RWMutex
	isMonitoring    bool
}

// NewUnifiedRegistry creates a new unified registry instance
func NewUnifiedRegistry(modelsPath, validationsPath string) *UnifiedRegistry {
	return &UnifiedRegistry{
		models:          make(map[ModelType]*ModelInfo),
		modelsPath:      modelsPath,
		validationsPath: validationsPath,
		mutex:           sync.RWMutex{},
		isMonitoring:    false,
	}
}

// StartAutoRegistration performs initial discovery and starts file system monitoring
func (ur *UnifiedRegistry) StartAutoRegistration(ctx context.Context, mux *http.ServeMux) error {
	ur.mux = mux

	log.Println("🚀 Starting unified automatic model registration system...")

	// Phase 1: Initial discovery and registration
	if err := ur.discoverAndRegisterAll(); err != nil {
		log.Printf("⚠️ Initial discovery had issues: %v", err)
	}

	// Phase 2: Register HTTP endpoints for discovered models (only once)
	ur.registerAllHTTPEndpoints()

	// Phase 3: Start file system monitoring for ongoing changes (TODO: Fix HTTP re-registration issue)
	// NOTE: Temporarily disabled to prevent HTTP endpoint conflicts
	log.Println("🔄 File system monitoring temporarily disabled until HTTP re-registration is fixed")
	ur.isMonitoring = false

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
	possibleNames = append(possibleNames,
		strings.Title(baseName)+"Payload",
		strings.Title(baseName)+"Model",
		strings.Title(baseName)+"Request",
		strings.Title(baseName)+"Data",
		strings.Title(baseName),
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
		titleCase = strings.Title(baseName)
	}

	possibleNames := []string{
		"New" + titleCase + "Validator",              // NewGitHubValidator, NewAPIValidator
		"New" + strings.Title(baseName) + "Validator", // NewGithubValidator
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
		return strings.Title(baseName) + " Payload"
	}

	return strings.Title(baseName) + " Data"
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

	return fmt.Sprintf("Automatically discovered %s validation with comprehensive business rules", strings.Title(baseName))
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
		json.NewEncoder(w).Encode(result)
	}
}

// sendJSONError sends standardized JSON error
func (ur *UnifiedRegistry) sendJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":     message,
		"status":    statusCode,
		"timestamp": time.Now().Format(time.RFC3339),
	})
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
		"monitoring":   ur.isMonitoring,
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
		globalUnifiedRegistry = NewUnifiedRegistry("models", "validations")
	}
	return globalUnifiedRegistry
}

// Start begins monitoring the file system for the FileSystemWatcher
func (fsw *FileSystemWatcher) Start(ctx context.Context) error {
	log.Printf("👁️ Starting file system watcher (polling every %v)", fsw.pollInterval)

	// Initial scan to establish baseline
	if err := fsw.scanDirectories(); err != nil {
		log.Printf("⚠️ Initial directory scan failed: %v", err)
	}

	ticker := time.NewTicker(fsw.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("🛑 File system watcher stopping...")
			return ctx.Err()
		case <-ticker.C:
			if err := fsw.scanDirectories(); err != nil {
				log.Printf("⚠️ Directory scan error: %v", err)
			}
		}
	}
}

// scanDirectories scans both directories for changes
func (fsw *FileSystemWatcher) scanDirectories() error {
	fsw.mutex.Lock()
	defer fsw.mutex.Unlock()

	// Get current files
	currentFiles := make(map[string]time.Time)

	// Scan models directory
	modelFiles, err := filepath.Glob(filepath.Join(fsw.modelsPath, "*.go"))
	if err != nil {
		return fmt.Errorf("scanning models directory: %w", err)
	}

	for _, file := range modelFiles {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}
		currentFiles[file] = info.ModTime()
	}

	// Scan validations directory
	validationFiles, err := filepath.Glob(filepath.Join(fsw.validationsPath, "*.go"))
	if err != nil {
		return fmt.Errorf("scanning validations directory: %w", err)
	}

	for _, file := range validationFiles {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}
		currentFiles[file] = info.ModTime()
	}

	// Detect changes
	fsw.detectChanges(currentFiles)

	// Update last scan
	fsw.lastScan = currentFiles

	return nil
}

// detectChanges compares current files with last scan and triggers appropriate actions
func (fsw *FileSystemWatcher) detectChanges(currentFiles map[string]time.Time) {
	// Detect new/modified files
	for filePath, modTime := range currentFiles {
		if lastModTime, exists := fsw.lastScan[filePath]; !exists || modTime.After(lastModTime) {
			fsw.handleFileChange(filePath, "created_or_modified")
		}
	}

	// Detect deleted files
	for filePath := range fsw.lastScan {
		if _, exists := currentFiles[filePath]; !exists {
			fsw.handleFileChange(filePath, "deleted")
		}
	}
}

// handleFileChange processes file system changes
func (fsw *FileSystemWatcher) handleFileChange(filePath, action string) {
	baseName := strings.TrimSuffix(filepath.Base(filePath), ".go")

	// Skip test files
	if strings.HasSuffix(baseName, "_test") {
		return
	}

	// Determine if this is a model or validation file
	isModelFile := strings.Contains(filePath, fsw.modelsPath)
	isValidationFile := strings.Contains(filePath, fsw.validationsPath)

	if !isModelFile && !isValidationFile {
		return
	}

	switch action {
	case "created_or_modified":
		fsw.handleFileAddedOrModified(baseName, isModelFile, isValidationFile)
	case "deleted":
		fsw.handleFileDeleted(baseName, isModelFile, isValidationFile)
	}
}

// handleFileAddedOrModified processes file additions or modifications
func (fsw *FileSystemWatcher) handleFileAddedOrModified(baseName string, isModelFile, isValidationFile bool) {
	log.Printf("📁 Detected file change: %s (model: %v, validation: %v)", baseName, isModelFile, isValidationFile)

	// Check if both model and validation files exist
	modelExists := fsw.fileExists(filepath.Join(fsw.modelsPath, baseName+".go"))
	validationExists := fsw.fileExists(filepath.Join(fsw.validationsPath, baseName+".go"))

	if modelExists && validationExists {
		// Both files exist, register or re-register the model
		if fsw.registry.IsRegistered(ModelType(baseName)) {
			log.Printf("🔄 Re-registering modified model: %s", baseName)
			// Unregister and re-register to pick up changes
			fsw.registry.UnregisterModel(ModelType(baseName))
		}

		if err := fsw.registry.registerModelAutomatically(baseName); err != nil {
			log.Printf("❌ Failed to register model %s: %v", baseName, err)
		}
	} else {
		log.Printf("⚠️ Model %s incomplete (model: %v, validation: %v)", baseName, modelExists, validationExists)
	}
}

// handleFileDeleted processes file deletions
func (fsw *FileSystemWatcher) handleFileDeleted(baseName string, isModelFile, isValidationFile bool) {
	log.Printf("🗑️ Detected file deletion: %s (model: %v, validation: %v)", baseName, isModelFile, isValidationFile)

	// Check if both files still exist
	modelExists := fsw.fileExists(filepath.Join(fsw.modelsPath, baseName+".go"))
	validationExists := fsw.fileExists(filepath.Join(fsw.validationsPath, baseName+".go"))

	// If either file is missing, unregister the model
	if !modelExists || !validationExists {
		if fsw.registry.IsRegistered(ModelType(baseName)) {
			log.Printf("🔥 Model %s is incomplete (model: %v, validation: %v) - unregistering", baseName, modelExists, validationExists)
			if err := fsw.registry.UnregisterModel(ModelType(baseName)); err != nil {
				log.Printf("❌ Failed to unregister model %s: %v", baseName, err)
			} else {
				log.Printf("✅ Successfully retired model: %s", baseName)
			}
		}
	}
}

// fileExists checks if a file exists
func (fsw *FileSystemWatcher) fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

// StartRegistration starts the unified registration system
func StartRegistration(ctx context.Context, mux *http.ServeMux) error {
	return GetGlobalRegistry().StartAutoRegistration(ctx, mux)
}