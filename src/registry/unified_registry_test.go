package registry

import (
	"net/http"
	"net/http/httptest"
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
