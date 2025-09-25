#!/bin/bash

# Comprehensive E2E Test Suite for Go Playground Data Validator
# Tests all aspects of the dynamic model registration system

set -e  # Exit on any error

echo "🧪 Starting Comprehensive E2E Test Suite"
echo "========================================"
echo ""

# Configuration
SERVER_PORT=8086
API_BASE="http://localhost:$SERVER_PORT"
SERVER_PID=""
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

# Test model validation functionality
test_model_validation() {
    local model_name=$1
    local valid_payload=$2
    local invalid_payload=$3

    # Test valid payload
    start_test "Testing valid payload for $model_name"
    response=$(curl -s -X POST "$API_BASE/validate/$model_name" \
        -H "Content-Type: application/json" \
        -d "$valid_payload")

    if echo "$response" | grep -q '"is_valid":true'; then
        pass_test "Valid payload validation passed for $model_name"
    else
        fail_test "Valid payload validation failed for $model_name"
        echo "Response: $response"
    fi

    # Test invalid payload
    start_test "Testing invalid payload for $model_name"
    response=$(curl -s -X POST "$API_BASE/validate/$model_name" \
        -H "Content-Type: application/json" \
        -d "$invalid_payload")

    if echo "$response" | grep -q '"is_valid":false'; then
        pass_test "Invalid payload validation passed for $model_name"
    else
        fail_test "Invalid payload validation failed for $model_name"
        echo "Response: $response"
    fi
}

# Create a test model
create_test_model() {
    local model_name=$1

    log_info "Creating test model: $model_name"

    # Create model file
    cat > "src/models/${model_name}.go" << EOF
package models

import "time"

// ${model_name^}Payload represents a test model for dynamic registration
type ${model_name^}Payload struct {
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

// ${model_name^}Validator handles validation for ${model_name} models
type ${model_name^}Validator struct {
    validator *validator.Validate
}

// New${model_name^}Validator creates a new ${model_name} validator
func New${model_name^}Validator() *${model_name^}Validator {
    return &${model_name^}Validator{
        validator: validator.New(),
    }
}

// ValidatePayload validates a ${model_name} payload
func (v *${model_name^}Validator) ValidatePayload(payload models.${model_name^}Payload) models.ValidationResult {
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

# Main test suite
main() {
    echo "🚀 Phase 1: Server Startup & Basic Health Checks"
    echo "================================================="

    # Start server
    log_info "Starting server on port $SERVER_PORT..."
    cd src && PORT=$SERVER_PORT go run main.go &
    SERVER_PID=$!
    cd ..
    wait_for_server

    echo ""
    echo "🔍 Phase 2: Basic Endpoint Testing"
    echo "=================================="

    # Test basic endpoints
    test_endpoint "Health endpoint" "$API_BASE/health" "200"
    test_endpoint "Models list endpoint" "$API_BASE/models" "200"
    test_endpoint "Swagger models endpoint" "$API_BASE/swagger/models" "200"
    test_endpoint "Swagger UI endpoint" "$API_BASE/swagger/" "200"

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

    # Test GitHub model validation
    github_valid='{"repository":{"full_name":"test/repo","name":"repo"},"pusher":{"name":"testuser"},"commits":[{"message":"test commit","author":{"name":"Test User","email":"test@example.com"}}]}'
    github_invalid='{"repository":{},"pusher":{},"commits":[]}'
    test_model_validation "github" "$github_valid" "$github_invalid"

    # Test Incident model validation
    incident_valid='{"title":"Test Incident","description":"A test incident","severity":"high","status":"open","reporter":"test@example.com","created_at":"2024-01-01T10:00:00Z"}'
    incident_invalid='{"title":"","description":"","severity":"invalid","status":"","reporter":"invalid-email"}'
    test_model_validation "incident" "$incident_valid" "$incident_invalid"

    echo ""
    echo "🗑️ Phase 5: Model Deletion & Recovery Testing (Incident Model)"
    echo "=============================================================="

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

    # Delete incident model
    log_info "Deleting incident model files..."
    rm -f src/models/incident.go src/validations/incident.go

    # Wait for file system watcher to detect deletion
    log_info "Waiting for file system watcher to detect deletion..."
    sleep 5

    # Check that incident model is unregistered
    check_model_not_registered "incident" "Incident model should be unregistered after deletion"

    # Test that incident endpoint returns appropriate response
    test_endpoint "Deleted incident model endpoint should return error" "$API_BASE/validate/incident" "404" "POST"

    # Restore incident model files
    log_info "Restoring incident model files..."
    if [ -f "$INCIDENT_MODEL_BACKUP" ]; then
        mv "$INCIDENT_MODEL_BACKUP" src/models/incident.go
    fi
    if [ -f "$INCIDENT_VALIDATION_BACKUP" ]; then
        mv "$INCIDENT_VALIDATION_BACKUP" src/validations/incident.go
    fi

    # Clear backup variables to prevent cleanup from trying to restore again
    INCIDENT_MODEL_BACKUP=""
    INCIDENT_VALIDATION_BACKUP=""

    # Wait for file system watcher to detect restoration
    log_info "Waiting for file system watcher to detect restoration..."
    sleep 5

    # Check that incident model is re-registered
    check_model_registered "incident" "Incident model should be re-registered after restoration"

    # Test incident model validation
    incident_valid='{"title":"Test Incident","description":"A test incident","severity":"high","status":"open","reporter":"test@example.com","created_at":"2024-01-01T10:00:00Z"}'
    incident_invalid='{"title":"","description":"","severity":"invalid","status":"","reporter":"invalid-email"}'
    test_model_validation "incident" "$incident_valid" "$incident_invalid"

    echo ""
    echo "🆕 Phase 6: Dynamic Model Creation Testing"
    echo "=========================================="

    # Create a new test model
    create_test_model "testmodel"

    # Wait for file system watcher to detect new model
    log_info "Waiting for file system watcher to detect new model..."
    sleep 5

    # Check that test model is registered
    check_model_registered "testmodel" "Test model should be auto-registered"

    # Test new model endpoint
    test_endpoint "New test model endpoint should be available" "$API_BASE/validate/testmodel" "400" "POST"

    # Test new model validation
    testmodel_valid='{"id":"test-123","name":"John Doe","email":"john@example.com","age":25,"is_active":true,"created_at":"2024-01-01T10:00:00Z","tags":["user","test"]}'
    testmodel_invalid='{"id":"","name":"A","email":"invalid-email","age":-5}'
    test_model_validation "testmodel" "$testmodel_valid" "$testmodel_invalid"

    echo ""
    echo "🔄 Phase 7: Dynamic Model Cleanup Testing"
    echo "========================================="

    # Delete test model
    log_info "Deleting test model files..."
    rm -f src/models/testmodel.go src/validations/testmodel.go

    # Wait for file system watcher to detect deletion
    log_info "Waiting for file system watcher to detect deletion..."
    sleep 5

    # Check that test model is unregistered
    check_model_not_registered "testmodel" "Test model should be unregistered after deletion"

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
    echo "🌐 Phase 9: HTTP Method Testing"
    echo "==============================="

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
        echo "✅ Dynamic model deletion/recovery: PASSED"
        echo "✅ Dynamic model creation/cleanup: PASSED"
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
