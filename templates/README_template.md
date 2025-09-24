# {{.ServiceName}} Model Integration Guide

> **Simple guide to adding {{.ServiceName}} validation to the modular validation server**

## Quick Start

Adding a new {{.ServiceName}} model to the validation system is easy:

1. **Create the model** → Define your data structure
2. **Create the validator** → Implement validation logic
3. **Register the model** → Add to the system
4. **Test it** → Verify it works

## 📋 What You'll Get

Once integrated, your {{.ServiceName}} model will have:

- ✅ **Automatic API endpoint**: `POST /validate/{{.EndpointPath}}`
- ✅ **Swagger documentation**: Auto-discovered and documented
- ✅ **Comprehensive validation**: Structural + business logic rules
- ✅ **Performance metrics**: Built-in timing and analytics
- ✅ **Error reporting**: Detailed validation error messages

## 🏗️ File Structure

Create these three files in your project:

```
src/
├── models/{{.VarName}}.go          # Data structures
├── validations/{{.VarName}}.go     # Validation logic
└── main.go                         # Registration (add to existing)
```

## 📝 Step 1: Define Your Model

**File**: `src/models/{{.VarName}}.go`

```go
package models

import "time"

// {{.MainStructName}} represents {{.Description}}
type {{.MainStructName}} struct {
    // Required fields
    ID        string    `json:"id" validate:"required,min=1,max=255"`
    Type      string    `json:"type" validate:"required,oneof={{.ValidTypes}}"`
    Timestamp time.Time `json:"timestamp" validate:"required"`

    // {{.ServiceName}}-specific fields (customize these)
    {{range .CustomFields}}
    {{.Name}} {{.Type}} `json:"{{.JSONTag}}" validate:"{{.ValidationRules}}"`{{end}}

    // Optional metadata
    Status   string                 `json:"status,omitempty" validate:"omitempty,oneof=active inactive"`
    Metadata map[string]interface{} `json:"metadata,omitempty"`
}
```

## ⚡ Step 2: Create Validator

**File**: `src/validations/{{.VarName}}.go`

```go
package validations

import (
    "time"
    "github.com/go-playground/validator/v10"
    "github-data-validator/models"
)

// {{.ServiceName}}Validator validates {{.ServiceName}} payloads
type {{.ServiceName}}Validator struct {
    validator *validator.Validate
}

// New{{.ServiceName}}Validator creates a new validator
func New{{.ServiceName}}Validator() *{{.ServiceName}}Validator {
    return &{{.ServiceName}}Validator{
        validator: validator.New(),
    }
}

// ValidatePayload validates a {{.ServiceName}} payload
func (v *{{.ServiceName}}Validator) ValidatePayload(payload interface{}) models.ValidationResult {
    // Convert to correct type
    {{.VarName}}Payload, ok := payload.(models.{{.MainStructName}})
    if !ok {
        return models.ValidationResult{
            IsValid: false,
            ModelType: "{{.MainStructName}}",
            Provider: "{{.VarName}}_validator",
            Timestamp: time.Now(),
            Errors: []models.ValidationError{{
                Field: "payload",
                Message: "Invalid payload type for {{.ServiceName}}",
                Code: "TYPE_MISMATCH",
                Severity: "error",
            }},
        }
    }

    result := models.ValidationResult{
        IsValid: true,
        ModelType: "{{.MainStructName}}",
        Provider: "{{.VarName}}_validator",
        Timestamp: time.Now(),
        Errors: []models.ValidationError{},
        Warnings: []models.ValidationWarning{},
    }

    // Structural validation
    if err := v.validator.Struct({{.VarName}}Payload); err != nil {
        result.IsValid = false
        if validationErrors, ok := err.(validator.ValidationErrors); ok {
            for _, fieldError := range validationErrors {
                result.Errors = append(result.Errors, models.ValidationError{
                    Field: fieldError.Field(),
                    Message: fieldError.Error(),
                    Code: fieldError.Tag(),
                    Severity: "error",
                })
            }
        }
    }

    // Add business logic validation here if needed
    // warnings := validateBusinessLogic({{.VarName}}Payload)
    // result.Warnings = append(result.Warnings, warnings...)

    return result
}
```

## 🔧 Step 3: Register the Model

