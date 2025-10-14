# Adding New Models Guide

## Overview

This guide shows you how to add new validation models to the Go Playground Data Validator. The system automatically discovers and registers new models - just add files and restart!

## Quick Start (3 Steps)

### Step 1: Create Model (`src/models/your_model.go`)

```go
package models

import "time"

type YourModelPayload struct {
    ID        string    `json:"id" validate:"required,min=3"`
    Name      string    `json:"name" validate:"required,min=2,max=100"`
    Email     string    `json:"email" validate:"required,email"`
    Status    string    `json:"status" validate:"required,oneof=active inactive"`
    CreatedAt time.Time `json:"created_at" validate:"required"`
}
```

### Step 2: Create Validator (`src/validations/your_model.go`)

```go
package validations

import (
    "goplayground-data-validator/models"
    "github.com/go-playground/validator/v10"
)

type YourModelValidator struct {
    validator *validator.Validate
}

func NewYourModelValidator() *YourModelValidator {
    return &YourModelValidator{validator: validator.New()}
}

func (v *YourModelValidator) ValidatePayload(payload models.YourModelPayload) models.ValidationResult {
    result := models.ValidationResult{
        IsValid:   true,
        ModelType: "your_model",
        Provider:  "go-playground",
        Errors:    []models.ValidationError{},
        Warnings:  []models.ValidationWarning{},
    }

    // Run validation
    if err := v.validator.Struct(payload); err != nil {
        result.IsValid = false
        // Handle errors (see examples below)
    }

    return result
}
```

### Step 3: Build and Test

```bash
# Build
make build

# Start server
./bin/validator

# Test
curl -X POST http://localhost:8080/validate/your_model \
  -H "Content-Type: application/json" \
  -d '{"id":"123","name":"Test","email":"test@example.com","status":"active","created_at":"2024-01-01T00:00:00Z"}'
```

That's it! Your model is now auto-registered with endpoint `/validate/your_model`

---

## Testing Your Model

### 1. Single Object Validation

**Using model-specific endpoint:**
```bash
curl -X POST http://localhost:8080/validate/your_model \
  -H "Content-Type: application/json" \
  -d '{
    "id": "model-001",
    "name": "Test Model",
    "email": "test@example.com",
    "status": "active",
    "created_at": "2024-01-01T00:00:00Z"
  }'
```

**Using generic endpoint:**
```bash
curl -X POST http://localhost:8080/validate \
  -H "Content-Type: application/json" \
  -d '{
    "model_type": "your_model",
    "payload": {
      "id": "model-001",
      "name": "Test Model",
      "email": "test@example.com",
      "status": "active",
      "created_at": "2024-01-01T00:00:00Z"
    }
  }'
```

**Response:**
```json
{
  "is_valid": true,
  "model_type": "your_model",
  "provider": "go-playground",
  "errors": [],
  "warnings": []
}
```

### 2. Array Validation

**Validate multiple records (without threshold):**
```bash
curl -X POST http://localhost:8080/validate \
  -H "Content-Type: application/json" \
  -d '{
    "model_type": "your_model",
    "data": [
      {"id":"001","name":"First","email":"first@example.com","status":"active","created_at":"2024-01-01T00:00:00Z"},
      {"id":"002","name":"Second","email":"second@example.com","status":"active","created_at":"2024-01-01T00:00:00Z"}
    ]
  }'
```

**Response:**
```json
{
  "batch_id": "auto_abc123",
  "status": "success",
  "total_records": 2,
  "valid_records": 2,
  "invalid_records": 0,
  "processing_time_ms": 5,
  "summary": {
    "success_rate": 100,
    "validation_errors": 0,
    "validation_warnings": 0,
    "total_records_processed": 2
  },
  "results": []
}
```

**Notes:**
- `results` array only includes invalid/warning records. Valid records are excluded.
- Without `threshold` parameter, status is always `"success"` for multiple records (returns validation details for each).
- To enforce minimum success rate, use `threshold` parameter (see next section).

### 3. Array Validation with Threshold

Use `threshold` parameter to enforce minimum success rate for batch validation. This is useful for data quality checks, import validation, or ensuring batch operations meet quality standards.

#### Example 1: Success Case (meets threshold)

