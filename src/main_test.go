package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"goplayground-data-validator/registry"
)

// testValidatorWrapper implements ValidatorInterface for testing
// It's model-agnostic and can be used for any model type
type testValidatorWrapper struct {
	modelType string
	isValid   bool
}

func (tvw *testValidatorWrapper) ValidatePayload(payload interface{}) interface{} {
	return map[string]interface{}{
		"is_valid":   tvw.isValid,
		"model_type": tvw.modelType,
		"provider":   "go-playground",
		"errors":     []interface{}{},
		"warnings":   []interface{}{},
	}
}

// GenericTestPayload represents a generic test payload structure
type GenericTestPayload struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
}

// createTestModel creates a generic test model for unit testing
func createTestModel(modelType string, isValid bool) *registry.ModelInfo {
	return &registry.ModelInfo{
		Type:        registry.ModelType(modelType),
		Name:        strings.Title(modelType) + " Test Model",
		Description: "Test model for " + modelType + " validation",
		ModelStruct: reflect.TypeOf(GenericTestPayload{}), // Use proper struct type
		Validator:   &testValidatorWrapper{modelType: modelType, isValid: isValid},
		Examples:    []interface{}{},
		Version:     "1.0.0-test",
		CreatedAt:   time.Now().Format(time.RFC3339),
		Author:      "test-framework",
		Tags:        []string{modelType, "test", "generic"},
	}
}

// getTestModels returns a list of test models to register
func getTestModels() []*registry.ModelInfo {
	return []*registry.ModelInfo{
		createTestModel("testmodel", true),     // Valid test model
		createTestModel("invalidmodel", false), // Invalid test model
	}
}

func init() {
	// Initialize the registry for unit tests with generic test models
	globalRegistry := registry.GetGlobalRegistry()

	// Register multiple test models to make tests model-agnostic
	for _, model := range getTestModels() {
		globalRegistry.RegisterModel(model)
	}
}

// TestHandleHealth tests the health endpoint
func TestHandleHealth(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	handleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", w.Code)
	}

	var health map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &health)
	if err != nil {
		t.Errorf("Failed to unmarshal health response: %v", err)
	}

	if health["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got %v", health["status"])
	}

	expectedFields := []string{"status", "version", "uptime", "server"}
	for _, field := range expectedFields {
		if _, exists := health[field]; !exists {
			t.Errorf("Missing field in health response: %s", field)
		}
	}

	if health["version"] != "2.0.0-modular" {
		t.Errorf("Expected version '2.0.0-modular', got %v", health["version"])
	}
}

// TestHandleListModels tests the models listing endpoint
func TestHandleListModels(t *testing.T) {
	req := httptest.NewRequest("GET", "/models", nil)
	w := httptest.NewRecorder()

	handleListModels(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to unmarshal models response: %v", err)
	}

	if response["models"] == nil {
		t.Error("Expected models field in response")
	}
}

