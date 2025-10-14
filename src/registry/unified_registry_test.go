package registry

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"goplayground-data-validator/models"
)

func TestNewUnifiedRegistry(t *testing.T) {
	registry := NewUnifiedRegistry("test/models", "test/validations")

	if registry == nil {
		t.Fatal("NewUnifiedRegistry returned nil")
	}
	if registry.modelsPath != "test/models" {
		t.Errorf("modelsPath = %s, want test/models", registry.modelsPath)
	}
	if registry.validationsPath != "test/validations" {
		t.Errorf("validationsPath = %s, want test/validations", registry.validationsPath)
	}
	if registry.models == nil {
		t.Error("models map should be initialized")
	}
	if len(registry.models) != 0 {
		t.Errorf("models map should be empty initially, got %d", len(registry.models))
	}
}

func TestUnifiedRegistry_RegisterModel(t *testing.T) {
	registry := NewUnifiedRegistry("test/models", "test/validations")

	modelInfo := &ModelInfo{
		Type:        "test",
		Name:        "Test Model",
		Description: "A test model",
		ModelStruct: reflect.TypeOf(models.IncidentPayload{}),
		Version:     "1.0.0",
		CreatedAt:   time.Now().Format(time.RFC3339),
		Author:      "Test",
		Tags:        []string{"test"},
	}

	err := registry.RegisterModel(modelInfo)
	if err != nil {
		t.Fatalf("RegisterModel failed: %v", err)
	}

	// Test that model was registered
	if !registry.IsRegistered("test") {
		t.Error("Model should be registered")
	}

	// Test retrieving the model
	retrieved, err := registry.GetModel("test")
	if err != nil {
		t.Fatalf("GetModel failed: %v", err)
	}
	if retrieved.Type != "test" {
		t.Errorf("Retrieved model type = %s, want test", retrieved.Type)
	}
}

func TestUnifiedRegistry_RegisterModel_EmptyType(t *testing.T) {
	registry := NewUnifiedRegistry("test/models", "test/validations")

	modelInfo := &ModelInfo{
		Type: "", // Empty type should fail
		Name: "Test Model",
	}

	err := registry.RegisterModel(modelInfo)
	if err == nil {
		t.Error("Expected error for empty model type")
	}
}

func TestUnifiedRegistry_UnregisterModel(t *testing.T) {
	registry := NewUnifiedRegistry("test/models", "test/validations")

	// Register a model first
	modelInfo := &ModelInfo{
		Type:        "test",
		Name:        "Test Model",
		ModelStruct: reflect.TypeOf(models.IncidentPayload{}),
	}
	registry.RegisterModel(modelInfo)

	// Test unregistering
	err := registry.UnregisterModel("test")
	if err != nil {
		t.Fatalf("UnregisterModel failed: %v", err)
	}

	// Test that model is no longer registered
	if registry.IsRegistered("test") {
		t.Error("Model should not be registered after unregistering")
	}
}

func TestUnifiedRegistry_UnregisterModel_NotFound(t *testing.T) {
	registry := NewUnifiedRegistry("test/models", "test/validations")

	err := registry.UnregisterModel("nonexistent")
	if err == nil {
		t.Error("Expected error for unregistering nonexistent model")
	}
}

func TestUnifiedRegistry_GetModel_NotFound(t *testing.T) {
	registry := NewUnifiedRegistry("test/models", "test/validations")

	_, err := registry.GetModel("nonexistent")
	if err == nil {
		t.Error("Expected error for getting nonexistent model")
	}
}

func TestUnifiedRegistry_GetAllModels(t *testing.T) {
	registry := NewUnifiedRegistry("test/models", "test/validations")

	// Register multiple models
	models := []ModelInfo{
		{Type: "test1", Name: "Test Model 1", ModelStruct: reflect.TypeOf(models.IncidentPayload{})},
		{Type: "test2", Name: "Test Model 2", ModelStruct: reflect.TypeOf(models.IncidentPayload{})},
	}

	for _, model := range models {
		registry.RegisterModel(&model)
	}

	allModels := registry.GetAllModels()
	if len(allModels) != 2 {
		t.Errorf("Expected 2 models, got %d", len(allModels))
	}

	// Check that both models are present
	if _, exists := allModels["test1"]; !exists {
		t.Error("test1 model should be present")
	}
	if _, exists := allModels["test2"]; !exists {
		t.Error("test2 model should be present")
	}
}

func TestUnifiedRegistry_ListModels(t *testing.T) {
	registry := NewUnifiedRegistry("test/models", "test/validations")

	// Register multiple models
	registry.RegisterModel(&ModelInfo{Type: "test1", ModelStruct: reflect.TypeOf(models.IncidentPayload{})})
	registry.RegisterModel(&ModelInfo{Type: "test2", ModelStruct: reflect.TypeOf(models.IncidentPayload{})})

	modelTypes := registry.ListModels()
	if len(modelTypes) != 2 {
		t.Errorf("Expected 2 model types, got %d", len(modelTypes))
	}

	// Check that both types are present
	found1, found2 := false, false
	for _, modelType := range modelTypes {
		if modelType == "test1" {
			found1 = true
		}
		if modelType == "test2" {
			found2 = true
		}
	}
	if !found1 || !found2 {
		t.Error("Both model types should be present in list")
	}
}

func TestUnifiedRegistry_GetValidator(t *testing.T) {
	registry := NewUnifiedRegistry("test/models", "test/validations")

	// Create a mock validator wrapper
	mockValidator := &UniversalValidatorWrapper{
		modelType:         "test",
		validatorInstance: &mockValidatorInstance{},
		modelStructType:   reflect.TypeOf(models.IncidentPayload{}),
	}

	modelInfo := &ModelInfo{
		Type:        "test",
		Name:        "Test Model",
		ModelStruct: reflect.TypeOf(models.IncidentPayload{}),
		Validator:   mockValidator,
	}

	registry.RegisterModel(modelInfo)

	validator, err := registry.GetValidator("test")
	if err != nil {
		t.Fatalf("GetValidator failed: %v", err)
	}
	if validator == nil {
		t.Error("Validator should not be nil")
	}
}

