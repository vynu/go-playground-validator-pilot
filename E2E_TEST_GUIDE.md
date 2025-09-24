# Complete End-to-End Testing Guide

## Overview

This project implements a comprehensive End-to-End (E2E) testing framework for the **modular Go Playground Validator multi-platform validation system**. The E2E tests validate all HTTP endpoints, registry-based validation capabilities, platform-specific validators, and Swagger integration across the modular validation server architecture.

## Test Architecture

### Test Structure

```
├── src/
│   ├── comprehensive_e2e_test.go      # Main E2E test suite
│   ├── main_test.go                   # Original test suite
│   └── utils.go                       # Utility functions
├── test_data/                         # Test data files
│   ├── sample_pull_request.json       # GitHub webhook payload
│   ├── gitlab_payload.json            # GitLab merge request payload
│   ├── bitbucket_payload.json         # Bitbucket pull request payload
│   ├── generic_json_payload.json      # Generic JSON validation data
│   ├── api_model_payload.json         # API model test data
│   ├── database_model_payload.json    # Database model test data
│   ├── validation_profiles.json       # Validation profile configurations
│   └── batch_payloads.json            # Batch validation test data
├── test_runner.sh                     # Comprehensive test runner script
├── curl_tests.sh                      # HTTP endpoint testing script
└── test_logs/                         # Test execution logs
```

### Test Categories

The E2E test suite covers **7 major test categories** for the modular server architecture:

1. **Modular Server Endpoints** - Tests platform-specific validation endpoints
2. **Registry-Based Validation** - Tests model registry and validation orchestration
3. **Multi-Platform Validation** - Tests GitHub, GitLab, Bitbucket, and Slack validation
4. **Swagger Integration** - Tests API documentation and dynamic model schemas
5. **Generic Validation** - Tests flexible validation with model type specification
6. **Error Handling** - Tests edge cases and validation error scenarios
7. **Performance & Concurrency** - Tests performance and concurrent request handling

## Running the Tests

### Method 1: Using the Test Runner Script (Recommended)

The `test_runner.sh` script provides the most comprehensive testing experience:

```bash
# Run all tests with coverage
./test_runner.sh

# Run only unit tests
./test_runner.sh --unit-only

# Run only E2E tests
./test_runner.sh --e2e-only

# Quick mode (skip performance tests)
./test_runner.sh --quick

# Verbose output
./test_runner.sh --verbose

# Skip coverage generation
./test_runner.sh --no-coverage
```

**Script Features:**
- Automated project structure verification
- Go dependency validation and tidying
- Build verification
- Comprehensive test execution with patterns
- Coverage report generation (HTML + console)
- Detailed logging with timestamps
- Error recovery and cleanup

### Method 2: Direct Go Test Commands

```bash
# Change to source directory
cd src

# Run all comprehensive E2E tests
go test -v -run "^TestComprehensive"

# Run specific test categories
go test -v -run "^TestComprehensive_OriginalServerEndpoints"
go test -v -run "^TestComprehensive_FlexibleServerEndpoints"
go test -v -run "^TestComprehensive_MultiModelValidation"
go test -v -run "^TestComprehensive_ValidationProfiles"
go test -v -run "^TestComprehensive_ProviderComparison"
go test -v -run "^TestComprehensive_ErrorHandling"
go test -v -run "^TestComprehensive_CORS"
go test -v -run "^TestComprehensive_Performance"
go test -v -run "^TestComprehensive_ConcurrentRequests"

# Run with coverage
go test -v -coverprofile=coverage.out -run "^TestComprehensive"
go tool cover -html=coverage.out -o coverage.html

# Run specific test with timeout
go test -v -timeout=30s -run "^TestComprehensive_Performance"
```

### Method 3: HTTP Endpoint Testing with cURL

```bash
# Test individual endpoints using cURL
./curl_tests.sh

# Test specific server
./curl_tests.sh original    # Test original server only
./curl_tests.sh flexible    # Test flexible server only
```

### Method 4: Individual Test Components

```bash
# Run only unit tests
go test -v -run "^Test[^E2E|^TestComprehensive]"

# Run original E2E tests
go test -v -run "^TestE2E"

# Run performance tests only
go test -v -run "Performance" -timeout=5m
```

## Test Results and Output

### Console Output Format

The tests provide detailed console output with several information levels:

