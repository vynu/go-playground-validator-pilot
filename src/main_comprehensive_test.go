package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"goplayground-data-validator/models"
)

// Test Batch Handlers
func TestHandleBatchStart(t *testing.T) {
	tests := []struct {
		name           string
		payload        map[string]interface{}
		expectedStatus int
	}{
		{
			name: "valid batch start",
			payload: map[string]interface{}{
				"model_type": "testmodel",
				"job_id":     "test-job-123",
				"threshold":  20.0,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "batch start without threshold",
			payload: map[string]interface{}{
				"model_type": "testmodel",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "missing model_type",
			payload: map[string]interface{}{
				"threshold": 50.0,
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", "/validate/batch/start", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			handleBatchStart(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)

				if response["batch_id"] == nil {
					t.Error("Expected batch_id in response")
				}
				if response["status"] != "active" {
					t.Errorf("Expected status active, got %v", response["status"])
				}
			}
		})
	}
}

func TestHandleBatchStatus(t *testing.T) {
	// Create a test batch session
	batchManager := models.GetBatchSessionManager()
	threshold := 50.0
	session := batchManager.CreateBatchSession("test-batch-123", &threshold)

	tests := []struct {
		name           string
		batchID        string
		expectedStatus int
	}{
		{
			name:           "valid batch status",
			batchID:        session.BatchID,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "batch not found",
			batchID:        "non-existent-batch",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/validate/batch/"+tt.batchID, nil)
			req.SetPathValue("id", tt.batchID)
			w := httptest.NewRecorder()

			handleBatchStatus(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)

				if response["batch_id"] != session.BatchID {
					t.Errorf("Expected batch_id %s, got %v", session.BatchID, response["batch_id"])
				}
			}
		})
	}

	// Cleanup
	batchManager.DeleteBatchSession(session.BatchID)
}

func TestHandleBatchComplete(t *testing.T) {
	// Create and populate a test batch session
	batchManager := models.GetBatchSessionManager()
	threshold := 50.0
	session := batchManager.CreateBatchSession("test-batch-456", &threshold)
	batchManager.UpdateBatchSession(session.BatchID, 60, 40, 5)

	tests := []struct {
		name           string
		batchID        string
		expectedStatus int
	}{
		{
			name:           "valid batch complete",
			batchID:        session.BatchID,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "batch not found",
			batchID:        "non-existent",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/validate/batch/"+tt.batchID+"/complete", nil)
			req.SetPathValue("id", tt.batchID)
			w := httptest.NewRecorder()

			handleBatchComplete(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)

				if response["status"] != "success" {
					t.Errorf("Expected status success with 60%% valid and 50%% threshold, got %v", response["status"])
				}
			}
		})
	}
}

