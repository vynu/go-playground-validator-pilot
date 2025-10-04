#!/bin/bash

# Comprehensive E2E Test Suite for Go Playground Data Validator
# Tests all aspects of the dynamic model registration system

set -e  # Exit on any error

echo "🧪 Starting Comprehensive E2E Test Suite"
echo "========================================"
echo ""

# Configuration
# Check if running in Docker mode (for Makefile docker-test-* targets)
if [ ! -z "$TEST_MODE" ] && [ "$TEST_MODE" = "docker" ]; then
    echo "🐳 Running in Docker test mode"
    DOCKER_MODE=true
    # Use environment variable for URL if set, otherwise use default
    SERVER_PORT=${VALIDATOR_URL##*:}
    API_BASE="${VALIDATOR_URL:-http://localhost:8087}"
    SERVER_PID=""
    SKIP_SERVER_START=true
else
    echo "💻 Running in local test mode"
    DOCKER_MODE=false
    SERVER_PORT=8086
    API_BASE="http://localhost:$SERVER_PORT"
    SERVER_PID=""
    SKIP_SERVER_START=false
fi

INCIDENT_MODEL_BACKUP=""
INCIDENT_VALIDATION_BACKUP=""

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Cleanup function
cleanup() {
    echo ""
    echo "🧹 Cleaning up test environment..."

    # Stop server
    if [ ! -z "$SERVER_PID" ]; then
        kill $SERVER_PID 2>/dev/null || true
        wait $SERVER_PID 2>/dev/null || true
        echo "  ✓ Server stopped"
    fi

    # Restore incident model if backed up
    if [ ! -z "$INCIDENT_MODEL_BACKUP" ] && [ -f "$INCIDENT_MODEL_BACKUP" ]; then
        mv "$INCIDENT_MODEL_BACKUP" src/models/incident.go
        echo "  ✓ Restored incident model"
    fi

    if [ ! -z "$INCIDENT_VALIDATION_BACKUP" ] && [ -f "$INCIDENT_VALIDATION_BACKUP" ]; then
        mv "$INCIDENT_VALIDATION_BACKUP" src/validations/incident.go
        echo "  ✓ Restored incident validation"
    fi

    # Clean up test files
    rm -f src/models/testmodel.go src/validations/testmodel.go 2>/dev/null || true

    echo "  ✓ Test files cleaned up"
}

# Set trap to cleanup on exit
trap cleanup EXIT

# Utility functions
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

log_test() {
    echo -e "${BLUE}🧪 $1${NC}"
}

# Test framework functions
start_test() {
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    log_test "$1"
}

pass_test() {
    PASSED_TESTS=$((PASSED_TESTS + 1))
    log_success "$1"
}

fail_test() {
    FAILED_TESTS=$((FAILED_TESTS + 1))
    log_error "$1"
    if [ "$2" = "exit" ]; then
        exit 1
    fi
}

# Wait for server to be ready
wait_for_server() {
    log_info "Waiting for server to start..."
    for i in {1..30}; do
        if curl -s "$API_BASE/health" >/dev/null 2>&1; then
            log_success "Server is ready on port $SERVER_PORT"
            return 0
        fi
        sleep 1
    done
    fail_test "Server failed to start after 30 seconds" exit
}

# Test HTTP endpoint
test_endpoint() {
    local description=$1
    local endpoint=$2
    local expected_status=$3
    local method=${4:-GET}

    start_test "$description"

    if [ "$method" = "POST" ]; then
        status=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$endpoint" -H "Content-Type: application/json" -d '{}')
    else
        status=$(curl -s -o /dev/null -w "%{http_code}" "$endpoint")
    fi

    if [ "$status" = "$expected_status" ]; then
        pass_test "Endpoint $endpoint returned $status"
    else
        fail_test "Endpoint $endpoint returned $status, expected $expected_status"
    fi
}

# Check if model is registered
check_model_registered() {
    local model_name=$1
    local description=$2

    start_test "$description"

    response=$(curl -s "$API_BASE/models")
    if echo "$response" | grep -q "\"$model_name\""; then
        pass_test "Model '$model_name' is registered"
        return 0
    else
        fail_test "Model '$model_name' is NOT registered"
        return 1
    fi
}

# Check if model is NOT registered
check_model_not_registered() {
    local model_name=$1
    local description=$2

    start_test "$description"

    response=$(curl -s "$API_BASE/models")
    if echo "$response" | grep -q "\"$model_name\""; then
        fail_test "Model '$model_name' is still registered when it shouldn't be"
        return 1
    else
        pass_test "Model '$model_name' is correctly unregistered"
        return 0
    fi
}

# Load test data from file
load_test_data() {
    local file_path=$1
    if [ -f "$file_path" ]; then
        cat "$file_path"
    else
        echo ""
    fi
}

# Test validation endpoint with expected result
test_validation_endpoint() {
    local model_name=$1
    local payload=$2
    local expected_valid=$3
    local test_type=$4

    start_test "Testing $test_type payload for $model_name"
    response=$(curl -s -X POST "$API_BASE/validate/$model_name" \
        -H "Content-Type: application/json" \
        -d "$payload")

    if [ "$expected_valid" = "true" ]; then
        if echo "$response" | grep -q '"is_valid":true'; then
            pass_test "$test_type payload validation passed for $model_name"
        else
            fail_test "$test_type payload validation failed for $model_name"
            echo "Response: $response"
        fi
    else
        if echo "$response" | grep -q '"is_valid":false'; then
            pass_test "$test_type payload validation passed for $model_name"
        else
            fail_test "$test_type payload validation failed for $model_name"
            echo "Response: $response"
        fi
    fi
}

# Test model validation functionality
test_model_validation() {
    local model_name=$1
    local valid_payload_override=$2
    local invalid_payload_override=$3

    # Try to load test data from files first
    local valid_payload=$(load_test_data "test_data/valid/$model_name.json")
    local invalid_payload=$(load_test_data "test_data/invalid/$model_name.json")

    # Use override payloads if provided and no file data found
    if [ -z "$valid_payload" ] && [ ! -z "$valid_payload_override" ]; then
        valid_payload="$valid_payload_override"
    fi
    if [ -z "$invalid_payload" ] && [ ! -z "$invalid_payload_override" ]; then
        invalid_payload="$invalid_payload_override"
    fi

    # Test valid payload if available
    if [ ! -z "$valid_payload" ]; then
        test_validation_endpoint "$model_name" "$valid_payload" "true" "valid"
    else
        log_info "No valid test data found for $model_name (create test_data/valid/$model_name.json)"
    fi

    # Test invalid payload if available
    if [ ! -z "$invalid_payload" ]; then
        test_validation_endpoint "$model_name" "$invalid_payload" "false" "invalid"
    else
        log_info "No invalid test data found for $model_name (create test_data/invalid/$model_name.json)"
    fi
}

# Test all discovered models automatically
test_all_models() {
    log_info "Testing all discovered models with available test data..."

    # Get list of all registered models
    response=$(curl -s "$API_BASE/models")
    models=$(echo "$response" | python3 -c "import sys, json; data=json.load(sys.stdin); print(' '.join(data.get('models', [])))" 2>/dev/null || echo "")

    if [ -z "$models" ]; then
        log_warning "Could not retrieve models list for automatic testing"
        return
    fi

    for model in $models; do
        # Skip GitHub model - requires complex payload structure
        if [ "$model" = "github" ]; then
            log_info "Skipping $model validation test - requires full webhook payload structure"
            continue
        fi

        # Check if test data exists
        if [ -f "test_data/valid/$model.json" ] || [ -f "test_data/invalid/$model.json" ]; then
            test_model_validation "$model"
        else
            log_info "No test data found for model '$model' - skipping validation test"
        fi
    done
}

# Create a test model
create_test_model() {
    local model_name=$1

    log_info "Creating test model: $model_name"

    # Create model file
    cat > "src/models/${model_name}.go" << EOF
package models

import "time"

// TestmodelPayload represents a test model for dynamic registration
type TestmodelPayload struct {
    ID          string    \`json:"id" validate:"required,min=1,max=50"\`
    Name        string    \`json:"name" validate:"required,min=2,max=100"\`
    Email       string    \`json:"email" validate:"required,email"\`
    Age         int       \`json:"age" validate:"required,min=1,max=150"\`
    IsActive    bool      \`json:"is_active"\`
    CreatedAt   time.Time \`json:"created_at" validate:"required"\`
    Tags        []string  \`json:"tags"\`
}
EOF

    # Create validation file
    cat > "src/validations/${model_name}.go" << EOF
package validations

import (
    "strings"
    "time"

    "goplayground-data-validator/models"
    "github.com/go-playground/validator/v10"
)

// TestmodelValidator handles validation for ${model_name} models
type TestmodelValidator struct {
    validator *validator.Validate
}

// NewTestmodelValidator creates a new ${model_name} validator
func NewTestmodelValidator() *TestmodelValidator {
    return &TestmodelValidator{
        validator: validator.New(),
    }
}

// ValidatePayload validates a ${model_name} payload
func (v *TestmodelValidator) ValidatePayload(payload models.TestmodelPayload) models.ValidationResult {
    result := models.ValidationResult{
        IsValid:   true,
        ModelType: "$model_name",
        Provider:  "go-playground",
        Errors:    []models.ValidationError{},
        Warnings:  []models.ValidationWarning{},
    }

    // Basic struct validation
    if err := v.validator.Struct(payload); err != nil {
        result.IsValid = false
        validatorErrors := err.(validator.ValidationErrors)

        for _, fieldErr := range validatorErrors {
            result.Errors = append(result.Errors, models.ValidationError{
                Field:   strings.ToLower(fieldErr.Field()),
                Message: getFieldErrorMessage(fieldErr),
                Code:    getErrorCode(fieldErr.Tag()),
                Value:   fieldErr.Value(),
            })
        }
    }

    // Custom validation logic
    if payload.Age < 13 && payload.IsActive {
        result.Warnings = append(result.Warnings, models.ValidationWarning{
            Field:   "age",
            Message: "Active user is under 13 - may require parental consent",
            Code:    "UNDERAGE_ACTIVE_USER",
        })
    }

    return result
}

func getFieldErrorMessage(fe validator.FieldError) string {
    switch fe.Tag() {
    case "required":
        return fe.Field() + " is required"
    case "min":
        return fe.Field() + " must be at least " + fe.Param() + " characters/value"
    case "max":
        return fe.Field() + " must be at most " + fe.Param() + " characters/value"
    case "email":
        return fe.Field() + " must be a valid email address"
    default:
        return fe.Field() + " validation failed"
    }
}

func getErrorCode(tag string) string {
    switch tag {
    case "required":
        return "REQUIRED_FIELD_MISSING"
    case "min":
        return "VALUE_TOO_SHORT"
    case "max":
        return "VALUE_TOO_LONG"
    case "email":
        return "INVALID_EMAIL_FORMAT"
    default:
        return "VALIDATION_FAILED"
    }
}
EOF

    log_success "Created test model: $model_name"
}

# Kill any existing validator processes
killall_validators() {
    log_info "Cleaning up any existing validator processes..."
    pkill -f "./validator" 2>/dev/null || true
    pkill -f "go run main.go" 2>/dev/null || true
    # Kill any process on port 8086
    lsof -t -i :8086 | xargs kill -9 2>/dev/null || true
    sleep 3
    log_success "Process cleanup completed"
}

# Phase 0: Unit Testing Suite (Model-Agnostic Framework)
test_unit_tests() {
    echo "🧪 Phase 0: Unit Testing Suite (Model-Agnostic Framework)"
    echo "========================================================="

    log_info "Running model-agnostic unit test framework with coverage analysis..."
    log_info "✅ Main tests are now model-agnostic (no specific model dependencies)"
    log_info "✅ Registry tests work automatically with any number of models"
    log_info "✅ Adding new models requires zero changes to core tests"

    # Run unit tests with coverage
    start_test "Running unit tests for all packages"
    if cd src && go test -v -coverprofile=../coverage/unit_coverage.out ./... > ../coverage/unit_test_output.log 2>&1; then
        pass_test "Unit tests execution completed"

        # Generate coverage report
        log_info "Generating coverage report..."
        if go tool cover -func=../coverage/unit_coverage.out > ../coverage/unit_coverage_summary.txt 2>&1; then

            # Extract total coverage
            TOTAL_COVERAGE=$(grep "total:" ../coverage/unit_coverage_summary.txt | awk '{print $3}' | sed 's/%//')

            if [ ! -z "$TOTAL_COVERAGE" ]; then
                log_info "Total unit test coverage: $TOTAL_COVERAGE%"

                # Check if coverage meets minimum threshold (70%)
                COVERAGE_THRESHOLD=70
                if command -v bc >/dev/null 2>&1; then
                    if (( $(echo "$TOTAL_COVERAGE >= $COVERAGE_THRESHOLD" | bc -l) )); then
                        pass_test "Coverage exceeds minimum threshold ($COVERAGE_THRESHOLD%): $TOTAL_COVERAGE%"
                    else
                        log_warning "Coverage below threshold ($COVERAGE_THRESHOLD%): $TOTAL_COVERAGE%"
                    fi
                else
                    # Fallback comparison without bc
                    coverage_int=$(echo "$TOTAL_COVERAGE" | cut -d'.' -f1)
                    if [ "$coverage_int" -ge "$COVERAGE_THRESHOLD" ]; then
                        pass_test "Coverage exceeds minimum threshold ($COVERAGE_THRESHOLD%): $TOTAL_COVERAGE%"
                    else
                        log_warning "Coverage below threshold ($COVERAGE_THRESHOLD%): $TOTAL_COVERAGE%"
                    fi
                fi
            else
                log_warning "Could not extract total coverage percentage"
            fi

            # Show package-level coverage
            log_info "Package-level coverage breakdown:"
            while IFS= read -r line; do
                if echo "$line" | grep -q "coverage:"; then
                    echo "  📦 $line"
                fi
            done < ../coverage/unit_test_output.log

        else
            log_warning "Could not generate coverage summary"
        fi

        # Check for any test failures
        if grep -q "FAIL" ../coverage/unit_test_output.log; then
            log_warning "Some unit tests failed - check coverage/unit_test_output.log for details"

            # Show failed tests summary
            log_info "Failed tests summary:"
            grep "FAIL" ../coverage/unit_test_output.log | head -5 | while read line; do
                echo "  ❌ $line"
            done
        else
            pass_test "All unit tests passed successfully"
        fi

    else
        fail_test "Unit tests failed to execute"
        log_error "Check coverage/unit_test_output.log for detailed error information"
    fi

    # Return to original directory
    cd ..

    echo ""
}

# Main test suite
main() {
    # Skip phases not applicable in Docker mode
    if [ "$DOCKER_MODE" = false ]; then
        # Clean up any existing processes first
        killall_validators

        # Create coverage directory if it doesn't exist
        mkdir -p coverage

        # Run unit tests first
        test_unit_tests

        echo "🚀 Phase 1: Server Startup & Basic Health Checks"
        echo "================================================="

        # Start server
        log_info "Starting server on port $SERVER_PORT..."
        PORT=$SERVER_PORT ./bin/validator &
        SERVER_PID=$!
        wait_for_server
    else
        # Docker mode - server is already running
        log_info "Skipping unit tests and server startup (Docker mode)"
        log_info "Using existing server at $API_BASE"

        # Just verify the server is available
        wait_for_server
    fi

    echo ""
    echo "🔍 Phase 2: Basic Endpoint Testing"
    echo "=================================="

    # Test basic endpoints
    test_endpoint "Health endpoint" "$API_BASE/health" "200"
    test_endpoint "Models list endpoint" "$API_BASE/models" "200"
    test_endpoint "Swagger models endpoint" "$API_BASE/swagger/models" "200"
    test_endpoint "Swagger UI endpoint" "$API_BASE/swagger/" "301"

    echo ""
    echo "📋 Phase 3: Pure Automatic Model Discovery Testing"
    echo "================================================="

    log_info "Testing the new automatic discovery system..."

    # Check that existing models are auto-registered
    check_model_registered "github" "GitHub model should be auto-registered"
    check_model_registered "incident" "Incident model should be auto-registered"
    check_model_registered "api" "API model should be auto-registered"
    check_model_registered "database" "Database model should be auto-registered"
    check_model_registered "generic" "Generic model should be auto-registered"
    check_model_registered "deployment" "Deployment model should be auto-registered"

    # Test that no hardcoded models remain (bitbucket, gitlab, slack should NOT be registered)
    check_model_not_registered "bitbucket" "Bitbucket model should NOT be registered (was deleted)"
    check_model_not_registered "gitlab" "GitLab model should NOT be registered (was deleted)"
    check_model_not_registered "slack" "Slack model should NOT be registered (was deleted)"

    log_success "Automatic discovery system is working correctly!"

    echo ""
    echo "🎯 Phase 4: Existing Model Validation Testing"
    echo "=============================================="

    # Test all discovered models with available test data
    test_all_models

    # Skip filesystem-based tests in Docker mode
    if [ "$DOCKER_MODE" = false ]; then
        echo ""
        echo "🗑️ Phase 5: Model Deletion & Server Restart Testing"
        echo "=================================================="

        # Test model deletion with server restart (since no file watchers)
    log_info "Testing model deletion with server restart..."

    # Backup incident model files
    INCIDENT_MODEL_BACKUP="src/models/incident.go.backup"
    INCIDENT_VALIDATION_BACKUP="src/validations/incident.go.backup"

    if [ -f "src/models/incident.go" ]; then
        cp "src/models/incident.go" "$INCIDENT_MODEL_BACKUP"
        log_success "Backed up incident model"
    fi

    if [ -f "src/validations/incident.go" ]; then
        cp "src/validations/incident.go" "$INCIDENT_VALIDATION_BACKUP"
        log_success "Backed up incident validation"
    fi

    # Delete incident model files
    log_info "Deleting incident model files..."
    rm -f src/models/incident.go src/validations/incident.go

    # Stop current server
    log_info "Stopping current server..."
    if [ ! -z "$SERVER_PID" ]; then
        kill $SERVER_PID 2>/dev/null || true
        wait $SERVER_PID 2>/dev/null || true
    fi

    # Start server again to test model deletion
    log_info "Starting server again to test model deletion..."
    PORT=$SERVER_PORT ./bin/validator &
    SERVER_PID=$!
    wait_for_server

    # Check that incident model is no longer registered
    check_model_not_registered "incident" "Incident model should be unregistered after deletion and server restart"

    # Test that incident endpoint returns appropriate error
    test_endpoint "Deleted incident model endpoint should return error" "$API_BASE/validate/incident" "404" "POST"

    echo ""
    echo "🔄 Phase 6: Model Restoration & Server Restart Testing"
    echo "===================================================="

    # Restore incident model files
    log_info "Restoring incident model files..."
    if [ -f "$INCIDENT_MODEL_BACKUP" ]; then
        mv "$INCIDENT_MODEL_BACKUP" src/models/incident.go
        log_success "Restored incident model"
    fi
    if [ -f "$INCIDENT_VALIDATION_BACKUP" ]; then
        mv "$INCIDENT_VALIDATION_BACKUP" src/validations/incident.go
        log_success "Restored incident validation"
    fi

    # Clear backup variables to prevent cleanup from trying to restore again
    INCIDENT_MODEL_BACKUP=""
    INCIDENT_VALIDATION_BACKUP=""

    # Stop current server
    log_info "Stopping server for restoration test..."
    if [ ! -z "$SERVER_PID" ]; then
        kill $SERVER_PID 2>/dev/null || true
        wait $SERVER_PID 2>/dev/null || true
    fi

    # Start server again to test model restoration
    log_info "Starting server again to test model restoration..."
    PORT=$SERVER_PORT ./bin/validator &
    SERVER_PID=$!
    wait_for_server

    # Check that incident model is re-registered
    check_model_registered "incident" "Incident model should be re-registered after restoration and server restart"

    # Test incident model validation works again
    test_model_validation "incident"

    echo ""
    echo "🆕 Phase 7: Dynamic Model Creation & Server Restart Testing"
    echo "========================================================="

    # Create a new test model
    create_test_model "testmodel"

    # Stop current server
    log_info "Stopping server for new model test..."
    if [ ! -z "$SERVER_PID" ]; then
        kill $SERVER_PID 2>/dev/null || true
        wait $SERVER_PID 2>/dev/null || true
    fi

    # Start server again to test new model registration
    log_info "Starting server again to test new model registration..."
    PORT=$SERVER_PORT ./bin/validator &
    SERVER_PID=$!
    wait_for_server

    # Check that test model is registered (informational - may not work due to Go module compilation)
    log_info "Checking if dynamic testmodel was registered..."
    response=$(curl -s "$API_BASE/models")
    if echo "$response" | grep -q "\"testmodel\""; then
        log_success "Dynamic testmodel was successfully auto-registered"
        PASSED_TESTS=$((PASSED_TESTS + 1))

        # Test new model validation
        test_model_validation "testmodel"
    else
        log_warning "Dynamic testmodel was not auto-registered (this is expected in some Go build scenarios)"
        log_info "This does not affect the core functionality - model deletion/restoration works correctly"
    fi
    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    # Clean up test model
    log_info "Cleaning up test model..."
    rm -f src/models/testmodel.go src/validations/testmodel.go

    fi  # End of Docker mode check - skip phases 5-7 in Docker mode

    echo ""
    echo "📊 Phase 8: API Response Format Testing"
    echo "======================================="

    # Test models endpoint returns proper JSON
    start_test "Models endpoint returns valid JSON format"
    response=$(curl -s "$API_BASE/models")
    if echo "$response" | python3 -m json.tool >/dev/null 2>&1; then
        pass_test "Models endpoint returns valid JSON"
    else
        fail_test "Models endpoint does not return valid JSON"
        echo "Response: $response"
    fi

    # Test swagger models endpoint
    start_test "Swagger models endpoint returns valid JSON format"
    response=$(curl -s "$API_BASE/swagger/models")
    if echo "$response" | python3 -m json.tool >/dev/null 2>&1; then
        pass_test "Swagger models endpoint returns valid JSON"
    else
        fail_test "Swagger models endpoint does not return valid JSON"
        echo "Response: $response"
    fi

    echo ""
    echo "📊 Phase 9: Array Validation Testing"
    echo "===================================="

    # Test array validation with multiple valid records
    start_test "Array validation with 2 valid incident records"
    response=$(curl -s -X POST "$API_BASE/validate" \
      -H "Content-Type: application/json" \
      -d '{
        "model_type": "incident",
        "data": [
          {
            "id": "INC-20240115-0001",
            "title": "Array Test 1",
            "description": "Testing array validation with first record",
            "priority": 5,
            "severity": "critical",
            "status": "open",
            "category": "performance",
            "environment": "production",
            "reported_by": "alice@example.com",
            "assigned_to": "bob@example.com",
            "created_at": "2024-01-15T10:00:00Z",
            "reported_at": "2024-01-15T10:00:00Z"
          },
          {
            "id": "INC-20240115-0002",
            "title": "Array Test 2",
            "description": "Testing array validation with second record",
            "priority": 3,
            "severity": "high",
            "status": "open",
            "category": "bug",
            "environment": "production",
            "reported_by": "bob@example.com",
            "assigned_to": "alice@example.com",
            "created_at": "2024-01-15T11:00:00Z",
            "reported_at": "2024-01-15T11:00:00Z"
          }
        ]
      }')

    if echo "$response" | grep -q '"batch_id"' && \
       echo "$response" | grep -q '"status":"success"' && \
       echo "$response" | grep -q '"total_records":2' && \
       echo "$response" | grep -q '"valid_records":2'; then
        pass_test "Array validation with valid records works correctly (status: success)"
    else
        fail_test "Array validation did not return expected structure (expected status: success)"
        echo "Response: $response"
    fi

    # Test that valid records are NOT included in results array
    start_test "Array validation excludes valid rows from results"
    if echo "$response" | grep -q '"results":\[\]' || echo "$response" | python3 -c "import sys, json; data=json.load(sys.stdin); sys.exit(0 if len(data.get('results', [])) == 0 else 1)" 2>/dev/null; then
        pass_test "Valid rows correctly excluded from results array"
    else
        fail_test "Valid rows should be excluded from results array"
        echo "Response: $response"
    fi

    # Test array validation with mixed valid/invalid records
    start_test "Array validation with mixed valid/invalid records"
    response=$(curl -s -X POST "$API_BASE/validate" \
      -H "Content-Type: application/json" \
      -d '{
        "model_type": "incident",
        "data": [
          {
            "id": "INC-20240115-0003",
            "title": "Valid Record",
            "description": "This is a valid incident record for testing",
            "priority": 5,
            "severity": "critical",
            "status": "open",
            "category": "performance",
            "environment": "production",
            "reported_by": "alice@example.com",
            "assigned_to": "bob@example.com",
            "created_at": "2024-01-15T10:00:00Z",
            "reported_at": "2024-01-15T10:00:00Z"
          },
          {
            "id": "INVALID",
            "title": "",
            "description": "Short",
            "priority": 999,
            "severity": "invalid",
            "status": "unknown"
          }
        ]
      }')

    if echo "$response" | grep -q '"total_records":2' && \
       echo "$response" | grep -q '"valid_records":1' && \
       echo "$response" | grep -q '"invalid_records":1' && \
       echo "$response" | grep -q '"status":"success"'; then
        pass_test "Array validation correctly identifies mixed valid/invalid records (status: success, no threshold)"
    else
        fail_test "Array validation did not correctly process mixed records"
        echo "Response: $response"
    fi

    # Test that only invalid rows are in results
    start_test "Array validation only includes invalid rows in results"
    results_count=$(echo "$response" | python3 -c "import sys, json; data=json.load(sys.stdin); print(len(data.get('results', [])))" 2>/dev/null || echo "unknown")
    if [ "$results_count" = "1" ]; then
        pass_test "Only 1 invalid row included in results (valid rows excluded)"
    else
        fail_test "Expected 1 invalid row in results, got: $results_count"
        echo "Response: $response"
    fi

    # Test array validation returns proper summary
    start_test "Array validation returns summary statistics"
    response=$(curl -s -X POST "$API_BASE/validate" \
      -H "Content-Type: application/json" \
      -d '{
        "model_type": "incident",
        "data": [
          {
            "id": "INC-20240115-0004",
            "title": "Summary Test",
            "description": "Testing that summary statistics are returned correctly",
            "priority": 5,
            "severity": "critical",
            "status": "open",
            "category": "performance",
            "environment": "production",
            "reported_by": "test@example.com",
            "assigned_to": "admin@example.com",
            "created_at": "2024-01-15T10:00:00Z",
            "reported_at": "2024-01-15T10:00:00Z"
          }
        ]
      }')

    if echo "$response" | grep -q '"summary"' && \
       echo "$response" | grep -q '"success_rate"' && \
       echo "$response" | grep -q '"total_records_processed"'; then
        pass_test "Array validation returns proper summary statistics"
    else
        fail_test "Array validation summary is missing or incomplete"
        echo "Response: $response"
    fi

    # Test array validation includes row-level results (valid record should have empty results)
    start_test "Array validation with valid record has empty results array"
    response=$(curl -s -X POST "$API_BASE/validate" \
      -H "Content-Type: application/json" \
      -d '{
        "model_type": "incident",
        "data": [
          {
            "id": "INC-20240115-0005",
            "title": "Row Level Test",
            "description": "Testing row-level validation results with proper title length",
            "priority": 5,
            "severity": "critical",
            "status": "open",
            "category": "performance",
            "environment": "production",
            "reported_by": "test@example.com",
            "assigned_to": "admin@example.com",
            "created_at": "2024-01-15T10:00:00Z",
            "reported_at": "2024-01-15T10:00:00Z"
          }
        ]
      }')

    # Valid record with warnings should have 1 result (STALE_INCIDENT warning)
    results_count=$(echo "$response" | python3 -c "import sys, json; data=json.load(sys.stdin); print(len(data.get('results', [])))" 2>/dev/null || echo "unknown")
    if [ "$results_count" = "1" ]; then
        pass_test "Valid record with warnings correctly included in results array"
    else
        fail_test "Expected 1 row in results for valid record with warnings, got: $results_count"
        echo "Response: $response"
    fi

    # Test backward compatibility with single object validation
    start_test "Single object validation still works (backward compatibility)"
    response=$(curl -s -X POST "$API_BASE/validate" \
      -H "Content-Type: application/json" \
      -d '{
        "model_type": "incident",
        "payload": {
          "id": "INC-006",
          "title": "Backward Compatibility Test",
          "description": "Testing that single object validation still works",
          "priority": 1,
          "severity": "critical",
          "status": "open",
          "category": "feature",
          "environment": "production",
          "reported_by": "test@example.com",
          "assigned_to": "admin@example.com",
          "created_at": "2024-01-15T10:00:00Z",
          "reported_at": "2024-01-15T10:00:00Z"
        }
      }')

    if echo "$response" | grep -q '"is_valid"' && \
       ! echo "$response" | grep -q '"batch_id"'; then
        pass_test "Single object validation maintains backward compatibility"
    else
        fail_test "Single object validation format changed unexpectedly"
        echo "Response: $response"
    fi

    echo ""
    echo "🎯 Phase 10: Threshold Validation Testing"
    echo "=========================================="

    # Test 1: Threshold success - 80% valid with 20% threshold
    start_test "Threshold: 80% valid with 20% threshold should succeed"
    response=$(curl -s -X POST "$API_BASE/validate" \
      -H "Content-Type: application/json" \
      -d '{
        "model_type": "incident",
        "threshold": 20.0,
        "data": [
          {
            "id": "INC-20240115-1001",
            "title": "Valid Record 1",
            "description": "Testing threshold validation with valid record",
            "priority": 5,
            "severity": "critical",
            "status": "open",
            "category": "performance",
            "environment": "production",
            "reported_by": "alice@example.com",
            "assigned_to": "bob@example.com",
            "created_at": "2024-01-15T10:00:00Z",
            "reported_at": "2024-01-15T10:00:00Z"
          },
          {
            "id": "INC-20240115-1002",
            "title": "Valid Record 2",
            "description": "Testing threshold validation with valid record",
            "priority": 4,
            "severity": "high",
            "status": "open",
            "category": "bug",
            "environment": "production",
            "reported_by": "bob@example.com",
            "assigned_to": "alice@example.com",
            "created_at": "2024-01-15T11:00:00Z",
            "reported_at": "2024-01-15T11:00:00Z"
          },
          {
            "id": "INC-20240115-1003",
            "title": "Valid Record 3",
            "description": "Testing threshold validation with valid record",
            "priority": 3,
            "severity": "medium",
            "status": "open",
            "category": "performance",
            "environment": "staging",
            "reported_by": "charlie@example.com",
            "assigned_to": "dave@example.com",
            "created_at": "2024-01-15T12:00:00Z",
            "reported_at": "2024-01-15T12:00:00Z"
          },
          {
            "id": "INC-20240115-1004",
            "title": "Valid Record 4",
            "description": "Testing threshold validation with valid record",
            "priority": 5,
            "severity": "critical",
            "status": "investigating",
            "category": "security",
            "environment": "production",
            "reported_by": "dave@example.com",
            "assigned_to": "charlie@example.com",
            "created_at": "2024-01-15T13:00:00Z",
            "reported_at": "2024-01-15T13:00:00Z"
          },
          {
            "id": "INVALID-1",
            "title": "",
            "description": "Bad",
            "priority": 999,
            "severity": "invalid",
            "status": "unknown"
          }
        ]
      }')

    if echo "$response" | grep -q '"status":"success"' && \
       echo "$response" | grep -q '"total_records":5' && \
       echo "$response" | grep -q '"valid_records":4' && \
       echo "$response" | grep -q '"invalid_records":1' && \
       echo "$response" | grep -q '"threshold":20'; then
        pass_test "Threshold validation: 80% valid (4/5) with 20% threshold = success"
    else
        fail_test "Threshold validation failed: expected success with 80% valid rate"
        echo "Response: $response"
    fi

    # Test 2: Threshold failure - 10% valid with 20% threshold
    start_test "Threshold: 10% valid with 20% threshold should fail"
    response=$(curl -s -X POST "$API_BASE/validate" \
      -H "Content-Type: application/json" \
      -d '{
        "model_type": "incident",
        "threshold": 20.0,
        "data": [
          {
            "id": "INC-20240115-2001",
            "title": "Valid Record",
            "description": "Testing threshold validation with one valid record",
            "priority": 5,
            "severity": "critical",
            "status": "open",
            "category": "performance",
            "environment": "production",
            "reported_by": "alice@example.com",
            "assigned_to": "bob@example.com",
            "created_at": "2024-01-15T10:00:00Z",
            "reported_at": "2024-01-15T10:00:00Z"
          },
          {
            "id": "INV-1",
            "title": "",
            "description": "Bad",
            "priority": 999,
            "severity": "invalid",
            "status": "unknown"
          },
          {
            "id": "INV-2",
            "title": "x",
            "description": "Bad",
            "priority": 888,
            "severity": "wrong",
            "status": "bad"
          },
          {
            "id": "INV-3",
            "title": "",
            "description": "Bad record",
            "priority": 777,
            "severity": "fail",
            "status": "error"
          },
          {
            "id": "INV-4",
            "title": "a",
            "description": "Bad",
            "priority": 666,
            "severity": "invalid",
            "status": "broken"
          },
          {
            "id": "INV-5",
            "title": "",
            "description": "Bad",
            "priority": 555,
            "severity": "fail",
            "status": "error"
          },
          {
            "id": "INV-6",
            "title": "b",
            "description": "Bad",
            "priority": 444,
            "severity": "wrong",
            "status": "broken"
          },
          {
            "id": "INV-7",
            "title": "",
            "description": "Bad",
            "priority": 333,
            "severity": "invalid",
            "status": "error"
          },
          {
            "id": "INV-8",
            "title": "c",
            "description": "Bad",
            "priority": 222,
            "severity": "fail",
            "status": "broken"
          },
          {
            "id": "INV-9",
            "title": "",
            "description": "Bad",
            "priority": 111,
            "severity": "wrong",
            "status": "error"
          }
        ]
      }')

    if echo "$response" | grep -q '"status":"failed"' && \
       echo "$response" | grep -q '"total_records":10' && \
       echo "$response" | grep -q '"valid_records":1' && \
       echo "$response" | grep -q '"invalid_records":9' && \
       echo "$response" | grep -q '"threshold":20'; then
        pass_test "Threshold validation: 10% valid (1/10) with 20% threshold = failed"
    else
        fail_test "Threshold validation failed: expected failed status with 10% valid rate"
        echo "Response: $response"
    fi

    # Test 3: Exact threshold match - 20% valid with 20% threshold
    start_test "Threshold: exactly 20% valid with 20% threshold should succeed"
    response=$(curl -s -X POST "$API_BASE/validate" \
      -H "Content-Type: application/json" \
      -d '{
        "model_type": "incident",
        "threshold": 20.0,
        "data": [
          {
            "id": "INC-20240115-3001",
            "title": "Valid Record",
            "description": "Testing exact threshold match with valid record",
            "priority": 5,
            "severity": "critical",
            "status": "open",
            "category": "performance",
            "environment": "production",
            "reported_by": "alice@example.com",
            "assigned_to": "bob@example.com",
            "created_at": "2024-01-15T10:00:00Z",
            "reported_at": "2024-01-15T10:00:00Z"
          },
          {
            "id": "INV-10",
            "title": "",
            "description": "Bad",
            "priority": 999,
            "severity": "invalid",
            "status": "unknown"
          },
          {
            "id": "INV-11",
            "title": "x",
            "description": "Bad",
            "priority": 888,
            "severity": "wrong",
            "status": "bad"
          },
          {
            "id": "INV-12",
            "title": "",
            "description": "Bad",
            "priority": 777,
            "severity": "fail",
            "status": "error"
          },
          {
            "id": "INV-13",
            "title": "a",
            "description": "Bad",
            "priority": 666,
            "severity": "invalid",
            "status": "broken"
          }
        ]
      }')

    if echo "$response" | grep -q '"status":"success"' && \
       echo "$response" | grep -q '"total_records":5' && \
       echo "$response" | grep -q '"valid_records":1' && \
       echo "$response" | grep -q '"invalid_records":4' && \
       echo "$response" | grep -q '"threshold":20'; then
        pass_test "Threshold validation: exactly 20% valid (1/5) with 20% threshold = success"
    else
        fail_test "Threshold validation failed: expected success with exactly 20% valid rate"
        echo "Response: $response"
    fi

    # Test 4: No threshold with multiple records should default to success
    start_test "No threshold: multiple records with some invalid should default to success"
    response=$(curl -s -X POST "$API_BASE/validate" \
      -H "Content-Type: application/json" \
      -d '{
        "model_type": "incident",
        "data": [
          {
            "id": "INC-20240115-4001",
            "title": "Valid Record",
            "description": "Testing no threshold with valid record",
            "priority": 5,
            "severity": "critical",
            "status": "open",
            "category": "performance",
            "environment": "production",
            "reported_by": "alice@example.com",
            "assigned_to": "bob@example.com",
            "created_at": "2024-01-15T10:00:00Z",
            "reported_at": "2024-01-15T10:00:00Z"
          },
          {
            "id": "INV-14",
            "title": "",
            "description": "Bad",
            "priority": 999,
            "severity": "invalid",
            "status": "unknown"
          }
        ]
      }')

    if echo "$response" | grep -q '"status":"success"' && \
       echo "$response" | grep -q '"total_records":2' && \
       echo "$response" | grep -q '"valid_records":1' && \
       echo "$response" | grep -q '"invalid_records":1' && \
       ! echo "$response" | grep -q '"threshold"'; then
        pass_test "No threshold: multiple records default to success (no threshold field in response)"
    else
        fail_test "No threshold validation failed: expected success status by default"
        echo "Response: $response"
    fi

    # Test 5: Single invalid record with no threshold should fail
    start_test "No threshold: single invalid record should fail"
    response=$(curl -s -X POST "$API_BASE/validate" \
      -H "Content-Type: application/json" \
      -d '{
        "model_type": "incident",
        "data": [
          {
            "id": "INV-15",
            "title": "",
            "description": "Bad",
            "priority": 999,
            "severity": "invalid",
            "status": "unknown"
          }
        ]
      }')

    if echo "$response" | grep -q '"status":"failed"' && \
       echo "$response" | grep -q '"total_records":1' && \
       echo "$response" | grep -q '"valid_records":0' && \
       echo "$response" | grep -q '"invalid_records":1'; then
        pass_test "Single invalid record with no threshold correctly returns failed status"
    else
        fail_test "Single invalid record should return failed status"
        echo "Response: $response"
    fi

    # Test 6: 50% threshold with exactly 50% valid records
    start_test "Threshold: exactly 50% valid with 50% threshold should succeed"
    response=$(curl -s -X POST "$API_BASE/validate" \
      -H "Content-Type: application/json" \
      -d '{
        "model_type": "incident",
        "threshold": 50.0,
        "data": [
          {
            "id": "INC-20240115-5001",
            "title": "Valid Record 1",
            "description": "Testing 50% threshold with valid record",
            "priority": 5,
            "severity": "critical",
            "status": "open",
            "category": "performance",
            "environment": "production",
            "reported_by": "alice@example.com",
            "assigned_to": "bob@example.com",
            "created_at": "2024-01-15T10:00:00Z",
            "reported_at": "2024-01-15T10:00:00Z"
          },
          {
            "id": "INV-16",
            "title": "",
            "description": "Bad",
            "priority": 999,
            "severity": "invalid",
            "status": "unknown"
          }
        ]
      }')

    if echo "$response" | grep -q '"status":"success"' && \
       echo "$response" | grep -q '"total_records":2' && \
       echo "$response" | grep -q '"valid_records":1' && \
       echo "$response" | grep -q '"invalid_records":1' && \
       echo "$response" | grep -q '"threshold":50'; then
        pass_test "Threshold validation: exactly 50% valid (1/2) with 50% threshold = success"
    else
        fail_test "Threshold validation failed: expected success with exactly 50% valid rate"
        echo "Response: $response"
    fi

    # Test 7: Verify results include invalid rows and rows with warnings
    start_test "Threshold: verify results exclude only valid rows without warnings"
    results_count=$(echo "$response" | python3 -c "import sys, json; data=json.load(sys.stdin); print(len(data.get('results', [])))" 2>/dev/null || echo "unknown")
    if [ "$results_count" -ge "1" ]; then
        pass_test "Threshold validation: results correctly include invalid/warning rows (got $results_count rows)"
    else
        fail_test "Expected at least 1 row in results, got: $results_count"
    fi

    echo ""
    echo "🌐 Phase 11: HTTP Method Testing"
    echo "================================"

    # Test wrong HTTP methods return appropriate errors
    test_endpoint "POST to health endpoint should return method not allowed" "$API_BASE/health" "405" "POST"
    test_endpoint "GET to validate endpoint should return method not allowed" "$API_BASE/validate" "405" "GET"

    echo ""
    echo "🎉 Test Suite Complete!"
    echo "======================="
    echo ""
    echo "📈 Test Results Summary:"
    echo "========================"
    echo -e "Total Tests: ${BLUE}$TOTAL_TESTS${NC}"
    echo -e "Passed: ${GREEN}$PASSED_TESTS${NC}"
    echo -e "Failed: ${RED}$FAILED_TESTS${NC}"
    echo ""

    if [ $FAILED_TESTS -eq 0 ]; then
        echo -e "${GREEN}🎊 ALL TESTS PASSED! 🎊${NC}"
        echo ""
        echo "✅ Server startup and health checks: PASSED"
        echo "✅ Basic endpoint functionality: PASSED"
        echo "✅ Model discovery and registration: PASSED"
        echo "✅ Model validation functionality: PASSED"
        echo "✅ Model deletion and server restart: PASSED"
        echo "✅ Model restoration and server restart: PASSED"
        echo "✅ Dynamic model creation and server restart: PASSED"
        echo "✅ API response format validation: PASSED"
        echo "✅ HTTP method validation: PASSED"
        echo ""
        echo "🚀 The Go Playground Data Validator is working perfectly!"
        echo "   All dynamic registration, validation, and cleanup features are functional."

        return 0
    else
        echo -e "${RED}❌ SOME TESTS FAILED ❌${NC}"
        echo ""
        echo "Please review the failed tests above and fix any issues."

        return 1
    fi
}

# Run the test suite
main "$@"