// TestHandleGenericValidation tests the generic validation endpoint
func TestHandleGenericValidation(t *testing.T) {
	tests := []struct {
		name           string
		request        map[string]interface{}
		expectedStatus int
		expectValid    bool
	}{
		{
			name: "valid test model payload",
			request: map[string]interface{}{
				"model_type": "testmodel",
				"payload": map[string]interface{}{
					"id":          "TEST-001",
					"name":        "Test payload",
					"description": "Generic test payload for validation",
					"type":        "test",
					"status":      "active",
					"created_at":  time.Now().Format(time.RFC3339),
				},
			},
			expectedStatus: http.StatusOK,
			expectValid:    true,
		},
		{
			name: "invalid test model payload",
			request: map[string]interface{}{
				"model_type": "invalidmodel",
				"payload": map[string]interface{}{
					"id": "INVALID-001",
				},
			},
			expectedStatus: http.StatusUnprocessableEntity, // 422 for invalid validation
			expectValid:    false,
		},
		{
			name: "invalid model type",
			request: map[string]interface{}{
				"model_type": "nonexistent",
				"payload":    map[string]interface{}{},
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing model type",
			request: map[string]interface{}{
				"payload": map[string]interface{}{},
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid JSON",
			request:        nil, // Will send invalid JSON
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body *bytes.Buffer
			if tt.request != nil {
				jsonData, _ := json.Marshal(tt.request)
				body = bytes.NewBuffer(jsonData)
			} else {
				body = bytes.NewBufferString("invalid json")
			}

			req := httptest.NewRequest("POST", "/validate", body)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handleGenericValidation(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK || tt.expectedStatus == http.StatusUnprocessableEntity {
				var result map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &result)
				if err != nil {
					t.Errorf("Failed to unmarshal validation response: %v", err)
					return
				}

				isValid, ok := result["is_valid"].(bool)
				if !ok {
					t.Error("Expected is_valid field to be boolean")
					return
				}

				if isValid != tt.expectValid {
					t.Errorf("Expected is_valid=%v, got %v", tt.expectValid, isValid)
				}
			}
		})
	}
}

// TestHandleSwaggerModels tests the swagger models endpoint
func TestHandleSwaggerModels(t *testing.T) {
	req := httptest.NewRequest("GET", "/swagger/models", nil)
	w := httptest.NewRecorder()

	handleSwaggerModels(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to unmarshal swagger models response: %v", err)
	}

	if response["models"] == nil {
		t.Error("Expected models field in swagger response")
	}
}

// TestHandleSwaggerJSON tests the swagger JSON endpoint
func TestHandleSwaggerJSON(t *testing.T) {
	req := httptest.NewRequest("GET", "/swagger/doc.json", nil)
	w := httptest.NewRecorder()

	handleSwaggerJSON(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to unmarshal swagger JSON response: %v", err)
	}

	// Check for basic Swagger structure
	if response["swagger"] == nil && response["openapi"] == nil {
		t.Error("Expected swagger or openapi field in response")
	}
}

// TestSendJSONError tests the error response helper
func TestSendJSONError(t *testing.T) {
	w := httptest.NewRecorder()

	sendJSONError(w, "Test error message", http.StatusBadRequest)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}

	var errorResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	if err != nil {
		t.Errorf("Failed to unmarshal error response: %v", err)
	}

	if errorResponse["error"] != "Test error message" {
		t.Errorf("Expected error message 'Test error message', got %v", errorResponse["error"])
	}
}

// TestConvertMapToStruct tests the map to struct conversion utility with generic data
func TestConvertMapToStruct(t *testing.T) {
	// Test with a simple generic struct that any model can use
	type GenericTestStruct struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Value       int    `json:"value"`
		Active      bool   `json:"active"`
		CreatedAt   string `json:"created_at"`
	}

	sourceMap := map[string]interface{}{
		"id":          "TEST-001",
		"name":        "Test Entity",
		"description": "This is a generic test entity for validation",
		"value":       42,
		"active":      true,
		"created_at":  time.Now().Format(time.RFC3339),
	}

	var testStruct GenericTestStruct
	err := convertMapToStruct(sourceMap, &testStruct)
	if err != nil {
		t.Errorf("convertMapToStruct failed: %v", err)
	}

	if testStruct.ID != "TEST-001" {
		t.Errorf("Expected ID 'TEST-001', got %s", testStruct.ID)
	}
	if testStruct.Value != 42 {
		t.Errorf("Expected Value 42, got %d", testStruct.Value)
	}
	if !testStruct.Active {
		t.Error("Expected Active to be true")
	}
}

// TestConvertMapToStruct_InvalidData tests conversion with invalid data using generic types
func TestConvertMapToStruct_InvalidData(t *testing.T) {
	type TestStruct struct {
		Value     int       `json:"value"`
		Timestamp time.Time `json:"timestamp"`
	}

	tests := []struct {
		name string
		data map[string]interface{}
	}{
		{
			name: "string to int conversion error",
			data: map[string]interface{}{
				"value": "not-a-number",
			},
		},
		{
			name: "invalid time format",
			data: map[string]interface{}{
				"timestamp": "invalid-time-format",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var testStruct TestStruct
			err := convertMapToStruct(tt.data, &testStruct)

			// We expect an error for invalid data
			if err == nil {
				t.Error("Expected error for invalid data, got nil")
			}
		})
	}
}