#### 1. Test Execution Summary
```
=== RUN   TestComprehensive_OriginalServerEndpoints
=== RUN   TestComprehensive_OriginalServerEndpoints/OriginalServer_AllEndpoints
--- PASS: TestComprehensive_OriginalServerEndpoints (0.01s)
    --- PASS: TestComprehensive_OriginalServerEndpoints/OriginalServer_AllEndpoints (0.00s)
```

#### 2. HTTP Request Logging
```
2025/09/21 12:29:00 GET /health 68.292µs - Request ID: comprehensive-test-1758475740010476000
2025/09/21 12:29:00 POST /validate/github 667.333µs - Request ID: comprehensive-test-1758475740011757000
```

#### 3. Performance Metrics
```
comprehensive_e2e_test.go:757: Endpoint /health responded in 259.542µs
comprehensive_e2e_test.go:757: Endpoint /metrics responded in 113µs
```

#### 4. Test Validation Results
```
✅ Original E2E Tests: PASSED
✅ Comprehensive E2E Tests: PASSED
✅ Multi-Model Validation: PASSED
✅ Validation Profiles: PASSED
```

### Log Files

Test execution generates several log files in the `test_logs/` directory:

#### Test Runner Logs
- **Location**: `test_logs/test_run_YYYYMMDD_HHMMSS.log`
- **Content**: Complete test execution transcript
- **Format**: Timestamped entries with color coding

```bash
# View latest test log
ls -la test_logs/
tail -f test_logs/test_run_20250921_122900.log
```

#### Coverage Reports
- **Location**: `coverage/`
- **Files**:
  - `unit_coverage.out` - Unit test coverage
  - `e2e_original_coverage.out` - Original E2E test coverage
  - `e2e_comprehensive_coverage.out` - Comprehensive E2E test coverage
  - `combined_coverage.out` - Merged coverage data
  - `coverage.html` - Interactive HTML coverage report

```bash
# Open coverage report in browser
open coverage/coverage.html

# View coverage summary
go tool cover -func=coverage/combined_coverage.out
```

### Test Data Validation Results

#### Multi-Model Validation Results
Each model type test provides specific validation feedback:

```go
// GitHub Payload Validation
✅ GitHub webhook payload validation - PASSED
   - Pull request structure validated
   - Repository information verified
   - User data validation successful
   - Business logic validation applied

// GitLab Payload Validation
✅ GitLab merge request payload validation - PASSED
   - Merge request structure validated
   - Project information verified
   - Author data validation successful

// Database Model Validation
✅ Database model payload validation - PASSED
   - SQL operation structure validated
   - Connection info verified
   - Record data validation successful
   - Constraint validation applied
```

#### Validation Profile Results
```go
// Strict Profile
✅ Strict validation profile - PASSED
   - All validation rules enforced
   - No warnings or errors allowed
   - Complete field validation

// Permissive Profile
✅ Permissive validation profile - PASSED
   - Flexible rule interpretation
   - Warnings allowed but tracked
   - Essential field validation

// Minimal Profile
✅ Minimal validation profile - PASSED
   - Basic structure validation only
   - Maximum flexibility
   - Core safety checks only
```

## Understanding Test Components

### 1. Modular Server Endpoints (8 core endpoints)

Tests the modular validation server functionality:

```go
endpoints := map[string]TestEndpoint{
    "/health":                    {Method: "GET", ExpectedStatus: 200},
    "/validate/github":           {Method: "POST", ExpectedStatus: 200, Body: githubPayload},
    "/validate/gitlab":           {Method: "POST", ExpectedStatus: 200, Body: gitlabPayload},
    "/validate/bitbucket":        {Method: "POST", ExpectedStatus: 200, Body: bitbucketPayload},
    "/validate/slack":            {Method: "POST", ExpectedStatus: 200, Body: slackPayload},
    "/validate":                  {Method: "POST", ExpectedStatus: 200, Body: genericPayload}, // Generic validation
    "/models":                    {Method: "GET", ExpectedStatus: 200},
    "/swagger/":                  {Method: "GET", ExpectedStatus: 200}, // Swagger UI
}
```

### 2. Registry-Based Validation

Tests the validation registry and model management system:

```go
// Registry validation tests:
// - Model type resolution
// - Validator instance management
// - Registry lookup and caching
// - Cross-model validation consistency
// - Error handling and fallbacks

registryTests := []RegistryTestCase{
    {ModelType: "github", Validator: "GitHubValidator", Expected: "pass"},
    {ModelType: "gitlab", Validator: "GitLabValidator", Expected: "pass"},
    {ModelType: "bitbucket", Validator: "BitbucketValidator", Expected: "pass"},
    {ModelType: "slack", Validator: "SlackValidator", Expected: "pass"},
}
```