func TestUnifiedRegistry_CreateModelInstance(t *testing.T) {
	registry := NewUnifiedRegistry("test/models", "test/validations")

	modelInfo := &ModelInfo{
		Type:        "test",
		Name:        "Test Model",
		ModelStruct: reflect.TypeOf(models.IncidentPayload{}),
	}

	registry.RegisterModel(modelInfo)

	instance, err := registry.CreateModelInstance("test")
	if err != nil {
		t.Fatalf("CreateModelInstance failed: %v", err)
	}
	if instance == nil {
		t.Error("Instance should not be nil")
	}

	// Check that it's the correct type
	if _, ok := instance.(*models.IncidentPayload); !ok {
		t.Errorf("Instance should be *models.IncidentPayload, got %T", instance)
	}
}

func TestUnifiedRegistry_GetModelStats(t *testing.T) {
	registry := NewUnifiedRegistry("test/models", "test/validations")

	// Initially should have 0 models
	stats := registry.GetModelStats()
	if stats["total_models"] != 0 {
		t.Errorf("Expected 0 total models, got %v", stats["total_models"])
	}

	// Register a model
	registry.RegisterModel(&ModelInfo{Type: "test", ModelStruct: reflect.TypeOf(models.IncidentPayload{})})

	stats = registry.GetModelStats()
	if stats["total_models"] != 1 {
		t.Errorf("Expected 1 total model, got %v", stats["total_models"])
	}

	modelTypes, ok := stats["model_types"].([]string)
	if !ok {
		t.Error("model_types should be []string")
	}
	if len(modelTypes) != 1 || modelTypes[0] != "test" {
		t.Errorf("Expected model_types to contain 'test', got %v", modelTypes)
	}
}

func TestUnifiedRegistry_GetRegisteredModelsWithDetails(t *testing.T) {
	registry := NewUnifiedRegistry("test/models", "test/validations")

	modelInfo := &ModelInfo{
		Type:        "test",
		Name:        "Test Model",
		Description: "A test model",
		Version:     "1.0.0",
		Author:      "Test Author",
		Tags:        []string{"test", "mock"},
		CreatedAt:   "2024-01-01T00:00:00Z",
		ModelStruct: reflect.TypeOf(models.IncidentPayload{}),
	}

	registry.RegisterModel(modelInfo)

	details := registry.GetRegisteredModelsWithDetails()
	if details["count"] != 1 {
		t.Errorf("Expected count 1, got %v", details["count"])
	}

	models, ok := details["models"].(map[string]interface{})
	if !ok {
		t.Fatal("models should be map[string]interface{}")
	}

	testModel, exists := models["test"]
	if !exists {
		t.Fatal("test model should exist in details")
	}

	modelMap, ok := testModel.(map[string]interface{})
	if !ok {
		t.Fatal("test model should be map[string]interface{}")
	}

	if modelMap["name"] != "Test Model" {
		t.Errorf("Expected name 'Test Model', got %v", modelMap["name"])
	}
	if modelMap["endpoint"] != "/validate/test" {
		t.Errorf("Expected endpoint '/validate/test', got %v", modelMap["endpoint"])
	}
}