**Request with 80% threshold, 100% valid records:**
```bash
curl -X POST http://localhost:8080/validate \
  -H "Content-Type: application/json" \
  -d '{
    "model_type": "your_model",
    "threshold": 80.0,
    "data": [
      {"id":"001","name":"Valid Record 1","email":"valid1@example.com","status":"active","created_at":"2024-01-01T00:00:00Z"},
      {"id":"002","name":"Valid Record 2","email":"valid2@example.com","status":"active","created_at":"2024-01-01T00:00:00Z"},
      {"id":"003","name":"Valid Record 3","email":"valid3@example.com","status":"active","created_at":"2024-01-01T00:00:00Z"}
    ]
  }'
```

**Response:**
```json
{
  "batch_id": "auto_abc123",
  "status": "success",
  "total_records": 3,
  "valid_records": 3,
  "invalid_records": 0,
  "threshold": 80.0,
  "summary": {
    "success_rate": 100.0,
    "validation_errors": 0,
    "total_records_processed": 3
  },
  "results": []
}
```

#### Example 2: Failure Case (below threshold)

**Request with 80% threshold, 50% valid records (below threshold):**
```bash
curl -X POST http://localhost:8080/validate \
  -H "Content-Type: application/json" \
  -d '{
    "model_type": "your_model",
    "threshold": 80.0,
    "data": [
      {"id":"001","name":"Valid","email":"valid@example.com","status":"active","created_at":"2024-01-01T00:00:00Z"},
      {"id":"002","name":"V","email":"invalid","status":"bad"}
    ]
  }'
```

**Response:**
```json
{
  "batch_id": "auto_xyz789",
  "status": "failed",
  "total_records": 2,
  "valid_records": 1,
  "invalid_records": 1,
  "threshold": 80.0,
  "summary": {
    "success_rate": 50.0,
    "validation_errors": 4,
    "total_records_processed": 2
  },
  "results": [
    {
      "row_index": 1,
      "record_identifier": "002",
      "is_valid": false,
      "errors": [
        {"field": "name", "message": "Must be at least 2 characters", "code": "VALUE_TOO_SHORT"},
        {"field": "email", "message": "Must be a valid email", "code": "INVALID_EMAIL"}
      ]
    }
  ]
}
```

**Threshold Logic:**
- `success_rate = (valid_records / total_records) * 100`
- Status = `"success"` if `success_rate >= threshold`, else `"failed"`
- HTTP Status: 200 OK for success, 422 Unprocessable Entity for failed

**Use Cases:**
- **Data Import**: Require 95% valid records before importing into production database
- **Batch Processing**: Ensure 90% of batch items are valid before proceeding
- **Quality Gates**: Enforce data quality thresholds in CI/CD pipelines

---

## Test Data Files

Create test files for E2E testing:

**Valid:** `test_data/single/valid/your_model.json`
```json
{
  "id": "model-001",
  "name": "Test Model",
  "email": "test@example.com",
  "status": "active",
  "created_at": "2024-01-01T00:00:00Z"
}
```

**Invalid:** `test_data/single/invalid/your_model.json`
```json
{
  "id": "x",
  "name": "",
  "email": "not-an-email",
  "status": "invalid"
}
```

**Array:** `test_data/arrays/valid/your_model.json`
```json
[
  {"id":"001","name":"First","email":"first@example.com","status":"active","created_at":"2024-01-01T00:00:00Z"},
  {"id":"002","name":"Second","email":"second@example.com","status":"active","created_at":"2024-01-01T00:00:00Z"}
]
```

---

## Validation Tags Reference

Common validation tags you can use:

```go
// Required
Field string `validate:"required"`

// Length
Field string `validate:"min=3,max=100"`

// Email
Email string `validate:"email"`

// URL
URL string `validate:"url"`

// Numeric ranges
Age int `validate:"min=18,max=100"`

// Enum values
Status string `validate:"oneof=active inactive suspended"`

// Arrays
Tags []string `validate:"dive,min=1,max=20"`

// Optional fields
Bio string `validate:"omitempty,max=500"`
```

---

## Adding Custom Validation

### Custom Validation Function

```go
func NewYourModelValidator() *YourModelValidator {
    v := validator.New()

    // Register custom validator
    v.RegisterValidation("username_format", func(fl validator.FieldLevel) bool {
        username := fl.Field().String()
        matched, _ := regexp.MatchString(`^[a-zA-Z][a-zA-Z0-9_]*$`, username)
        return matched
    })

    return &YourModelValidator{validator: v}
}

// Use in model
type YourModelPayload struct {
    Username string `json:"username" validate:"required,username_format"`
}
```