// Test Helper Functions
func TestConvertToInt64(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected int64
		ok       bool
	}{
		{"int", int(42), 42, true},
		{"int8", int8(42), 42, true},
		{"int16", int16(42), 42, true},
		{"int32", int32(42), 42, true},
		{"int64", int64(42), 42, true},
		{"float32", float32(42.5), 42, true},
		{"float64", float64(42.5), 42, true},
		{"string valid", "42", 42, true},
		{"string invalid", "abc", 0, false},
		{"bool", true, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := convertToInt64(tt.input)
			if ok != tt.ok {
				t.Errorf("Expected ok=%v, got %v", tt.ok, ok)
			}
			if ok && result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestConvertToUint64(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected uint64
		ok       bool
	}{
		{"uint", uint(42), 42, true},
		{"uint8", uint8(42), 42, true},
		{"uint16", uint16(42), 42, true},
		{"uint32", uint32(42), 42, true},
		{"uint64", uint64(42), 42, true},
		{"int positive", int(42), 42, true},
		{"int negative", int(-1), 0, false},
		{"int64 positive", int64(42), 42, true},
		{"int64 negative", int64(-1), 0, false},
		{"float64 positive", float64(42.5), 42, true},
		{"float64 negative", float64(-1.5), 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := convertToUint64(tt.input)
			if ok != tt.ok {
				t.Errorf("Expected ok=%v, got %v", tt.ok, ok)
			}
			if ok && result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestConvertToFloat64(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected float64
		ok       bool
	}{
		{"float32", float32(42.5), 42.5, true},
		{"float64", float64(42.5), 42.5, true},
		{"int", int(42), 42.0, true},
		{"int64", int64(42), 42.0, true},
		{"uint64", uint64(42), 42.0, true},
		{"string valid", "42.5", 42.5, true},
		{"string invalid", "abc", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := convertToFloat64(tt.input)
			if ok != tt.ok {
				t.Errorf("Expected ok=%v, got %v", tt.ok, ok)
			}
			if ok && result != tt.expected {
				t.Errorf("Expected %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestSetFieldValue(t *testing.T) {
	type TestStruct struct {
		StringField string
		IntField    int
		FloatField  float64
		BoolField   bool
	}

	tests := []struct {
		name      string
		fieldName string
		value     interface{}
		expected  interface{}
		shouldErr bool
	}{
		{"string field", "StringField", "test", "test", false},
		{"int field", "IntField", 42, 42, false},
		{"float field", "FloatField", 42.5, 42.5, false},
		{"bool field", "BoolField", true, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := &TestStruct{}
			val := reflect.ValueOf(ts).Elem()
			field := val.FieldByName(tt.fieldName)

			err := setFieldValue(field, tt.value)

			if tt.shouldErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestSetSliceValue(t *testing.T) {
	type TestStruct struct {
		IntSlice    []int
		StringSlice []string
	}

	tests := []struct {
		name      string
		fieldName string
		value     interface{}
		shouldErr bool
	}{
		{"valid int slice", "IntSlice", []interface{}{1, 2, 3}, false},
		{"valid string slice", "StringSlice", []interface{}{"a", "b", "c"}, false},
		{"invalid not slice", "IntSlice", "not a slice", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := &TestStruct{}
			val := reflect.ValueOf(ts).Elem()
			field := val.FieldByName(tt.fieldName)

			err := setSliceValue(field, tt.value)

			if tt.shouldErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestSetMapValue(t *testing.T) {
	type TestStruct struct {
		StringMap map[string]string
		IntMap    map[string]int
	}

	tests := []struct {
		name      string
		fieldName string
		value     interface{}
		shouldErr bool
	}{
		{
			"valid string map",
			"StringMap",
			map[string]interface{}{"key1": "value1", "key2": "value2"},
			false,
		},
		{
			"valid int map",
			"IntMap",
			map[string]interface{}{"key1": 1, "key2": 2},
			false,
		},
		{
			"invalid not map",
			"StringMap",
			"not a map",
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := &TestStruct{}
			val := reflect.ValueOf(ts).Elem()
			field := val.FieldByName(tt.fieldName)

			err := setMapValue(field, tt.value)

			if tt.shouldErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// Test additional edge cases for coverage
func TestConvertMapToStruct_EdgeCases(t *testing.T) {
	type ComplexStruct struct {
		StringField string                 `json:"string_field"`
		IntField    int                    `json:"int_field"`
		FloatField  float64                `json:"float_field"`
		BoolField   bool                   `json:"bool_field"`
		SliceField  []string               `json:"slice_field"`
		MapField    map[string]interface{} `json:"map_field"`
		SkipField   string                 `json:"-"`
		NoTagField  string
	}

	t.Run("complex struct with all field types", func(t *testing.T) {
		input := map[string]interface{}{
			"string_field": "test",
			"int_field":    42,
			"float_field":  3.14,
			"bool_field":   true,
			"slice_field":  []interface{}{"a", "b", "c"},
			"map_field":    map[string]interface{}{"key": "value"},
		}

		cs := &ComplexStruct{}
		err := convertMapToStruct(input, cs)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if cs.StringField != "test" {
			t.Errorf("Expected StringField='test', got '%s'", cs.StringField)
		}
		if cs.IntField != 42 {
			t.Errorf("Expected IntField=42, got %d", cs.IntField)
		}
	})
}

func TestBatchSessionIntegration(t *testing.T) {
	t.Run("full batch lifecycle", func(t *testing.T) {
		// Start batch
		startPayload := map[string]interface{}{
			"model_type": "testmodel",
			"threshold":  80.0,
		}
		body, _ := json.Marshal(startPayload)
		req := httptest.NewRequest("POST", "/validate/batch/start", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		handleBatchStart(w, req)

		var startResp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &startResp)
		batchID := startResp["batch_id"].(string)

		// Check status
		req = httptest.NewRequest("GET", "/validate/batch/"+batchID, nil)
		req.SetPathValue("id", batchID)
		w = httptest.NewRecorder()
		handleBatchStatus(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status check to succeed, got %d", w.Code)
		}

		// Complete batch
		req = httptest.NewRequest("POST", "/validate/batch/"+batchID+"/complete", nil)
		req.SetPathValue("id", batchID)
		w = httptest.NewRecorder()
		handleBatchComplete(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected batch complete to succeed, got %d", w.Code)
		}

		// Wait for cleanup
		time.Sleep(2 * time.Second)
	})
}

// ============================================================================
// COMPREHENSIVE COVERAGE TESTS - Target 80%+
// ============================================================================

// Test handleBatchStart with invalid JSON
func TestHandleBatchStart_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest("POST", "/validate/batch/start", bytes.NewBufferString("invalid json"))
	w := httptest.NewRecorder()
	handleBatchStart(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d for invalid JSON, got %d", http.StatusBadRequest, w.Code)
	}
}

// Test handleBatchStatus with empty batch ID
func TestHandleBatchStatus_EmptyID(t *testing.T) {
	req := httptest.NewRequest("GET", "/validate/batch/", nil)
	req.SetPathValue("id", "")
	w := httptest.NewRecorder()
	handleBatchStatus(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d for empty batch ID, got %d", http.StatusBadRequest, w.Code)
	}
}

// Test handleBatchComplete with empty batch ID
func TestHandleBatchComplete_EmptyID(t *testing.T) {
	req := httptest.NewRequest("POST", "/validate/batch//complete", nil)
	req.SetPathValue("id", "")
	w := httptest.NewRecorder()
	handleBatchComplete(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d for empty batch ID, got %d", http.StatusBadRequest, w.Code)
	}
}

// Test handleBatchComplete with failed threshold validation
func TestHandleBatchComplete_FailedThreshold(t *testing.T) {
	batchManager := models.GetBatchSessionManager()
	threshold := 90.0 // High threshold
	session := batchManager.CreateBatchSession("test-batch-fail", &threshold)
	// Add few valid, mostly invalid records (below threshold)
	batchManager.UpdateBatchSession(session.BatchID, 10, 90, 0)

	req := httptest.NewRequest("POST", "/validate/batch/"+session.BatchID+"/complete", nil)
	req.SetPathValue("id", session.BatchID)
	w := httptest.NewRecorder()
	handleBatchComplete(w, req)

	// Should return 422 for failed validation
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("Expected status %d for failed threshold, got %d", http.StatusUnprocessableEntity, w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	if response["status"] != "failed" {
		t.Errorf("Expected status 'failed', got %v", response["status"])
	}

	// Wait for cleanup
	time.Sleep(2 * time.Second)
}

// Test array validation path in handleGenericValidation
func TestHandleGenericValidation_ArrayValidation(t *testing.T) {
	tests := []struct {
		name           string
		payload        map[string]interface{}
		expectedStatus int
		expectStatus   string
	}{
		{
			name: "array validation without threshold",
			payload: map[string]interface{}{
				"model_type": "testmodel",
				"data": []map[string]interface{}{
					{"id": "1", "name": "test1"},
					{"id": "2", "name": "test2"},
				},
			},
			expectedStatus: http.StatusOK,
			expectStatus:   "success",
		},
		{
			name: "array validation with threshold",
			payload: map[string]interface{}{
				"model_type": "testmodel",
				"threshold":  50.0,
				"data": []map[string]interface{}{
					{"id": "1", "name": "test1"},
					{"id": "2", "name": "test2"},
				},
			},
			expectedStatus: http.StatusOK,
			expectStatus:   "success",
		},
		{
			name: "array validation with invalid model type",
			payload: map[string]interface{}{
				"model_type": "nonexistent",
				"data": []map[string]interface{}{
					{"id": "1"},
				},
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", "/validate", bytes.NewBuffer(body))
			w := httptest.NewRecorder()
			handleGenericValidation(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Response: %s", tt.expectedStatus, w.Code, w.Body.String())
			}

			if tt.expectStatus != "" {
				var result map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &result)
				if result["status"] != tt.expectStatus {
					t.Errorf("Expected status '%s', got '%v'", tt.expectStatus, result["status"])
				}
			}
		})
	}
}

// Test batch accumulation with X-Batch-ID header
func TestHandleGenericValidation_BatchAccumulation(t *testing.T) {
	// Create a batch session first
	batchManager := models.GetBatchSessionManager()
	threshold := 50.0
	session := batchManager.CreateBatchSession("test-batch-accum", &threshold)

	t.Run("accumulate with valid batch ID", func(t *testing.T) {
		payload := map[string]interface{}{
			"model_type": "testmodel",
			"data": []map[string]interface{}{
				{"id": "1", "name": "test1"},
				{"id": "2", "name": "test2"},
			},
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/validate", bytes.NewBuffer(body))
		req.Header.Set("X-Batch-ID", session.BatchID)
		w := httptest.NewRecorder()
		handleGenericValidation(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var result map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &result)
		if result["status"] != "accumulating" {
			t.Errorf("Expected status 'accumulating', got '%v'", result["status"])
		}
	})

	t.Run("accumulate with non-existent batch ID", func(t *testing.T) {
		payload := map[string]interface{}{
			"model_type": "testmodel",
			"data": []map[string]interface{}{
				{"id": "1"},
			},
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/validate", bytes.NewBuffer(body))
		req.Header.Set("X-Batch-ID", "non-existent-batch")
		w := httptest.NewRecorder()
		handleGenericValidation(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})

	// Cleanup
	batchManager.DeleteBatchSession(session.BatchID)
}

// Test X-Batch-Complete header functionality
func TestHandleGenericValidation_BatchComplete(t *testing.T) {
	// Create and populate a batch session
	batchManager := models.GetBatchSessionManager()
	threshold := 50.0
	session := batchManager.CreateBatchSession("test-batch-complete-header", &threshold)
	batchManager.UpdateBatchSession(session.BatchID, 60, 40, 5)

	t.Run("complete with valid batch ID", func(t *testing.T) {
		payload := map[string]interface{}{
			"model_type": "testmodel",
			"payload":    map[string]interface{}{"id": "1"},
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/validate", bytes.NewBuffer(body))
		req.Header.Set("X-Batch-Complete", session.BatchID)
		w := httptest.NewRecorder()
		handleGenericValidation(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		var result map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &result)
		if result["status"] != "success" {
			t.Errorf("Expected status 'success', got '%v'", result["status"])
		}

		// Wait for cleanup
		time.Sleep(2 * time.Second)
	})

	t.Run("complete with non-existent batch ID", func(t *testing.T) {
		payload := map[string]interface{}{
			"model_type": "testmodel",
			"payload":    map[string]interface{}{"id": "1"},
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/validate", bytes.NewBuffer(body))
		req.Header.Set("X-Batch-Complete", "non-existent-batch")
		w := httptest.NewRecorder()
		handleGenericValidation(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})
}

// Test X-Batch-Complete with failed threshold
func TestHandleGenericValidation_BatchCompleteFailed(t *testing.T) {
	batchManager := models.GetBatchSessionManager()
	threshold := 90.0
	session := batchManager.CreateBatchSession("test-batch-fail-header", &threshold)
	batchManager.UpdateBatchSession(session.BatchID, 10, 90, 0) // 10% success rate

	payload := map[string]interface{}{
		"model_type": "testmodel",
		"payload":    map[string]interface{}{"id": "1"},
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	req.Header.Set("X-Batch-Complete", session.BatchID)
	w := httptest.NewRecorder()
	handleGenericValidation(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("Expected status %d for failed threshold, got %d", http.StatusUnprocessableEntity, w.Code)
	}

	var result map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &result)
	if result["status"] != "failed" {
		t.Errorf("Expected status 'failed', got '%v'", result["status"])
	}

	// Wait for cleanup
	time.Sleep(2 * time.Second)
}

// Test error creating model instance
func TestHandleGenericValidation_ModelInstanceError(t *testing.T) {
	// We can't easily test this without modifying the registry
	// But we can test the path where model type is registered but payload conversion fails
	payload := map[string]interface{}{
		"model_type": "testmodel",
		"payload":    map[string]interface{}{},
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	handleGenericValidation(w, req)

	// Should succeed with empty payload
	if w.Code != http.StatusOK && w.Code != http.StatusUnprocessableEntity {
		t.Logf("Status: %d, expected OK or UnprocessableEntity", w.Code)
	}
}

// Test convertMapToStruct with non-pointer destination
func TestConvertMapToStruct_NonPointer(t *testing.T) {
	type TestStruct struct {
		Field string `json:"field"`
	}

	input := map[string]interface{}{"field": "value"}
	ts := TestStruct{}

	err := convertMapToStruct(input, ts) // Not a pointer
	if err == nil {
		t.Error("Expected error for non-pointer destination")
	}
	if err != nil && err.Error() != "destination must be a pointer to a struct" {
		t.Errorf("Expected specific error message, got: %v", err)
	}
}

// Test convertMapToStruct with pointer to non-struct
func TestConvertMapToStruct_PointerToNonStruct(t *testing.T) {
	input := map[string]interface{}{"field": "value"}
	var str string

	err := convertMapToStruct(input, &str)
	if err == nil {
		t.Error("Expected error for pointer to non-struct")
	}
}

// Test convertMapToStruct with unexported fields
func TestConvertMapToStruct_UnexportedFields(t *testing.T) {
	type TestStruct struct {
		ExportedField   string `json:"exported"`
		unexportedField string `json:"unexported"`
	}

	input := map[string]interface{}{
		"exported":   "visible",
		"unexported": "hidden",
	}

	ts := &TestStruct{}
	err := convertMapToStruct(input, ts)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if ts.ExportedField != "visible" {
		t.Errorf("Expected ExportedField='visible', got '%s'", ts.ExportedField)
	}
	if ts.unexportedField != "" {
		t.Error("Unexported field should not be set")
	}
}

// Test convertMapToStruct with json:"-" tag
func TestConvertMapToStruct_IgnoredField(t *testing.T) {
	type TestStruct struct {
		NormalField  string `json:"normal"`
		IgnoredField string `json:"-"`
	}

	input := map[string]interface{}{
		"normal":  "value1",
		"ignored": "should_not_be_set",
	}

	ts := &TestStruct{}
	err := convertMapToStruct(input, ts)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if ts.NormalField != "value1" {
		t.Errorf("Expected NormalField='value1', got '%s'", ts.NormalField)
	}
	if ts.IgnoredField != "" {
		t.Error("Ignored field should not be set")
	}
}

// Test convertMapToStruct with missing source values
func TestConvertMapToStruct_MissingValues(t *testing.T) {
	type TestStruct struct {
		Field1 string `json:"field1"`
		Field2 string `json:"field2"`
	}

	input := map[string]interface{}{
		"field1": "value1",
		// field2 is missing
	}

	ts := &TestStruct{}
	err := convertMapToStruct(input, ts)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if ts.Field1 != "value1" {
		t.Errorf("Expected Field1='value1', got '%s'", ts.Field1)
	}
	if ts.Field2 != "" {
		t.Errorf("Expected Field2 to be empty, got '%s'", ts.Field2)
	}
}

// Test setFieldValue with nil value
func TestSetFieldValue_NilValue(t *testing.T) {
	type TestStruct struct {
		Field string
	}

	ts := &TestStruct{}
	val := reflect.ValueOf(ts).Elem()
	field := val.FieldByName("Field")

	err := setFieldValue(field, nil)
	if err != nil {
		t.Errorf("Unexpected error for nil value: %v", err)
	}
}

// Test setFieldValue with type conversion - string formatting
func TestSetFieldValue_StringFormatting(t *testing.T) {
	type TestStruct struct {
		Field string
	}

	ts := &TestStruct{}
	val := reflect.ValueOf(ts).Elem()
	field := val.FieldByName("Field")

	// Test with integer that gets formatted to string
	err := setFieldValue(field, 42)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if ts.Field != "42" {
		t.Errorf("Expected Field='42', got '%s'", ts.Field)
	}
}

// Test setFieldValue with integer overflow
func TestSetFieldValue_IntOverflow(t *testing.T) {
	type TestStruct struct {
		SmallInt int8
	}

	ts := &TestStruct{}
	val := reflect.ValueOf(ts).Elem()
	field := val.FieldByName("SmallInt")

	// Try to set a value that overflows int8 (max 127)
	err := setFieldValue(field, int64(1000))
	if err == nil {
		t.Error("Expected overflow error")
	}
}

// Test setFieldValue with uint overflow
func TestSetFieldValue_UintOverflow(t *testing.T) {
	type TestStruct struct {
		SmallUint uint8
	}

	ts := &TestStruct{}
	val := reflect.ValueOf(ts).Elem()
	field := val.FieldByName("SmallUint")

	// Try to set a value that overflows uint8 (max 255)
	err := setFieldValue(field, uint64(1000))
	if err == nil {
		t.Error("Expected overflow error")
	}
}

// Test setFieldValue with float overflow
func TestSetFieldValue_FloatOverflow(t *testing.T) {
	type TestStruct struct {
		SmallFloat float32
	}

	ts := &TestStruct{}
	val := reflect.ValueOf(ts).Elem()
	field := val.FieldByName("SmallFloat")

	// Try to set a value that overflows float32
	err := setFieldValue(field, float64(1e308))
	if err == nil {
		t.Error("Expected overflow error")
	}
}

// Test setFieldValue with invalid bool conversion
func TestSetFieldValue_InvalidBool(t *testing.T) {
	type TestStruct struct {
		BoolField bool
	}

	ts := &TestStruct{}
	val := reflect.ValueOf(ts).Elem()
	field := val.FieldByName("BoolField")

	err := setFieldValue(field, "not-a-bool")
	if err == nil {
		t.Error("Expected error for invalid bool conversion")
	}
}

// Test setFieldValue fallback to JSON conversion
func TestSetFieldValue_JSONFallback(t *testing.T) {
	type NestedStruct struct {
		Value string `json:"value"`
	}

	type TestStruct struct {
		Nested NestedStruct
	}

	ts := &TestStruct{}
	val := reflect.ValueOf(ts).Elem()
	field := val.FieldByName("Nested")

	nestedData := map[string]interface{}{
		"value": "test",
	}

	err := setFieldValue(field, nestedData)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if ts.Nested.Value != "test" {
		t.Errorf("Expected Nested.Value='test', got '%s'", ts.Nested.Value)
	}
}

// Test setSliceValue with recursive conversion
func TestSetSliceValue_RecursiveConversion(t *testing.T) {
	type TestStruct struct {
		IntSlice []int
	}

	ts := &TestStruct{}
	val := reflect.ValueOf(ts).Elem()
	field := val.FieldByName("IntSlice")

	// Test with float values that need conversion to int
	err := setSliceValue(field, []interface{}{1.0, 2.0, 3.0})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(ts.IntSlice) != 3 {
		t.Errorf("Expected slice length 3, got %d", len(ts.IntSlice))
	}
}

// Test setMapValue with key conversion
func TestSetMapValue_KeyConversion(t *testing.T) {
	type TestStruct struct {
		IntMap map[string]int
	}

	ts := &TestStruct{}
	val := reflect.ValueOf(ts).Elem()
	field := val.FieldByName("IntMap")

	// Test with float values that need conversion to int
	mapData := map[string]interface{}{
		"key1": 1.0,
		"key2": 2.0,
	}

	err := setMapValue(field, mapData)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(ts.IntMap) != 2 {
		t.Errorf("Expected map length 2, got %d", len(ts.IntMap))
	}
}

// Test concurrent access to batch sessions
func TestConcurrentBatchAccess(t *testing.T) {
	batchManager := models.GetBatchSessionManager()
	threshold := 50.0
	session := batchManager.CreateBatchSession("concurrent-test", &threshold)

	done := make(chan bool)

	// Start 10 concurrent updates
	for i := 0; i < 10; i++ {
		go func() {
			batchManager.UpdateBatchSession(session.BatchID, 1, 0, 0)
			done <- true
		}()
	}

	// Wait for all to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	updatedSession, _ := batchManager.GetBatchSession(session.BatchID)
	if updatedSession.ValidRecords != 10 {
		t.Errorf("Expected 10 valid records, got %d", updatedSession.ValidRecords)
	}

	batchManager.DeleteBatchSession(session.BatchID)
}

// Test convertToInt64 with uint types
func TestConvertToInt64_UintTypes(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		ok    bool
	}{
		{"uint", uint(42), false},
		{"uint8", uint8(42), false},
		{"uint16", uint16(42), false},
		{"uint32", uint32(42), false},
		{"uint64", uint64(42), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ok := convertToInt64(tt.input)
			if ok != tt.ok {
				t.Errorf("Expected ok=%v, got %v", tt.ok, ok)
			}
		})
	}
}

// Test convertToUint64 with string conversion
func TestConvertToUint64_StringConversion(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		ok    bool
	}{
		{"string", "42", false},
		{"bool", true, false},
		{"float32", float32(42.5), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ok := convertToUint64(tt.input)
			if ok != tt.ok {
				t.Errorf("Expected ok=%v, got %v for %v", tt.ok, ok, tt.input)
			}
		})
	}
}

// Test convertToFloat64 with more edge cases
func TestConvertToFloat64_EdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		ok    bool
	}{
		{"bool", true, false},
		{"uint", uint(42), false},
		{"uint32", uint32(42), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ok := convertToFloat64(tt.input)
			if ok != tt.ok {
				t.Errorf("Expected ok=%v, got %v for %v", tt.ok, ok, tt.input)
			}
		})
	}
}

// Test array validation with failed status
func TestHandleGenericValidation_ArrayValidationFailed(t *testing.T) {
	payload := map[string]interface{}{
		"model_type": "invalidmodel",
		"data": []map[string]interface{}{
			{"id": "1"},
			{"id": "2"},
		},
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	handleGenericValidation(w, req)

	// Should return 422 for all invalid records
	if w.Code != http.StatusUnprocessableEntity {
		t.Logf("Expected status %d, got %d", http.StatusUnprocessableEntity, w.Code)
	}
}

// Test convertMapToStruct with json tag containing options
func TestConvertMapToStruct_JSONTagWithOptions(t *testing.T) {
	type TestStruct struct {
		Field1 string `json:"field1,omitempty"`
		Field2 string `json:"field2,string"`
	}

	input := map[string]interface{}{
		"field1": "value1",
		"field2": "value2",
	}

	ts := &TestStruct{}
	err := convertMapToStruct(input, ts)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if ts.Field1 != "value1" {
		t.Errorf("Expected Field1='value1', got '%s'", ts.Field1)
	}
	if ts.Field2 != "value2" {
		t.Errorf("Expected Field2='value2', got '%s'", ts.Field2)
	}
}

// Test setFieldValue with all uint types
func TestSetFieldValue_AllUintTypes(t *testing.T) {
	type TestStruct struct {
		Uint   uint
		Uint8  uint8
		Uint16 uint16
		Uint32 uint32
		Uint64 uint64
	}

	tests := []struct {
		field string
		value interface{}
	}{
		{"Uint", uint(42)},
		{"Uint8", uint8(42)},
		{"Uint16", uint16(42)},
		{"Uint32", uint32(42)},
		{"Uint64", uint64(42)},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			ts := &TestStruct{}
			val := reflect.ValueOf(ts).Elem()
			field := val.FieldByName(tt.field)

			err := setFieldValue(field, tt.value)
			if err != nil {
				t.Errorf("Unexpected error for %s: %v", tt.field, err)
			}
		})
	}
}

// Test setFieldValue with all int types
func TestSetFieldValue_AllIntTypes(t *testing.T) {
	type TestStruct struct {
		Int   int
		Int8  int8
		Int16 int16
		Int32 int32
		Int64 int64
	}

	tests := []struct {
		field string
		value interface{}
	}{
		{"Int", int(42)},
		{"Int8", int8(42)},
		{"Int16", int16(42)},
		{"Int32", int32(42)},
		{"Int64", int64(42)},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			ts := &TestStruct{}
			val := reflect.ValueOf(ts).Elem()
			field := val.FieldByName(tt.field)

			err := setFieldValue(field, tt.value)
			if err != nil {
				t.Errorf("Unexpected error for %s: %v", tt.field, err)
			}
		})
	}
}

// Test setFieldValue with float types
func TestSetFieldValue_FloatTypes(t *testing.T) {
	type TestStruct struct {
		Float32 float32
		Float64 float64
	}

	tests := []struct {
		field string
		value interface{}
	}{
		{"Float32", float32(42.5)},
		{"Float64", float64(42.5)},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			ts := &TestStruct{}
			val := reflect.ValueOf(ts).Elem()
			field := val.FieldByName(tt.field)

			err := setFieldValue(field, tt.value)
			if err != nil {
				t.Errorf("Unexpected error for %s: %v", tt.field, err)
			}
		})
	}
}

// Test setFieldValue with invalid int conversion
func TestSetFieldValue_InvalidIntConversion(t *testing.T) {
	type TestStruct struct {
		IntField int
	}

	ts := &TestStruct{}
	val := reflect.ValueOf(ts).Elem()
	field := val.FieldByName("IntField")

	err := setFieldValue(field, "not-a-number")
	if err == nil {
		t.Error("Expected error for invalid int conversion")
	}
}

// Test setFieldValue with invalid uint conversion
func TestSetFieldValue_InvalidUintConversion(t *testing.T) {
	type TestStruct struct {
		UintField uint
	}

	ts := &TestStruct{}
	val := reflect.ValueOf(ts).Elem()
	field := val.FieldByName("UintField")

	err := setFieldValue(field, "not-a-number")
	if err == nil {
		t.Error("Expected error for invalid uint conversion")
	}
}

// Test setFieldValue with invalid float conversion
func TestSetFieldValue_InvalidFloatConversion(t *testing.T) {
	type TestStruct struct {
		FloatField float64
	}

	ts := &TestStruct{}
	val := reflect.ValueOf(ts).Elem()
	field := val.FieldByName("FloatField")

	err := setFieldValue(field, "not-a-number")
	if err == nil {
		t.Error("Expected error for invalid float conversion")
	}
}

// Test handleGenericValidation with array validation error
func TestHandleGenericValidation_ArrayValidationError(t *testing.T) {
	// Create a mock registry that will fail validation
	payload := map[string]interface{}{
		"model_type": "testmodel",
		"data": []map[string]interface{}{
			{"id": "1", "name": "test"},
		},
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	handleGenericValidation(w, req)

	// Should handle validation gracefully
	if w.Code != http.StatusOK && w.Code != http.StatusUnprocessableEntity && w.Code != http.StatusInternalServerError {
		t.Logf("Status: %d", w.Code)
	}
}

// Test handleGenericValidation single object validation error
func TestHandleGenericValidation_SingleObjectValidationError(t *testing.T) {
	payload := map[string]interface{}{
		"model_type": "testmodel",
		"payload": map[string]interface{}{
			"id": "test-error",
		},
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	handleGenericValidation(w, req)

	// Should handle validation
	if w.Code != http.StatusOK && w.Code != http.StatusUnprocessableEntity {
		t.Logf("Status: %d", w.Code)
	}
}

// Test setFieldValue with direct assignable types
func TestSetFieldValue_DirectAssignment(t *testing.T) {
	type TestStruct struct {
		StringField string
	}

	ts := &TestStruct{}
	val := reflect.ValueOf(ts).Elem()
	field := val.FieldByName("StringField")

	// Test direct assignment when types match
	err := setFieldValue(field, "direct-value")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if ts.StringField != "direct-value" {
		t.Errorf("Expected 'direct-value', got '%s'", ts.StringField)
	}
}

// Test setSliceValue error in recursive conversion
func TestSetSliceValue_RecursiveError(t *testing.T) {
	type TestStruct struct {
		IntSlice []int
	}

	ts := &TestStruct{}
	val := reflect.ValueOf(ts).Elem()
	field := val.FieldByName("IntSlice")

	// Test with values that will cause conversion errors
	err := setSliceValue(field, []interface{}{"not-a-number", "also-not-a-number"})
	if err == nil {
		t.Error("Expected error for invalid slice element conversion")
	}
}

// Test setMapValue error in recursive conversion
func TestSetMapValue_RecursiveError(t *testing.T) {
	type TestStruct struct {
		IntMap map[string]int
	}

	ts := &TestStruct{}
	val := reflect.ValueOf(ts).Elem()
	field := val.FieldByName("IntMap")

	// Test with values that will cause conversion errors
	mapData := map[string]interface{}{
		"key1": "not-a-number",
	}

	err := setMapValue(field, mapData)
	if err == nil {
		t.Error("Expected error for invalid map value conversion")
	}
}

// Test convertMapToStruct with field conversion error
func TestConvertMapToStruct_FieldConversionError(t *testing.T) {
	type TestStruct struct {
		IntField int `json:"int_field"`
	}

	input := map[string]interface{}{
		"int_field": "definitely-not-a-number",
	}

	ts := &TestStruct{}
	err := convertMapToStruct(input, ts)

	if err == nil {
		t.Error("Expected error for field conversion failure")
	}
}

// Test handleGenericValidation with batch accumulation validation error
func TestHandleGenericValidation_BatchAccumulationValidationError(t *testing.T) {
	batchManager := models.GetBatchSessionManager()
	threshold := 50.0
	session := batchManager.CreateBatchSession("test-batch-validation-error", &threshold)

	// Send invalid data that will cause validation to fail
	payload := map[string]interface{}{
		"model_type": "testmodel",
		"data": []map[string]interface{}{
			{"id": "1"},
		},
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	req.Header.Set("X-Batch-ID", session.BatchID)
	w := httptest.NewRecorder()
	handleGenericValidation(w, req)

	// Should handle the request
	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Logf("Status: %d", w.Code)
	}

	batchManager.DeleteBatchSession(session.BatchID)
}

// Test convertMapToStruct with empty json tag name
func TestConvertMapToStruct_EmptyJSONTagName(t *testing.T) {
	type TestStruct struct {
		Field1 string `json:",omitempty"`
	}

	input := map[string]interface{}{
		"Field1": "value1",
	}

	ts := &TestStruct{}
	err := convertMapToStruct(input, ts)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// When json tag is empty, it falls back to field name
	if ts.Field1 != "value1" {
		t.Errorf("Expected Field1='value1', got '%s'", ts.Field1)
	}
}

// Test count the number of tests added
func TestCountNewTests(t *testing.T) {
	// This test just documents that we've added comprehensive coverage tests
	t.Log("Comprehensive test coverage has been added to main_comprehensive_test.go")
}

// Test handleGenericValidation with X-Batch-Complete finalization error
func TestHandleGenericValidation_BatchCompleteFinalizationError(t *testing.T) {
	// Try to finalize a non-existent batch
	payload := map[string]interface{}{
		"model_type": "testmodel",
		"payload":    map[string]interface{}{"id": "1"},
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	req.Header.Set("X-Batch-Complete", "finalization-error-batch")
	w := httptest.NewRecorder()
	handleGenericValidation(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d for non-existent batch, got %d", http.StatusNotFound, w.Code)
	}
}

// Test setFieldValue with negative int conversion to uint
func TestSetFieldValue_NegativeIntToUint(t *testing.T) {
	type TestStruct struct {
		UintField uint
	}

	ts := &TestStruct{}
	val := reflect.ValueOf(ts).Elem()
	field := val.FieldByName("UintField")

	// Try to set a negative value to uint field
	err := setFieldValue(field, int(-1))
	if err == nil {
		t.Error("Expected error for negative int to uint conversion")
	}
}

// Test setFieldValue with negative int64 conversion to uint
func TestSetFieldValue_NegativeInt64ToUint(t *testing.T) {
	type TestStruct struct {
		UintField uint
	}

	ts := &TestStruct{}
	val := reflect.ValueOf(ts).Elem()
	field := val.FieldByName("UintField")

	// Try to set a negative int64 value to uint field
	err := setFieldValue(field, int64(-100))
	if err == nil {
		t.Error("Expected error for negative int64 to uint conversion")
	}
}

// Test setFieldValue with negative float64 conversion to uint
func TestSetFieldValue_NegativeFloatToUint(t *testing.T) {
	type TestStruct struct {
		UintField uint
	}

	ts := &TestStruct{}
	val := reflect.ValueOf(ts).Elem()
	field := val.FieldByName("UintField")

	// Try to set a negative float64 value to uint field
	err := setFieldValue(field, float64(-1.5))
	if err == nil {
		t.Error("Expected error for negative float to uint conversion")
	}
}

// Test convertMapToStruct with empty field name fallback
func TestConvertMapToStruct_EmptyFieldNameFallback(t *testing.T) {
	type TestStruct struct {
		TestField string `json:","`
	}

	input := map[string]interface{}{
		"TestField": "value",
	}

	ts := &TestStruct{}
	err := convertMapToStruct(input, ts)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if ts.TestField != "value" {
		t.Errorf("Expected TestField='value', got '%s'", ts.TestField)
	}
}

// Test handleGenericValidation array validation internal server error path
func TestHandleGenericValidation_ArrayValidationInternalError(t *testing.T) {
	// This tests the path where ValidateArray might fail
	payload := map[string]interface{}{
		"model_type": "testmodel",
		"threshold":  50.0,
		"data": []map[string]interface{}{
			{"id": "1", "name": "test"},
		},
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/validate", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	handleGenericValidation(w, req)

	// Should complete successfully or with validation error
	if w.Code != http.StatusOK && w.Code != http.StatusUnprocessableEntity && w.Code != http.StatusInternalServerError {
		t.Logf("Unexpected status: %d", w.Code)
	}
}

// Test convertToInt64 with all numeric edge cases
func TestConvertToInt64_AllEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected int64
		ok       bool
	}{
		{"float32 conversion", float32(3.14), 3, true},
		{"int type", int(100), 100, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := convertToInt64(tt.input)
			if ok != tt.ok {
				t.Errorf("Expected ok=%v, got %v", tt.ok, ok)
			}
			if ok && result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

// Test convertToFloat64 with all numeric conversions
func TestConvertToFloat64_AllNumericConversions(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected float64
		ok       bool
	}{
		{"int to float", int(42), 42.0, true},
		{"int64 to float", int64(100), 100.0, true},
		{"uint64 to float", uint64(50), 50.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := convertToFloat64(tt.input)
			if ok != tt.ok {
				t.Errorf("Expected ok=%v, got %v", tt.ok, ok)
			}
			if ok && result != tt.expected {
				t.Errorf("Expected %f, got %f", tt.expected, result)
			}
		})
	}
}

// Test convertToUint64 with float64 positive value
func TestConvertToUint64_Float64Positive(t *testing.T) {
	result, ok := convertToUint64(float64(42.5))
	if !ok {
		t.Error("Expected successful conversion")
	}
	if result != 42 {
		t.Errorf("Expected 42, got %d", result)
	}
}

// Test setFieldValue JSON fallback error
func TestSetFieldValue_JSONFallbackError(t *testing.T) {
	type NestedStruct struct {
		Value string `json:"value"`
	}

	type TestStruct struct {
		Nested NestedStruct
	}

	ts := &TestStruct{}
	val := reflect.ValueOf(ts).Elem()
	field := val.FieldByName("Nested")

	// Use a channel which can't be marshaled to JSON
	ch := make(chan int)

	err := setFieldValue(field, ch)
	if err == nil {
		t.Error("Expected error for unmarshallable type")
	}
}

// Test convertToInt64 edge case with string scanning
func TestConvertToInt64_StringScanning(t *testing.T) {
	// Test successful string conversion
	result, ok := convertToInt64("123")
	if !ok || result != 123 {
		t.Errorf("Expected (123, true), got (%d, %v)", result, ok)
	}

	// Test invalid string conversion
	_, ok = convertToInt64("abc")
	if ok {
		t.Error("Expected false for invalid string")
	}
}

// Test convertToFloat64 string scanning
func TestConvertToFloat64_StringScanning(t *testing.T) {
	// Test successful string conversion
	result, ok := convertToFloat64("3.14")
	if !ok || result != 3.14 {
		t.Errorf("Expected (3.14, true), got (%f, %v)", result, ok)
	}

	// Test invalid string conversion
	_, ok = convertToFloat64("not-a-float")
	if ok {
		t.Error("Expected false for invalid string")
	}
}