### 3. Multi-Platform Validation (4 primary platforms)

Tests validation across different platform webhook types:

| Platform | Description | Test Data File | Validator |
|----------|-------------|----------------|-----------|
| **GitHub** | GitHub webhook payloads | `sample_pull_request.json` | `GitHubValidator` |
| **GitLab** | GitLab merge request payloads | `gitlab_payload.json` | `GitLabValidator` |
| **Bitbucket** | Bitbucket pull request payloads | `bitbucket_payload.json` | `BitbucketValidator` |
| **Slack** | Slack webhook payloads | `slack_payload.json` | `SlackValidator` |

### 4. Swagger Integration

Tests API documentation and dynamic schema generation:

```go
swaggerEndpoints := map[string]TestEndpoint{
    "/swagger/":                  {Method: "GET", ExpectedStatus: 200}, // Swagger UI
    "/swagger/doc.json":          {Method: "GET", ExpectedStatus: 200}, // OpenAPI spec
    "/swagger/models":            {Method: "GET", ExpectedStatus: 200}, // Dynamic model schemas
}

// Validates:
// - OpenAPI specification generation
// - Dynamic model schema creation
// - API documentation completeness
// - Interactive Swagger UI functionality
```

### 5. Generic Validation

Tests flexible validation with explicit model type specification:

```go
genericValidationCases := []GenericTestCase{
    {
        ModelType: "github",
        Payload:   githubPayload,
        Expected:  ValidationResult{IsValid: true},
    },
    {
        ModelType: "gitlab",
        Payload:   gitlabPayload,
        Expected:  ValidationResult{IsValid: true},
    },
    {
        ModelType: "invalid",
        Payload:   anyPayload,
        Expected:  ValidationResult{IsValid: false},
    },
}

// Tests POST /validate endpoint with model_type in request body:
// {
//   "model_type": "github",
//   "payload": { /* actual webhook data */ }
// }
```

### 6. Error Handling and Edge Cases

Tests system behavior under various error conditions:

```go
errorScenarios := []ErrorTestCase{
    {Name: "Invalid JSON", Payload: `{"invalid": json}`, ExpectedStatus: 400},
    {Name: "Missing Model Type", Payload: validJSON, ExpectedStatus: 400},
    {Name: "Unknown Model Type", ModelType: "unknown", ExpectedStatus: 400},
    {Name: "Validation Failures", Payload: invalidData, ExpectedStatus: 422},
    {Name: "Server Errors", Trigger: "internal-error", ExpectedStatus: 500},
}

// Validates:
// - Proper HTTP status codes for different error types
// - Structured error response formats
// - Error message clarity and usefulness
// - System stability under error conditions
```

## Performance Testing Details

### Performance Metrics Collected

1. **Response Time Measurements**
   - Individual endpoint response times
   - P99 latency tracking
   - Performance regression detection

2. **Throughput Testing**
   - Concurrent request handling
   - Load balancing verification
   - Resource utilization monitoring

3. **Memory Usage**
   - Memory allocation tracking
   - Garbage collection impact
   - Memory leak detection

### Performance Test Results Format

```go
// Example performance output
comprehensive_e2e_test.go:757: Endpoint /health responded in 259.542µs
comprehensive_e2e_test.go:757: Endpoint /metrics responded in 113µs
comprehensive_e2e_test.go:757: Endpoint /models responded in 108µs
```

### Concurrent Request Testing

Tests system behavior under concurrent load:

```go
const numRequests = 10
// Spawns 10 concurrent requests to both servers
// Validates:
// - No race conditions
// - Consistent response times
// - Error-free concurrent processing
// - Resource cleanup
```

## Error Handling and Diagnostics

### Common Test Failures

#### 1. Connection Refused
```
Error: dial tcp [::1]:8080: connect: connection refused
Solution: Ensure servers are running before test execution
```

#### 2. Validation Configuration Warnings
```
Warning: Failed to load configuration: json: cannot unmarshal...
Impact: Non-blocking, tests continue with default configuration
```

#### 3. Test Data File Missing
```
Error: open test_data/sample_pull_request.json: no such file or directory
Solution: Verify test_data directory exists and contains required files
```

