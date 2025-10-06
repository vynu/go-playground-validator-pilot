# Code Execution Flow Guide

> **For new developers**: This guide explains how the validation server works from startup to request processing.

## Table of Contents
1. [Quick Overview](#quick-overview)
2. [System Architecture](#system-architecture)
3. [Directory Structure](#directory-structure)
4. [Startup Flow](#startup-flow)
5. [Request Processing](#request-processing)
6. [Validation Flow](#validation-flow)
7. [Array Validation (Detailed Flow)](#array-validation-detailed-flow)
8. [Batch Processing (Multi-Request Sessions)](#batch-processing-multi-request-sessions)
9. [Concurrency & Thread Safety](#concurrency--thread-safety)
10. [Adding New Models](#adding-new-models)
11. [Key Interfaces](#key-interfaces)
12. [Quick Reference](#quick-reference)

---

## Quick Overview

This is a **Go validation server** that automatically discovers and registers validation models at startup. You drop model files in `src/models/` and validator files in `src/validations/` - the system auto-discovers everything.

**Key Features:**
- 🚀 **Zero Configuration** - Auto-discovers models and validators
- 🎯 **Single & Batch Validation** - Validate one record or thousands
- 📊 **Threshold Support** - Set minimum success rates for batches
- 🔌 **Auto-Generated Endpoints** - REST endpoints created automatically
- ⚡ **Built on go-playground/validator** - Industry-standard validation

---

## System Architecture

```
┌──────────────────────────────────────────────────────────────┐
│                     HTTP REQUEST                              │
│  POST /validate {"model_type":"incident", "data":[...]}      │
└───────────────────────┬──────────────────────────────────────┘
                        │
                        ▼
┌──────────────────────────────────────────────────────────────┐
│                  HTTP ROUTER (net/http)                       │
│  Routes: /health, /validate, /models, /validate/{type}       │
└───────────────────────┬──────────────────────────────────────┘
                        │
                        ▼
┌──────────────────────────────────────────────────────────────┐
│              UNIFIED REGISTRY SYSTEM                          │
│  • Model Discovery (AST parsing)                             │
│  • Validator Matching (reflection)                           │
│  • Endpoint Generation                                       │
└───────────────────────┬──────────────────────────────────────┘
                        │
                        ▼
┌──────────────────────────────────────────────────────────────┐
│                VALIDATION PROCESSING                          │
│  Single Record:  ValidatePayload(model, data)                │
│  Array:          ValidateArray(model, records, threshold)    │
└───────────────────────┬──────────────────────────────────────┘
                        │
                        ▼
┌──────────────────────────────────────────────────────────────┐
│          go-playground/validator + Custom Logic               │
│  • Struct tag validation (required, min, max, etc.)          │
│  • Custom business rules (ID format, consistency, etc.)      │
│  • Warnings generation (best practices)                      │
└───────────────────────┬──────────────────────────────────────┘
                        │
                        ▼
┌──────────────────────────────────────────────────────────────┐
│                  JSON RESPONSE                                │
│  {"is_valid":true, "errors":[], "warnings":[], ...}          │
└──────────────────────────────────────────────────────────────┘
```

---

## Directory Structure

```
src/
├── main.go                      # Entry point, HTTP setup
├── models/                      # Data models with validation tags
│   ├── incident.go              # IncidentPayload struct
│   ├── api.go                   # APIRequest struct
│   ├── github.go                # GitHubPayload struct
│   ├── validation_result.go     # Common result types
│   └── batch_session.go         # Batch processing support
│
├── validations/                 # Validation logic
│   ├── base_validator.go        # Shared validation framework
│   ├── incident.go              # NewIncidentValidator()
│   ├── api.go                   # NewAPIValidator()
│   └── github.go                # NewGitHubValidator()
│
├── registry/                    # Auto-discovery system
│   ├── model_registry.go        # Core types and interfaces
│   ├── unified_registry.go      # Auto-discovery engine
│   └── dynamic_registry.go      # Runtime utilities
│
└── config/
    └── constants.go             # Error codes, thresholds

test_data/                       # Sample payloads for testing
e2e_test_suite.sh               # End-to-end tests
```

---

## Startup Flow

### Step 1: Server Initialization
**File**: `src/main.go:25-37`

```go
func main() {
    log.Println("Starting Modular Validation Server...")
    startModularServer()
}
```

### Step 2: HTTP Routes Setup
**File**: `src/main.go:40-76`

```go
func startModularServer() {
    mux := http.NewServeMux()

    // Core endpoints
    mux.HandleFunc("GET /health", handleHealth)
    mux.HandleFunc("POST /validate", handleGenericValidation)
    mux.HandleFunc("GET /models", handleListModels)

    // Swagger docs
    mux.Handle("/swagger/", httpswagger.WrapHandler)

    // Start auto-discovery
    ctx := context.Background()
    go registry.StartRegistration(ctx, mux)

    // Start server on port 8080 (or $PORT)
    server := &http.Server{Addr: ":" + port, Handler: mux}
    server.ListenAndServe()
}
```

### Step 3: Auto-Discovery
**File**: `src/registry/unified_registry.go:52-70`

```go
func (ur *UnifiedRegistry) StartAutoRegistration(ctx, mux) error {
    // Phase 1: Discover all models
    ur.discoverAndRegisterAll()

    // Phase 2: Create HTTP endpoints
    ur.registerAllHTTPEndpoints()

    return nil
}
```

**Discovery Process** (`unified_registry.go:72-110`):
```go
func (ur *UnifiedRegistry) discoverAndRegisterAll() error {
    // 1. Scan src/models/*.go
    modelFiles := glob("src/models/*.go")

    for each modelFile {
        baseName := "incident" // from incident.go

        // 2. Check validator exists: src/validations/incident.go
        if validatorExists(baseName) {
            // 3. Auto-register
            ur.registerModelAutomatically(baseName)
        }
    }
}
```

**Registration** (`unified_registry.go:122-159`):
```go
func (ur *UnifiedRegistry) registerModelAutomatically(baseName) error {
    // 1. Find struct: IncidentPayload (from models/incident.go)
    modelStruct := discoverModelStruct(baseName) // uses AST

    // 2. Find constructor: NewIncidentValidator (from validations/incident.go)
    validator := createValidatorInstance(baseName) // uses reflection

    // 3. Register in map
    ur.models["incident"] = &ModelInfo{
        Name: "Incident Report",
        ModelStruct: modelStruct,
        Validator: validator,
    }
}
```

**Endpoint Creation** (`unified_registry.go:443-466`):
```go
func (ur *UnifiedRegistry) registerAllHTTPEndpoints() {
    for modelType, modelInfo := range ur.models {
        path := "/validate/" + modelType  // "/validate/incident"

        // Create handler
        handler := ur.createDynamicHandler(modelType, modelInfo)

        // Register route
        ur.mux.HandleFunc("POST " + path, handler)

        log.Printf("✅ Registered endpoint: POST %s", path)
    }
}
```

**System Ready**: All models discovered, all endpoints registered!

---

## Request Processing

### Single Record Validation

**Request**:
```bash
POST /validate
{
  "model_type": "incident",
  "payload": {
    "id": "INC-20250106-0001",
    "title": "Critical payment processing bug requiring urgent attention",
    "severity": "critical",
    "priority": 5,
    ...
  }
}
```

**Flow** (`src/main.go:133-200`):
```go
func handleGenericValidation(w, r) {
    // 1. Parse request
    var request struct {
        ModelType string
        Payload   map[string]interface{}
    }
    json.NewDecoder(r.Body).Decode(&request)

    // 2. Get registry
    registry := registry.GetGlobalRegistry()

    // 3. Create model instance
    modelInstance := registry.CreateModelInstance("incident")
    // Returns: &models.IncidentPayload{}

    // 4. Convert map → struct
    convertMapToStruct(request.Payload, modelInstance)

    // 5. Validate
    result := registry.ValidatePayload("incident", modelInstance)

    // 6. Return JSON
    json.NewEncoder(w).Encode(result)
}
```

### Array Validation

**Request**:
```bash
POST /validate
{
  "model_type": "incident",
  "threshold": 80.0,
  "data": [
    {"id": "INC-20250106-0001", ...},
    {"id": "INC-20250106-0002", ...},
    {"id": "INC-20250106-0003", ...}
  ]
}
```

**Flow** (`src/main.go:201-214`):
```go
func handleGenericValidation(w, r) {
    // ... parse request ...

    // Array validation path
    if len(request.Data) > 0 {
        result := registry.ValidateArray(
            modelType,
            request.Data,
            request.Threshold,  // Optional: 80.0
        )

        json.NewEncoder(w).Encode(result)
        return
    }
}
```

---

## Validation Flow

### Step 1: Validator Lookup
**File**: `src/registry/unified_registry.go:363-383`

```go
func (ur *UnifiedRegistry) ValidatePayload(modelType, payload) (interface{}, error) {
    // Get registered model info
    modelInfo := ur.models[modelType]

    // Call validator
    result := modelInfo.Validator.ValidatePayload(payload)

    return result, nil
}
```

### Step 2: Struct Validation
**File**: `src/validations/incident.go:26-48`

```go
func (iv *IncidentValidator) ValidatePayload(payload) ValidationResult {
    result := ValidationResult{IsValid: true, Errors: []}

    // 1. go-playground/validator struct validation
    err := iv.validator.Struct(payload)
    if err != nil {
        result.IsValid = false
        for _, ve := range err.(validator.ValidationErrors) {
            result.Errors = append(result.Errors, ValidationError{
                Field:   ve.Field(),    // "Title"
                Message: "Field must be at least 10 characters long",
                Code:    "VALIDATION_FAILED",
                Value:   ve.Value(),
            })
        }
    }

    return result
}
```

**Model Tags** (`src/models/incident.go:7-22`):
```go
type IncidentPayload struct {
    ID       string `json:"id" validate:"required,min=3,max=50"`
    Title    string `json:"title" validate:"required,min=10,max=200"`
    Severity string `json:"severity" validate:"required,oneof=low medium high critical"`
    Priority int    `json:"priority" validate:"required,min=1,max=5"`
    ...
}
```

### Step 3: Custom Validation
**File**: `src/validations/incident.go:50-73`

```go
func (iv *IncidentValidator) ValidatePayload(payload) ValidationResult {
    // ... after struct validation ...

    if result.IsValid {
        // Custom Validation 1: ID format (INC-YYYYMMDD-NNNN)
        if err := iv.validateIncidentIDFormat(payload.ID); err != nil {
            result.IsValid = false
            result.Errors = append(result.Errors, ValidationError{
                Field:   "id",
                Message: "incident ID must follow format INC-YYYYMMDD-NNNN",
                Code:    "INVALID_ID_FORMAT",
            })
        }

        // Custom Validation 2: Priority vs Severity consistency
        if err := iv.validatePrioritySeverityConsistency(
            payload.Priority, payload.Severity,
        ); err != nil {
            result.Errors = append(result.Errors, ...)
        }
    }

    return result
}
```

### Step 4: Business Warnings
**File**: `src/validations/incident.go:142-207`

```go
func (iv *IncidentValidator) validateBusinessLogic(payload) []ValidationWarning {
    warnings := []ValidationWarning{}

    // Warning 1: Critical incidents should be assigned
    if payload.Severity == "critical" && payload.AssignedTo == "" {
        warnings = append(warnings, ValidationWarning{
            Field:      "assigned_to",
            Message:    "Critical incident should be assigned immediately",
            Code:       "CRITICAL_INCIDENT_UNASSIGNED",
            Suggestion: "Assign to on-call engineer",
        })
    }

    // Warning 2: Production issues should have high priority
    if payload.Environment == "production" && payload.Priority < 3 {
        warnings = append(warnings, ValidationWarning{
            Field:      "priority",
            Message:    "Production incident has low priority",
            Code:       "PRODUCTION_LOW_PRIORITY",
        })
    }

    return warnings
}
```

### Step 5: Result Response

**Response**:
```json
{
  "is_valid": false,
  "model_type": "incident",
  "provider": "go-playground",
  "timestamp": "2025-10-06T14:30:00Z",
  "processing_duration": "5ms",
  "errors": [
    {
      "field": "id",
      "message": "incident ID must follow format INC-YYYYMMDD-NNNN (e.g., INC-20240924-0001), got: INC-123",
      "code": "INVALID_ID_FORMAT",
      "value": "INC-123"
    }
  ],
  "warnings": [
    {
      "field": "assigned_to",
      "message": "Critical incident should be assigned to an engineer immediately",
      "code": "CRITICAL_INCIDENT_UNASSIGNED",
      "suggestion": "Assign to on-call engineer or escalation team"
    }
  ]
}
```

---

## Array Validation (Detailed Flow)

Array validation allows you to validate multiple records in a single request, with optional quality gate thresholds.

### Array Validation Flow Diagram

```
POST /validate with "data": [...]
        ▼
┌──────────────────────────────────────┐
│ 1. Detect Array vs Single Object    │
│    if len(request.Data) > 0          │
└──────────┬───────────────────────────┘
           ▼
┌──────────────────────────────────────┐
│ 2. Call ValidateArray()              │
│    - modelType: "incident"           │
│    - records: array of maps          │
│    - threshold: 80.0 (optional)      │
└──────────┬───────────────────────────┘
           ▼
┌──────────────────────────────────────┐
│ 3. Initialize Result                 │
│    - Status: "success"               │
│    - TotalRecords: len(records)      │
│    - ValidRecords: 0                 │
│    - InvalidRecords: 0               │
└──────────┬───────────────────────────┘
           ▼
┌──────────────────────────────────────┐
│ 4. Loop Through Records              │
│    for i, record := range records    │
└──────────┬───────────────────────────┘
           ▼
┌──────────────────────────────────────┐
│ 5. Per-Record Processing             │
│    a) CreateModelInstance()          │
│    b) convertMapToStruct()           │
│    c) ValidatePayload()              │
│    d) Increment counters             │
└──────────┬───────────────────────────┘
           ▼
┌──────────────────────────────────────┐
│ 6. Apply Threshold Check (if set)   │
│    successRate = (valid/total)*100   │
│    if rate >= threshold: "success"   │
│    else: "failed"                    │
└──────────┬───────────────────────────┘
           ▼
┌──────────────────────────────────────┐
│ 7. Return ArrayValidationResult      │
│    - HTTP 200: threshold met         │
│    - HTTP 422: threshold not met     │
└──────────────────────────────────────┘
```

### Without Threshold (Basic Batch)
**File**: `src/registry/unified_registry.go:436-515`

```go
func (ur *UnifiedRegistry) ValidateArray(modelType, records, threshold) (*ArrayValidationResult, error) {
    result := &ArrayValidationResult{
        Status:       "success",  // Always success without threshold
        TotalRecords: len(records),
    }

    // Validate each record
    for i, record := range records {
        modelInstance := ur.CreateModelInstance(modelType)
        convertMapToStruct(record, modelInstance)

        validationResult := ur.ValidatePayload(modelType, modelInstance)

        if validationResult.IsValid {
            result.ValidRecords++
        } else {
            result.InvalidRecords++
            // Include in results
            result.Results = append(result.Results, RowValidationResult{
                RowIndex: i,
                IsValid:  false,
                Errors:   validationResult.Errors,
            })
        }
    }

    return result, nil
}
```

**Response**:
```json
{
  "status": "success",
  "total_records": 5,
  "valid_records": 3,
  "invalid_records": 2,
  "results": [
    {
      "row_index": 1,
      "is_valid": false,
      "errors": [...]
    },
    {
      "row_index": 3,
      "is_valid": false,
      "errors": [...]
    }
  ]
}
```

### With Threshold (Quality Gate)

```go
func (ur *UnifiedRegistry) ValidateArray(modelType, records, threshold) (*ArrayValidationResult, error) {
    // ... validate all records ...

    // Calculate success rate
    successRate := (float64(validCount) / float64(totalCount)) * 100.0

    // Apply threshold check
    if threshold != nil {
        result.Threshold = *threshold
        result.SuccessRate = successRate

        if successRate >= *threshold {
            result.Status = "success"  // HTTP 200
        } else {
            result.Status = "failed"   // HTTP 422
        }
    }

    return result, nil
}
```

**Request** (80% threshold):
```json
{
  "model_type": "incident",
  "threshold": 80.0,
  "data": [...]
}
```

**Response** (Success - 100% >= 80%):
```json
{
  "status": "success",
  "threshold": 80.0,
  "success_rate": 100.0,
  "total_records": 5,
  "valid_records": 5,
  "invalid_records": 0
}
```

**Response** (Failure - 60% < 80%):
```json
{
  "status": "failed",
  "threshold": 80.0,
  "success_rate": 60.0,
  "total_records": 5,
  "valid_records": 3,
  "invalid_records": 2,
  "results": [...]
}
```

**Use Cases**:
- **Data Import**: Require 95% valid before importing to database
- **CI/CD Gates**: Enforce 100% valid test data in pipelines
- **Batch Processing**: Accept batches with 90%+ success rate

---

## Batch Processing (Multi-Request Sessions)

Batch processing allows validating large datasets across **multiple HTTP requests** using a session-based approach. This is ideal for streaming data or chunked uploads.

### Batch Processing Flow Diagram

```
Client                          Server
  │                               │
  │ 1. POST /validate/batch/start │
  │ ──────────────────────────────>│
  │                               │ Create BatchSession
  │                               │ Generate batch_id
  │ <──────────────────────────────│
  │   {batch_id: "abc-123"}       │
  │                               │
  │ 2. POST /validate             │
  │    X-Batch-ID: abc-123        │
  │    {data: [chunk1]}           │
  │ ──────────────────────────────>│
  │                               │ Validate chunk1
  │                               │ Update session counters
  │ <──────────────────────────────│
  │                               │
  │ 3. POST /validate             │
  │    X-Batch-ID: abc-123        │
  │    {data: [chunk2]}           │
  │ ──────────────────────────────>│
  │                               │ Validate chunk2
  │                               │ Update session counters
  │ <──────────────────────────────│
  │                               │
  │ 4. GET /validate/batch/{id}   │
  │ ──────────────────────────────>│
  │                               │ Return current stats
  │ <──────────────────────────────│
  │   {valid: 150, invalid: 10}   │
  │                               │
  │ 5. POST /validate/batch/{id}/complete │
  │ ──────────────────────────────>│
  │                               │ Finalize session
  │                               │ Apply threshold check
  │ <──────────────────────────────│
  │   {status: "success"}         │
  │                               │
```

### Step 1: Start Batch Session
**Endpoint**: `POST /validate/batch/start`
**File**: `src/main.go:1053-1088`

```go
func handleBatchStart(w, r) {
    var request struct {
        ModelType string   `json:"model_type"`  // "incident"
        JobID     string   `json:"job_id"`      // Optional client ID
        Threshold *float64 `json:"threshold"`   // 80.0
    }

    // Parse request
    json.NewDecoder(r.Body).Decode(&request)

    // Generate unique batch ID
    batchID := models.GenerateBatchID(request.JobID)
    // Example: "batch-incident-1696723456-abc123"

    // Create session
    batchManager := models.GetBatchSessionManager()
    session := batchManager.CreateBatchSession(batchID, request.Threshold)

    // Return session info
    return {
        "batch_id": session.BatchID,
        "status": "active",
        "started_at": session.StartedAt,
        "expires_at": session.StartedAt + 30min,
        "threshold": session.Threshold
    }
}
```

**Request**:
```bash
curl -X POST http://localhost:8080/validate/batch/start \
  -H "Content-Type: application/json" \
  -d '{
    "model_type": "incident",
    "job_id": "import-2025-01",
    "threshold": 95.0
  }'
```

**Response**:
```json
{
  "batch_id": "batch-incident-1704537600-import-2025-01",
  "status": "active",
  "started_at": "2025-01-06T14:00:00Z",
  "expires_at": "2025-01-06T14:30:00Z",
  "threshold": 95.0,
  "message": "Batch session created. Use X-Batch-ID header to add data."
}
```

### Step 2: Send Data Chunks
**Endpoint**: `POST /validate` (with `X-Batch-ID` header)
**File**: `src/main.go:133-214`

```go
func handleGenericValidation(w, r) {
    // Check for batch ID in header
    batchID := r.Header.Get("X-Batch-ID")

    if batchID != "" {
        // Batch mode: accumulate results
        result := registry.ValidateArray(modelType, request.Data, nil)

        // Update session counters
        batchManager := models.GetBatchSessionManager()
        batchManager.UpdateBatchSession(
            batchID,
            result.ValidRecords,
            result.InvalidRecords,
            result.WarningRecords,
        )

        return {
            "chunk_processed": true,
            "valid_records": result.ValidRecords,
            "invalid_records": result.InvalidRecords
        }
    }

    // Normal mode: single request
    // ... standard validation ...
}
```

**Request (Chunk 1)**:
```bash
curl -X POST http://localhost:8080/validate \
  -H "Content-Type: application/json" \
  -H "X-Batch-ID: batch-incident-1704537600-import-2025-01" \
  -d '{
    "model_type": "incident",
    "data": [
      {"id": "INC-20250106-0001", ...},
      {"id": "INC-20250106-0002", ...},
      {"id": "INC-20250106-0003", ...}
    ]
  }'
```

**Response**:
```json
{
  "chunk_processed": true,
  "valid_records": 3,
  "invalid_records": 0
}
```

### Step 3: Check Status (Optional)
**Endpoint**: `GET /validate/batch/{id}`
**File**: `src/main.go:1091-1107`

```go
func handleBatchStatus(w, r) {
    batchID := r.PathValue("id")

    batchManager := models.GetBatchSessionManager()
    session, exists := batchManager.GetBatchSession(batchID)
    if !exists {
        return 404
    }

    return session.GetStatus()
}
```

**Request**:
```bash
curl http://localhost:8080/validate/batch/batch-incident-1704537600-import-2025-01
```

**Response**:
```json
{
  "batch_id": "batch-incident-1704537600-import-2025-01",
  "total_records": 150,
  "valid_records": 145,
  "invalid_records": 5,
  "warning_records": 12,
  "threshold": 95.0,
  "started_at": "2025-01-06T14:00:00Z",
  "last_updated": "2025-01-06T14:05:23Z",
  "is_final": false
}
```

### Step 4: Complete Batch
**Endpoint**: `POST /validate/batch/{id}/complete`
**File**: `src/main.go:1110-1151`

```go
func handleBatchComplete(w, r) {
    batchID := r.PathValue("id")

    batchManager := models.GetBatchSessionManager()

    // Finalize and apply threshold check
    status, err := batchManager.FinalizeBatchSession(batchID)
    // status = "success" or "failed"

    session, _ := batchManager.GetBatchSession(batchID)

    // Set HTTP status
    if status == "failed" {
        w.WriteHeader(422) // Unprocessable Entity
    } else {
        w.WriteHeader(200)
    }

    // Return final results
    return {
        "batch_id": session.BatchID,
        "status": status,
        "total_records": session.TotalRecords,
        "valid_records": session.ValidRecords,
        "invalid_records": session.InvalidRecords,
        "threshold": session.Threshold
    }

    // Auto-cleanup after 1 second
    go func() {
        time.Sleep(1 * time.Second)
        batchManager.DeleteBatchSession(batchID)
    }()
}
```

**Request**:
```bash
curl -X POST http://localhost:8080/validate/batch/batch-incident-1704537600-import-2025-01/complete
```

**Response (Success - 96.67% >= 95%)**:
```json
{
  "batch_id": "batch-incident-1704537600-import-2025-01",
  "status": "success",
  "total_records": 150,
  "valid_records": 145,
  "invalid_records": 5,
  "warning_records": 12,
  "threshold": 95.0,
  "started_at": "2025-01-06T14:00:00Z",
  "completed_at": "2025-01-06T14:10:45Z",
  "message": "Batch validation completed with status: success"
}
```

**Response (Failed - 90% < 95%)** (HTTP 422):
```json
{
  "batch_id": "batch-incident-1704537600-import-2025-01",
  "status": "failed",
  "total_records": 100,
  "valid_records": 90,
  "invalid_records": 10,
  "threshold": 95.0,
  "started_at": "2025-01-06T14:00:00Z",
  "completed_at": "2025-01-06T14:10:45Z",
  "message": "Batch validation completed with status: failed"
}
```

### Batch Session Management
**File**: `src/models/validation_result.go:138-290`

```go
type BatchSession struct {
    BatchID        string
    TotalRecords   int
    ValidRecords   int
    InvalidRecords int
    WarningRecords int
    Threshold      *float64
    StartedAt      time.Time
    LastUpdated    time.Time
    IsFinal        bool
    mutex          sync.RWMutex  // Thread-safe updates
}

type BatchSessionManager struct {
    sessions map[string]*BatchSession
    mutex    sync.RWMutex  // Protect concurrent access
}

// Global singleton
func GetBatchSessionManager() *BatchSessionManager {
    // Returns global instance
}

// Key methods
func (bsm *BatchSessionManager) CreateBatchSession(batchID, threshold)
func (bsm *BatchSessionManager) GetBatchSession(batchID)
func (bsm *BatchSessionManager) UpdateBatchSession(batchID, valid, invalid, warnings)
func (bsm *BatchSessionManager) FinalizeBatchSession(batchID) (status, error)
func (bsm *BatchSessionManager) DeleteBatchSession(batchID)
func (bsm *BatchSessionManager) CleanupExpiredBatches()  // Auto-cleanup routine
```

### Session Lifecycle

```
┌──────────────────────────────────────────────────────────┐
│                  BATCH SESSION STATES                    │
├──────────────────────────────────────────────────────────┤
│                                                          │
│  1. CREATED (POST /batch/start)                         │
│     - IsFinal: false                                     │
│     - Counters: 0                                        │
│     - Expires: 30 minutes                                │
│                                                          │
│  2. ACTIVE (POST /validate with X-Batch-ID)             │
│     - IsFinal: false                                     │
│     - Counters: incrementing                             │
│     - LastUpdated: updating                              │
│                                                          │
│  3. FINALIZED (POST /batch/{id}/complete)               │
│     - IsFinal: true                                      │
│     - Status: "success" or "failed"                      │
│     - Threshold check applied                            │
│                                                          │
│  4. DELETED (auto-cleanup after 1 second)               │
│     - Session removed from memory                        │
│                                                          │
└──────────────────────────────────────────────────────────┘
```

### Auto-Cleanup Routine
**File**: `src/models/validation_result.go:283-290`

```go
func (bsm *BatchSessionManager) StartCleanupRoutine() {
    go func() {
        ticker := time.NewTicker(10 * time.Minute)
        for range ticker.C {
            bsm.CleanupExpiredBatches()  // Remove sessions > 30min old
        }
    }()
}
```

Runs every 10 minutes, removes sessions older than 30 minutes.

### Batch vs Array: When to Use Which?

| Feature | Array Validation | Batch Processing |
|---------|------------------|------------------|
| **Request Count** | Single request | Multiple requests |
| **Max Records** | ~1000-10000 (memory limit) | Unlimited (streaming) |
| **Use Case** | Small to medium datasets | Large datasets, streaming |
| **Session State** | Stateless | Stateful (session tracking) |
| **Threshold** | Per-request | Per-batch (accumulated) |
| **API Calls** | 1 call | 3+ calls (start, chunks, complete) |
| **Network** | One large payload | Multiple small payloads |
| **Example** | Validating 500 records | Importing 1M records in 1000 chunks |

### Complete Batch Example

**Scenario**: Import 10,000 incident records with 95% quality threshold

```bash
# Step 1: Start batch
BATCH_ID=$(curl -X POST http://localhost:8080/validate/batch/start \
  -H "Content-Type: application/json" \
  -d '{"model_type":"incident","threshold":95.0}' \
  | jq -r '.batch_id')

# Step 2: Send data in chunks (100 records each)
for chunk in chunk_*.json; do
  curl -X POST http://localhost:8080/validate \
    -H "Content-Type: application/json" \
    -H "X-Batch-ID: $BATCH_ID" \
    -d @$chunk
done

# Step 3: Check status
curl http://localhost:8080/validate/batch/$BATCH_ID

# Step 4: Complete and get final result
curl -X POST http://localhost:8080/validate/batch/$BATCH_ID/complete
```

**Output**:
```
Chunk 1: 98 valid, 2 invalid
Chunk 2: 100 valid, 0 invalid
...
Chunk 100: 97 valid, 3 invalid

Final: 9,750 valid / 10,000 total = 97.5% (>= 95% threshold)
Status: SUCCESS ✅
```

---

## Concurrency & Thread Safety

The system handles **thousands of simultaneous requests** safely using Go's native concurrency and mutex protection.

### Goroutine-Per-Request Model

```
HTTP Server (Port 8080)
        │
        ├─ Request 1 arrives → Goroutine 1 spawned
        ├─ Request 2 arrives → Goroutine 2 spawned
        ├─ Request 3 arrives → Goroutine 3 spawned
        └─ ...
```

**File**: `src/main.go:84-91`

```go
server := &http.Server{
    Addr:         ":" + port,
    Handler:      mux,
    ReadTimeout:  15 * time.Second,
    WriteTimeout: 15 * time.Second,
    IdleTimeout:  60 * time.Second,
}
// Go automatically creates goroutine per request
```

### Thread-Safe Registry

**File**: `src/registry/unified_registry.go:34-40`

```go
type UnifiedRegistry struct {
    models map[ModelType]*ModelInfo  // Shared state
    mutex  sync.RWMutex              // Read-Write mutex
}
```

**Concurrent Reads** (multiple goroutines reading simultaneously):
```go
func (ur *UnifiedRegistry) GetModel(modelType) (*ModelInfo, error) {
    ur.mutex.RLock()        // ← Shared read lock
    defer ur.mutex.RUnlock()

    model := ur.models[modelType]  // Safe concurrent read
    return model, nil
}
```

**Exclusive Writes** (blocks all access):
```go
func (ur *UnifiedRegistry) RegisterModel(info *ModelInfo) error {
    ur.mutex.Lock()         // ← Exclusive write lock
    defer ur.mutex.Unlock()

    ur.models[info.Type] = info  // Safe write
    return nil
}
```

### Stateless Validators (No Locking Needed)

**File**: `src/validations/base_validator.go:16-20`

```go
type BaseValidator struct {
    validator *validator.Validate  // Read-only after init
    modelType string               // Immutable
    provider  string                // Immutable
}
// ✅ No shared mutable state = no mutex needed
```

### Batch Session Two-Level Locking

**File**: `src/models/validation_result.go:138-156`

```go
type BatchSession struct {
    BatchID     string
    // ... counters ...
    mutex       sync.RWMutex  // ← Per-session lock
}

type BatchSessionManager struct {
    sessions map[string]*BatchSession
    mutex    sync.RWMutex  // ← Manager-level lock
}
```

**Why two levels?**
- Manager lock protects the session map (create/get/delete)
- Session lock protects individual session data (update counters)
- Allows concurrent updates to **different** batch sessions

**Example**: Updating different batches concurrently

```go
// Client A updates batch-001
// Client B updates batch-002
// Both can proceed simultaneously!

func UpdateBatchSession(batchID, valid, invalid, warnings) {
    bsm.mutex.Lock()        // Brief lock to get session
    session := bsm.sessions[batchID]
    bsm.mutex.Unlock()      // Release manager lock quickly

    session.mutex.Lock()    // Lock only this session
    session.ValidRecords += valid
    session.mutex.Unlock()
}
```

### Singleton with sync.Once

**File**: `src/models/validation_result.go:158-171`

```go
var (
    globalBatchManager *BatchSessionManager
    batchManagerOnce   sync.Once
)

func GetBatchSessionManager() *BatchSessionManager {
    batchManagerOnce.Do(func() {
        // Executes exactly once, even with 1000 concurrent calls
        globalBatchManager = &BatchSessionManager{
            sessions: make(map[string]*BatchSession),
        }
    })
    return globalBatchManager
}
```

### Concurrent Request Scenario

**3 Requests Arrive Simultaneously**:

```
T0: All 3 requests hit server
    ├─ Request 1: POST /validate {"model_type":"github", ...}
    ├─ Request 2: POST /validate {"model_type":"incident", ...}
    └─ Request 3: POST /validate {"model_type":"api", ...}

T1: Registry lookups (CONCURRENT - RLock allows this)
    ├─ Goroutine 1: registry.GetModel("github") ← RLock()
    ├─ Goroutine 2: registry.GetModel("incident") ← RLock()
    └─ Goroutine 3: registry.GetModel("api") ← RLock()

T2: Model instance creation (ISOLATED)
    ├─ Goroutine 1: &GitHubPayload{}
    ├─ Goroutine 2: &IncidentPayload{}
    └─ Goroutine 3: &APIRequest{}

T3: Validation (PARALLEL)
    ├─ Goroutine 1: GitHubValidator.ValidatePayload()
    ├─ Goroutine 2: IncidentValidator.ValidatePayload()
    └─ Goroutine 3: APIValidator.ValidatePayload()

T4: Responses (INDEPENDENT)
    ├─ Response 1: {"is_valid": true, ...}
    ├─ Response 2: {"is_valid": false, ...}
    └─ Response 3: {"is_valid": true, ...}
```

**No contention** because:
1. Registry reads use `RLock()` (shared access)
2. Each request gets its own model instance
3. Validators are stateless (no shared state)

### Safety Guarantees

✅ **No Race Conditions**
- All shared state protected by mutexes
- Verified with `go test -race`

✅ **No Deadlocks**
- Consistent lock ordering (manager → session)
- `defer` ensures locks always released

✅ **No Response Mixing**
- Each request has isolated context
- Per-request model instances

✅ **Handles 1000+ req/s**
- RWMutex allows unlimited concurrent reads
- Stateless validators enable parallel processing

**For detailed concurrency analysis**, see [CONCURRENT_VALIDATION.md](CONCURRENT_VALIDATION.md).

---

## Adding New Models

### 1. Create Model Struct
**File**: `src/models/order.go`

```go
package models

type OrderPayload struct {
    OrderID    string  `json:"order_id" validate:"required,min=5"`
    CustomerID string  `json:"customer_id" validate:"required"`
    Amount     float64 `json:"amount" validate:"required,gt=0"`
    Status     string  `json:"status" validate:"required,oneof=pending paid shipped"`
}
```

**Validation Tags**:
- `required` - Field cannot be empty
- `min=5` - Minimum length/value
- `max=100` - Maximum length/value
- `oneof=a b c` - Must be one of these values
- `gt=0` - Greater than 0
- `email` - Must be valid email
- `url` - Must be valid URL
- `ip` - Must be valid IP address

### 2. Create Validator
**File**: `src/validations/order.go`

```go
package validations

import (
    "goplayground-data-validator/models"
)

type OrderValidator struct {
    *BaseValidator
}

func NewOrderValidator() *OrderValidator {
    return &OrderValidator{
        BaseValidator: NewBaseValidator("order", "order-validator"),
    }
}

func (ov *OrderValidator) ValidatePayload(payload interface{}) models.ValidationResult {
    order, ok := payload.(models.OrderPayload)
    if !ok {
        return ov.createInvalidTypeResult("OrderPayload")
    }

    // Use base validator framework
    return ov.ValidateWithBusinessLogic(order, func(p interface{}) []models.ValidationWarning {
        return ov.validateOrderBusinessLogic(p.(models.OrderPayload))
    })
}

func (ov *OrderValidator) validateOrderBusinessLogic(order models.OrderPayload) []models.ValidationWarning {
    var warnings []models.ValidationWarning

    // Custom business rule
    if order.Status == "paid" && order.Amount < 1.0 {
        warnings = append(warnings, models.ValidationWarning{
            Field:      "amount",
            Message:    "Paid orders typically have amount >= $1.00",
            Code:       "LOW_AMOUNT_WARNING",
            Suggestion: "Verify this is not a test order",
        })
    }

    return warnings
}
```

### 3. Register Model
**File**: `src/registry/unified_registry.go`

Add to `getKnownModelTypes()`:
```go
func (ur *UnifiedRegistry) getKnownModelTypes() map[string]reflect.Type {
    return map[string]reflect.Type{
        "OrderPayload": reflect.TypeOf(models.OrderPayload{}),  // ADD THIS
        // ... existing models ...
    }
}
```

Add to `getKnownValidatorConstructors()`:
```go
func (ur *UnifiedRegistry) getKnownValidatorConstructors() map[string]func() interface{} {
    return map[string]func() interface{}{
        "NewOrderValidator": func() interface{} { return validations.NewOrderValidator() },  // ADD THIS
        // ... existing validators ...
    }
}
```

### 4. Test It!

**Build and run**:
```bash
go build -o bin/validator src/main.go
PORT=8080 ./bin/validator
```

**Test request**:
```bash
curl -X POST http://localhost:8080/validate \
  -H "Content-Type: application/json" \
  -d '{
    "model_type": "order",
    "payload": {
      "order_id": "ORD-12345",
      "customer_id": "CUST-001",
      "amount": 99.99,
      "status": "paid"
    }
  }'
```

**Auto-generated endpoint**:
```bash
curl -X POST http://localhost:8080/validate/order \
  -H "Content-Type: application/json" \
  -d '{
    "order_id": "ORD-12345",
    "customer_id": "CUST-001",
    "amount": 99.99,
    "status": "paid"
  }'
```

---

## Key Interfaces

### ModelInfo
**File**: `src/registry/model_registry.go:18-24`

```go
type ModelInfo struct {
    Name        string                  // "Incident Report"
    Description string                  // "Incident reporting payload"
    Endpoint    string                  // "/validate/incident"
    Version     string                  // "1.0.0"
    ModelStruct reflect.Type            // reflect.TypeOf(IncidentPayload{})
    Validator   UniversalValidatorIface // Validator instance
}
```

### UniversalValidatorIface
**File**: `src/registry/model_registry.go:27-30`

```go
type UniversalValidatorIface interface {
    ValidatePayload(payload interface{}) interface{}
}
```

All validators must implement this interface.

### ValidationResult
**File**: `src/models/validation_result.go:9-20`

```go
type ValidationResult struct {
    IsValid            bool                  `json:"is_valid"`
    ModelType          string                `json:"model_type"`
    Provider           string                `json:"provider"`
    Timestamp          time.Time             `json:"timestamp"`
    ProcessingDuration time.Duration         `json:"processing_duration"`
    Errors             []ValidationError     `json:"errors"`
    Warnings           []ValidationWarning   `json:"warnings"`
    PerformanceMetrics *PerformanceMetrics   `json:"performance_metrics,omitempty"`
}
```

### ArrayValidationResult
**File**: `src/models/validation_result.go:48-67`

```go
type ArrayValidationResult struct {
    Status          string                   `json:"status"`           // "success" or "failed"
    Threshold       *float64                 `json:"threshold"`        // 80.0
    SuccessRate     float64                  `json:"success_rate"`     // 100.0
    TotalRecords    int                      `json:"total_records"`    // 5
    ValidRecords    int                      `json:"valid_records"`    // 5
    InvalidRecords  int                      `json:"invalid_records"`  // 0
    Results         []RowValidationResult    `json:"results"`          // Only invalid/warning records
}
```

---

## Quick Reference

### Available Endpoints
```
# Core Endpoints
GET  /health                        - Server health check
GET  /models                        - List all registered models
POST /validate                      - Generic validation (single or array)

# Batch Processing (Multi-Request Sessions)
POST /validate/batch/start          - Start new batch session
POST /validate                      - Add chunk (with X-Batch-ID header)
GET  /validate/batch/{id}           - Get batch status
POST /validate/batch/{id}/complete  - Finalize batch and get results

# Model-Specific Endpoints (Auto-Generated)
POST /validate/incident             - Incident-specific endpoint
POST /validate/api                  - API validation endpoint
POST /validate/github               - GitHub webhook validation
POST /validate/database             - Database query validation
POST /validate/deployment           - Deployment validation

# Documentation
GET  /swagger/                      - Swagger UI documentation
```

### Common Validation Tags
```go
validate:"required"                    // Cannot be empty
validate:"min=10,max=200"             // Length constraints
validate:"oneof=low medium high"      // Enum values
validate:"email"                      // Email format
validate:"url"                        // URL format
validate:"ip"                         // IP address
validate:"gte=0,lte=100"             // Number range
validate:"omitempty,min=3"           // Optional but if present min=3
```

### HTTP Status Codes
- `200 OK` - Validation completed (check `is_valid` field)
- `400 Bad Request` - Invalid JSON or missing model_type
- `404 Not Found` - Unknown model type
- `422 Unprocessable Entity` - Threshold not met (array validation)
- `500 Internal Server Error` - Server error

### Testing Commands
```bash
# Build
go build -o bin/validator src/main.go

# Run
PORT=8080 ./bin/validator

# Unit tests
go test ./src/... -v

# E2E tests
./e2e_test_suite.sh

# Docker
make docker-build
make docker-run
```

---

**For more details**:
- See `ADDING_NEW_MODELS_GUIDE.md` for step-by-step model creation
- See `E2E_TEST_GUIDE.md` for comprehensive testing guide
- See `THRESHOLD_VALIDATION_SUMMARY.md` for threshold validation examples
