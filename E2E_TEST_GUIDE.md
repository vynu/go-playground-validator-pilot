# E2E Test Guide

## Table of Contents
1. [Overview](#overview)
2. [Running E2E Tests](#running-e2e-tests)
3. [Test Data Structure](#test-data-structure)
4. [Test Phases](#test-phases)
5. [Docker E2E Tests](#docker-e2e-tests)
6. [Makefile Commands](#makefile-commands)
7. [Troubleshooting](#troubleshooting)

---

## Overview

The E2E (End-to-End) test suite validates the complete functionality of the Go Playground Data Validator system, including:
- ✅ Unit tests with coverage analysis
- ✅ Automatic model discovery
- ✅ Validation endpoints
- ✅ Array/batch validation
- ✅ Threshold-based validation
- ✅ Docker container testing
- ✅ Server lifecycle management

---

## Running E2E Tests

### Prerequisites

1. **Go environment** properly configured
2. **Build the binary** first
3. **Test data** available in `test_data/` directory
4. **Port 8086** available (or configure custom port)

### Quick Start

#### Option 1: Using Make (Recommended)
```bash
# Build and run E2E tests
make test-e2e

# This internally runs:
# 1. make build (builds binary to bin/validator)
# 2. PORT=8086 ./e2e_test_suite.sh
```

#### Option 2: Manual Execution
```bash
# 1. Build the validator binary
make build

# 2. Run E2E tests
./e2e_test_suite.sh

# 3. Or specify custom port
PORT=9090 ./e2e_test_suite.sh
```

### Expected Output

```
🧪 Starting Comprehensive E2E Test Suite
========================================

💻 Running in local test mode
✅ Process cleanup completed

🧪 Phase 0: Unit Testing Suite (Model-Agnostic Framework)
=========================================================
✅ Unit tests execution completed
ℹ️  Total unit test coverage: 84.6%
✅ Coverage exceeds minimum threshold (70%): 84.6%

🚀 Phase 1: Server Startup & Basic Health Checks
=================================================
✅ Server is ready on port 8086

[... continues through all phases ...]

🎉 Test Suite Complete!
📈 Test Results Summary:
Total Tests: 35
Passed: 35
Failed: 0

✅ ALL TESTS PASSED! 🎊
```

**Note**: Test count increased from 22 to 35 with the addition of Phase 10 (Threshold Validation with Test Data Files) which includes 5 comprehensive threshold tests and 8 new inline threshold tests from Phase 9.

---

## Test Data Structure

### Directory Layout

```
test_data/
├── single/               # Single record validation
│   ├── valid/           # Valid test payloads
│   │   ├── api.json
│   │   ├── incident.json
│   │   ├── github.json
│   │   ├── database.json
│   │   ├── deployment.json
│   │   └── generic.json
│   └── invalid/         # Invalid test payloads
│       ├── api.json
│       ├── incident.json
│       └── ...
├── arrays/              # Array validation
│   ├── valid/          # Valid arrays
│   │   ├── api.json
│   │   └── incident.json
│   └── mixed/          # Mixed valid/invalid
│       └── incident.json
└── batch/               # Batch with threshold
    ├── valid/
    └── mixed/
```

### Adding New Test Data

When adding a new model, create corresponding test files:

1. **Valid payload**: `test_data/single/valid/{model}.json`
```json
{
  "id": "MODEL-001",
  "name": "Test Entry",
  "description": "Valid test payload"
}
```

2. **Invalid payload**: `test_data/single/invalid/{model}.json`
```json
{
  "id": "INVALID",
  "name": "X"
}
```

3. **Array payload**: `test_data/arrays/valid/{model}.json`
```json
[
  { "id": "MODEL-001", "name": "First" },
  { "id": "MODEL-002", "name": "Second" }
]
```

The test suite will automatically discover and test these files!

---

## Test Phases

### Phase 0: Unit Testing Suite

**Purpose**: Run comprehensive unit tests before E2E testing

**What it tests**:
- All packages (models, validations, registry, main, config)
- Code coverage analysis
- Test failures detection

**Coverage Files Created**:
- `coverage/unit_coverage.out` - Coverage profile
- `coverage/unit_coverage_summary.txt` - Summary statistics
- `coverage/unit_test_output.log` - Test output

**Commands**:
```bash
cd src && go test -v -coverprofile=../coverage/unit_coverage.out ./...
go tool cover -func=../coverage/unit_coverage.out
```

### Phase 1: Server Startup

**Purpose**: Verify server can start and respond

**What it tests**:
- Server startup on configured port
- Health endpoint accessibility
- Server ready state

### Phase 2: Basic Endpoint Testing

**Purpose**: Test core API endpoints

**Endpoints tested**:
- `GET /health` → 200 OK
- `GET /models` → 200 OK
- `GET /swagger/models` → 200 OK
- `GET /swagger/` → 301 Redirect

### Phase 3: Automatic Model Discovery

**Purpose**: Verify automatic model registration

**What it tests**:
- Expected models are registered (github, incident, api, database, generic, deployment)
- Deleted models are NOT registered (bitbucket, gitlab, slack)

### Phase 4: Model Validation Testing

**Purpose**: Test validation with test_data payloads

**What it tests**:
- Reads test data from `test_data/single/valid/` and `test_data/single/invalid/`
- Validates payloads against registered models
- Reports validation results

**Example Request**:
```bash
curl -X POST http://localhost:8086/validate/incident \
  -H "Content-Type: application/json" \
  -d @test_data/single/valid/incident.json
```

### Phase 8: API Response Format

**Purpose**: Validate JSON response formats

**What it tests**:
- `/models` endpoint returns valid JSON
- `/swagger/models` endpoint returns valid JSON

### Phase 9: Array Validation

**Purpose**: Test batch/array validation functionality

**What it tests**:
- Array validation with multiple records
- Row-level validation results
- Summary statistics (success rate, error counts)
- Valid row filtering (only invalid/warning rows in results)
- Backward compatibility with single object validation

**Example Request**:
```bash
curl -X POST http://localhost:8086/validate \
  -H "Content-Type: application/json" \
  -d '{
    "model_type": "incident",
    "data": [
      { "id": "INC-20240115-0001", ... },
      { "id": "INC-20240115-0002", ... }
    ]
  }'
```

**Response Structure**:
```json
{
  "batch_id": "auto_abc123",
  "status": "success",
  "total_records": 2,
  "valid_records": 2,
  "invalid_records": 0,
  "warning_records": 0,
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

**Note**: `results` array only includes invalid or warning records. Valid records are excluded.

### Phase 10: Threshold Validation with Test Data Files

**Purpose**: Test threshold validation using test_data files for comprehensive scenario coverage

**What it tests**:
- Threshold validation with real test data files
- Success cases (meeting threshold requirements)
- Failure cases (below threshold requirements)
- No threshold behavior (always success for arrays)

**Test Data Location**: `test_data/arrays/threshold/`

**Test Cases**:

1. **Incident Success Case** (100% valid, 80% threshold)
   - File: `incident_success_80.json`
   - Expected: `status: "success"`, `threshold: 80`
   - 5 valid incident records

2. **Incident Failure Case** (60% valid, 80% threshold)
   - File: `incident_failure_80.json`
   - Expected: `status: "failed"`, `threshold: 80`
   - 3 valid + 2 invalid records (60% < 80%)

3. **API Success Case** (100% valid, 80% threshold)
   - File: `api_success_80.json`
   - Expected: `status: "success"`, `threshold: 80`
   - 4 valid API records

4. **API Failure Case** (50% valid, 80% threshold)
   - File: `api_failure_80.json`
   - Expected: `status: "failed"`, `threshold: 80`
   - 2 valid + 2 invalid records (50% < 80%)

5. **No Threshold Test** (mixed results, no threshold)
   - File: `incident_failure_80.json` (reused)
   - Expected: `status: "success"` (no threshold enforcement)
   - Always succeeds for multiple records without threshold

**Example Request**:
```bash
curl -X POST http://localhost:8086/validate \
  -H "Content-Type: application/json" \
  -d '{
    "model_type": "incident",
    "threshold": 80.0,
    "data": [...from test_data file...]
  }'
```

**Response Structure**:
```json
{
  "batch_id": "auto_xyz",
  "status": "success|failed",
  "total_records": 5,
  "valid_records": 5,
  "invalid_records": 0,
  "threshold": 80.0,
  "summary": {
    "success_rate": 100.0,
    "validation_errors": 0
  }
}
```

### Phase 11: HTTP Method Testing

**Purpose**: Verify incorrect HTTP methods return errors

**What it tests**:
- POST to /health → 405 Method Not Allowed
- GET to /validate → 405 Method Not Allowed

---

## Docker E2E Tests

### Using Makefile (Recommended)

#### Test Against Distroless Image
```bash
# Build Docker image and run E2E tests
make docker-test-e2e

# This internally:
# 1. Builds distroless Docker image
# 2. Starts container on port 8087
# 3. Runs E2E tests with TEST_MODE=docker
# 4. Cleans up container
```

#### Test Against Alpine Image
```bash
# Build Alpine image and run E2E tests
make docker-test-e2e-alpine
```

#### Test Using Docker Compose
```bash
# Use docker-compose stack for testing
make docker-test-compose
```

### Manual Docker Testing

```bash
# 1. Build Docker image
make docker-build

# 2. Start container
docker run -d --name validator-test -p 8087:8080 go-playground-validator:latest

# 3. Run E2E tests against container
VALIDATOR_URL=http://localhost:8087 TEST_MODE=docker ./e2e_test_suite.sh

# 4. Cleanup
docker stop validator-test && docker rm validator-test
```

### Docker Test Mode Differences

When `TEST_MODE=docker`:
- ✅ Skips unit tests (already tested in CI)
- ✅ Skips server startup (uses existing container)
- ✅ Tests against `VALIDATOR_URL` environment variable
- ✅ Skips server lifecycle tests (restart, deletion)
- ✅ Focuses on API endpoint validation

---

## Makefile Commands

### Build Commands
```bash
make build                  # Build binary for current platform
make build-linux           # Build for Linux
make build-all             # Build for all platforms
```

### Test Commands
```bash
make test                  # Run unit tests
make test-coverage         # Run tests with coverage report
make test-race             # Run tests with race detection
make test-e2e              # Run E2E test suite (build + e2e)
make test-all              # Run all tests (unit + race + e2e)
```

### Docker Test Commands
```bash
make docker-test-e2e           # E2E tests against distroless Docker
make docker-test-e2e-alpine    # E2E tests against Alpine Docker
make docker-test-compose       # E2E tests using docker-compose
```

### Coverage Commands
```bash
# Generate coverage report (outputs to root directory)
make test-coverage

# Files created:
# - coverage.out (coverage profile)
# - coverage.html (HTML report)

# View coverage
open coverage.html
```

### Clean Commands
```bash
make clean                 # Clean binaries and test artifacts
make clean-docker         # Clean Docker artifacts
make clean-all            # Clean everything
```

---

## Troubleshooting

### Common Issues

#### 1. Port Already in Use
**Error**: `Server failed to start after 30 seconds`

**Solution**:
```bash
# Using Makefile
make kill-port PORT=8086

# Or manually
lsof -ti :8086 | xargs kill -9
```

#### 2. Binary Not Found
**Error**: `./bin/validator: No such file or directory`

**Solution**:
```bash
make build
```

#### 3. Permission Denied
**Error**: `Permission denied: ./e2e_test_suite.sh`

**Solution**:
```bash
chmod +x e2e_test_suite.sh
```

#### 4. Docker Build Fails
**Error**: `undefined: models.IncidentPayload`

**Cause**: E2E tests deleted model files during testing

**Solution**:
```bash
# Restore model files
cd src && git checkout HEAD -- models/incident.go validations/incident.go
cd ..

# Rebuild Docker
make docker-build
```

#### 5. Coverage Files in Wrong Location
**Issue**: Coverage files created in `src/` directory

**Solution**: The Makefile and e2e_test_suite.sh have been updated to create coverage files in the root `coverage/` directory. If you still see files in `src/`, update your scripts:

```bash
# Ensure these are updated:
# - e2e_test_suite.sh (lines 430-485)
# - Makefile (test-coverage target)
```

### Manual Cleanup

If tests fail unexpectedly:

```bash
# 1. Kill all validator processes
pkill -f "./bin/validator"
pkill -f "validator"

# 2. Kill processes on test port
make kill-port PORT=8086

# 3. Remove test files
rm -f src/models/testmodel.go src/validations/testmodel.go

# 4. Clean build artifacts
make clean

# 5. Restore any backed up files
cd src && git status
git checkout -- models/incident.go validations/incident.go
```

### Debugging Test Failures

#### Enable Debug Mode
```bash
# Add at top of e2e_test_suite.sh
set -x  # Enable debug output
```

#### Check Logs
```bash
# View coverage test output
cat coverage/unit_test_output.log

# View coverage summary
cat coverage/unit_coverage_summary.txt
```

#### Run Specific Validation
```bash
# Test single endpoint manually
curl -X POST http://localhost:8086/validate/incident \
  -H "Content-Type: application/json" \
  -d @test_data/single/valid/incident.json
```

---

## Test Results Interpretation

### Success Output
```
🎉 Test Suite Complete!
📈 Test Results Summary:
Total Tests: 35
Passed: 35
Failed: 0

✅ ALL TESTS PASSED! 🎊
```

**Test Breakdown**:
- Phase 0: Unit Tests (1 test)
- Phases 1-8: Core functionality (17 tests)
- Phase 9: Array Validation (8 tests)
- Phase 10: Threshold Validation with Test Data (5 tests)
- Phase 11: HTTP Method Testing (2 tests)
- Additional: Model deletion/lifecycle (2 tests)

### Partial Failure Output
```
🎉 Test Suite Complete!
📈 Test Results Summary:
Total Tests: 35
Passed: 33
Failed: 2

❌ SOME TESTS FAILED ❌

Please review the failed tests above and fix any issues.
```

### Coverage Analysis
```
ℹ️  Total unit test coverage: 84.6%
✅ Coverage exceeds minimum threshold (70%): 84.6%

Package-level coverage breakdown:
  📦 ok  	goplayground-data-validator/models    coverage: 81.5% of statements
  📦 ok  	goplayground-data-validator/validations    coverage: 84.6% of statements
  📦 ok  	goplayground-data-validator/registry    coverage: 95.5% of statements
  📦 ok  	goplayground-data-validator    coverage: 79.6% of statements
  📦 ok  	goplayground-data-validator/config    coverage: 100.0% of statements
```

---

## CI/CD Integration

### GitHub Actions Example
```yaml
name: E2E Tests
on: [push, pull_request]

jobs:
  e2e-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Run E2E Tests
        run: make test-e2e

      - name: Docker E2E Tests
        run: make docker-test-e2e

      - name: Upload Coverage
        uses: actions/upload-artifact@v3
        with:
          name: coverage-report
          path: coverage/
```

### Jenkins Pipeline Example
```groovy
pipeline {
    agent any
    stages {
        stage('Build') {
            steps {
                sh 'make build'
            }
        }
        stage('E2E Tests') {
            steps {
                sh 'make test-e2e'
            }
        }
        stage('Docker E2E Tests') {
            steps {
                sh 'make docker-test-e2e'
            }
        }
    }
    post {
        always {
            publishHTML([
                reportDir: 'coverage',
                reportFiles: 'coverage.html',
                reportName: 'Coverage Report'
            ])
        }
    }
}
```

---

## Best Practices

### 1. Always Build Before Testing
```bash
make build && ./e2e_test_suite.sh
# Or use
make test-e2e  # Does both
```

### 2. Check Coverage Reports
```bash
make test-coverage
open coverage.html
```

### 3. Test Docker Builds
```bash
# Test both distroless and alpine
make docker-test-e2e
make docker-test-e2e-alpine
```

### 4. Clean Between Test Runs
```bash
make clean
make build
make test-e2e
```

### 5. Verify Port Availability
```bash
make check-port PORT=8086
```

---

## Summary

The E2E test suite provides comprehensive validation of the Go Playground Data Validator system:

✅ **Automated Testing**: One command (`make test-e2e`) runs complete test suite
✅ **Coverage Analysis**: Unit tests with 80%+ coverage requirement
✅ **Docker Testing**: Validates containerized deployment
✅ **Test Data Driven**: Easy to add new test cases via `test_data/` directory
✅ **CI/CD Ready**: Integrates with GitHub Actions, Jenkins, etc.
✅ **Clean Artifacts**: All outputs go to root `coverage/` and `bin/` directories

**Quick Commands Reference**:
```bash
make test-e2e              # Local E2E tests
make docker-test-e2e       # Docker E2E tests
make test-coverage         # Coverage report
make clean                 # Clean artifacts
```