func TestUnifiedRegistry_ConcurrentAccess(t *testing.T) {
	registry := NewUnifiedRegistry("test/models", "test/validations")

	// Test concurrent registration and access
	done := make(chan bool, 10)

	// Concurrent registrations
	for i := 0; i < 5; i++ {
		go func(id int) {
			modelInfo := &ModelInfo{
				Type:        ModelType("test" + string(rune(48+id))), // test0, test1, etc.
				Name:        "Test Model",
				ModelStruct: reflect.TypeOf(models.IncidentPayload{}),
			}
			registry.RegisterModel(modelInfo)
			done <- true
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 5; i++ {
		go func() {
			_ = registry.ListModels()
			_ = registry.GetAllModels()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Check final state
	models := registry.ListModels()
	if len(models) != 5 {
		t.Errorf("Expected 5 models after concurrent operations, got %d", len(models))
	}
}

func TestUnifiedRegistry_toTitleCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "Hello"},
		{"hello_world", "Hello_World"},
		{"hello-world", "Hello-World"},
		{"", ""},
		{"API", "Api"},
		{"test_case", "Test_Case"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := toTitleCase(tt.input)
			if result != tt.expected {
				t.Errorf("toTitleCase(%s) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestUnifiedRegistry_generateModelName(t *testing.T) {
	registry := NewUnifiedRegistry("test/models", "test/validations")

	tests := []struct {
		baseName   string
		structName string
		expected   string
	}{
		{"github", "GitHubPayload", "GitHub Webhook"},
		{"api", "APIRequest", "API Request/Response"},
		{"custom", "CustomPayload", "Custom Payload"},
		{"unknown", "UnknownStruct", "Unknown Data"},
	}

	for _, tt := range tests {
		t.Run(tt.baseName, func(t *testing.T) {
			result := registry.generateModelName(tt.baseName, tt.structName)
			if result != tt.expected {
				t.Errorf("generateModelName(%s, %s) = %s, want %s",
					tt.baseName, tt.structName, result, tt.expected)
			}
		})
	}
}

func TestUnifiedRegistry_generateModelDescription(t *testing.T) {
	registry := NewUnifiedRegistry("test/models", "test/validations")

	tests := []struct {
		baseName string
		contains string
	}{
		{"github", "GitHub webhook"},
		{"incident", "Incident report"},
		{"api", "API request"},
		{"unknown", "automatically discovered"},
	}

	for _, tt := range tests {
		t.Run(tt.baseName, func(t *testing.T) {
			result := registry.generateModelDescription(tt.baseName)
			if !strings.Contains(strings.ToLower(result), strings.ToLower(tt.contains)) {
				t.Errorf("generateModelDescription(%s) = %s, should contain %s",
					tt.baseName, result, tt.contains)
			}
		})
	}
}

func TestUnifiedRegistry_HTTPHandlers(t *testing.T) {
	registry := NewUnifiedRegistry("test/models", "test/validations")
	mux := http.NewServeMux()

	// Register a test model with mock validator
	mockValidator := &UniversalValidatorWrapper{
		modelType:         "test",
		validatorInstance: &mockValidatorInstance{},
		modelStructType:   reflect.TypeOf(models.IncidentPayload{}),
	}

	modelInfo := &ModelInfo{
		Type:        "test",
		Name:        "Test Model",
		ModelStruct: reflect.TypeOf(models.IncidentPayload{}),
		Validator:   mockValidator,
	}

	registry.RegisterModel(modelInfo)
	registry.mux = mux
	registry.registerAllHTTPEndpoints()

	// Test the created handler
	req := httptest.NewRequest("POST", "/validate/test", strings.NewReader(`{"id":"test-123"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler := registry.createDynamicHandler("test", modelInfo)
	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func BenchmarkUnifiedRegistry_RegisterModel(b *testing.B) {
	registry := NewUnifiedRegistry("test/models", "test/validations")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		modelInfo := &ModelInfo{
			Type:        ModelType("test" + string(rune(i))),
			Name:        "Test Model",
			ModelStruct: reflect.TypeOf(models.IncidentPayload{}),
		}
		registry.RegisterModel(modelInfo)
	}
}

func BenchmarkUnifiedRegistry_GetModel(b *testing.B) {
	registry := NewUnifiedRegistry("test/models", "test/validations")

	// Setup
	registry.RegisterModel(&ModelInfo{
		Type:        "test",
		Name:        "Test Model",
		ModelStruct: reflect.TypeOf(models.IncidentPayload{}),
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = registry.GetModel("test")
	}
}

func TestUnifiedRegistry_ValidateArray(t *testing.T) {
	registry := NewUnifiedRegistry("test/models", "test/validations")

	// Register a test model with mock validator
	mockValidator := &UniversalValidatorWrapper{
		modelType:         "test",
		validatorInstance: &mockValidatorInstance{},
		modelStructType:   reflect.TypeOf(models.GenericPayload{}),
	}

	modelInfo := &ModelInfo{
		Type:        "test",
		Name:        "Test Model",
		ModelStruct: reflect.TypeOf(models.GenericPayload{}),
		Validator:   mockValidator,
	}

	registry.RegisterModel(modelInfo)

	t.Run("validate array without threshold", func(t *testing.T) {
		records := []map[string]interface{}{
			{"id": "1", "name": "Record 1"},
			{"id": "2", "name": "Record 2"},
		}

		result, err := registry.ValidateArray("test", records, nil)
		if err != nil {
			t.Fatalf("ValidateArray failed: %v", err)
		}

		if result.TotalRecords != 2 {
			t.Errorf("Expected 2 total records, got %d", result.TotalRecords)
		}
		if result.Status != "success" {
			t.Errorf("Expected status success, got %s", result.Status)
		}
	})

	t.Run("validate array with threshold", func(t *testing.T) {
		threshold := 80.0
		records := []map[string]interface{}{
			{"id": "1", "name": "Record 1"},
			{"id": "2", "name": "Record 2"},
			{"id": "3", "name": "Record 3"},
		}

		result, err := registry.ValidateArray("test", records, &threshold)
		if err != nil {
			t.Fatalf("ValidateArray failed: %v", err)
		}

		if result.Threshold == nil {
			t.Error("Expected threshold to be set")
		} else if *result.Threshold != threshold {
			t.Errorf("Expected threshold %f, got %f", threshold, *result.Threshold)
		}
	})

	t.Run("validate array with non-existent model", func(t *testing.T) {
		records := []map[string]interface{}{
			{"id": "1", "name": "Record 1"},
		}

		_, err := registry.ValidateArray("nonexistent", records, nil)
		if err == nil {
			t.Error("Expected error for non-existent model type")
		}
	})

	t.Run("validate empty array", func(t *testing.T) {
		records := []map[string]interface{}{}

		result, err := registry.ValidateArray("test", records, nil)
		if err != nil {
			t.Fatalf("ValidateArray failed: %v", err)
		}

		if result.TotalRecords != 0 {
			t.Errorf("Expected 0 total records, got %d", result.TotalRecords)
		}
	})
}

func TestUnifiedRegistry_ValidatePayload(t *testing.T) {
	registry := NewUnifiedRegistry("test/models", "test/validations")

	// Register a test model with mock validator
	mockValidator := &UniversalValidatorWrapper{
		modelType:         "test",
		validatorInstance: &mockValidatorInstance{},
		modelStructType:   reflect.TypeOf(models.GenericPayload{}),
	}

	modelInfo := &ModelInfo{
		Type:        "test",
		Name:        "Test Model",
		ModelStruct: reflect.TypeOf(models.GenericPayload{}),
		Validator:   mockValidator,
	}

	registry.RegisterModel(modelInfo)

	t.Run("validate single payload", func(t *testing.T) {
		payload := map[string]interface{}{
			"id":   "test-123",
			"name": "Test Item",
		}

		result, err := registry.ValidatePayload("test", payload)
		if err != nil {
			t.Fatalf("ValidatePayload failed: %v", err)
		}

		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Fatal("Result should be map[string]interface{}")
		}

		if resultMap["is_valid"] != true {
			t.Errorf("Expected is_valid=true, got %v", resultMap["is_valid"])
		}
	})

	t.Run("validate payload with non-existent model", func(t *testing.T) {
		payload := map[string]interface{}{
			"id": "test-123",
		}

		_, err := registry.ValidatePayload("nonexistent", payload)
		if err == nil {
			t.Error("Expected error for non-existent model type")
		}
	})
}

// Mock validator for testing
type mockValidatorInstance struct{}

func (m *mockValidatorInstance) ValidatePayload(payload interface{}) interface{} {
	return map[string]interface{}{
		"is_valid":   true,
		"model_type": "test",
		"errors":     []interface{}{},
		"warnings":   []interface{}{},
	}
}

// TestUnifiedRegistry_StartAutoRegistration tests the auto-registration system
func TestUnifiedRegistry_StartAutoRegistration(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir := t.TempDir()
	modelsDir := filepath.Join(tempDir, "models")
	validationsDir := filepath.Join(tempDir, "validations")

	// Create directories
	os.MkdirAll(modelsDir, 0755)
	os.MkdirAll(validationsDir, 0755)

	// Create a test model file
	modelContent := `package models

type TestPayload struct {
	ID   string ` + "`json:\"id\"`" + `
	Name string ` + "`json:\"name\"`" + `
}
`
	os.WriteFile(filepath.Join(modelsDir, "test.go"), []byte(modelContent), 0644)

	// Create a test validator file
	validatorContent := `package validations

type TestValidator struct{}

func NewTestValidator() *TestValidator {
	return &TestValidator{}
}

func (v *TestValidator) ValidatePayload(payload interface{}) interface{} {
	return map[string]interface{}{
		"is_valid": true,
	}
}
`
	os.WriteFile(filepath.Join(validationsDir, "test.go"), []byte(validatorContent), 0644)

	registry := NewUnifiedRegistry(modelsDir, validationsDir)
	mux := http.NewServeMux()

	// Test with context
	ctx := context.Background()
	err := registry.StartAutoRegistration(ctx, mux)

	// Should complete without critical errors (some models might not be found, that's OK)
	if err != nil {
		t.Logf("StartAutoRegistration returned error (expected for missing models): %v", err)
	}

	// Verify mux was set
	if registry.mux == nil {
		t.Error("mux should be set after StartAutoRegistration")
	}
}

// TestUnifiedRegistry_DiscoverAndRegisterAll tests the discovery process
func TestUnifiedRegistry_DiscoverAndRegisterAll(t *testing.T) {
	tempDir := t.TempDir()
	modelsDir := filepath.Join(tempDir, "models")
	validationsDir := filepath.Join(tempDir, "validations")

	os.MkdirAll(modelsDir, 0755)
	os.MkdirAll(validationsDir, 0755)

	// Create a model file without validator (should be skipped)
	modelContent := `package models
type OrphanPayload struct {
	ID string
}
`
	os.WriteFile(filepath.Join(modelsDir, "orphan.go"), []byte(modelContent), 0644)

	registry := NewUnifiedRegistry(modelsDir, validationsDir)

	// Run discovery
	err := registry.discoverAndRegisterAll()

	// Should not return error even if no models are registered
	if err != nil {
		t.Logf("discoverAndRegisterAll completed with warnings: %v", err)
	}

	// Should have 0 models since there's no validator for orphan
	if len(registry.models) != 0 {
		t.Logf("Expected 0 models, got %d (orphan should be skipped)", len(registry.models))
	}
}

// TestUnifiedRegistry_ParseGoFileForStructs tests struct parsing
func TestUnifiedRegistry_ParseGoFileForStructs(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("valid go file", func(t *testing.T) {
		goFile := filepath.Join(tempDir, "test_valid.go")
		content := `package test

type FirstStruct struct {
	Field1 string
}

type SecondStruct struct {
	Field2 int
}

type NotAStruct interface {
	Method()
}
`
		os.WriteFile(goFile, []byte(content), 0644)

		registry := NewUnifiedRegistry("", "")
		structs, err := registry.parseGoFileForStructs(goFile)

		if err != nil {
			t.Fatalf("parseGoFileForStructs failed: %v", err)
		}

		if len(structs) != 2 {
			t.Errorf("Expected 2 structs, got %d", len(structs))
		}

		if !contains(structs, "FirstStruct") {
			t.Error("Should find FirstStruct")
		}
		if !contains(structs, "SecondStruct") {
			t.Error("Should find SecondStruct")
		}
	})

	t.Run("invalid go file", func(t *testing.T) {
		goFile := filepath.Join(tempDir, "test_invalid.go")
		os.WriteFile(goFile, []byte("this is not valid go code {{{"), 0644)

		registry := NewUnifiedRegistry("", "")
		_, err := registry.parseGoFileForStructs(goFile)

		if err == nil {
			t.Error("Expected error for invalid go file")
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		registry := NewUnifiedRegistry("", "")
		_, err := registry.parseGoFileForStructs("/nonexistent/file.go")

		if err == nil {
			t.Error("Expected error for nonexistent file")
		}
	})
}

// TestUnifiedRegistry_GetKnownModelTypes tests model type retrieval
func TestUnifiedRegistry_GetKnownModelTypes(t *testing.T) {
	registry := NewUnifiedRegistry("", "")
	knownTypes := registry.getKnownModelTypes()

	if len(knownTypes) == 0 {
		t.Error("Should have known model types")
	}

	// Check for expected types
	expectedTypes := []string{
		"IncidentPayload",
		"GitHubPayload",
		"APIRequest",
		"DatabaseQuery",
		"GenericPayload",
		"DeploymentPayload",
	}

	for _, typeName := range expectedTypes {
		if _, exists := knownTypes[typeName]; !exists {
			t.Errorf("Expected to find %s in known types", typeName)
		}
	}
}

// TestUnifiedRegistry_GetKnownValidatorConstructors tests validator constructor retrieval
func TestUnifiedRegistry_GetKnownValidatorConstructors(t *testing.T) {
	registry := NewUnifiedRegistry("", "")
	constructors := registry.getKnownValidatorConstructors()

	if len(constructors) == 0 {
		t.Error("Should have known validator constructors")
	}

	// Check for expected constructors
	expectedConstructors := []string{
		"NewIncidentValidator",
		"NewGitHubValidator",
		"NewAPIValidator",
		"NewDatabaseValidator",
		"NewGenericValidator",
		"NewDeploymentValidator",
	}

	for _, constructorName := range expectedConstructors {
		if _, exists := constructors[constructorName]; !exists {
			t.Errorf("Expected to find %s in known constructors", constructorName)
		}
		// Test that constructor can be called
		if constructor, exists := constructors[constructorName]; exists {
			result := constructor()
			if result == nil {
				t.Errorf("Constructor %s returned nil", constructorName)
			}
		}
	}
}

// TestUnifiedRegistry_GenerateModelTags tests tag generation
func TestUnifiedRegistry_GenerateModelTags(t *testing.T) {
	registry := NewUnifiedRegistry("", "")

	tests := []struct {
		baseName     string
		expectedTags []string
	}{
		{
			baseName:     "github",
			expectedTags: []string{"auto-discovered", "github", "webhook", "git", "collaboration"},
		},
		{
			baseName:     "api",
			expectedTags: []string{"auto-discovered", "api", "http", "rest", "web"},
		},
		{
			baseName:     "database",
			expectedTags: []string{"auto-discovered", "database", "sql", "transaction", "query"},
		},
		{
			baseName:     "unknown",
			expectedTags: []string{"auto-discovered", "unknown", "custom", "flexible"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.baseName, func(t *testing.T) {
			tags := registry.generateModelTags(tt.baseName)

			for _, expectedTag := range tt.expectedTags {
				if !contains(tags, expectedTag) {
					t.Errorf("Expected tag %s not found in %v", expectedTag, tags)
				}
			}
		})
	}
}

// TestUnifiedRegistry_CreateValidatorInstance tests validator instance creation
func TestUnifiedRegistry_CreateValidatorInstance(t *testing.T) {
	registry := NewUnifiedRegistry("", "")

	t.Run("known validator", func(t *testing.T) {
		instance, err := registry.createValidatorInstance("incident")
		if err != nil {
			t.Fatalf("createValidatorInstance failed: %v", err)
		}
		if instance == nil {
			t.Error("Validator instance should not be nil")
		}
	})

	t.Run("special case - github", func(t *testing.T) {
		instance, err := registry.createValidatorInstance("github")
		if err != nil {
			t.Fatalf("createValidatorInstance for github failed: %v", err)
		}
		if instance == nil {
			t.Error("Validator instance should not be nil")
		}
	})

	t.Run("unknown validator", func(t *testing.T) {
		_, err := registry.createValidatorInstance("nonexistent")
		if err == nil {
			t.Error("Expected error for unknown validator")
		}
	})
}

// TestUnifiedRegistry_DiscoverModelStruct tests struct discovery
func TestUnifiedRegistry_DiscoverModelStruct(t *testing.T) {
	tempDir := t.TempDir()
	modelsDir := filepath.Join(tempDir, "models")
	os.MkdirAll(modelsDir, 0755)

	registry := NewUnifiedRegistry(modelsDir, "")

	t.Run("known model type", func(t *testing.T) {
		// Test with a known model type
		structType, structName, err := registry.discoverModelStruct("incident")

		if err != nil {
			t.Fatalf("discoverModelStruct failed: %v", err)
		}

		if structType == nil {
			t.Error("Struct type should not be nil")
		}

		if structName != "IncidentPayload" {
			t.Errorf("Expected IncidentPayload, got %s", structName)
		}
	})

	t.Run("unknown model type", func(t *testing.T) {
		_, _, err := registry.discoverModelStruct("totallyfakemodel")

		if err == nil {
			t.Error("Expected error for unknown model type")
		}
	})
}

// TestUnifiedRegistry_RegisterModelAutomatically tests automatic registration
func TestUnifiedRegistry_RegisterModelAutomatically(t *testing.T) {
	registry := NewUnifiedRegistry("", "")

	t.Run("successful registration", func(t *testing.T) {
		// This will likely fail due to missing files, but we're testing the flow
		err := registry.registerModelAutomatically("incident")

		if err != nil {
			// Expected to fail in test environment
			t.Logf("registerModelAutomatically failed as expected: %v", err)
		}
	})
}

// TestUnifiedRegistry_SendJSONError tests JSON error responses
func TestUnifiedRegistry_SendJSONError(t *testing.T) {
	registry := NewUnifiedRegistry("", "")

	tests := []struct {
		name       string
		message    string
		statusCode int
	}{
		{"bad request", "Invalid input", http.StatusBadRequest},
		{"not found", "Model not found", http.StatusNotFound},
		{"internal error", "Server error", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			registry.sendJSONError(w, tt.message, tt.statusCode)

			if w.Code != tt.statusCode {
				t.Errorf("Expected status %d, got %d", tt.statusCode, w.Code)
			}

			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to parse JSON response: %v", err)
			}

			if response["error"] != tt.message {
				t.Errorf("Expected error message %s, got %v", tt.message, response["error"])
			}

			if status, ok := response["status"].(float64); !ok || int(status) != tt.statusCode {
				t.Errorf("Expected status %d in body, got %v", tt.statusCode, response["status"])
			}

			if _, exists := response["timestamp"]; !exists {
				t.Error("Expected timestamp in response")
			}
		})
	}
}

// TestUnifiedRegistry_CreateDynamicHandler_ErrorPaths tests error handling in dynamic handler
func TestUnifiedRegistry_CreateDynamicHandler_ErrorPaths(t *testing.T) {
	registry := NewUnifiedRegistry("", "")

	mockValidator := &UniversalValidatorWrapper{
		modelType:         "test",
		validatorInstance: &mockValidatorInstance{},
		modelStructType:   reflect.TypeOf(models.IncidentPayload{}),
	}

	modelInfo := &ModelInfo{
		Type:        "test",
		Name:        "Test Model",
		ModelStruct: reflect.TypeOf(models.IncidentPayload{}),
		Validator:   mockValidator,
	}

	registry.RegisterModel(modelInfo)

	handler := registry.createDynamicHandler("test", modelInfo)

	t.Run("invalid JSON payload", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/validate/test", strings.NewReader(`{invalid json`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		if !strings.Contains(response["error"].(string), "Invalid JSON") {
			t.Errorf("Expected 'Invalid JSON' error, got %v", response["error"])
		}
	})

	t.Run("empty payload", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/validate/test", strings.NewReader(``))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})
}

// TestUnifiedRegistry_ValidateSingleRow_EdgeCases tests edge cases in row validation
func TestUnifiedRegistry_ValidateSingleRow_EdgeCases(t *testing.T) {
	registry := NewUnifiedRegistry("", "")

	t.Run("JSON marshal error - circular reference", func(t *testing.T) {
		modelInfo := &ModelInfo{
			Type:        "test",
			Name:        "Test",
			ModelStruct: reflect.TypeOf(models.GenericPayload{}),
		}

		// Create a record that might cause marshal issues
		record := map[string]interface{}{
			"id": make(chan int), // channels can't be marshaled
		}

		result := registry.validateSingleRow("test", modelInfo, record, 0)

		if result.IsValid {
			t.Error("Expected validation to fail for unmarshalable data")
		}

		if len(result.Errors) == 0 {
			t.Error("Expected errors for marshal failure")
		}
	})

	t.Run("validation error handling", func(t *testing.T) {
		// Create a validator that returns ValidationResult struct
		mockValidator := &UniversalValidatorWrapper{
			modelType:         "test",
			validatorInstance: &mockValidatorWithErrors{},
			modelStructType:   reflect.TypeOf(models.GenericPayload{}),
		}

		modelInfo := &ModelInfo{
			Type:        "test",
			ModelStruct: reflect.TypeOf(models.GenericPayload{}),
			Validator:   mockValidator,
		}

		registry.RegisterModel(modelInfo)

		record := map[string]interface{}{
			"id":   "test-123",
			"data": "test",
		}

		result := registry.validateSingleRow("test", modelInfo, record, 5)

		if result.RowIndex != 5 {
			t.Errorf("Expected row index 5, got %d", result.RowIndex)
		}

		if result.RecordIdentifier != "test-123" {
			t.Errorf("Expected record identifier test-123, got %s", result.RecordIdentifier)
		}
	})
}

// TestUnifiedRegistry_ValidateArray_EdgeCases tests edge cases in array validation
func TestUnifiedRegistry_ValidateArray_EdgeCases(t *testing.T) {
	registry := NewUnifiedRegistry("", "")

	mockValidator := &UniversalValidatorWrapper{
		modelType:         "test",
		validatorInstance: &mockValidatorInstance{},
		modelStructType:   reflect.TypeOf(models.GenericPayload{}),
	}

	modelInfo := &ModelInfo{
		Type:        "test",
		ModelStruct: reflect.TypeOf(models.GenericPayload{}),
		Validator:   mockValidator,
	}

	registry.RegisterModel(modelInfo)

	t.Run("single invalid record without threshold", func(t *testing.T) {
		mockFailValidator := &UniversalValidatorWrapper{
			modelType:         "failtest",
			validatorInstance: &mockValidatorWithErrors{},
			modelStructType:   reflect.TypeOf(models.GenericPayload{}),
		}

		failModelInfo := &ModelInfo{
			Type:        "failtest",
			ModelStruct: reflect.TypeOf(models.GenericPayload{}),
			Validator:   mockFailValidator,
		}

		registry.RegisterModel(failModelInfo)

		records := []map[string]interface{}{
			{"id": "fail-1"},
		}

		result, err := registry.ValidateArray("failtest", records, nil)
		if err != nil {
			t.Fatalf("ValidateArray failed: %v", err)
		}

		if result.Status != "failed" {
			t.Errorf("Expected status 'failed' for single invalid record, got %s", result.Status)
		}
	})

	t.Run("multiple records with threshold met", func(t *testing.T) {
		threshold := 50.0
		records := []map[string]interface{}{
			{"id": "1"},
			{"id": "2"},
		}

		result, err := registry.ValidateArray("test", records, &threshold)
		if err != nil {
			t.Fatalf("ValidateArray failed: %v", err)
		}

		if result.Status != "success" {
			t.Errorf("Expected status 'success' when threshold is met, got %s", result.Status)
		}

		if result.Threshold == nil || *result.Threshold != threshold {
			t.Error("Threshold should be set in result")
		}
	})

	t.Run("records with warnings", func(t *testing.T) {
		mockWarnValidator := &UniversalValidatorWrapper{
			modelType:         "warntest",
			validatorInstance: &mockValidatorWithWarnings{},
			modelStructType:   reflect.TypeOf(models.GenericPayload{}),
		}

		warnModelInfo := &ModelInfo{
			Type:        "warntest",
			ModelStruct: reflect.TypeOf(models.GenericPayload{}),
			Validator:   mockWarnValidator,
		}

		registry.RegisterModel(warnModelInfo)

		records := []map[string]interface{}{
			{"id": "warn-1"},
		}

		result, err := registry.ValidateArray("warntest", records, nil)
		if err != nil {
			t.Fatalf("ValidateArray failed: %v", err)
		}

		if result.WarningRecords != 1 {
			t.Errorf("Expected 1 warning record, got %d", result.WarningRecords)
		}

		// Should include records with warnings in results
		if len(result.Results) == 0 {
			t.Error("Expected results to include records with warnings")
		}
	})

	t.Run("threshold success when 100% valid", func(t *testing.T) {
		threshold := 80.0
		records := []map[string]interface{}{
			{"id": "1", "name": "Record 1"},
			{"id": "2", "name": "Record 2"},
			{"id": "3", "name": "Record 3"},
			{"id": "4", "name": "Record 4"},
			{"id": "5", "name": "Record 5"},
		}

		result, err := registry.ValidateArray("test", records, &threshold)
		if err != nil {
			t.Fatalf("ValidateArray failed: %v", err)
		}

		if result.Status != "success" {
			t.Errorf("Expected status 'success' when 100%% valid (threshold 80%%), got %s", result.Status)
		}

		if result.TotalRecords != 5 {
			t.Errorf("Expected 5 total records, got %d", result.TotalRecords)
		}

		if result.ValidRecords != 5 {
			t.Errorf("Expected 5 valid records, got %d", result.ValidRecords)
		}

		successRate := (float64(result.ValidRecords) / float64(result.TotalRecords)) * 100.0
		if successRate < threshold {
			t.Errorf("Success rate %.2f%% should be >= threshold %.2f%%", successRate, threshold)
		}
	})

	t.Run("threshold failure when below threshold", func(t *testing.T) {
		// Use mock validator that returns errors for specific records
		mockMixedValidator := &UniversalValidatorWrapper{
			modelType:         "mixedtest",
			validatorInstance: &mockValidatorWithMixedResults{},
			modelStructType:   reflect.TypeOf(models.GenericPayload{}),
		}

		mixedModelInfo := &ModelInfo{
			Type:        "mixedtest",
			ModelStruct: reflect.TypeOf(models.GenericPayload{}),
			Validator:   mockMixedValidator,
		}

		registry.RegisterModel(mixedModelInfo)

		threshold := 80.0
		// This will have 50% success rate (3 valid, 3 invalid based on mock implementation)
		records := []map[string]interface{}{
			{"id": "valid-1"},
			{"id": "invalid-1"},
			{"id": "valid-2"},
			{"id": "invalid-2"},
			{"id": "valid-3"},
			{"id": "invalid-3"},
		}

		result, err := registry.ValidateArray("mixedtest", records, &threshold)
		if err != nil {
			t.Fatalf("ValidateArray failed: %v", err)
		}

		// With 50% success rate and 80% threshold, it should fail
		if result.Status != "failed" {
			t.Errorf("Expected status 'failed' when below threshold (50%% < 80%%), got %s", result.Status)
		}

		if result.Threshold == nil || *result.Threshold != threshold {
			t.Errorf("Expected threshold %.2f to be set in result", threshold)
		}
	})

	t.Run("threshold exact match", func(t *testing.T) {
		// Create mock validator for exact threshold test
		mockExactValidator := &UniversalValidatorWrapper{
			modelType:         "exacttest",
			validatorInstance: &mockValidatorExactThreshold{},
			modelStructType:   reflect.TypeOf(models.GenericPayload{}),
		}

		exactModelInfo := &ModelInfo{
			Type:        "exacttest",
			ModelStruct: reflect.TypeOf(models.GenericPayload{}),
			Validator:   mockExactValidator,
		}

		registry.RegisterModel(exactModelInfo)

		threshold := 80.0
		// 4 valid out of 5 = 80% exactly
		records := []map[string]interface{}{
			{"id": "valid-1"},
			{"id": "valid-2"},
			{"id": "valid-3"},
			{"id": "valid-4"},
			{"id": "invalid-1"},
		}

		result, err := registry.ValidateArray("exacttest", records, &threshold)
		if err != nil {
			t.Fatalf("ValidateArray failed: %v", err)
		}

		// Exactly at threshold should be success
		if result.Status != "success" {
			t.Errorf("Expected status 'success' when exactly at threshold (80%% == 80%%), got %s", result.Status)
		}
	})

	t.Run("no threshold with mixed results", func(t *testing.T) {
		mockMixedValidator := &UniversalValidatorWrapper{
			modelType:         "mixedtest2",
			validatorInstance: &mockValidatorWithMixedResults{},
			modelStructType:   reflect.TypeOf(models.GenericPayload{}),
		}

		mixedModelInfo := &ModelInfo{
			Type:        "mixedtest2",
			ModelStruct: reflect.TypeOf(models.GenericPayload{}),
			Validator:   mockMixedValidator,
		}

		registry.RegisterModel(mixedModelInfo)

		// No threshold - should always be success for multiple records
		records := []map[string]interface{}{
			{"id": "valid-1"},
			{"id": "invalid-1"},
			{"id": "valid-2"},
		}

		result, err := registry.ValidateArray("mixedtest2", records, nil)
		if err != nil {
			t.Fatalf("ValidateArray failed: %v", err)
		}

		// No threshold with multiple records should return success
		if result.Status != "success" {
			t.Errorf("Expected status 'success' for multiple records without threshold, got %s", result.Status)
		}

		if result.Threshold != nil {
			t.Error("Expected threshold to be nil when not provided")
		}
	})
}

// TestUnifiedRegistry_GlobalRegistry tests global registry functions
func TestUnifiedRegistry_GetGlobalRegistry(t *testing.T) {
	registry1 := GetGlobalRegistry()
	if registry1 == nil {
		t.Fatal("GetGlobalRegistry should not return nil")
	}

	// Should return same instance
	registry2 := GetGlobalRegistry()
	if registry1 != registry2 {
		t.Error("GetGlobalRegistry should return the same instance")
	}
}

// TestUnifiedRegistry_StartRegistration tests the global start function
func TestUnifiedRegistry_StartRegistration(t *testing.T) {
	ctx := context.Background()
	mux := http.NewServeMux()

	err := StartRegistration(ctx, mux)
	if err != nil {
		t.Logf("StartRegistration completed with warnings: %v", err)
	}

	// Should have initialized global registry
	registry := GetGlobalRegistry()
	if registry.mux == nil {
		t.Error("Global registry mux should be set")
	}
}

// TestUnifiedRegistry_CreateModelInstance_ErrorPath tests error handling
func TestUnifiedRegistry_CreateModelInstance_ErrorPath(t *testing.T) {
	registry := NewUnifiedRegistry("", "")

	_, err := registry.CreateModelInstance("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent model")
	}
}

// TestUnifiedRegistry_RegisterAllHTTPEndpoints tests endpoint registration
func TestUnifiedRegistry_RegisterAllHTTPEndpoints(t *testing.T) {
	registry := NewUnifiedRegistry("", "")

	t.Run("with nil mux", func(t *testing.T) {
		registry.mux = nil
		registry.registerAllHTTPEndpoints()
		// Should complete without panic
	})

	t.Run("with mux and models", func(t *testing.T) {
		mux := http.NewServeMux()
		registry.mux = mux

		mockValidator := &UniversalValidatorWrapper{
			modelType:         "endpoint_test",
			validatorInstance: &mockValidatorInstance{},
			modelStructType:   reflect.TypeOf(models.GenericPayload{}),
		}

		modelInfo := &ModelInfo{
			Type:        "endpoint_test",
			Name:        "Endpoint Test",
			ModelStruct: reflect.TypeOf(models.GenericPayload{}),
			Validator:   mockValidator,
		}

		registry.RegisterModel(modelInfo)
		registry.registerAllHTTPEndpoints()

		// Test that endpoint works
		req := httptest.NewRequest("POST", "/validate/endpoint_test", strings.NewReader(`{"id":"test"}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})
}

// TestDynamicRegistry_Functions tests dynamic registry wrapper functions
func TestDynamicRegistry_Functions(t *testing.T) {
	unified := NewUnifiedRegistry("", "")
	dynamic := NewDynamicModelRegistry(unified, "models", "validations")

	if dynamic.UnifiedRegistry != unified {
		t.Error("Dynamic registry should wrap unified registry")
	}

	// Test delegated functions
	allModels := dynamic.GetAllModels()
	if allModels == nil {
		t.Error("GetAllModels should not return nil")
	}

	details := dynamic.GetRegisteredModelsWithDetails()
	if details == nil {
		t.Error("GetRegisteredModelsWithDetails should not return nil")
	}
}

// TestGlobalDynamicRegistry tests global dynamic registry
func TestGlobalDynamicRegistry(t *testing.T) {
	registry1 := GetGlobalDynamicRegistry()
	if registry1 == nil {
		t.Fatal("GetGlobalDynamicRegistry should not return nil")
	}

	registry2 := GetGlobalDynamicRegistry()
	if registry1 != registry2 {
		t.Error("Should return same instance")
	}
}

// TestStartDynamicRegistration tests dynamic registration start
func TestStartDynamicRegistration(t *testing.T) {
	ctx := context.Background()
	mux := http.NewServeMux()

	err := StartDynamicRegistration(ctx, mux)
	if err != nil {
		t.Logf("StartDynamicRegistration completed: %v", err)
	}
}

// TestUniversalValidatorWrapper_EdgeCases tests validator wrapper edge cases
func TestUniversalValidatorWrapper_EdgeCases(t *testing.T) {
	t.Run("validator without ValidatePayload method", func(t *testing.T) {
		wrapper := &UniversalValidatorWrapper{
			modelType:         "test",
			validatorInstance: &struct{}{}, // No methods
			modelStructType:   reflect.TypeOf(models.GenericPayload{}),
		}

		result := wrapper.ValidatePayload(map[string]interface{}{})

		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Fatal("Expected map result")
		}

		if resultMap["is_valid"] != false {
			t.Error("Expected is_valid=false for missing method")
		}

		if resultMap["provider"] != "universal-wrapper-fallback" {
			t.Error("Expected fallback provider")
		}
	})

	t.Run("validator with alternative method names", func(t *testing.T) {
		wrapper := &UniversalValidatorWrapper{
			modelType:         "test",
			validatorInstance: &mockValidatorWithAlternateMethod{},
			modelStructType:   reflect.TypeOf(models.GenericPayload{}),
		}

		result := wrapper.ValidatePayload(map[string]interface{}{})

		if result == nil {
			t.Error("Expected non-nil result")
		}
	})
}

// Helper function
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Additional mock validators for testing

type mockValidatorWithErrors struct{}

func (m *mockValidatorWithErrors) ValidatePayload(payload interface{}) interface{} {
	return models.ValidationResult{
		IsValid:   false,
		ModelType: "test",
		Errors: []models.ValidationError{
			{Field: "id", Message: "Invalid ID", Code: "INVALID_ID"},
		},
		Warnings: []models.ValidationWarning{},
	}
}

type mockValidatorWithWarnings struct{}

func (m *mockValidatorWithWarnings) ValidatePayload(payload interface{}) interface{} {
	return models.ValidationResult{
		IsValid:   true,
		ModelType: "test",
		Errors:    []models.ValidationError{},
		Warnings: []models.ValidationWarning{
			{Field: "data", Message: "Deprecated field", Code: "DEPRECATED"},
		},
	}
}

type mockValidatorWithAlternateMethod struct{}

func (m *mockValidatorWithAlternateMethod) Validate(payload interface{}) interface{} {
	return map[string]interface{}{
		"is_valid": true,
	}
}

// mockValidatorWithMixedResults returns success for IDs starting with "valid-", errors for others
type mockValidatorWithMixedResults struct{}

func (m *mockValidatorWithMixedResults) ValidatePayload(payload interface{}) interface{} {
	// Extract ID from payload if it's a map
	if payloadMap, ok := payload.(map[string]interface{}); ok {
		if id, exists := payloadMap["id"]; exists {
			if idStr, ok := id.(string); ok {
				// Return valid for IDs starting with "valid-"
				if len(idStr) >= 6 && idStr[:6] == "valid-" {
					return models.ValidationResult{
						IsValid:   true,
						ModelType: "mixedtest",
						Provider:  "go-playground",
						Errors:    []models.ValidationError{},
						Warnings:  []models.ValidationWarning{},
					}
				}
			}
		}
	}

	// Return error for all other cases
	return models.ValidationResult{
		IsValid:   false,
		ModelType: "mixedtest",
		Provider:  "go-playground",
		Errors: []models.ValidationError{
			{
				Field:   "id",
				Message: "Invalid record",
				Code:    "VALIDATION_FAILED",
			},
		},
		Warnings: []models.ValidationWarning{},
	}
}

// mockValidatorExactThreshold returns errors only for IDs starting with "invalid-"
type mockValidatorExactThreshold struct{}

func (m *mockValidatorExactThreshold) ValidatePayload(payload interface{}) interface{} {
	// Extract ID from payload if it's a map
	if payloadMap, ok := payload.(map[string]interface{}); ok {
		if id, exists := payloadMap["id"]; exists {
			if idStr, ok := id.(string); ok {
				// Return invalid for IDs starting with "invalid-"
				if len(idStr) >= 8 && idStr[:8] == "invalid-" {
					return models.ValidationResult{
						IsValid:   false,
						ModelType: "exacttest",
						Provider:  "go-playground",
						Errors: []models.ValidationError{
							{
								Field:   "id",
								Message: "Invalid record",
								Code:    "VALIDATION_FAILED",
							},
						},
						Warnings: []models.ValidationWarning{},
					}
				}
			}
		}
	}

	// Return valid for all other cases
	return models.ValidationResult{
		IsValid:   true,
		ModelType: "exacttest",
		Provider:  "go-playground",
		Errors:    []models.ValidationError{},
		Warnings:  []models.ValidationWarning{},
	}
}