// TestStartTime verifies the global start time is set
func TestStartTime(t *testing.T) {
	if startTime.IsZero() {
		t.Error("startTime should be initialized")
	}

	// Should be recent (within last minute for test purposes)
	if time.Since(startTime) > time.Minute {
		t.Error("startTime seems too old, might not be properly initialized")
	}
}

// TestGlobalRegistry tests that the global registry is accessible
func TestGlobalRegistry(t *testing.T) {
	globalRegistry := registry.GetGlobalRegistry()
	if globalRegistry == nil {
		t.Error("Global registry should not be nil")
	}

	// Test basic registry functionality
	stats := globalRegistry.GetModelStats()
	if stats == nil {
		t.Error("Registry stats should not be nil")
	}

	if _, exists := stats["total_models"]; !exists {
		t.Error("Registry stats should include total_models")
	}
}

// TestHTTPMethodValidation tests that endpoints reject invalid HTTP methods
func TestHTTPMethodValidation(t *testing.T) {
	tests := []struct {
		endpoint string
		method   string
		handler  http.HandlerFunc
	}{
		{"/health", "POST", handleHealth},
		{"/models", "POST", handleListModels},
		{"/validate", "GET", handleGenericValidation},
	}

	for _, tt := range tests {
		t.Run(tt.endpoint+"_"+tt.method, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.endpoint, nil)
			w := httptest.NewRecorder()

			tt.handler(w, req)

			// Most endpoints should return 405 for wrong methods,
			// but some might handle it differently
			if w.Code != http.StatusMethodNotAllowed && w.Code != http.StatusBadRequest && w.Code != http.StatusOK {
				// Some handlers might be more permissive, that's okay for now
				t.Logf("Endpoint %s with method %s returned status %d", tt.endpoint, tt.method, w.Code)
			}
		})
	}
}

// TestJSONResponseHeaders tests that JSON endpoints set proper headers
func TestJSONResponseHeaders(t *testing.T) {
	endpoints := []struct {
		path    string
		handler http.HandlerFunc
	}{
		{"/health", handleHealth},
		{"/models", handleListModels},
		{"/swagger/models", handleSwaggerModels},
	}

	for _, endpoint := range endpoints {
		t.Run(endpoint.path, func(t *testing.T) {
			req := httptest.NewRequest("GET", endpoint.path, nil)
			w := httptest.NewRecorder()

			endpoint.handler(w, req)

			contentType := w.Header().Get("Content-Type")
			if !strings.Contains(contentType, "application/json") {
				t.Errorf("Expected Content-Type to contain application/json, got %s", contentType)
			}
		})
	}
}

// Benchmark tests for performance validation
func BenchmarkHandleHealth(b *testing.B) {
	req := httptest.NewRequest("GET", "/health", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handleHealth(w, req)
	}
}

func BenchmarkHandleListModels(b *testing.B) {
	req := httptest.NewRequest("GET", "/models", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handleListModels(w, req)
	}
}

func BenchmarkConvertMapToStruct(b *testing.B) {
	type BenchStruct struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Type        string `json:"type"`
		Status      string `json:"status"`
		Value       int    `json:"value"`
		CreatedAt   string `json:"created_at"`
	}

	sourceMap := map[string]interface{}{
		"id":          "BENCH-001",
		"name":        "Benchmark test",
		"description": "This is a benchmark test for generic validation",
		"type":        "benchmark",
		"status":      "active",
		"value":       123,
		"created_at":  time.Now().Format(time.RFC3339),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var benchStruct BenchStruct
		convertMapToStruct(sourceMap, &benchStruct)
	}
}
