# Go Playground Data Validator

> A production-ready, auto-discovering validation server built on `go-playground/validator` with support for single record, array, and multi-request batch validation.

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

---

## Table of Contents
- [Features vs GreatExpectations](#features-vs-greatexpectations)
- [Architecture Overview](#architecture-overview)
- [Quick Start](#quick-start)
- [API Endpoints](#api-endpoints)
- [Testing](#testing)
- [Error Codes & Best Practices](#error-codes--best-practices)
- [Documentation](#documentation)
- [Project Structure](#project-structure)

---

## Features vs GreatExpectations

| Feature | Go Playground Validator | GreatExpectations |
|---------|------------------------|-------------------|
| **Language** | Go (compiled, fast) | Python (interpreted) |
| **Performance** | ~0-15ms per validation | ~100-500ms per validation |
| **Deployment** | Single binary (12MB distroless) | Python + dependencies (~200MB+) |
| **Memory Usage** | ~30-50MB | ~150-300MB |
| **Auto-Discovery** | ✅ Zero config - auto-discovers models | ❌ Manual configuration required |
| **Real-time API** | ✅ REST API with <15ms response | ⚠️ Batch-oriented, slower |
| **Struct Validation** | ✅ Native Go struct tags | ❌ Python dictionaries/DataFrames |
| **Custom Rules** | ✅ Go functions (compiled) | ✅ Python expectations (interpreted) |
| **Threshold Support** | ✅ Built-in quality gates | ✅ Validation result stores |
| **Array Validation** | ✅ Single request, 1000s of records | ⚠️ Batch processing via files |
| **Batch Processing** | ✅ Multi-request sessions, unlimited scale | ✅ File-based batch jobs |
| **Concurrent Safety** | ✅ Thread-safe, handles 1000+ req/s | ⚠️ Single-threaded by default |
| **Docker Image** | 12MB (distroless) | 200MB+ (Python base) |
| **Startup Time** | <1 second | ~5-10 seconds |
| **HTTP Status Codes** | ✅ RESTful (200, 422, 404, etc.) | ⚠️ Primarily JSON responses |
| **Swagger Docs** | ✅ Auto-generated OpenAPI | ❌ Not available |
| **Use Case** | Real-time validation API, microservices | Data quality testing, batch ETL |

**Key Advantages**:
- 🚀 **10-50x faster** - Compiled Go vs interpreted Python
- 📦 **17x smaller** - 12MB vs 200MB+ Docker images
- ⚡ **Zero Config** - Models auto-discovered at startup
- 🔌 **Production Ready** - RESTful API, health checks, metrics
- 🎯 **Type Safe** - Compile-time type checking with Go structs

**When to Use This**:
- Real-time data validation in APIs/microservices
- High-throughput validation (1000+ records/second)
- Resource-constrained environments (Kubernetes, edge)
- Type-safe validation with compile-time guarantees

**When to Use GreatExpectations**:
- Data science workflows (Pandas/Spark integration)
- Complex statistical expectations (distributions, correlations)
- Jupyter notebook integration
- Existing Python data pipelines

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                        CLIENT REQUEST                            │
│  POST /validate {"model_type":"incident", "data":[...]}         │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                    HTTP ROUTER (Go net/http)                     │
│  • Health Check    • Generic Validation    • Batch Sessions     │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│              UNIFIED REGISTRY (Auto-Discovery)                   │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │ AST Parser   │  │  Reflection  │  │  HTTP Routes │          │
│  │ (Go Files)   │  │  (Validators)│  │  (Dynamic)   │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                  VALIDATION ENGINE                               │
│  ┌─────────────────────┐    ┌───────────────────────┐          │
│  │ go-playground       │    │ Custom Business Logic │          │
│  │ Struct Tags         │───>│ • ID formats          │          │
│  │ (required, min,     │    │ • Consistency checks  │          │
│  │  max, oneof, etc.)  │    │ • Warnings/Suggestions│          │
│  └─────────────────────┘    └───────────────────────┘          │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                       JSON RESPONSE                              │
│  {"is_valid":true, "errors":[], "warnings":[], "metrics":{...}} │
└─────────────────────────────────────────────────────────────────┘
```

**Key Components**:
1. **Auto-Discovery**: Scans `src/models/` and `src/validations/` at startup
2. **Registry**: Maps model types to validators using reflection
3. **Validation**: Two-layer approach (struct tags + custom rules)
4. **Response**: Standardized JSON with errors, warnings, and metrics

---

## Quick Start

### Prerequisites
- Go 1.21+ ([install](https://go.dev/doc/install))
- Make (optional, for convenience commands)

### Installation & Build

```bash
# Clone repository
git clone https://github.com/your-org/go-playground-validator-pilot.git
cd go-playground-validator-pilot

# Install dependencies
go mod tidy

# Build binary
go build -o bin/validator src/main.go

# Run server
PORT=8080 ./bin/validator
```

**Server output**:
```
2025/10/06 14:30:00 Starting Modular Validation Server...
2025/10/06 14:30:00 🚀 Starting unified automatic model registration system...
2025/10/06 14:30:00 ✅ Registered model: incident -> Incident Report
2025/10/06 14:30:00 ✅ Registered model: api -> API Request/Response
2025/10/06 14:30:00 ✅ Registered model: github -> GitHub Webhook
2025/10/06 14:30:00 🚀 Modular server starting on port 8080
```

### Docker

```bash
# Build Docker image
make docker-build

# Run container (distroless - 12MB image)
make docker-run

# Run with alpine (debug shell available)
make docker-run-alpine
```

### Quick Test

```bash
# Health check
curl http://localhost:8080/health

# Simple validation
curl -X POST http://localhost:8080/validate \
  -H "Content-Type: application/json" \
  -d '{
    "model_type": "incident",
    "payload": {
      "id": "INC-20250106-0001",
      "title": "Critical payment processing bug requiring immediate attention",
      "description": "Payment gateway failing for all credit card transactions in production",
      "severity": "critical",
      "status": "open",
      "priority": 5,
      "category": "bug",
      "environment": "production",
      "reported_by": "ops@example.com",
      "reported_at": "2025-01-06T14:30:00Z"
    }
  }'
```

**Response**:
```json
{
  "is_valid": true,
  "model_type": "incident",
  "provider": "go-playground",
  "timestamp": "2025-01-06T14:30:05Z",
  "processing_duration": "5ms",
  "errors": [],
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

## API Endpoints

### Core Endpoints

#### 1. Health Check
```bash
GET /health
```

**Response**:
```json
{
  "status": "healthy",
  "version": "2.0.0-modular",
  "uptime": "2h34m12s",
  "server": "modular-validation-server"
}
```

#### 2. List Models
```bash
GET /models
```

**Response**:
```json
{
  "models": {
    "incident": {
      "name": "Incident Report",
      "description": "Incident reporting payload validation",
      "endpoint": "/validate/incident",
      "version": "1.0.0"
    },
    "api": {
      "name": "API Request/Response",
      "endpoint": "/validate/api"
    }
  },
  "count": 6
}
```

#### 3. Generic Validation (Single Record)
```bash
POST /validate
Content-Type: application/json

{
  "model_type": "incident",
  "payload": {
    "id": "INC-20250106-0001",
    "title": "Critical bug in payment system",
    ...
  }
}
```

**Test with sample data**:
```bash
curl -X POST http://localhost:8080/validate \
  -H "Content-Type: application/json" \
  -d @test_data/valid/incident_valid.json
```

**Response (Valid)**:
```json
{
  "is_valid": true,
  "model_type": "incident",
  "provider": "go-playground",
  "errors": [],
  "warnings": []
}
```

**Response (Invalid)**:
```json
{
  "is_valid": false,
  "model_type": "incident",
  "errors": [
    {
      "field": "id",
      "message": "incident ID must follow format INC-YYYYMMDD-NNNN (e.g., INC-20240924-0001), got: INC-123",
      "code": "INVALID_ID_FORMAT",
      "value": "INC-123"
    },
    {
      "field": "title",
      "message": "Field must be at least 10 characters long",
      "code": "VALIDATION_FAILED",
      "value": "Bug"
    }
  ]
}
```

### Array Validation (Batch - Single Request)

#### 4. Array Validation Without Threshold
```bash
POST /validate
{
  "model_type": "incident",
  "data": [
    {"id": "INC-20250106-0001", ...},
    {"id": "INC-20250106-0002", ...},
    {"id": "INC-20250106-0003", ...}
  ]
}
```

**Test**:
```bash
curl -X POST http://localhost:8080/validate \
  -H "Content-Type: application/json" \
  -d @test_data/arrays/incident_array.json
```

**Response** (HTTP 200):
```json
{
  "status": "success",
  "total_records": 3,
  "valid_records": 2,
  "invalid_records": 1,
  "results": [
    {
      "row_index": 2,
      "record_identifier": "INC-20250106-0003",
      "is_valid": false,
      "errors": [...]
    }
  ]
}
```

#### 5. Array Validation With Threshold (Quality Gate)
```bash
POST /validate
{
  "model_type": "incident",
  "threshold": 80.0,
  "data": [...]
}
```

**Test**:
```bash
curl -X POST http://localhost:8080/validate \
  -H "Content-Type: application/json" \
  -d @test_data/arrays/threshold/incident_success_80.json
```

**Response (Success - 100% >= 80%)** (HTTP 200):
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

**Response (Failed - 60% < 80%)** (HTTP 422):
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

### Batch Processing (Multi-Request Sessions)

#### 6. Start Batch Session
```bash
POST /validate/batch/start
{
  "model_type": "incident",
  "job_id": "import-2025-01",
  "threshold": 95.0
}
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

#### 7. Add Data to Batch (Chunked Upload)
```bash
POST /validate
X-Batch-ID: batch-incident-1704537600-import-2025-01
{
  "model_type": "incident",
  "data": [...]
}
```

**Response**:
```json
{
  "chunk_processed": true,
  "valid_records": 98,
  "invalid_records": 2
}
```

#### 8. Check Batch Status
```bash
GET /validate/batch/{batch_id}
```

**Response**:
```json
{
  "batch_id": "batch-incident-1704537600-import-2025-01",
  "total_records": 150,
  "valid_records": 145,
  "invalid_records": 5,
  "threshold": 95.0,
  "is_final": false
}
```

#### 9. Complete Batch
```bash
POST /validate/batch/{batch_id}/complete
```

**Response (Success)**:
```json
{
  "batch_id": "batch-incident-1704537600-import-2025-01",
  "status": "success",
  "total_records": 150,
  "valid_records": 145,
  "invalid_records": 5,
  "threshold": 95.0,
  "message": "Batch validation completed with status: success"
}
```

### Model-Specific Endpoints (Auto-Generated)

```bash
# Each registered model gets its own endpoint
POST /validate/incident      # Incident validation
POST /validate/api           # API request/response validation
POST /validate/github        # GitHub webhook validation
POST /validate/database      # Database query validation
POST /validate/deployment    # Deployment validation
```

**Example**:
```bash
curl -X POST http://localhost:8080/validate/incident \
  -H "Content-Type: application/json" \
  -d '{
    "id": "INC-20250106-0001",
    "title": "Critical bug",
    ...
  }'
```

### Swagger Documentation

```bash
GET /swagger/              # Swagger UI
GET /swagger/doc.json      # OpenAPI spec
GET /swagger/models        # Dynamic model schemas
```

Open in browser: `http://localhost:8080/swagger/`

---

## Testing

### Unit Tests

```bash
# Run all unit tests
go test ./src/... -v

# Run with coverage
go test ./src/... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run specific package
go test ./src/validations/... -v

# Run specific test
go test ./src/validations/... -run TestIncidentValidator -v
```

**Test files**:
- `src/models/*_test.go` - Model structure tests
- `src/validations/*_test.go` - Validation logic tests
- `src/registry/*_test.go` - Registry system tests

### End-to-End Tests

```bash
# Run complete E2E test suite (35 tests)
./e2e_test_suite.sh

# Run with custom port
PORT=8086 ./e2e_test_suite.sh
```

**Test coverage**:
- Phase 0: Unit tests
- Phase 1-8: Core validation functionality
- Phase 9: Array validation (8 tests)
- Phase 10: Threshold validation with test data files (5 tests)
- Phase 11: HTTP method testing

**See**: [E2E_TEST_GUIDE.md](E2E_TEST_GUIDE.md) for detailed test documentation.

### Testing with Sample Data

```bash
# Test data directory structure
test_data/
├── valid/
│   ├── incident_valid.json
│   ├── api_valid.json
│   └── github_valid.json
├── invalid/
│   ├── incident_invalid.json
│   └── api_invalid.json
└── arrays/
    ├── incident_array.json
    └── threshold/
        ├── incident_success_80.json
        └── incident_failure_80.json

# Test with valid data
curl -X POST http://localhost:8080/validate \
  -H "Content-Type: application/json" \
  -d @test_data/valid/incident_valid.json

# Test with invalid data
curl -X POST http://localhost:8080/validate \
  -H "Content-Type: application/json" \
  -d @test_data/invalid/incident_invalid.json

# Test array validation
curl -X POST http://localhost:8080/validate \
  -H "Content-Type: application/json" \
  -d @test_data/arrays/incident_array.json

# Test threshold validation
curl -X POST http://localhost:8080/validate \
  -H "Content-Type: application/json" \
  -d @test_data/arrays/threshold/incident_success_80.json
```

### Docker Testing

```bash
# Build and run tests in Docker
make docker-build
make docker-run

# Run E2E tests against Docker container
./e2e_test_suite.sh
```

---

## Error Codes & Best Practices

### Standard Error Codes

| Code | Description | Example |
|------|-------------|---------|
| `VALIDATION_FAILED` | Generic validation failure | Field doesn't meet requirements |
| `REQUIRED_FIELD_MISSING` | Required field is empty | `"title": ""` |
| `VALUE_TOO_SHORT` | String/number below minimum | `"title": "Bug"` (min: 10 chars) |
| `VALUE_TOO_LONG` | String/number exceeds maximum | Title > 200 characters |
| `INVALID_FORMAT` | Format doesn't match pattern | Invalid email/URL/IP |
| `INVALID_ENUM_VALUE` | Value not in allowed list | `"severity": "urgent"` (allowed: low/medium/high/critical) |
| `INVALID_ID_FORMAT` | Custom ID format check | `"id": "INC-123"` (expected: INC-YYYYMMDD-NNNN) |
| `PRIORITY_SEVERITY_MISMATCH` | Business logic violation | Priority 1 with severity "critical" |

### HTTP Status Codes

| Code | Meaning | When |
|------|---------|------|
| `200 OK` | Validation completed | Single record or array (check `is_valid` field) |
| `200 OK` | Threshold met | Array validation with threshold passed |
| `400 Bad Request` | Invalid request | Missing `model_type`, malformed JSON |
| `404 Not Found` | Resource not found | Unknown model type or batch ID |
| `405 Method Not Allowed` | Wrong HTTP method | GET on POST endpoint |
| `422 Unprocessable Entity` | Threshold not met | Array validation failed threshold check |
| `500 Internal Server Error` | Server error | Unexpected server failure |

### Best Practices

1. **Always check `is_valid` field** in response (HTTP 200 doesn't mean valid data)
2. **Use threshold validation** for data imports/batch processing (quality gates)
3. **Review warnings** even when `is_valid: true` (best practice suggestions)
4. **Use batch sessions** for large datasets (>10,000 records)
5. **Leverage test data** in `test_data/` for integration testing
6. **Monitor processing_duration** - validation should be <15ms typically
7. **Handle 422 status** when using thresholds (failed quality gate)

### Performance Metrics

Every validation response includes:
```json
{
  "processing_duration": "5ms",
  "performance_metrics": {
    "validation_duration": "5ms",
    "field_count": 10,
    "rule_count": 25,
    "memory_usage": 1024
  }
}
```

Warnings added if validation > 100ms (configurable in `src/config/constants.go`).

---

## Documentation

Comprehensive guides are available in separate markdown files:

### For Users

- **[ADDING_NEW_MODELS_GUIDE.md](ADDING_NEW_MODELS_GUIDE.md)** - Step-by-step guide to add new validation models
  - Creating model structs with validation tags
  - Writing validators with custom business logic
  - Registering models in the auto-discovery system
  - Testing new models

- **[E2E_TEST_GUIDE.md](E2E_TEST_GUIDE.md)** - Complete end-to-end testing guide
  - Running the E2E test suite (35 tests across 11 phases)
  - Understanding test phases and coverage
  - Writing new E2E tests
  - Debugging test failures

- **[THRESHOLD_VALIDATION_SUMMARY.md](THRESHOLD_VALIDATION_SUMMARY.md)** - Threshold validation deep-dive
  - How threshold validation works (success rate calculation)
  - Array validation with/without thresholds
  - Test data files and examples
  - Use cases (data import, CI/CD gates, batch processing)

### For Developers

- **[CODE_EXECUTION_FLOW_GUIDE.md](CODE_EXECUTION_FLOW_GUIDE.md)** - Complete code execution flow (CRITICAL for new developers)
  - System architecture and startup flow
  - Auto-discovery mechanism (AST parsing, reflection)
  - Request processing pipeline (single, array, batch)
  - Validation flow (struct tags → custom rules → warnings)
  - Array validation detailed flow
  - Batch processing multi-request sessions
  - Adding new models (with code examples)
  - Key interfaces and data structures

- **[UNIT_TESTING_GUIDE.md](UNIT_TESTING_GUIDE.md)** - Unit testing best practices
  - Writing tests for validators
  - Testing model structures
  - Mocking dependencies
  - Coverage requirements

- **[CONCURRENT_VALIDATION.md](CONCURRENT_VALIDATION.md)** - Concurrency and thread safety
  - Batch session management (mutex locks)
  - Concurrent request handling
  - Performance optimization techniques

---

## Project Structure

```
.
├── src/
│   ├── main.go                      # Entry point, HTTP setup
│   ├── models/                      # Data models with validation tags
│   │   ├── incident.go              # IncidentPayload struct
│   │   ├── api.go                   # APIRequest struct
│   │   ├── github.go                # GitHubPayload struct
│   │   ├── database.go              # DatabaseQuery struct
│   │   ├── deployment.go            # DeploymentPayload struct
│   │   ├── generic.go               # GenericPayload struct
│   │   └── validation_result.go     # Result types, BatchSession
│   │
│   ├── validations/                 # Validation logic
│   │   ├── base_validator.go        # Shared validation framework
│   │   ├── incident.go              # Incident validator with custom rules
│   │   ├── api.go                   # API validator
│   │   ├── github.go                # GitHub webhook validator
│   │   ├── database.go              # Database query validator
│   │   ├── deployment.go            # Deployment validator
│   │   └── generic.go               # Generic validator
│   │
│   ├── registry/                    # Auto-discovery system
│   │   ├── model_registry.go        # Core types and interfaces
│   │   ├── unified_registry.go      # Auto-discovery engine (AST + reflection)
│   │   └── dynamic_registry.go      # Runtime utilities
│   │
│   └── config/
│       └── constants.go             # Error codes, thresholds
│
├── test_data/                       # Sample validation payloads
│   ├── valid/                       # Valid test cases
│   ├── invalid/                     # Invalid test cases
│   └── arrays/                      # Array validation tests
│       └── threshold/               # Threshold validation tests
│
├── e2e_test_suite.sh               # End-to-end test suite (35 tests)
├── Makefile                         # Build and deployment commands
├── Dockerfile                       # Multi-stage Docker build
├── docker-compose.yml              # Docker Compose setup
├── go.mod                          # Go module dependencies
├── go.sum                          # Dependency checksums
│
└── Documentation/
    ├── README.md                    # This file
    ├── CODE_EXECUTION_FLOW_GUIDE.md # Complete code flow (for developers)
    ├── ADDING_NEW_MODELS_GUIDE.md   # Adding new models (for users)
    ├── E2E_TEST_GUIDE.md            # E2E testing guide
    ├── THRESHOLD_VALIDATION_SUMMARY.md
    ├── UNIT_TESTING_GUIDE.md
    └── CONCURRENT_VALIDATION.md
```

---

## Quick Command Reference

```bash
# Build
go build -o bin/validator src/main.go

# Run
PORT=8080 ./bin/validator

# Test
go test ./src/... -v                # Unit tests
./e2e_test_suite.sh                # E2E tests

# Docker
make docker-build                   # Build image
make docker-run                     # Run distroless (12MB)
make docker-run-alpine             # Run alpine (with shell)

# Development
go mod tidy                        # Update dependencies
go test ./src/... -coverprofile=coverage.out
go tool cover -html=coverage.out   # View coverage

# API Testing
curl http://localhost:8080/health                      # Health check
curl http://localhost:8080/models                      # List models
curl -X POST http://localhost:8080/validate -d @test_data/valid/incident_valid.json
```

---

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP server port |
| `SERVER_MODE` | `modular` | Server mode (always modular, legacy deprecated) |

---

## Contributing

1. Read [CODE_EXECUTION_FLOW_GUIDE.md](CODE_EXECUTION_FLOW_GUIDE.md) to understand the architecture
2. Follow [ADDING_NEW_MODELS_GUIDE.md](ADDING_NEW_MODELS_GUIDE.md) to add new models
3. Write unit tests (see [UNIT_TESTING_GUIDE.md](UNIT_TESTING_GUIDE.md))
4. Run E2E tests: `./e2e_test_suite.sh`
5. Ensure all tests pass before submitting PR

---

## License

MIT License - See [LICENSE](LICENSE) file for details.

---

## Support

- **Documentation**: See markdown files in root directory
- **Issues**: [GitHub Issues](https://github.com/your-org/go-playground-validator-pilot/issues)
- **Code Flow**: [CODE_EXECUTION_FLOW_GUIDE.md](CODE_EXECUTION_FLOW_GUIDE.md) (start here for new developers)

---

**Built with ❤️ using [go-playground/validator](https://github.com/go-playground/validator)**