**Add to**: `src/main.go` (in the startModularServer function)

```go
// Add this import at the top
import "github-data-validator/validations"

// Add this line in startModularServer() before server.ListenAndServe()
mux.HandleFunc("POST /validate/{{.EndpointPath}}", handle{{.ServiceName}}Validation)

// Add this handler function
func handle{{.ServiceName}}Validation(w http.ResponseWriter, r *http.Request) {
    var payload models.{{.MainStructName}}

    if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
        sendJSONError(w, "Invalid JSON payload", http.StatusBadRequest)
        return
    }

    // Create validator and validate
    validator := validations.New{{.ServiceName}}Validator()
    result := validator.ValidatePayload(payload)

    // Return result
    w.Header().Set("Content-Type", "application/json")
    if !result.IsValid {
        w.WriteHeader(http.StatusUnprocessableEntity)
    }
    json.NewEncoder(w).Encode(result)
}
```

## 🧪 Step 4: Test Your Integration

### Start the server:
```bash
cd src
go run main.go
```

### Test your endpoint:
```bash
curl -X POST http://localhost:8080/validate/{{.EndpointPath}} \
  -H "Content-Type: application/json" \
  -d '{
    "id": "test-123",
    "type": "{{.ExampleType}}",
    "timestamp": "2024-01-01T00:00:00Z"{{range .ExampleFields}},
    "{{.JSONTag}}": {{.ExampleValue}}{{end}}
  }'
```

### Expected response:
```json
{
  "is_valid": true,
  "model_type": "{{.MainStructName}}",
  "provider": "{{.VarName}}_validator",
  "timestamp": "2024-01-01T00:00:00Z",
  "errors": [],
  "warnings": []
}
```

## 📚 Check Documentation

Your new model will automatically appear in:

- **Models list**: `GET http://localhost:8080/models`
- **Swagger docs**: `GET http://localhost:8080/swagger/models`
- **API documentation**: Auto-generated OpenAPI spec

## 🎯 Advanced Features

### Custom Validation Rules

Add custom validators to your validator:

```go
func New{{.ServiceName}}Validator() *{{.ServiceName}}Validator {
    v := validator.New()

    // Add custom validator
    v.RegisterValidation("{{.VarName}}_id", validate{{.ServiceName}}ID)

    return &{{.ServiceName}}Validator{validator: v}
}

func validate{{.ServiceName}}ID(fl validator.FieldLevel) bool {
    id := fl.Field().String()
    // Add your custom validation logic here
    return len(id) > 0 && strings.HasPrefix(id, "{{.VarName}}_")
}
```

### Business Logic Warnings

Add business rule checks:

```go
func validateBusinessLogic(payload models.{{.MainStructName}}) []models.ValidationWarning {
    var warnings []models.ValidationWarning

    // Example: Check for suspicious patterns
    if strings.Contains(payload.ID, "test") {
        warnings = append(warnings, models.ValidationWarning{
            Field: "ID",
            Message: "ID contains 'test' - might be test data",
            Code: "TEST_DATA_DETECTED",
            Category: "data-quality",
        })
    }

    return warnings
}
```

### Using with Generic Endpoint

Your model also works with the generic validation endpoint:

```bash
curl -X POST http://localhost:8080/validate \
  -H "Content-Type: application/json" \
  -d '{
    "model_type": "{{.VarName}}",
    "payload": {
      "id": "test-123",
      "type": "{{.ExampleType}}",
      "timestamp": "2024-01-01T00:00:00Z"
    }
  }'
```

## 🔍 Troubleshooting

### Common Issues:

**"Model not found"** → Make sure you added the handler in main.go
**"Invalid JSON"** → Check your JSON syntax
**"Validation failed"** → Check required fields and validation rules
**"Type mismatch"** → Ensure payload matches your struct definition

### Debug Tips:

1. **Check server logs** for registration messages
2. **Verify model appears** in `/models` endpoint
3. **Test with minimal payload** first
4. **Check validation tags** on struct fields

## 🚀 Next Steps

- **Add to CI/CD**: Include your model in automated tests
- **Monitor performance**: Check validation timing metrics
- **Extend functionality**: Add more business logic rules
- **Update documentation**: Add examples and use cases

---

**Need help?** Check the main project README or existing model implementations for more examples.