### Business Logic Warnings

```go
func (v *YourModelValidator) ValidatePayload(payload models.YourModelPayload) models.ValidationResult {
    // ... standard validation ...

    // Add warnings (even if validation passed)
    var warnings []models.ValidationWarning

    if payload.Status == "inactive" && payload.LastLogin.IsZero() {
        warnings = append(warnings, models.ValidationWarning{
            Field:      "status",
            Message:    "Inactive account with no login history",
            Code:       "SUSPICIOUS_ACCOUNT",
            Suggestion: "Review account for potential cleanup",
        })
    }

    result.Warnings = warnings
    return result
}
```

---

## Unit Testing

**Create test:** `src/models/your_model_test.go`
```go
package models

import (
    "testing"
    "github.com/go-playground/validator/v10"
)

func TestYourModel_Valid(t *testing.T) {
    v := validator.New()

    payload := YourModelPayload{
        ID:     "test-123",
        Name:   "Test",
        Email:  "test@example.com",
        Status: "active",
    }

    err := v.Struct(payload)
    if err != nil {
        t.Errorf("Expected valid, got error: %v", err)
    }
}
```

**Run tests:**
```bash
# Test specific model
go test ./models -run TestYourModel -v

# Test all with coverage
make test-coverage
```

---

## E2E Testing

**Run E2E test suite:**
```bash
# Build and test
make test-e2e

# Docker test
make docker-test-e2e
```

The E2E suite automatically:
- ✅ Discovers your model
- ✅ Tests with `test_data/` files
- ✅ Validates endpoints
- ✅ Reports results

---

## Best Practices

### ✅ Model Design
```go
// Good: Clear names, proper types
type UserPayload struct {
    ID        string    `json:"id" validate:"required,min=3"`
    Email     string    `json:"email" validate:"required,email"`
    CreatedAt time.Time `json:"created_at" validate:"required"`
    UpdatedAt time.Time `json:"updated_at" validate:"required"`
}

// Avoid: Unclear names, missing validation
type User struct {
    I  string `json:"i"`
    E  string `json:"e"`
    CA string `json:"ca"`
}
```

### ✅ Validation
```go
// Good: Layered validation
1. Struct validation (go-playground tags)
2. Custom field validation
3. Business logic warnings

// Avoid: All logic in one function
```

### ✅ Error Messages
```go
// Good: Helpful, specific
"Email must be in valid format (e.g., user@example.com)"
"Age must be between 18 and 100"

// Avoid: Generic, vague
"Invalid input"
"Validation failed"
```

---

## Troubleshooting

### Model Not Discovered
**Check:**
- ✅ Model file in `src/models/`
- ✅ Struct name ends with `Payload`
- ✅ Validator file in `src/validations/`
- ✅ Constructor named `New{Model}Validator`
- ✅ Server restarted

### Validation Not Working
**Check:**
- ✅ `ValidatePayload` method signature correct
- ✅ Custom validators registered in constructor
- ✅ Import paths correct
- ✅ Build successful (`make build`)

### Test Data Not Found
**Check:**
- ✅ Files in `test_data/single/valid/` and `test_data/single/invalid/`
- ✅ Filename matches model name (case-sensitive)
- ✅ JSON syntax valid

---

## Quick Command Reference

```bash
# Build
make build

# Run server
./bin/validator
# Or with custom port
PORT=9090 ./bin/validator

# Test endpoints
curl http://localhost:8080/models        # List all models
curl http://localhost:8080/health        # Health check

# Run tests
make test              # Unit tests
make test-coverage     # With coverage
make test-e2e          # E2E tests
make docker-test-e2e   # Docker E2E

# Clean
make clean             # Clean artifacts
```

---

## Summary

Adding a new model requires:

1. **Create model** (`src/models/your_model.go`)
2. **Create validator** (`src/validations/your_model.go`)
3. **Build and test** (`make build && ./bin/validator`)

The system automatically:
- ✅ Discovers your model
- ✅ Registers HTTP endpoints
- ✅ Enables validation (single, array, batch)
- ✅ Integrates with E2E tests

**Happy validating! 🚀**
