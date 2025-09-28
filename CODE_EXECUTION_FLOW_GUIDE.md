# Go Playground Validator - Code Execution Flow Guide

## Table of Contents
1. [Overview](#overview)
2. [Architecture Overview](#architecture-overview)
3. [Directory Structure](#directory-structure)
4. [Application Flow](#application-flow)
5. [Request Processing Pipeline](#request-processing-pipeline)
6. [Model Registration System](#model-registration-system)
7. [Validation Framework](#validation-framework)
8. [API Endpoints](#api-endpoints)
9. [Error Handling](#error-handling)
10. [Performance Monitoring](#performance-monitoring)
11. [Testing Architecture](#testing-architecture)
12. [Contributing Guidelines](#contributing-guidelines)

## Overview

This Go validation server is a **modular, auto-discovering validation platform** that dynamically registers models and their validators at startup. It uses reflection, Go AST parsing, and conventional naming patterns to create a unified validation system without requiring manual configuration.

### Key Features
- 🚀 **Automatic Model Discovery**: Scans filesystem for Go structs and validators
- 🔄 **Dynamic HTTP Endpoint Registration**: Creates REST endpoints automatically
- 🎯 **Universal Validation Interface**: Works with any validator that follows Go conventions
- 📊 **Performance Monitoring**: Built-in metrics and performance tracking
- 🔧 **Clean Architecture**: Separation of concerns with modular design

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                    HTTP REQUEST LAYER                          │
├─────────────────────────────────────────────────────────────────┤
│  GET /health  │  POST /validate  │  GET /models  │  POST /validate/{type} │
└─────────────────────────────────────────────────────────────────┘
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                   ROUTING & MIDDLEWARE                         │
├─────────────────────────────────────────────────────────────────┤
│  HTTP Multiplexer (net/http)  │  Request Validation  │  CORS    │
└─────────────────────────────────────────────────────────────────┘
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                   UNIFIED REGISTRY SYSTEM                      │
├─────────────────────────────────────────────────────────────────┤
│   Model Discovery  │  Validator Registration  │  HTTP Endpoints │
│   ┌─────────────┐  │  ┌────────────────────┐  │  ┌─────────────┐ │
│   │ AST Parser  │  │  │ Reflection Utils   │  │  │ Route Gen   │ │
│   │ File Scanner│  │  │ Constructor Search │  │  │ Handler Gen │ │
│   └─────────────┘  │  └────────────────────┘  │  └─────────────┘ │
└─────────────────────────────────────────────────────────────────┘
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                  VALIDATION PROCESSING                         │
├─────────────────────────────────────────────────────────────────┤
│   Model Resolution  │  Payload Conversion  │  Business Logic    │
│   ┌─────────────┐   │  ┌────────────────┐   │  ┌──────────────┐  │
│   │ Type Lookup │   │  │ JSON → Struct  │   │  │ Custom Rules │  │
│   │ Validator   │   │  │ Map → Struct   │   │  │ Field Logic  │  │
│   │ Instance    │   │  │ Reflection     │   │  │ Context      │  │
│   └─────────────┘   │  └────────────────┘   │  └──────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                   CORE VALIDATION ENGINE                       │
├─────────────────────────────────────────────────────────────────┤
│  go-playground/validator  │  BaseValidator  │  Performance Metrics │
│  ┌─────────────────────┐  │  ┌───────────┐  │  ┌─────────────────┐ │
│  │ Struct Validation   │  │  │ Common    │  │  │ Duration        │ │
│  │ Custom Tags         │  │  │ Utilities │  │  │ Memory Usage    │ │
│  │ Error Formatting    │  │  │ Metadata  │  │  │ Field Count     │ │
│  └─────────────────────┘  │  └───────────┘  │  └─────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                     RESPONSE GENERATION                        │
├─────────────────────────────────────────────────────────────────┤
│  Validation Result  │  Error Messages  │  Performance Data  │  JSON │
│  ┌───────────────┐  │  ┌─────────────┐  │  ┌──────────────┐  │  ┌───┐ │
│  │ IsValid       │  │  │ Field Errors│  │  │ Duration     │  │  │ ▼ │ │
│  │ ModelType     │  │  │ Warnings    │  │  │ Rule Count   │  │  │ ▼ │ │
│  │ Provider      │  │  │ Suggestions │  │  │ Memory       │  │  │ ▼ │ │
│  └───────────────┘  │  └─────────────┘  │  └──────────────┘  │  └───┘ │
└─────────────────────────────────────────────────────────────────┘
```

## Directory Structure

```
src/
├── main.go                     # Application entry point
├── config/
│   └── constants.go           # Configuration constants and thresholds
├── models/                    # Data model definitions
│   ├── api.go                # API request/response models
│   ├── github.go             # GitHub webhook models
│   ├── incident.go           # Incident report models
│   ├── database.go           # Database operation models
│   ├── deployment.go         # Deployment models
│   ├── generic.go            # Generic payload models
│   └── *_test.go             # Model unit tests
├── validations/               # Validation logic
│   ├── base_validator.go     # Common validation framework
│   ├── api.go                # API-specific validation rules
│   ├── github.go             # GitHub validation logic
│   ├── incident.go           # Incident validation rules
│   ├── database.go           # Database validation logic
│   ├── deployment.go         # Deployment validation rules
│   ├── generic.go            # Generic validation logic
│   └── *_test.go             # Validation unit tests
└── registry/                  # Model registration system
    ├── model_registry.go     # Core registry types and interfaces
    ├── unified_registry.go   # Unified auto-discovery system
    ├── dynamic_registry.go   # Dynamic registration utilities
    └── *_test.go             # Registry unit tests
```

## Application Flow

### 1. Server Startup (`main.go:24-36`)

```go
func main() {
    serverMode := os.Getenv("SERVER_MODE")

    // Always use modular server (legacy mode deprecated)
    log.Println("Starting Modular Validation Server...")
    startModularServer()
}
```

### 2. HTTP Server Initialization (`main.go:39-98`)

```go
func startModularServer() {
    port := os.Getenv("PORT") // Default: 8080
    mux := http.NewServeMux()

    // Register core endpoints
    mux.HandleFunc("GET /health", handleHealth)
    mux.HandleFunc("POST /validate", handleGenericValidation)
    mux.HandleFunc("GET /models", handleListModels)

    // Swagger documentation
    mux.Handle("/swagger/", httpswagger.WrapHandler)
    mux.HandleFunc("GET /swagger/doc.json", handleSwaggerJSON)
    mux.HandleFunc("GET /swagger/models", handleSwaggerModels)

    // Start unified registration system
    ctx := context.Background()
    go registry.StartRegistration(ctx, mux)

    // Start HTTP server with optimized timeouts
    server := &http.Server{
        Addr:         ":" + port,
        Handler:      mux,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    server.ListenAndServe()
}
```

### 3. Registry Initialization (`unified_registry.go:52-70`)

```go
func (ur *UnifiedRegistry) StartAutoRegistration(ctx context.Context, mux *http.ServeMux) error {
    ur.mux = mux

    // Phase 1: Discover and register all models
    ur.discoverAndRegisterAll()

    // Phase 2: Register HTTP endpoints for discovered models
    ur.registerAllHTTPEndpoints()

    // System ready
    return nil
}
```

## Request Processing Pipeline

### Step 1: HTTP Request Reception

When a request arrives at any endpoint:

1. **Route Matching**: Go's `http.ServeMux` matches the route pattern
2. **Method Validation**: HTTP method is validated against registered handlers
3. **Content-Type Check**: JSON content-type validation for POST requests

### Step 2: Request Dispatch

#### Generic Validation (`/validate`) - `main.go:121-176`

```go
func handleGenericValidation(w http.ResponseWriter, r *http.Request) {
    var request struct {
        ModelType string                 `json:"model_type"`
        Payload   map[string]interface{} `json:"payload"`
    }

    // 1. Parse JSON request
    json.NewDecoder(r.Body).Decode(&request)

    // 2. Lookup model in registry
    globalRegistry := registry.GetGlobalRegistry()
    modelType := registry.ModelType(request.ModelType)

    // 3. Create model instance
    modelInstance, err := globalRegistry.CreateModelInstance(modelType)

    // 4. Convert map to struct
    convertMapToStruct(request.Payload, modelInstance)

    // 5. Validate using registry
    result, err := globalRegistry.ValidatePayload(modelType, modelValue)

    // 6. Return JSON response
    json.NewEncoder(w).Encode(result)
}
```

#### Model-Specific Validation (`/validate/{type}`) - `unified_registry.go:468-497`

```go
func (ur *UnifiedRegistry) createDynamicHandler(modelType ModelType, modelInfo *ModelInfo) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // 1. Create model instance using reflection
        modelInstance := reflect.New(modelInfo.ModelStruct).Interface()

        // 2. Parse JSON directly into struct
        json.NewDecoder(r.Body).Decode(modelInstance)

        // 3. Extract struct value
        modelValue := reflect.ValueOf(modelInstance).Elem().Interface()

        // 4. Validate using registered validator
        result, err := ur.ValidatePayload(modelType, modelValue)

        // 5. Return response
        json.NewEncoder(w).Encode(result)
    }
}
```

### Step 3: Payload Conversion

#### Map to Struct Conversion (`main.go:216-261`)

```go
func convertMapToStruct(src map[string]interface{}, dest interface{}) error {
    destValue := reflect.ValueOf(dest).Elem()
    destType := destValue.Type()

    for i := 0; i < destValue.NumField(); i++ {
        field := destValue.Field(i)
        fieldType := destType.Field(i)

        // Get JSON tag name
        jsonTag := fieldType.Tag.Get("json")
        fieldName := strings.Split(jsonTag, ",")[0]

        // Get value from source map
        srcValue := src[fieldName]

        // Convert and set value with type safety
        setFieldValue(field, srcValue)
    }
}
```

### Step 4: Validation Execution

#### Universal Validator Wrapper (`model_registry.go:38-84`)

```go
func (uvw *UniversalValidatorWrapper) ValidatePayload(payload interface{}) interface{} {
    validatorValue := reflect.ValueOf(uvw.validatorInstance)

    // Try multiple method names
    methodNames := []string{"ValidatePayload", "Validate", "ValidateRequest"}

    for _, methodName := range methodNames {
        validateMethod := validatorValue.MethodByName(methodName)
        if validateMethod.IsValid() {
            // Call validator method
            results := validateMethod.Call([]reflect.Value{reflect.ValueOf(payload)})
            return results[0].Interface()
        }
    }

    // Fallback error
    return createValidationError()
}
```

#### Base Validation Framework (`base_validator.go:108-156`)

```go
func (bv *BaseValidator) ValidateWithBusinessLogic(
    payload interface{},
    businessLogicFunc func(interface{}) []models.ValidationWarning,
) models.ValidationResult {
    start := time.Now()
    result := bv.CreateValidationResult()

    // 1. Struct validation using go-playground/validator
    if err := bv.validator.Struct(payload); err != nil {
        result.IsValid = false
        // Convert validator errors to standard format
        for _, ve := range err.(validator.ValidationErrors) {
            result.Errors = append(result.Errors, models.ValidationError{
                Field:   ve.Field(),
                Message: FormatValidationError(ve, bv.modelType),
                Code:    GetErrorCode(ve.Tag()),
                Value:   fmt.Sprintf("%v", ve.Value()),
            })
        }
    }

    // 2. Apply business logic validation
    if result.IsValid && businessLogicFunc != nil {
        businessWarnings := businessLogicFunc(payload)
        result.Warnings = append(result.Warnings, businessWarnings...)
    }

    // 3. Add performance metrics
    bv.AddPerformanceMetrics(&result, start)

    return result
}
```

## Model Registration System

### Automatic Discovery Process (`unified_registry.go:72-120`)

```go
func (ur *UnifiedRegistry) discoverAndRegisterAll() error {
    // 1. Scan models directory for Go files
    modelFiles, _ := filepath.Glob(filepath.Join(ur.modelsPath, "*.go"))

    for _, modelFile := range modelFiles {
        baseName := strings.TrimSuffix(filepath.Base(modelFile), ".go")

        // Skip test files
        if strings.HasSuffix(baseName, "_test") { continue }

        // 2. Check if corresponding validator exists
        validatorFile := filepath.Join(ur.validationsPath, baseName+".go")
        if _, err := os.Stat(validatorFile); os.IsNotExist(err) {
            continue
        }

        // 3. Register model automatically
        ur.registerModelAutomatically(baseName)
    }
}
```

### Model Struct Discovery (`unified_registry.go:162-193`)

```go
func (ur *UnifiedRegistry) discoverModelStruct(baseName string) (reflect.Type, string, error) {
    // 1. Parse Go file using AST
    modelFile := filepath.Join(ur.modelsPath, baseName+".go")
    discoveredStructs := ur.parseGoFileForStructs(modelFile)

    // 2. Try naming conventions
    titleCase := toTitleCase(baseName)
    possibleNames := append(discoveredStructs,
        titleCase+"Payload",
        titleCase+"Model",
        titleCase+"Request",
        titleCase+"Data",
        titleCase,
    )

    // 3. Match against known types
    knownTypes := ur.getKnownModelTypes()
    for _, name := range possibleNames {
        if structType, exists := knownTypes[name]; exists {
            return structType, name, nil
        }
    }
}
```

### Validator Constructor Discovery (`unified_registry.go:216-254`)

```go
func (ur *UnifiedRegistry) createValidatorInstance(baseName string) (interface{}, error) {
    // Handle special naming cases
    specialCases := map[string]string{
        "github": "GitHub",
        "api":    "API",
    }

    titleCase := specialCases[baseName]
    if titleCase == "" {
        titleCase = toTitleCase(baseName)
    }

    // Try constructor patterns
    possibleNames := []string{
        "New" + titleCase + "Validator",     // NewGitHubValidator
        "New" + toTitleCase(baseName) + "Validator", // NewGithubValidator
    }

    knownValidators := ur.getKnownValidatorConstructors()
    for _, constructorName := range possibleNames {
        if constructor, exists := knownValidators[constructorName]; exists {
            return constructor(), nil
        }
    }
}
```

### HTTP Endpoint Registration (`unified_registry.go:443-466`)

```go
func (ur *UnifiedRegistry) registerAllHTTPEndpoints() {
    for modelType, modelInfo := range ur.models {
        endpointPath := "/validate/" + string(modelType)

        // Create closure to capture variables
        func(mt ModelType, mi *ModelInfo, path string) {
            ur.mux.HandleFunc("POST "+path, ur.createDynamicHandler(mt, mi))
            log.Printf("✅ Registered endpoint: POST %s -> %s", path, mi.Name)
        }(modelType, modelInfo, endpointPath)
    }
}
```

## Validation Framework

### Model Definitions

Models are defined in `src/models/` with comprehensive validation tags:

```go
// Example: API Request Model (models/api.go:9-31)
type APIRequest struct {
    Method        string                 `json:"method" validate:"required,oneof=GET POST PUT PATCH DELETE"`
    URL           string                 `json:"url" validate:"required,url"`
    Headers       map[string]string      `json:"headers" validate:"omitempty"`
    QueryParams   map[string]interface{} `json:"query_params" validate:"omitempty"`
    Body          interface{}            `json:"body" validate:"omitempty"`
    Timestamp     time.Time              `json:"timestamp" validate:"required"`
    RequestID     string                 `json:"request_id" validate:"omitempty,min=1,max=255"`
    UserAgent     string                 `json:"user_agent" validate:"omitempty,max=1000"`
    RemoteIP      string                 `json:"remote_ip" validate:"omitempty,ip"`
    Authorization *APIAuthorization      `json:"authorization,omitempty" validate:"omitempty"`
}
```

### Validator Implementation Pattern

Each model has a corresponding validator in `src/validations/`:

```go
// Example: API Validator (validations/api.go)
type APIValidator struct {
    *BaseValidator
}

func NewAPIValidator() *APIValidator {
    return &APIValidator{
        BaseValidator: NewBaseValidator("api", "api-validator"),
    }
}

func (av *APIValidator) ValidatePayload(payload interface{}) models.ValidationResult {
    apiRequest, ok := payload.(models.APIRequest)
    if !ok {
        return av.createInvalidTypeResult("APIRequest")
    }

    return av.ValidateWithBusinessLogic(apiRequest, func(p interface{}) []models.ValidationWarning {
        return av.validateAPIBusinessLogic(p.(models.APIRequest))
    })
}

func (av *APIValidator) validateAPIBusinessLogic(api models.APIRequest) []models.ValidationWarning {
    var warnings []models.ValidationWarning

    // Custom business logic
    if api.Method == "POST" && api.Body == nil {
        warnings = append(warnings, models.ValidationWarning{
            Field:   "body",
            Message: "POST requests typically include a request body",
            Code:    "MISSING_BODY_WARNING",
        })
    }

    return warnings
}
```

## API Endpoints

### Core System Endpoints

#### Health Check
```
GET /health
```
**Response:**
```json
{
    "status": "healthy",
    "version": "2.0.0-modular",
    "uptime": "2h34m12s",
    "server": "modular-validation-server"
}
```

#### List Models
```
GET /models
```
**Response:**
```json
{
    "models": {
        "github": {
            "name": "GitHub Webhook",
            "description": "GitHub webhook payload validation...",
            "endpoint": "/validate/github",
            "version": "1.0.0"
        }
    },
    "count": 6
}
```

#### Generic Validation
```
POST /validate
Content-Type: application/json

{
    "model_type": "github",
    "payload": {
        "action": "push",
        "repository": {
            "name": "test-repo"
        }
    }
}
```

#### Model-Specific Validation
```
POST /validate/github
Content-Type: application/json

{
    "action": "push",
    "repository": {
        "name": "test-repo",
        "full_name": "user/test-repo"
    }
}
```

### Dynamic Endpoint Generation

All model-specific endpoints are generated automatically:
- `/validate/github` - GitHub webhook validation
- `/validate/api` - API request/response validation
- `/validate/incident` - Incident report validation
- `/validate/database` - Database operation validation
- `/validate/deployment` - Deployment validation
- `/validate/generic` - Generic payload validation

## Error Handling

### Validation Error Structure

```json
{
    "is_valid": false,
    "model_type": "github",
    "provider": "github-validator",
    "timestamp": "2024-09-27T10:30:00Z",
    "processing_duration": "15ms",
    "errors": [
        {
            "field": "repository.name",
            "message": "Field 'name' is required",
            "code": "REQUIRED_FIELD_MISSING",
            "value": ""
        }
    ],
    "warnings": [
        {
            "field": "action",
            "message": "Action 'opened' is less common than 'push'",
            "code": "UNCOMMON_VALUE_WARNING",
            "suggestion": "Consider using standard GitHub webhook actions"
        }
    ]
}
```

### Error Code Standards (`config/constants.go:13-22`)

```go
const (
    ErrCodeValidationFailed = "VALIDATION_FAILED"
    ErrCodeRequiredMissing  = "REQUIRED_FIELD_MISSING"
    ErrCodeValueTooShort    = "VALUE_TOO_SHORT"
    ErrCodeValueTooLong     = "VALUE_TOO_LONG"
    ErrCodeInvalidFormat    = "INVALID_FORMAT"
    ErrCodeInvalidEmail     = "INVALID_EMAIL_FORMAT"
    ErrCodeInvalidURL       = "INVALID_URL_FORMAT"
    ErrCodeInvalidEnum      = "INVALID_ENUM_VALUE"
)
```

### HTTP Error Responses

```go
func sendJSONError(w http.ResponseWriter, message string, status int) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "error":  message,
        "status": status,
    })
}
```

## Performance Monitoring

### Built-in Metrics (`base_validator.go:44-64`)

```go
func (bv *BaseValidator) AddPerformanceMetrics(result *models.ValidationResult, start time.Time) {
    duration := time.Since(start)
    result.ProcessingDuration = duration

    result.PerformanceMetrics = &models.PerformanceMetrics{
        ValidationDuration: duration,
        FieldCount:        bv.countStructFields(result.ModelType),
        RuleCount:         bv.getRuleCount(),
        MemoryUsage:       getApproximateMemoryUsage(),
    }

    // Performance warning for slow validations
    if config.IsSlowValidation(duration) {
        result.Warnings = append(result.Warnings, models.ValidationWarning{
            Field:   "performance",
            Message: fmt.Sprintf("Validation took %v (longer than expected)", duration),
            Code:    config.ErrCodeValidationFailed,
        })
    }
}
```

### Performance Thresholds (`config/constants.go:7-10`)

```go
const (
    SlowValidationThreshold = 100 * time.Millisecond
)

func IsSlowValidation(duration time.Duration) bool {
    return duration > SlowValidationThreshold
}
```

## Testing Architecture

### Unit Testing Structure

```
src/
├── models/
│   ├── api_test.go           # Model structure tests
│   ├── github_test.go        # GitHub model tests
│   └── incident_test.go      # Incident model tests
├── validations/
│   ├── base_validator_test.go # Base framework tests
│   ├── api_test.go           # API validation tests
│   └── incident_test.go      # Incident validation tests
└── registry/
    └── unified_registry_test.go # Registry system tests
```

### E2E Testing

The project includes comprehensive end-to-end testing:
- `e2e_test_suite.sh` - Complete system testing
- `test_data/` - Sample validation payloads
- `coverage/` - Coverage reports

### Running Tests

```bash
# Unit tests
go test ./src/... -v

# E2E tests
./e2e_test_suite.sh

# Coverage report
go test ./src/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Contributing Guidelines

### Adding a New Model

1. **Create Model Structure** (`src/models/newmodel.go`):
```go
package models

type NewModelPayload struct {
    Field1 string `json:"field1" validate:"required,min=1"`
    Field2 int    `json:"field2" validate:"gte=0"`
}
```

2. **Create Validator** (`src/validations/newmodel.go`):
```go
package validations

type NewModelValidator struct {
    *BaseValidator
}

func NewNewModelValidator() *NewModelValidator {
    return &NewModelValidator{
        BaseValidator: NewBaseValidator("newmodel", "newmodel-validator"),
    }
}

func (nv *NewModelValidator) ValidatePayload(payload interface{}) models.ValidationResult {
    newModel, ok := payload.(models.NewModelPayload)
    if !ok {
        return nv.createInvalidTypeResult("NewModelPayload")
    }

    return nv.ValidateWithBusinessLogic(newModel, nv.validateNewModelBusinessLogic)
}

func (nv *NewModelValidator) validateNewModelBusinessLogic(nm models.NewModelPayload) []models.ValidationWarning {
    var warnings []models.ValidationWarning
    // Add custom validation logic
    return warnings
}
```

3. **Register Model** (`src/registry/unified_registry.go`):

Add to `getKnownModelTypes()`:
```go
"NewModelPayload": reflect.TypeOf(models.NewModelPayload{}),
```

Add to `getKnownValidatorConstructors()`:
```go
"NewNewModelValidator": func() interface{} { return validations.NewNewModelValidator() },
```

4. **Test the Model**:
```bash
# Start server
go run src/main.go

# Test endpoint
curl -X POST http://localhost:8080/validate/newmodel \
  -H "Content-Type: application/json" \
  -d '{"field1": "test", "field2": 42}'
```

### Code Style Guidelines

1. **Naming Conventions**:
   - Models: `{Type}Payload` struct in `models/{type}.go`
   - Validators: `{Type}Validator` in `validations/{type}.go`
   - Constructors: `New{Type}Validator()` function

2. **Error Handling**:
   - Use standardized error codes from `config/constants.go`
   - Include helpful error messages and suggestions
   - Add context information to errors

3. **Performance**:
   - Pre-allocate slices when size is known
   - Use efficient type conversion methods
   - Add performance metrics to new validators

4. **Testing**:
   - Write unit tests for all new models and validators
   - Include both valid and invalid test cases
   - Test edge cases and error conditions

5. **Documentation**:
   - Add comprehensive comments to public APIs
   - Update this guide when adding new concepts
   - Include examples in validation logic

### Development Workflow

1. **Setup Development Environment**:
```bash
cd /path/to/project
go mod tidy
```

2. **Run in Development Mode**:
```bash
# Start server with hot reload (if using air or similar)
go run src/main.go

# Or with specific port
PORT=3000 go run src/main.go
```

3. **Testing During Development**:
```bash
# Quick validation test
./simple_test.sh

# Full test suite
./e2e_test_suite.sh

# Unit tests only
go test ./src/... -v -short
```

4. **Build for Production**:
```bash
go build -o validator src/main.go
```

This guide provides a complete understanding of the codebase architecture and execution flow. The system is designed to be easily extensible while maintaining high performance and code quality standards.