# Unit Testing Guide for Go Playground Data Validator

## Table of Contents
1. [Overview & Coverage Requirements](#overview--coverage-requirements)
2. [Testing New Models](#testing-new-models)
3. [Testing New Validators](#testing-new-validators)
4. [Testing Main Package (HTTP Handlers & Core Logic)](#testing-main-package-http-handlers--core-logic)
5. [Testing Registry Package](#testing-registry-package)
6. [Best Practices](#best-practices)
7. [Running Tests](#running-tests)
8. [Common Patterns](#common-patterns)
9. [Coverage Analysis](#coverage-analysis)
10. [Quick Reference Checklist](#quick-reference-checklist)

---

## Overview & Coverage Requirements

This guide provides comprehensive instructions for adding unit tests for new models and validations in the Go Playground Data Validator project. The testing framework is **model-agnostic** and designed to be easy to extend.

### Current Coverage Status
```
Overall Coverage: 84.6% (Exceeds 80% requirement)
├── models:      81.5% (1,571 lines)
├── validations: 84.6% (3,133 lines)
├── registry:    95.5% (914 lines)
├── main:        79.6% (1,151 lines)
└── config:     100.0% (27 lines)
```

### Test Coverage Requirements

#### Jenkins Pipeline Requirements
- **Minimum**: 80% overall weighted coverage
- **Calculation**: Weighted by lines of code per package
- **Formula**: `Σ(package_lines × package_coverage) / total_lines`

#### Package-Specific Targets
| Package | Target Coverage | Priority |
|---------|----------------|----------|
| models | ≥80% | High |
| validations | ≥80% | High |
| registry | ≥70% | Medium |
| main | ≥75% | High |
| config | 100% | Low (small package) |

### Key Testing Libraries
- **Standard library**: `testing` package
- **Assertions**: `github.com/stretchr/testify/assert`
- **HTTP testing**: `net/http/httptest`
- **Reflection**: `reflect` package for type testing
- **Validator**: `github.com/go-playground/validator/v10`

### Testing Architecture

#### Easy to Add (No Modifications Required)
- **Models Package** (`src/models/`)
- **Validations Package** (`src/validations/`)

#### Model-Agnostic (No Specific Model Dependencies)
- **Main Package** (`src/main_test.go`)
- **Registry Package** (`src/registry/`)

---

## Testing New Models

### Location
All model tests go in: `src/models/` directory with `_test.go` suffix

### File Structure for New Models
When adding a new model called `MyModel`, create these files:

```
src/
├── models/
│   ├── my_model.go          # Model definition
│   └── my_model_test.go     # Model validation tests
├── validations/
│   ├── my_model.go          # Custom validator
│   └── my_model_test.go     # Validator tests
└── main_test.go             # NO CHANGES NEEDED!
```

### Pattern 1: Table-Driven Validation Tests

```go
package models

import (
    "encoding/json"
    "testing"
    "github.com/go-playground/validator/v10"
)

// TestMyModelPayload_Validation tests all validation scenarios
func TestMyModelPayload_Validation(t *testing.T) {
    tests := []struct {
        name        string
        payload     MyModelPayload
        expectValid bool
        description string
    }{
        {
            name: "valid my_model payload",
            payload: MyModelPayload{
                ID:    "MYMODEL-001",
                Name:  "Valid Model",
                Type:  "standard",
                // ... other required fields
            },
            expectValid: true,
            description: "All required fields with valid values",
        },
        {
            name: "missing required ID",
            payload: MyModelPayload{
                Name: "Missing ID Model",
                Type: "standard",
            },
            expectValid: false,
            description: "Should fail when required ID is missing",
        },
        // Add more test cases...
    }

    v := validator.New()
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := v.Struct(tt.payload)

            if tt.expectValid && err != nil {
                t.Errorf("Expected valid payload, got error: %v", err)
            }
            if !tt.expectValid && err == nil {
                t.Error("Expected validation error, got none")
            }
        })
    }
}
```

### Pattern 2: JSON Marshaling Tests

```go
// TestMyModelPayload_JSONMarshaling tests JSON serialization
func TestMyModelPayload_JSONMarshaling(t *testing.T) {
    payload := MyModelPayload{
        ID:   "MYMODEL-001",
        Name: "Test Model",
        Type: "standard",
    }

    jsonData, err := json.Marshal(payload)
    if err != nil {
        t.Errorf("Failed to marshal MyModelPayload: %v", err)
    }

    var unmarshaled MyModelPayload
    err = json.Unmarshal(jsonData, &unmarshaled)
    if err != nil {
        t.Errorf("Failed to unmarshal MyModelPayload: %v", err)
    }

    if unmarshaled.ID != payload.ID {
        t.Errorf("Expected ID %s, got %s", payload.ID, unmarshaled.ID)
    }
}
```

### Model-Specific Test Cases

#### 1. Required Fields
```go
func TestModel_RequiredFields(t *testing.T) {
    model := YourModel{
        // Omit required fields
    }

    err := validate.Struct(model)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "required")
}
```

#### 2. Field Validation Tags
```go
func TestModel_FieldValidation(t *testing.T) {
    tests := []struct {
        name  string
        field string
        value interface{}
        valid bool
    }{
        {"email valid", "Email", "test@example.com", true},
        {"email invalid", "Email", "invalid-email", false},
        {"url valid", "URL", "https://example.com", true},
        {"url invalid", "URL", "not-a-url", false},
        {"min length", "Name", "ab", false}, // min=3
        {"max length", "Name", "a"*101, false}, // max=100
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test individual field validation
        })
    }
}
```

#### 3. Enum Validation
```go
func TestModel_EnumValidation(t *testing.T) {
    validStatuses := []string{"pending", "active", "inactive", "archived"}
    invalidStatuses := []string{"invalid", "unknown", ""}

    for _, status := range validStatuses {
        model := YourModel{Status: status}
        err := validate.Struct(model)
        assert.NoError(t, err, "Status %s should be valid", status)
    }

    for _, status := range invalidStatuses {
        model := YourModel{Status: status}
        err := validate.Struct(model)
        assert.Error(t, err, "Status %s should be invalid", status)
    }
}
```

#### 4. Custom Validators
```go
func TestModel_CustomValidation(t *testing.T) {
    // Test custom validation tags like github_username, hexcolor, etc.
    tests := []struct {
        name     string
        username string
        valid    bool
    }{
        {"valid username", "octocat", true},
        {"valid with hyphen", "octo-cat", true},
        {"invalid special chars", "octo@cat", false},
        {"invalid spaces", "octo cat", false},
        {"too short", "ab", false}, // min=3
        {"too long", strings.Repeat("a", 40), false}, // max=39
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            model := GitHubPayload{Username: tt.username}
            err := validate.Struct(model)
            if tt.valid {
                assert.NoError(t, err)
            } else {
                assert.Error(t, err)
            }
        })
    }
}
```

### Testing Array Validation and Threshold Features

#### Testing Batch Session Manager
```go
// TestBatchSessionManager tests the batch session tracking
func TestBatchSessionManager(t *testing.T) {
    manager := models.GetBatchSessionManager()

    tests := []struct {
        name           string
        threshold      *float64
        validRecords   int
        invalidRecords int
        expectStatus   string
    }{
        {
            name:           "80% valid with 20% threshold - success",
            threshold:      floatPtr(20.0),
            validRecords:   80,
            invalidRecords: 20,
            expectStatus:   "success",
        },
        {
            name:           "10% valid with 20% threshold - failed",
            threshold:      floatPtr(20.0),
            validRecords:   10,
            invalidRecords: 90,
            expectStatus:   "failed",
        },
        {
            name:           "exactly 20% valid with 20% threshold - success",
            threshold:      floatPtr(20.0),
            validRecords:   20,
            invalidRecords: 80,
            expectStatus:   "success",
        },
        {
            name:           "no threshold with mixed records - success",
            threshold:      nil,
            validRecords:   50,
            invalidRecords: 50,
            expectStatus:   "success",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            batchID := "test_batch_" + tt.name
            session := manager.CreateBatchSession(batchID, tt.threshold)

            if session == nil {
                t.Fatal("CreateBatchSession returned nil")
            }

            manager.UpdateBatchSession(batchID, tt.validRecords, tt.invalidRecords, 0)
            status, err := manager.FinalizeBatchSession(batchID)

            if err != nil {
                t.Fatalf("FinalizeBatchSession failed: %v", err)
            }

            if status != tt.expectStatus {
                t.Errorf("Expected status %s, got %s", tt.expectStatus, status)
            }
        })
    }
}

// Helper function for creating float64 pointers
func floatPtr(f float64) *float64 {
    return &f
}
```

#### Testing Result Filtering
```go
// TestArrayValidation_ResultFiltering tests that valid rows are excluded
func TestArrayValidation_ResultFiltering(t *testing.T) {
    // Create test data with mixed valid/invalid records
    records := []map[string]interface{}{
        {
            "id":    "VALID-001",
            "name":  "Valid Record 1",
            "email": "valid1@example.com",
            "age":   25,
        },
        {
            "id":    "VALID-002",
            "name":  "Valid Record 2",
            "email": "valid2@example.com",
            "age":   30,
        },
        {
            "id":    "INVALID",
            "name":  "x", // Too short
            "email": "invalid-email",
            "age":   5, // Below minimum
        },
    }

    // Validate array
    result, err := registry.ValidateArray("mymodel", records, nil)
    if err != nil {
        t.Fatalf("ValidateArray failed: %v", err)
    }

    // Check that only invalid rows are in results
    if len(result.Results) != 1 {
        t.Errorf("Expected 1 invalid row in results, got %d", len(result.Results))
    }

    // Check that the invalid row is the correct one
    if len(result.Results) > 0 && result.Results[0].RecordIdentifier != "INVALID" {
        t.Errorf("Expected invalid row to have ID 'INVALID', got %s",
            result.Results[0].RecordIdentifier)
    }

    // Verify counts
    if result.TotalRecords != 3 {
        t.Errorf("Expected 3 total records, got %d", result.TotalRecords)
    }
    if result.ValidRecords != 2 {
        t.Errorf("Expected 2 valid records, got %d", result.ValidRecords)
    }
    if result.InvalidRecords != 1 {
        t.Errorf("Expected 1 invalid record, got %d", result.InvalidRecords)
    }
}
```

---

## Testing New Validators

### Location
All validator tests go in: `src/validations/` directory

### File Naming Convention
- Main validator tests: `validations/<validator_name>_test.go`
- Comprehensive tests: `validations/comprehensive_validators_test.go`

### Pattern: Basic Validator Test Structure

```go
package validations

import (
    "strings"
    "testing"
    "time"
    "goplayground-data-validator/models"
)

func TestNewMyModelValidator(t *testing.T) {
    validator := NewMyModelValidator()
    if validator == nil {
        t.Error("NewMyModelValidator should not return nil")
    }
}

func TestMyModelValidator_ValidatePayload(t *testing.T) {
    tests := []struct {
        name           string
        payload        models.MyModelPayload
        expectValid    bool
        expectErrors   int
        expectWarnings int
        checkErrorField string
    }{
        {
            name: "valid payload",
            payload: getValidMyModelPayload(),
            expectValid: true,
            expectErrors: 0,
            expectWarnings: 0,
        },
        {
            name: "invalid custom validation",
            payload: func() models.MyModelPayload {
                p := getValidMyModelPayload()
                p.ID = "INVALID-FORMAT" // Violates custom ID format
                return p
            }(),
            expectValid: false,
            expectErrors: 1,
            checkErrorField: "id",
        },
        // Add more test cases for custom validations...
    }

    validator := NewMyModelValidator()
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := validator.ValidatePayload(tt.payload)

            if result.IsValid != tt.expectValid {
                t.Errorf("Expected IsValid=%v, got %v", tt.expectValid, result.IsValid)
            }

            if len(result.Errors) != tt.expectErrors {
                t.Errorf("Expected %d errors, got %d", tt.expectErrors, len(result.Errors))
            }

            if len(result.Warnings) != tt.expectWarnings {
                t.Errorf("Expected %d warnings, got %d", tt.expectWarnings, len(result.Warnings))
            }

            if tt.checkErrorField != "" && len(result.Errors) > 0 {
                found := false
                for _, err := range result.Errors {
                    if err.Field == tt.checkErrorField {
                        found = true
                        break
                    }
                }
                if !found {
                    t.Errorf("Expected error in field %s", tt.checkErrorField)
                }
            }
        })
    }
}

// Helper function to create valid test payload
func getValidMyModelPayload() models.MyModelPayload {
    return models.MyModelPayload{
        ID:        "MYMODEL-20240927-0001",
        Name:      "Valid test model payload",
        Type:      "standard",
        Status:    "active",
        CreatedAt: time.Now(),
        // ... other fields
    }
}
```

### Validator-Specific Test Cases

#### 1. Business Logic Validation
```go
func TestValidator_BusinessLogic(t *testing.T) {
    validator := NewGitHubValidator()

    tests := []struct {
        name           string
        payload        models.GitHubPayload
        expectWarnings bool
        warningMsg     string
    }{
        {
            name: "large changeset warning",
            payload: models.GitHubPayload{
                // Valid required fields...
                FilesChanged:    101, // Threshold: 100
                LinesAdded:      1500,
                LinesDeleted:    500,
            },
            expectWarnings: true,
            warningMsg:     "large changeset",
        },
        {
            name: "WIP detection in title",
            payload: models.GitHubPayload{
                // Valid required fields...
                PRTitle: "WIP: New feature",
            },
            expectWarnings: true,
            warningMsg:     "work in progress",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := validator.ValidatePayload(tt.payload)

            if tt.expectWarnings {
                assert.NotEmpty(t, result.Warnings)
                assert.Contains(t, result.Warnings[0].Message, tt.warningMsg)
            } else {
                assert.Empty(t, result.Warnings)
            }
        })
    }
}
```

#### 2. Custom Validation Methods
```go
// Test custom validation methods
func TestMyModelValidator_validateCustomRules(t *testing.T) {
    validator := NewMyModelValidator()

    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"valid format", "MYMODEL-20240927-0001", false},
        {"invalid format", "INVALID-FORMAT", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validator.validateMyModelIDFormat(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("validateMyModelIDFormat() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

#### 3. Business Logic Warnings
```go
// Test business logic warnings
func TestMyModelValidator_validateBusinessLogic(t *testing.T) {
    validator := NewMyModelValidator()

    tests := []struct {
        name            string
        payload         models.MyModelPayload
        expectWarnings  int
        warningContains string
    }{
        {
            name: "no warnings",
            payload: getValidMyModelPayload(),
            expectWarnings: 0,
        },
        {
            name: "business logic warning",
            payload: func() models.MyModelPayload {
                p := getValidMyModelPayload()
                p.Status = "deprecated" // Triggers business warning
                return p
            }(),
            expectWarnings: 1,
            warningContains: "deprecated",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            warnings := validator.validateBusinessLogic(tt.payload)

            if len(warnings) != tt.expectWarnings {
                t.Errorf("Expected %d warnings, got %d", tt.expectWarnings, len(warnings))
            }

            if tt.warningContains != "" && len(warnings) > 0 {
                found := false
                for _, warning := range warnings {
                    if strings.Contains(strings.ToLower(warning.Message), strings.ToLower(tt.warningContains)) {
                        found = true
                        break
                    }
                }
                if !found {
                    t.Errorf("Expected warning containing '%s'", tt.warningContains)
                }
            }
        })
    }
}
```

#### 4. Error Formatting
```go
func TestValidator_ErrorFormatting(t *testing.T) {
    validator := NewYourValidator()

    // Test that errors are properly formatted with:
    // - Field name
    // - Error message
    // - Error code (if applicable)

    payload := models.YourModel{
        // Invalid field
        Email: "invalid-email",
    }

    result := validator.ValidatePayload(payload)

    assert.False(t, result.IsValid)
    assert.NotEmpty(t, result.Errors)
    assert.Equal(t, "Email", result.Errors[0].Field)
    assert.Contains(t, result.Errors[0].Message, "email")
}
```

#### 5. Edge Cases
```go
func TestValidator_EdgeCases(t *testing.T) {
    validator := NewYourValidator()

    tests := []struct {
        name    string
        payload interface{}
        isValid bool
    }{
        {
            name:    "nil payload",
            payload: nil,
            isValid: false,
        },
        {
            name:    "empty struct",
            payload: models.YourModel{},
            isValid: false,
        },
        {
            name: "zero values",
            payload: models.YourModel{
                ID:    "",
                Count: 0,
            },
            isValid: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := validator.ValidatePayload(tt.payload)
            assert.Equal(t, tt.isValid, result.IsValid)
        })
    }
}
```

#### 6. Threshold Validation
```go
// TestMyModelValidator_ThresholdValidation tests threshold behavior
func TestMyModelValidator_ThresholdValidation(t *testing.T) {
    validator := NewMyModelValidator()

    tests := []struct {
        name            string
        payloads        []models.MyModelPayload
        threshold       *float64
        expectStatus    string
        expectValid     int
        expectInvalid   int
    }{
        {
            name: "batch with 80% valid, 20% threshold",
            payloads: []models.MyModelPayload{
                getValidMyModelPayload(),
                getValidMyModelPayload(),
                getValidMyModelPayload(),
                getValidMyModelPayload(),
                getInvalidMyModelPayload(),
            },
            threshold:     floatPtr(20.0),
            expectStatus:  "success",
            expectValid:   4,
            expectInvalid: 1,
        },
        {
            name: "batch with 10% valid, 20% threshold",
            payloads: []models.MyModelPayload{
                getValidMyModelPayload(),
                getInvalidMyModelPayload(),
                getInvalidMyModelPayload(),
                getInvalidMyModelPayload(),
                getInvalidMyModelPayload(),
            },
            threshold:     floatPtr(20.0),
            expectStatus:  "failed",
            expectValid:   1,
            expectInvalid: 4,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            validCount := 0
            invalidCount := 0

            for _, payload := range tt.payloads {
                result := validator.ValidatePayload(payload)
                if result.IsValid {
                    validCount++
                } else {
                    invalidCount++
                }
            }

            if validCount != tt.expectValid {
                t.Errorf("Expected %d valid, got %d", tt.expectValid, validCount)
            }
            if invalidCount != tt.expectInvalid {
                t.Errorf("Expected %d invalid, got %d", tt.expectInvalid, invalidCount)
            }

            // Calculate success rate
            successRate := (float64(validCount) / float64(len(tt.payloads))) * 100.0

            var status string
            if tt.threshold != nil {
                if successRate >= *tt.threshold {
                    status = "success"
                } else {
                    status = "failed"
                }
            } else {
                status = "success"
            }

            if status != tt.expectStatus {
                t.Errorf("Expected status %s, got %s", tt.expectStatus, status)
            }
        })
    }
}
```

---

## Testing Main Package (HTTP Handlers & Core Logic)

### Location
`main_comprehensive_test.go` in the `src/` directory

### Key Features
- Model-agnostic tests using generic test models
- Uses `testmodel` and `invalidmodel` for testing
- No imports from `models` package required
- Tests both valid and invalid scenarios

### HTTP Handler Testing

#### 1. Batch Start Handler
```go
func TestHandleBatchStart(t *testing.T) {
    tests := []struct {
        name           string
        payload        map[string]interface{}
        expectedStatus int
    }{
        {
            name: "valid batch start",
            payload: map[string]interface{}{
                "model_type": "github",
                "job_id":     "test-job-123",
                "threshold":  20.0,
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
        {
            name: "invalid threshold",
            payload: map[string]interface{}{
                "model_type": "github",
                "threshold":  150.0, // > 100
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

            assert.Equal(t, tt.expectedStatus, w.Code)

            if tt.expectedStatus == http.StatusOK {
                var response map[string]interface{}
                json.Unmarshal(w.Body.Bytes(), &response)

                assert.NotNil(t, response["batch_id"])
                assert.Equal(t, "active", response["status"])
            }
        })
    }
}
```

#### 2. Error Path Testing
```go
func TestHandleBatchStart_ErrorPaths(t *testing.T) {
    t.Run("invalid JSON", func(t *testing.T) {
        req := httptest.NewRequest("POST", "/validate/batch/start",
            bytes.NewBufferString("invalid json"))
        w := httptest.NewRecorder()

        handleBatchStart(w, req)

        assert.Equal(t, http.StatusBadRequest, w.Code)
    })

    t.Run("empty body", func(t *testing.T) {
        req := httptest.NewRequest("POST", "/validate/batch/start", nil)
        w := httptest.NewRecorder()

        handleBatchStart(w, req)

        assert.Equal(t, http.StatusBadRequest, w.Code)
    })
}
```

#### 3. Batch Complete Handler
```go
func TestHandleBatchComplete(t *testing.T) {
    batchManager := models.GetBatchSessionManager()

    tests := []struct {
        name           string
        setupBatch     func() string
        expectedStatus int
        expectPass     bool
    }{
        {
            name: "batch passes threshold",
            setupBatch: func() string {
                threshold := 90.0
                session := batchManager.CreateBatchSession("test-pass", &threshold)
                // 95% pass rate (95 valid, 5 invalid)
                batchManager.UpdateBatchSession(session.BatchID, 95, 5, 0)
                return session.BatchID
            },
            expectedStatus: http.StatusOK,
            expectPass:     true,
        },
        {
            name: "batch fails threshold",
            setupBatch: func() string {
                threshold := 90.0
                session := batchManager.CreateBatchSession("test-fail", &threshold)
                // 80% pass rate (80 valid, 20 invalid)
                batchManager.UpdateBatchSession(session.BatchID, 80, 20, 0)
                return session.BatchID
            },
            expectedStatus: http.StatusUnprocessableEntity,
            expectPass:     false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            batchID := tt.setupBatch()

            req := httptest.NewRequest("POST", "/validate/batch/"+batchID+"/complete", nil)
            req.SetPathValue("id", batchID)
            w := httptest.NewRecorder()

            handleBatchComplete(w, req)

            assert.Equal(t, tt.expectedStatus, w.Code)

            var response map[string]interface{}
            json.Unmarshal(w.Body.Bytes(), &response)

            if tt.expectPass {
                assert.Equal(t, "completed", response["status"])
                assert.Equal(t, true, response["threshold_passed"])
            } else {
                assert.Equal(t, "failed", response["status"])
                assert.Equal(t, false, response["threshold_passed"])
            }
        })
    }
}
```

### Type Conversion Testing

#### 1. Numeric Type Conversions
```go
func TestConvertToInt64(t *testing.T) {
    tests := []struct {
        name      string
        value     interface{}
        expected  int64
        expectErr bool
    }{
        {"int", 123, 123, false},
        {"int32", int32(456), 456, false},
        {"int64", int64(789), 789, false},
        {"float64", float64(100.5), 100, false},
        {"string number", "999", 999, false},
        {"string invalid", "abc", 0, true},
        {"nil", nil, 0, true},
        {"bool", true, 0, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := convertToInt64(tt.value)

            if tt.expectErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.expected, result)
            }
        })
    }
}
```

#### 2. Overflow Detection
```go
func TestConvertToUint64_Overflow(t *testing.T) {
    tests := []struct {
        name      string
        value     interface{}
        expectErr bool
    }{
        {"negative int", -1, true},
        {"negative float", -10.5, true},
        {"max uint64", uint64(math.MaxUint64), false},
        {"overflow from string", "18446744073709551616", true}, // > MaxUint64
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := convertToUint64(tt.value)

            if tt.expectErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### Field Setter Testing

```go
func TestSetFieldValue(t *testing.T) {
    type TestStruct struct {
        StringField string
        IntField    int
        BoolField   bool
        FloatField  float64
    }

    tests := []struct {
        name       string
        fieldName  string
        value      interface{}
        expectPass bool
    }{
        {"set string", "StringField", "test", true},
        {"set int", "IntField", 123, true},
        {"set bool", "BoolField", true, true},
        {"set float", "FloatField", 45.67, true},
        {"invalid field", "NonExistent", "value", false},
        {"type mismatch", "IntField", "not-an-int", false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            obj := &TestStruct{}
            err := setFieldValue(reflect.ValueOf(obj).Elem(), tt.fieldName, tt.value)

            if tt.expectPass {
                assert.NoError(t, err)
            } else {
                assert.Error(t, err)
            }
        })
    }
}
```

### Concurrent Testing

```go
func TestConcurrentBatchAccess(t *testing.T) {
    batchManager := models.GetBatchSessionManager()
    threshold := 50.0
    session := batchManager.CreateBatchSession("concurrent-test", &threshold)

    const numGoroutines = 100
    done := make(chan bool, numGoroutines)

    // Spawn multiple goroutines updating the same batch
    for i := 0; i < numGoroutines; i++ {
        go func() {
            batchManager.UpdateBatchSession(session.BatchID, 1, 0, 0)
            done <- true
        }()
    }

    // Wait for all goroutines to complete
    for i := 0; i < numGoroutines; i++ {
        <-done
    }

    // Verify final state
    finalSession := batchManager.GetBatchSession(session.BatchID)
    assert.Equal(t, int32(numGoroutines), finalSession.ValidCount)
}
```

---

## Testing Registry Package

### Location
`registry/unified_registry_test.go`

### Registry Test Patterns

#### 1. Model Registration
```go
func TestUnifiedRegistry_RegisterModel(t *testing.T) {
    registry := NewUnifiedRegistry("test/models", "test/validations")

    modelInfo := &ModelInfo{
        Type:        "test_model",
        Name:        "Test Model",
        Description: "Test model description",
        ModelStruct: reflect.TypeOf(models.GenericPayload{}),
        Version:     "1.0.0",
        CreatedAt:   time.Now().Format(time.RFC3339),
        Author:      "Test Author",
        Tags:        []string{"test", "example"},
    }

    err := registry.RegisterModel(modelInfo)
    assert.NoError(t, err)

    // Verify registration
    assert.True(t, registry.IsRegistered("test_model"))

    // Retrieve and verify
    retrieved, err := registry.GetModel("test_model")
    assert.NoError(t, err)
    assert.Equal(t, "test_model", retrieved.Type)
    assert.Equal(t, "Test Model", retrieved.Name)
}
```

#### 2. Validator Registration
```go
func TestUnifiedRegistry_RegisterValidator(t *testing.T) {
    registry := NewUnifiedRegistry("test/models", "test/validations")

    // Create a mock validator
    mockValidator := &mockValidatorImpl{}

    err := registry.RegisterValidator("test_model", mockValidator)
    assert.NoError(t, err)

    // Verify validator can be retrieved
    validator, err := registry.GetValidator("test_model")
    assert.NoError(t, err)
    assert.NotNil(t, validator)
}

// Mock validator implementation
type mockValidatorImpl struct{}

func (m *mockValidatorImpl) ValidatePayload(payload interface{}) interface{} {
    return models.ValidationResult{
        IsValid:   true,
        ModelType: "test_model",
    }
}
```

#### 3. Auto-Discovery Testing
```go
func TestUnifiedRegistry_StartAutoRegistration(t *testing.T) {
    // Create temporary directories for testing
    tempDir := t.TempDir()
    modelsDir := filepath.Join(tempDir, "models")
    validationsDir := filepath.Join(tempDir, "validations")

    os.MkdirAll(modelsDir, 0755)
    os.MkdirAll(validationsDir, 0755)

    // Create test model file
    modelContent := `package models
type TestPayload struct {
    ID string ` + "`json:\"id\"`" + `
}`
    os.WriteFile(filepath.Join(modelsDir, "test.go"), []byte(modelContent), 0644)

    // Create test validator file
    validatorContent := `package validations
type TestValidator struct{}
func (v *TestValidator) ValidatePayload(payload interface{}) interface{} {
    return nil
}`
    os.WriteFile(filepath.Join(validationsDir, "test_validator.go"),
        []byte(validatorContent), 0644)

    // Test auto-registration
    registry := NewUnifiedRegistry(modelsDir, validationsDir)
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

    go registry.StartAutoRegistration(ctx, 500*time.Millisecond)

    time.Sleep(1 * time.Second)

    // Verify models were discovered
    models := registry.ListModels()
    assert.NotEmpty(t, models)
}
```

#### 4. HTTP Handler Testing
```go
func TestUnifiedRegistry_ServeHTTP(t *testing.T) {
    registry := NewUnifiedRegistry("test/models", "test/validations")

    // Register a test model
    modelInfo := &ModelInfo{
        Type: "test",
        Name: "Test Model",
    }
    registry.RegisterModel(modelInfo)

    // Test GET /models
    req := httptest.NewRequest("GET", "/models", nil)
    w := httptest.NewRecorder()

    registry.ServeHTTP(w, req)

    assert.Equal(t, http.StatusOK, w.Code)

    var response []ModelInfo
    json.Unmarshal(w.Body.Bytes(), &response)

    assert.NotEmpty(t, response)
    assert.Equal(t, "test", response[0].Type)
}
```

#### 5. Error Handling
```go
func TestUnifiedRegistry_ErrorHandling(t *testing.T) {
    registry := NewUnifiedRegistry("test/models", "test/validations")

    t.Run("get non-existent model", func(t *testing.T) {
        _, err := registry.GetModel("non_existent")
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "not found")
    })

    t.Run("register model with empty type", func(t *testing.T) {
        modelInfo := &ModelInfo{Type: ""}
        err := registry.RegisterModel(modelInfo)
        assert.Error(t, err)
    })

    t.Run("unregister non-existent model", func(t *testing.T) {
        err := registry.UnregisterModel("non_existent")
        assert.Error(t, err)
    })
}
```

---

## Best Practices

### 1. Use Table-Driven Tests
**Why**: Easier to add new test cases, better organization, reduces code duplication

```go
tests := []struct {
    name     string
    input    interface{}
    expected interface{}
    wantErr  bool
}{
    {"case1", input1, expected1, false},
    {"case2", input2, expected2, true},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        // Test logic
    })
}
```

### 2. Test Error Paths
**Why**: Error handling is often undertested but critical for reliability

```go
// Good: Test both success and failure
func TestFunction(t *testing.T) {
    t.Run("success case", func(t *testing.T) { /* ... */ })
    t.Run("error: nil input", func(t *testing.T) { /* ... */ })
    t.Run("error: invalid format", func(t *testing.T) { /* ... */ })
}
```

### 3. Use Descriptive Test Names
**Why**: Makes test failures easier to debug

```go
// Bad
func TestValidate1(t *testing.T) { /* ... */ }

// Good
func TestGitHubValidator_ValidatePayload_MissingRequiredField(t *testing.T) { /* ... */ }
```

### 4. Test Boundary Conditions
**Why**: Edge cases often reveal bugs

```go
// Test min/max values, empty strings, nil, zero values
tests := []struct {
    name  string
    value int
    valid bool
}{
    {"minimum", 0, true},
    {"just below min", -1, false},
    {"maximum", 100, true},
    {"just above max", 101, false},
}
```

### 5. Use Test Helpers
**Why**: Reduces duplication and improves readability

```go
func createTestPayload(t *testing.T, overrides map[string]interface{}) models.GitHubPayload {
    t.Helper()

    payload := models.GitHubPayload{
        // Default valid payload
        Username:  "octocat",
        RepoName:  "test-repo",
        PRNumber:  123,
        // ...
    }

    // Apply overrides
    for key, value := range overrides {
        // Set field using reflection
    }

    return payload
}

// Usage
func TestSomething(t *testing.T) {
    payload := createTestPayload(t, map[string]interface{}{
        "Username": "invalid@user",
    })
    // Test with modified payload
}
```

### 6. Mock External Dependencies
**Why**: Unit tests should be isolated and fast

```go
type mockValidator struct {
    validateFunc func(interface{}) interface{}
}

func (m *mockValidator) ValidatePayload(payload interface{}) interface{} {
    if m.validateFunc != nil {
        return m.validateFunc(payload)
    }
    return models.ValidationResult{IsValid: true}
}
```

### 7. Test Concurrent Access
**Why**: Catches race conditions and deadlocks

```go
func TestConcurrentAccess(t *testing.T) {
    const numGoroutines = 100
    done := make(chan bool, numGoroutines)

    for i := 0; i < numGoroutines; i++ {
        go func() {
            // Concurrent operation
            done <- true
        }()
    }

    for i := 0; i < numGoroutines; i++ {
        <-done
    }
}
```

### 8. Clean Up Resources
**Why**: Prevents test pollution and resource leaks

```go
func TestWithTempDir(t *testing.T) {
    // Use t.TempDir() - automatically cleaned up
    tempDir := t.TempDir()

    // Or manual cleanup
    cleanup := setupTest()
    defer cleanup()
}
```

### 9. Avoid Test Interdependencies
**Why**: Tests should be runnable in any order

```go
// Bad: Depends on global state from another test
var globalSession *Session

func TestCreateSession(t *testing.T) {
    globalSession = CreateSession()
}

func TestUseSession(t *testing.T) {
    // Assumes TestCreateSession ran first!
    result := globalSession.DoSomething()
}

// Good: Each test is independent
func TestCreateSession(t *testing.T) {
    session := CreateSession()
    assert.NotNil(t, session)
}

func TestUseSession(t *testing.T) {
    session := CreateSession()
    result := session.DoSomething()
    assert.NotNil(t, result)
}
```

### 10. Use Assertions Wisely
**Why**: Better error messages and test readability

```go
// Use testify/assert for better error messages
assert.Equal(t, expected, actual, "optional message")
assert.NotNil(t, obj)
assert.Contains(t, haystack, needle)
assert.Error(t, err)
assert.NoError(t, err)

// For multiple assertions, use require to fail fast
require.NoError(t, err) // Stops test if error
assert.Equal(t, expected, result) // Won't run if above fails
```

---

## Running Tests

### Quick Reference - Common Commands

```bash
# Navigate to src directory
cd src/

# Run all tests
go test ./...

# Run all tests with coverage
go test ./... -cover

# Run all tests with verbose output
go test ./... -v

# Run all tests with race detector
go test ./... -race

# Generate coverage profile and view in browser
go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out
```

### Package-Specific Tests

#### Run Tests for Models Package
```bash
cd src/
go test ./models
go test ./models -cover          # With coverage
go test ./models -v              # Verbose output
go test ./models -cover -v       # Both coverage and verbose
```

#### Run Tests for Validations Package
```bash
cd src/
go test ./validations
go test ./validations -cover
go test ./validations -v
go test ./validations -cover -v
```

#### Run Tests for Registry Package
```bash
cd src/
go test ./registry
go test ./registry -cover
go test ./registry -v
go test ./registry -cover -v
```

#### Run Tests for Main Package
```bash
cd src/
go test .                    # Main package (current directory)
go test . -cover
go test . -v
go test . -cover -v
```

#### Run Tests for Config Package
```bash
cd src/
go test ./config
go test ./config -cover
```

### Coverage Reports

#### Generate Coverage Profile
```bash
cd src/
go test ./... -coverprofile=coverage.out
```

#### View Coverage in Browser (HTML Report)
```bash
cd src/
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

This will open your browser showing:
- **Green**: Covered code
- **Red**: Uncovered code
- **Gray**: Non-executable code

#### View Coverage in Terminal (Function-Level)
```bash
cd src/
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out
```

Output shows coverage per function:
```
goplayground-data-validator/models/github.go:15:    NewGitHubPayload        100.0%
goplayground-data-validator/models/github.go:25:    Validate                85.7%
goplayground-data-validator/validations/github.go:30: ValidatePayload       92.3%
```

#### Calculate Overall Weighted Coverage
```bash
cd src/

# Get coverage for each package
go test ./models -cover
go test ./validations -cover
go test ./registry -cover
go test . -cover
go test ./config -cover
```

**Example Output:**
```
ok      goplayground-data-validator/models         0.123s  coverage: 81.5% of statements
ok      goplayground-data-validator/validations    0.456s  coverage: 84.6% of statements
ok      goplayground-data-validator/registry       0.234s  coverage: 95.5% of statements
ok      goplayground-data-validator                0.345s  coverage: 79.6% of statements
ok      goplayground-data-validator/config         0.012s  coverage: 100.0% of statements
```

To calculate weighted overall coverage:
```
Overall = (1571×0.815 + 3133×0.846 + 914×0.955 + 1151×0.796 + 27×1.0) / 6796
        = 84.6%
```

### Run Specific Tests

#### Run a Single Test Function
```bash
cd src/
go test ./validations -run TestGitHubValidator_ValidatePayload_Valid
go test ./models -run TestGitHubPayload_Validation
go test . -run TestHandleBatchStart
```

#### Run Tests Matching a Pattern
```bash
cd src/
go test ./validations -run TestGitHub        # All tests starting with TestGitHub
go test ./models -run Validation             # All tests containing "Validation"
go test . -run Batch                         # All batch-related tests
```

#### Run Tests with Regex Pattern
```bash
cd src/
go test ./validations -run "TestGitHub.*Valid"  # Tests matching pattern
go test ./models -run "Test.*Enum"               # All enum tests
```

### Advanced Test Options

#### Run Tests with Race Detector
```bash
cd src/
go test ./... -race
```

Detects data races in concurrent code. Important for:
- Batch session management
- Concurrent registry access
- Thread-safe operations

#### Run Tests in Short Mode
```bash
cd src/
go test ./... -short
```

Skips long-running tests marked with:
```go
func TestLongRunning(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping long-running test in short mode")
    }
    // Test logic
}
```

#### Run Tests with Timeout
```bash
cd src/
go test ./... -timeout 30s
```

Fails tests that run longer than specified timeout.

#### Run Tests with Count (No Cache)
```bash
cd src/
go test ./... -count=1
```

Forces tests to run without using cached results.

#### Run Tests in Parallel
```bash
cd src/
go test ./... -parallel 4
```

Runs tests in parallel (default is GOMAXPROCS).

### Continuous Integration

#### Full CI Test Suite
```bash
cd src/

# Run all tests with coverage and race detection
go test ./... -cover -race -v

# Generate coverage report
go test ./... -coverprofile=coverage.out

# Check if coverage meets threshold (80%)
go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//'
# If result >= 80.0, tests pass CI requirements
```

#### Jenkins/CI Pipeline Command
```bash
cd src/
go test ./... -coverprofile=coverage.out -race -v && \
go tool cover -func=coverage.out
```

### Benchmarking (Optional)

#### Run Benchmarks
```bash
cd src/
go test ./... -bench=.
go test ./validations -bench=BenchmarkValidation
```

#### Run Benchmarks with Memory Stats
```bash
cd src/
go test ./... -bench=. -benchmem
```

### Test Output Formats

#### JSON Output
```bash
cd src/
go test ./... -json
```

Useful for parsing test results in CI/CD pipelines.

#### Verbose Output
```bash
cd src/
go test ./... -v
```

Shows individual test results:
```
=== RUN   TestGitHubValidator_ValidatePayload_Valid
--- PASS: TestGitHubValidator_ValidatePayload_Valid (0.00s)
=== RUN   TestGitHubValidator_ValidatePayload_Invalid
--- PASS: TestGitHubValidator_ValidatePayload_Invalid (0.00s)
```

### Troubleshooting Failed Tests

#### Run Only Failed Tests
After a test failure, re-run only the failed test:
```bash
cd src/
go test ./validations -run TestGitHubValidator_ValidatePayload_Invalid -v
```

#### Get More Debug Information
```bash
cd src/
go test ./validations -v -count=1 -run TestFailingTest
```

#### Check Test Cache
```bash
# Clear test cache
go clean -testcache

# Run tests without cache
cd src/
go test ./... -count=1
```

### Model-Specific Test Examples

#### Testing Individual Models
```bash
# Test all incident-related functionality
go test -v ./... -run ".*[Ii]ncident.*"

# Test all GitHub-related functionality
go test -v ./... -run ".*[Gg]it[Hh]ub.*"

# Test all API-related functionality
go test -v ./... -run ".*API.*"
```

#### Testing Validation Scenarios
```bash
# Test all validation scenarios
go test -v ./... -run ".*[Vv]alidation.*"

# Test all error scenarios
go test -v ./... -run ".*[Ee]rror.*"

# Test all business logic scenarios
go test -v ./... -run ".*[Bb]usiness.*"
```

---

## Common Patterns

### Pattern 1: Testing HTTP Handlers

```go
import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestHTTPHandler(t *testing.T) {
    // 1. Create request
    payload := map[string]interface{}{"key": "value"}
    body, _ := json.Marshal(payload)
    req := httptest.NewRequest("POST", "/endpoint", bytes.NewBuffer(body))

    // 2. Create response recorder
    w := httptest.NewRecorder()

    // 3. Call handler
    yourHandler(w, req)

    // 4. Assert response
    assert.Equal(t, http.StatusOK, w.Code)

    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)
    assert.Equal(t, "expected", response["key"])
}
```

### Pattern 2: Testing with Time

```go
func TestTimeDependent(t *testing.T) {
    // Use fixed time for reproducibility
    fixedTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

    payload := models.Payload{
        CreatedAt: fixedTime,
    }

    // Or test relative times
    now := time.Now()
    payload.CreatedAt = now.Add(-24 * time.Hour)

    result := validator.ValidatePayload(payload)
    assert.True(t, result.IsValid)
}
```

### Pattern 3: Testing Reflection-Based Code

```go
func TestReflectionLogic(t *testing.T) {
    type TestStruct struct {
        Field1 string
        Field2 int
    }

    obj := &TestStruct{}
    val := reflect.ValueOf(obj).Elem()

    // Test field access
    field := val.FieldByName("Field1")
    assert.True(t, field.IsValid())
    assert.Equal(t, reflect.String, field.Kind())

    // Test field modification
    field.SetString("test value")
    assert.Equal(t, "test value", obj.Field1)
}
```

### Pattern 4: Testing with Context

```go
func TestWithContext(t *testing.T) {
    // Test with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

    result := functionWithContext(ctx)
    assert.NotNil(t, result)

    // Test context cancellation
    ctx2, cancel2 := context.WithCancel(context.Background())
    cancel2() // Cancel immediately

    err := functionThatChecksCancellation(ctx2)
    assert.ErrorIs(t, err, context.Canceled)
}
```

### Pattern 5: Testing File Operations

```go
func TestFileOperations(t *testing.T) {
    // Use t.TempDir() for automatic cleanup
    tempDir := t.TempDir()

    testFile := filepath.Join(tempDir, "test.txt")
    content := []byte("test content")

    err := os.WriteFile(testFile, content, 0644)
    assert.NoError(t, err)

    readContent, err := os.ReadFile(testFile)
    assert.NoError(t, err)
    assert.Equal(t, content, readContent)
}
```

### Pattern 6: Performance Testing

```go
func BenchmarkMyModelValidator(b *testing.B) {
    validator := NewMyModelValidator()
    payload := getValidMyModelPayload()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        validator.ValidatePayload(payload)
    }
}
```

### Pattern 7: Custom Validators for Complex Fields

```go
func TestValidateComplexField(t *testing.T) {
    v := validator.New()
    v.RegisterValidation("custom_rule", func(fl validator.FieldLevel) bool {
        // Custom validation logic
        return true
    })

    // Test with custom rule
}
```

---

## Coverage Analysis

### Understanding Coverage Metrics

#### Statement Coverage
- Percentage of code statements executed during tests
- Calculated by Go's coverage tool
- Formula: `(executed_statements / total_statements) × 100`

#### Package Coverage
Individual coverage per package:
```
models:      81.5%  <- Model struct validation
validations: 84.6%  <- Validator logic
registry:    95.5%  <- Registry operations
main:        79.6%  <- HTTP handlers
config:      100.0% <- Configuration constants
```

#### Weighted Coverage
Overall project coverage weighted by code volume:
```
Overall = Σ(package_lines × package_coverage) / total_lines
        = (1571×0.815 + 3133×0.846 + 914×0.955 + 1151×0.796 + 27×1.0) / 6796
        = 84.6%
```

### Identifying Coverage Gaps

#### 1. Generate Coverage Profile
```bash
go test ./... -coverprofile=coverage.out
```

#### 2. View Uncovered Code
```bash
go tool cover -html=coverage.out
```

This opens a browser showing:
- **Green**: Covered code
- **Red**: Uncovered code
- **Gray**: Not executable (comments, declarations)

#### 3. Function-Level Coverage
```bash
go tool cover -func=coverage.out
```

Output shows coverage per function:
```
models/github.go:15:    NewGitHubPayload        100.0%
models/github.go:25:    Validate                85.7%
validations/github.go:30: ValidatePayload       92.3%
```

#### 4. Find Low-Coverage Functions
```bash
go tool cover -func=coverage.out | grep -v "100.0%" | sort -k3 -n
```

### Increasing Coverage

#### Priority Areas (to reach 80%+)

1. **Error Paths** (often untested)
   - Invalid input handling
   - Network errors
   - File I/O errors
   - Timeout handling

2. **Edge Cases**
   - Nil values
   - Empty collections
   - Boundary values (min/max)
   - Concurrent access

3. **Business Logic Branches**
   - All enum values
   - Conditional validation rules
   - Warning generation
   - Custom validators

4. **HTTP Handlers**
   - All HTTP methods
   - Request header variations
   - Invalid payloads
   - Response formatting

#### Example: Increasing Validator Coverage

**Before (71.5% coverage):**
```go
// Only basic happy path tested
func TestValidator_Valid(t *testing.T) {
    result := validator.ValidatePayload(validPayload)
    assert.True(t, result.IsValid)
}
```

**After (84.6% coverage):**
```go
// Added tests for:
// - Invalid payloads (missing fields, wrong types)
// - Business logic branches (WIP detection, large changesets)
// - Custom validators (username, hexcolor, etc.)
// - Error formatting
// - Warning generation
// - Edge cases (nil, empty, zero values)
```

---

## Quick Reference Checklist

### For New Models:
- [ ] Valid payload passes validation
- [ ] Missing required fields are caught
- [ ] Invalid field values are rejected
- [ ] Enum fields validate correctly
- [ ] Custom validation tags work
- [ ] Min/max constraints are enforced
- [ ] Format validations (email, URL, etc.) work
- [ ] Edge cases (empty, nil, zero values)
- [ ] JSON marshaling/unmarshaling
- [ ] Array validation scenarios
- [ ] Batch validation with threshold parameter

### For New Validators:
- [ ] Constructor creates valid instance (`NewXValidator`)
- [ ] Valid payloads return `IsValid: true`
- [ ] Invalid payloads return proper errors
- [ ] Error messages are descriptive
- [ ] Business logic is tested
- [ ] Warnings are generated appropriately
- [ ] Custom validation functions work
- [ ] Edge cases handled
- [ ] Performance tests (if applicable)
- [ ] Batch session manager tests (if using threshold validation)

### For HTTP Handlers:
- [ ] Valid requests return 200 OK
- [ ] Invalid JSON returns 400 Bad Request
- [ ] Missing parameters handled
- [ ] Error responses formatted correctly
- [ ] Headers processed correctly
- [ ] Concurrent requests handled safely

### For Core Logic:
- [ ] Type conversions work correctly
- [ ] Overflow detection works
- [ ] Field setters handle all types
- [ ] Error paths tested
- [ ] Concurrent access safe
- [ ] Resource cleanup verified

### Integration Tests (Automatic)
- [ ] Main tests run without modification
- [ ] Registry tests work automatically
- [ ] E2E tests include new model (add test data to `test_data/`)
- [ ] Array validation tests with valid/invalid mix
- [ ] Threshold validation tests with different percentages

---

## Summary

### Key Takeaways:
1. Aim for **≥80% overall coverage** (Jenkins requirement)
2. Use **table-driven tests** for comprehensive coverage
3. **Test error paths** and edge cases, not just happy paths
4. Use **mock objects** to isolate unit tests
5. Test **concurrent access** for shared resources
6. Use `go test ./... -coverprofile=coverage.out` to generate coverage
7. Use `go tool cover -html=coverage.out` to find gaps
8. The testing framework is **model-agnostic** - no changes to main tests needed

### Coverage Achievement:
- Started at: **~25% overall**
- Current: **84.6% overall** (Exceeds requirement)
- Exceeds Jenkins requirement by **4.6 percentage points**

### Test Statistics:
- **154 test functions** added
- **~2,500+ lines** of test code
- **All packages** now meet or exceed targets

### Benefits of This Approach

#### Easy Model Addition
- **2 files**: Just add `my_model_test.go` in models and validations packages
- **No main code changes**: Registry and main tests are model-agnostic
- **No interdependencies**: Tests are isolated and focused

#### Maintainable
- **Clear patterns**: Consistent testing structure across all models
- **Generic framework**: Main code tests work for any new model
- **Self-documenting**: Tests serve as documentation for model behavior

#### Comprehensive Coverage
- **Multiple test levels**: Unit → Integration → E2E
- **Business logic testing**: Custom validators and warnings
- **Performance testing**: Benchmark tests for critical paths

---

**Happy Testing!**

This model-agnostic testing framework makes adding new models and validations effortless while maintaining high test coverage and quality standards.