#### 4. Test Timeout
```
Error: test timed out after 30s
Solution: Increase timeout or investigate performance issues
```

### Debugging Test Issues

#### Enable Verbose Logging
```bash
go test -v -run "^TestComprehensive" 2>&1 | tee debug.log
```

#### Check Server Logs
```bash
# Original server logs
curl -s http://localhost:8080/health | jq

# Flexible server logs
curl -s http://localhost:8081/health | jq
```

#### Validate Test Data
```bash
# Verify JSON test data is valid
jq . test_data/sample_pull_request.json
jq . test_data/validation_profiles.json
```

#### Check Resource Usage
```bash
# Monitor during test execution
top -pid $(pgrep -f "go test")
netstat -an | grep ":808[01]"
```

## Integration with CI/CD

### GitHub Actions Integration

```yaml
name: E2E Tests
on: [push, pull_request]

jobs:
  e2e-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
        with:
          go-version: '1.25'

      - name: Run E2E Tests
        run: |
          chmod +x test_runner.sh
          ./test_runner.sh --verbose

      - name: Upload Coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage/combined_coverage.out

      - name: Archive Test Logs
        uses: actions/upload-artifact@v3
        if: always()
        with:
          name: test-logs
          path: test_logs/
```

### Jenkins Pipeline Integration

```groovy
pipeline {
    agent any
    stages {
        stage('E2E Tests') {
            steps {
                sh './test_runner.sh --verbose'
            }
            post {
                always {
                    publishHTML([
                        allowMissing: false,
                        alwaysLinkToLastBuild: true,
                        keepAll: true,
                        reportDir: 'coverage',
                        reportFiles: 'coverage.html',
                        reportName: 'Coverage Report'
                    ])
                    archiveArtifacts artifacts: 'test_logs/**/*', allowEmptyArchive: true
                }
            }
        }
    }
}
```

## Best Practices

### 1. Test Data Management

- **Keep test data realistic** - Use production-like payloads
- **Version test data** - Track changes to test datasets
- **Validate test data** - Ensure JSON files are well-formed
- **Document test scenarios** - Explain what each test data file validates

### 2. Test Execution

- **Run tests in isolation** - Each test should be independent
- **Clean up resources** - Ensure tests don't leave hanging connections
- **Use timeouts** - Prevent tests from hanging indefinitely
- **Parallel execution** - Run independent tests concurrently when possible

### 3. Error Handling

- **Graceful failure** - Tests should fail with clear error messages
- **Retry logic** - Implement retries for flaky network operations
- **Resource cleanup** - Always clean up even when tests fail
- **Detailed logging** - Log enough information to debug failures

### 4. Performance Testing

- **Establish baselines** - Track performance over time
- **Set realistic limits** - Use production-like performance expectations
- **Monitor resources** - Track CPU, memory, and network usage
- **Test under load** - Validate behavior under concurrent requests

## Troubleshooting Guide

### Test Execution Issues

| Issue | Symptoms | Solution |
|-------|----------|----------|
| **Port conflicts** | Connection refused errors | Check if ports 8080/8081 are available |
| **Permission denied** | Script execution fails | `chmod +x test_runner.sh curl_tests.sh` |
| **Missing dependencies** | Import errors | `go mod tidy && go mod verify` |
| **Test data corruption** | JSON parse errors | Validate JSON files with `jq` |
| **Memory issues** | Test timeouts/crashes | Increase test timeout, check system resources |
| **Network timeouts** | HTTP request failures | Check network connectivity, firewall settings |

### Performance Issues

| Issue | Symptoms | Solution |
|-------|----------|----------|
| **Slow response times** | High latency measurements | Profile application, optimize bottlenecks |
| **Failed concurrent tests** | Race condition errors | Review thread safety, add synchronization |
| **Memory leaks** | Increasing memory usage | Profile memory allocation, fix leaks |
| **Resource exhaustion** | System becomes unresponsive | Limit concurrent operations, add resource limits |

## Conclusion

The E2E testing framework provides comprehensive validation of the entire multi-model validation system. By following this guide, you can:

- **Execute complete test suites** covering all functionality
- **Understand test results** and performance metrics
- **Debug test failures** efficiently
- **Integrate testing** into CI/CD pipelines
- **Maintain test quality** over time

The testing framework ensures that all validation endpoints, multi-model capabilities, and business logic work correctly across different validation profiles and providers, giving confidence in the system's reliability and performance.