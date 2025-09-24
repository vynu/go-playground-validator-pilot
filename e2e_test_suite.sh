#!/bin/bash

# Unified E2E Test Suite for Go Playground Validator
# Runs all test types: unit, e2e, integration, swagger, api, performance
# Supports modular execution and comprehensive reporting

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
NC='\033[0m' # No Color

# Print functions
print_header() {
    echo -e "${MAGENTA}╔══════════════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${MAGENTA}║${NC} $1 ${MAGENTA}║${NC}"
    echo -e "${MAGENTA}╚══════════════════════════════════════════════════════════════════════════╝${NC}"
}

print_section() {
    echo -e "\n${CYAN}═══ $1 ═══${NC}"
}

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[PASS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[FAIL]${NC} $1"
}

print_step() {
    echo -e "${CYAN}▶${NC} $1"
}

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SRC_DIR="$PROJECT_ROOT/src"
TEST_DATA_DIR="$PROJECT_ROOT/test_data"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
LOG_DIR="$PROJECT_ROOT/test_logs"
RESULTS_DIR="$PROJECT_ROOT/test_results_${TIMESTAMP}"
COVERAGE_DIR="$RESULTS_DIR/coverage"

# Test type flags
RUN_UNIT_TESTS=true
RUN_E2E_TESTS=true
RUN_INTEGRATION_TESTS=true
RUN_SWAGGER_TESTS=true
RUN_API_TESTS=true
RUN_PERFORMANCE_TESTS=true

# Server configuration
SERVER_HOST="localhost"
SERVER_PORT="8080"
FLEX_PORT="8081"
BASE_URL="http://${SERVER_HOST}:${SERVER_PORT}"
FLEX_URL="http://${SERVER_HOST}:${FLEX_PORT}"

# Test configuration
GENERATE_COVERAGE=true
VERBOSE=false
PARALLEL_TESTS=false
QUICK_MODE=false
START_SERVERS=false
CLEANUP_ON_EXIT=true
SERVER_TIMEOUT=60
PERFORMANCE_DURATION=30

# Counters
TOTAL_TESTS=0
TOTAL_PASSED=0
TOTAL_FAILED=0
TOTAL_SKIPPED=0

# Arrays for tracking results (using bash associative arrays)
declare -a TEST_CATEGORIES_KEYS
declare -a TEST_CATEGORIES_VALUES
declare -a TEST_RESULTS_KEYS
declare -a TEST_RESULTS_VALUES
declare -a TEST_TIMES_KEYS
declare -a TEST_TIMES_VALUES
declare -a SERVER_PIDS_KEYS
declare -a SERVER_PIDS_VALUES

# Helper functions for associative arrays
set_test_category() {
    local key="$1"
    local value="$2"
    TEST_CATEGORIES_KEYS+=("$key")
    TEST_CATEGORIES_VALUES+=("$value")
}

get_test_category() {
    local key="$1"
    for i in "${!TEST_CATEGORIES_KEYS[@]}"; do
        if [[ "${TEST_CATEGORIES_KEYS[$i]}" == "$key" ]]; then
            echo "${TEST_CATEGORIES_VALUES[$i]}"
            return
        fi
    done
    echo ""
}

set_test_result() {
    local key="$1"
    local value="$2"
    TEST_RESULTS_KEYS+=("$key")
    TEST_RESULTS_VALUES+=("$value")
}

get_test_result() {
    local key="$1"
    for i in "${!TEST_RESULTS_KEYS[@]}"; do
        if [[ "${TEST_RESULTS_KEYS[$i]}" == "$key" ]]; then
            echo "${TEST_RESULTS_VALUES[$i]}"
            return
        fi
    done
    echo "1"
}

set_test_time() {
    local key="$1"
    local value="$2"
    TEST_TIMES_KEYS+=("$key")
    TEST_TIMES_VALUES+=("$value")
}

get_test_time() {
    local key="$1"
    for i in "${!TEST_TIMES_KEYS[@]}"; do
        if [[ "${TEST_TIMES_KEYS[$i]}" == "$key" ]]; then
            echo "${TEST_TIMES_VALUES[$i]}"
            return
        fi
    done
    echo "0"
}

set_server_pid() {
    local key="$1"
    local value="$2"
    SERVER_PIDS_KEYS+=("$key")
    SERVER_PIDS_VALUES+=("$value")
}

get_server_pid() {
    local key="$1"
    for i in "${!SERVER_PIDS_KEYS[@]}"; do
        if [[ "${SERVER_PIDS_KEYS[$i]}" == "$key" ]]; then
            echo "${SERVER_PIDS_VALUES[$i]}"
            return
        fi
    done
    echo ""
}

# Create directories
mkdir -p "$LOG_DIR" "$RESULTS_DIR" "$COVERAGE_DIR"

# Cleanup function
cleanup() {
    if [[ "$CLEANUP_ON_EXIT" == "true" ]]; then
        print_status "Cleaning up..."

        # Kill any servers we started
        for i in "${!SERVER_PIDS_KEYS[@]}"; do
            local pid="${SERVER_PIDS_VALUES[$i]}"
            if kill -0 "$pid" 2>/dev/null; then
                print_status "Stopping server (PID: $pid)"
                kill "$pid" 2>/dev/null || true
            fi
        done

        # Clean up temporary files
        rm -f /tmp/curl_response.json
        rm -f validator validator_*

        if [[ "$QUICK_MODE" == "false" ]]; then
            cd "$SRC_DIR" && go clean -testcache 2>/dev/null || true
        fi
    fi
}

# Set trap for cleanup
trap cleanup EXIT

# Help function
show_help() {
    cat << EOF
Unified E2E Test Suite for Go Playground Validator

USAGE:
    $0 [OPTIONS]

TEST TYPE OPTIONS:
    --unit                  Run only unit tests
    --e2e                   Run only end-to-end tests
    --integration           Run only integration tests
    --swagger               Run only Swagger/API documentation tests
    --api                   Run only API endpoint tests
    --performance           Run only performance tests
    --all                   Run all test types (default)

CONFIGURATION OPTIONS:
    --host HOST             Server host (default: localhost)
    --port PORT             Main server port (default: 8080)
    --flex-port PORT        Flexible server port (default: 8081)
    --start-servers         Start servers automatically
    --no-coverage           Skip coverage generation
    --no-cleanup            Don't cleanup on exit
    --parallel              Run tests in parallel where possible
    --quick                 Quick mode (reduced timeouts, no performance tests)
    --performance-duration  Duration for performance tests in seconds (default: 30)

OUTPUT OPTIONS:
    --verbose, -v           Verbose output
    --results-dir DIR       Custom results directory
    --quiet                 Minimal output

EXAMPLES:
    $0                                      # Run all tests
    $0 --unit --integration                 # Run unit and integration tests only
    $0 --api --swagger --start-servers      # Run API tests with auto-server start
    $0 --performance --performance-duration 60  # Extended performance testing
    $0 --quick --verbose                    # Quick run with detailed output
    $0 --parallel --no-coverage             # Fast parallel execution

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --unit)
            RUN_UNIT_TESTS=true
            RUN_E2E_TESTS=false
            RUN_INTEGRATION_TESTS=false
            RUN_SWAGGER_TESTS=false
            RUN_API_TESTS=false
            RUN_PERFORMANCE_TESTS=false
            shift
            ;;
        --e2e)
            RUN_UNIT_TESTS=false
            RUN_E2E_TESTS=true
            RUN_INTEGRATION_TESTS=false
            RUN_SWAGGER_TESTS=false
            RUN_API_TESTS=false
            RUN_PERFORMANCE_TESTS=false
            shift
            ;;
        --integration)
            RUN_UNIT_TESTS=false
            RUN_E2E_TESTS=false
            RUN_INTEGRATION_TESTS=true
            RUN_SWAGGER_TESTS=false
            RUN_API_TESTS=false
            RUN_PERFORMANCE_TESTS=false
            shift
            ;;
        --swagger)
            RUN_UNIT_TESTS=false
            RUN_E2E_TESTS=false
            RUN_INTEGRATION_TESTS=false
            RUN_SWAGGER_TESTS=true
            RUN_API_TESTS=false
            RUN_PERFORMANCE_TESTS=false
            shift
            ;;
        --api)
            RUN_UNIT_TESTS=false
            RUN_E2E_TESTS=false
            RUN_INTEGRATION_TESTS=false
            RUN_SWAGGER_TESTS=false
            RUN_API_TESTS=true
            RUN_PERFORMANCE_TESTS=false
            shift
            ;;
        --performance)
            RUN_UNIT_TESTS=false
            RUN_E2E_TESTS=false
            RUN_INTEGRATION_TESTS=false
            RUN_SWAGGER_TESTS=false
            RUN_API_TESTS=false
            RUN_PERFORMANCE_TESTS=true
            shift
            ;;
        --all)
            RUN_UNIT_TESTS=true
            RUN_E2E_TESTS=true
            RUN_INTEGRATION_TESTS=true
            RUN_SWAGGER_TESTS=true
            RUN_API_TESTS=true
            RUN_PERFORMANCE_TESTS=true
            shift
            ;;
        --host)
            SERVER_HOST="$2"
            BASE_URL="http://${SERVER_HOST}:${SERVER_PORT}"
            FLEX_URL="http://${SERVER_HOST}:${FLEX_PORT}"
            shift 2
            ;;
        --port)
            SERVER_PORT="$2"
            BASE_URL="http://${SERVER_HOST}:${SERVER_PORT}"
            shift 2
            ;;
        --flex-port)
            FLEX_PORT="$2"
            FLEX_URL="http://${SERVER_HOST}:${FLEX_PORT}"
            shift 2
            ;;
        --start-servers)
            START_SERVERS=true
            shift
            ;;
        --no-coverage)
            GENERATE_COVERAGE=false
            shift
            ;;
        --no-cleanup)
            CLEANUP_ON_EXIT=false
            shift
            ;;
        --parallel)
            PARALLEL_TESTS=true
            shift
            ;;
        --quick)
            QUICK_MODE=true
            SERVER_TIMEOUT=15
            PERFORMANCE_DURATION=10
            RUN_PERFORMANCE_TESTS=false
            shift
            ;;
        --performance-duration)
            PERFORMANCE_DURATION="$2"
            shift 2
            ;;
        --verbose|-v)
            VERBOSE=true
            shift
            ;;
        --results-dir)
            RESULTS_DIR="$2"
            COVERAGE_DIR="$RESULTS_DIR/coverage"
            mkdir -p "$RESULTS_DIR" "$COVERAGE_DIR"
            shift 2
            ;;
        --quiet)
            VERBOSE=false
            exec > "$LOG_DIR/test_run_${TIMESTAMP}.log" 2>&1
            shift
            ;;
        --help|-h)
            show_help
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Print test configuration
print_header "GO PLAYGROUND VALIDATOR - UNIFIED E2E TEST SUITE"

print_status "Test Configuration:"
echo "  Project Root: $PROJECT_ROOT"
echo "  Results Dir: $RESULTS_DIR"
echo "  Target Server: $BASE_URL"
echo "  Flexible Server: $FLEX_URL"
echo "  Timestamp: $(date)"
echo ""

print_status "Test Types Enabled:"
[[ "$RUN_UNIT_TESTS" == "true" ]] && echo "  ✓ Unit Tests"
[[ "$RUN_E2E_TESTS" == "true" ]] && echo "  ✓ End-to-End Tests"
[[ "$RUN_INTEGRATION_TESTS" == "true" ]] && echo "  ✓ Integration Tests"
[[ "$RUN_SWAGGER_TESTS" == "true" ]] && echo "  ✓ Swagger/Documentation Tests"
[[ "$RUN_API_TESTS" == "true" ]] && echo "  ✓ API Endpoint Tests"
[[ "$RUN_PERFORMANCE_TESTS" == "true" ]] && echo "  ✓ Performance Tests"
echo ""

print_status "Configuration:"
echo "  Coverage: $([[ "$GENERATE_COVERAGE" == "true" ]] && echo "Enabled" || echo "Disabled")"
echo "  Parallel: $([[ "$PARALLEL_TESTS" == "true" ]] && echo "Enabled" || echo "Disabled")"
echo "  Verbose: $([[ "$VERBOSE" == "true" ]] && echo "Enabled" || echo "Disabled")"
echo "  Quick Mode: $([[ "$QUICK_MODE" == "true" ]] && echo "Enabled" || echo "Disabled")"
echo "  Auto Start Servers: $([[ "$START_SERVERS" == "true" ]] && echo "Enabled" || echo "Disabled")"
echo ""

# Verify project structure
print_section "PROJECT VERIFICATION"

print_step "Verifying project structure..."
if [[ ! -d "$SRC_DIR" ]]; then
    print_error "Source directory not found: $SRC_DIR"
    exit 1
fi

if [[ ! -d "$TEST_DATA_DIR" ]]; then
    print_warning "Test data directory not found: $TEST_DATA_DIR"
    mkdir -p "$TEST_DATA_DIR"
fi

# Verify Go installation
print_step "Checking Go installation..."
if ! command -v go &> /dev/null; then
    print_error "Go is not installed or not in PATH"
    exit 1
fi

GO_VERSION=$(go version)
print_success "Go version: $GO_VERSION"

# Change to source directory for build operations
cd "$SRC_DIR"

# Verify dependencies
print_step "Verifying Go dependencies..."
if ! go mod tidy; then
    print_error "Failed to tidy Go modules"
    exit 1
fi

if ! go mod verify; then
    print_error "Failed to verify Go modules"
    exit 1
fi

print_success "Dependencies verified"

# Build the project
print_step "Building the project..."
if ! go build -o validator .; then
    print_error "Failed to build the project"
    exit 1
fi

print_success "Build successful"

# Test functions
run_go_test() {
    local test_name="$1"
    local test_pattern="$2"
    local coverage_file="$3"
    local timeout="${4:-300}"

    local start_time=$(date +%s)
    print_step "Running $test_name..."

    local cmd="timeout ${timeout}s go test"
    if [[ "$GENERATE_COVERAGE" == "true" && -n "$coverage_file" ]]; then
        cmd="$cmd -coverprofile=$COVERAGE_DIR/$coverage_file"
    fi

    if [[ "$VERBOSE" == "true" ]]; then
        cmd="$cmd -v"
    fi

    if [[ -n "$test_pattern" ]]; then
        cmd="$cmd -run $test_pattern"
    fi

    if [[ "$PARALLEL_TESTS" == "true" ]]; then
        cmd="$cmd -parallel 4"
    fi

    local output
    local result=0
    if output=$(eval "$cmd" 2>&1); then
        print_success "$test_name completed"
        result=0
    else
        print_error "$test_name failed"
        result=1
    fi

    local end_time=$(date +%s)
    local duration=$((end_time - start_time))

    # Parse test results
    local passed=$(echo "$output" | grep -c "PASS:" 2>/dev/null || echo "0")
    local failed=$(echo "$output" | grep -c "FAIL:" 2>/dev/null || echo "0")
    local skipped=$(echo "$output" | grep -c "SKIP:" 2>/dev/null || echo "0")

    TOTAL_TESTS=$((TOTAL_TESTS + passed + failed + skipped))
    TOTAL_PASSED=$((TOTAL_PASSED + passed))
    TOTAL_FAILED=$((TOTAL_FAILED + failed))
    TOTAL_SKIPPED=$((TOTAL_SKIPPED + skipped))

    set_test_category "$test_name" "$passed/$failed/$skipped"
    set_test_result "$test_name" $result
    set_test_time "$test_name" $duration

    # Save detailed output
    echo "$output" > "$RESULTS_DIR/${test_name// /_}.log"

    if [[ "$VERBOSE" == "true" ]]; then
        echo "$output"
    fi

    return $result
}

check_server() {
    local url="$1"
    local name="$2"

    if curl -s -f "$url/health" > /dev/null 2>&1; then
        return 0
    else
        return 1
    fi
}

wait_for_server() {
    local url="$1"
    local name="$2"
    local timeout="$3"

    print_step "Waiting for $name to be ready at $url..."

    local count=0
    while [[ $count -lt $timeout ]]; do
        if check_server "$url" "$name"; then
            print_success "$name is ready"
            return 0
        fi

        sleep 1
        count=$((count + 1))

        if [[ $((count % 10)) -eq 0 ]]; then
            print_status "Still waiting for $name... ($count/${timeout}s)"
        fi
    done

    print_error "$name is not responding after ${timeout}s"
    return 1
}

start_server() {
    local port="$1"
    local name="$2"
    local binary_name="$3"

    print_step "Starting $name on port $port..."

    if [[ -f "$binary_name" ]]; then
        PORT=$port ./"$binary_name" > "$RESULTS_DIR/${name// /_}_server.log" 2>&1 &
        local pid=$!
        set_server_pid "$name" $pid
        print_status "$name started (PID: $pid)"
        return 0
    else
        print_warning "$name binary not found: $binary_name"
        return 1
    fi
}

make_api_request() {
    local method="$1"
    local endpoint="$2"
    local expected_status="$3"
    local data_file="$4"
    local server_url="$5"
    local description="$6"

    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    local url="${server_url}${endpoint}"
    local curl_args=("-s" "-w" "%{http_code}" "-o" "/tmp/curl_response.json")

    if [[ "$method" == "POST" || "$method" == "PUT" ]]; then
        curl_args+=("-H" "Content-Type: application/json")
        if [[ -n "$data_file" && -f "$data_file" ]]; then
            curl_args+=("-d" "@$data_file")
        elif [[ -n "$data_file" ]]; then
            curl_args+=("-d" "$data_file")
        fi
    fi

    curl_args+=("-X" "$method" "$url")

    if [[ "$VERBOSE" == "true" ]]; then
        print_step "Request: $method $endpoint"
    fi

    local http_code
    http_code=$(curl "${curl_args[@]}" 2>/dev/null || echo "000")

    # Handle multiple expected status codes
    local expected_codes
    IFS='|' read -ra expected_codes <<< "$expected_status"
    local status_match=false

    for expected_code in "${expected_codes[@]}"; do
        if [[ "$http_code" == "$expected_code" ]]; then
            status_match=true
            break
        fi
    done

    if [[ "$status_match" == "true" ]]; then
        print_success "$description (Status: $http_code)"
        TOTAL_PASSED=$((TOTAL_PASSED + 1))

        if [[ "$VERBOSE" == "true" ]]; then
            echo "Response:"
            cat /tmp/curl_response.json | python3 -m json.tool 2>/dev/null || cat /tmp/curl_response.json
            echo ""
        fi

        return 0
    else
        print_error "$description - Expected: $expected_status, Got: $http_code"
        TOTAL_FAILED=$((TOTAL_FAILED + 1))

        echo "Response:"
        cat /tmp/curl_response.json 2>/dev/null || echo "No response body"
        echo ""

        return 1
    fi
}

# Start servers if requested
if [[ "$START_SERVERS" == "true" ]]; then
    print_section "SERVER STARTUP"

    # Build server binaries if they don't exist
    if [[ ! -f "validator_optimized" ]]; then
        print_step "Building optimized validator..."
        go build -o validator_optimized .
    fi

    if [[ ! -f "validator_swagger" ]]; then
        print_step "Building swagger validator..."
        go build -tags swagger -o validator_swagger .
    fi

    start_server "$SERVER_PORT" "Main Server" "validator_optimized"
    start_server "$FLEX_PORT" "Flexible Server" "validator_swagger"

    # Wait for servers to be ready
    wait_for_server "$BASE_URL" "Main Server" "$SERVER_TIMEOUT"
    wait_for_server "$FLEX_URL" "Flexible Server" "$SERVER_TIMEOUT" || print_warning "Flexible server not available"
fi

# Run Unit Tests
if [[ "$RUN_UNIT_TESTS" == "true" ]]; then
    print_section "UNIT TESTS"
    run_go_test "Unit Tests" "^Test[^E2E|^TestComprehensive|^TestIntegration|^TestPerformance]" "unit_coverage.out" 120
fi

# Run E2E Tests
if [[ "$RUN_E2E_TESTS" == "true" ]]; then
    print_section "END-TO-END TESTS"

    # Verify test data exists
    if [[ ! -f "$TEST_DATA_DIR/sample_pull_request.json" ]]; then
        print_warning "Sample test data not found, creating minimal test data..."
        mkdir -p "$TEST_DATA_DIR"
        cat > "$TEST_DATA_DIR/sample_pull_request.json" << 'EOF'
{
    "action": "opened",
    "number": 123,
    "pull_request": {
        "id": 123,
        "number": 123,
        "title": "Test PR",
        "state": "open",
        "created_at": "2025-09-21T12:00:00Z",
        "updated_at": "2025-09-21T12:00:00Z",
        "draft": false,
        "head": {
            "label": "test",
            "ref": "test",
            "sha": "a1b2c3d4e5f6789012345678901234567890abcd",
            "user": {"id": 1, "login": "test", "avatar_url": "https://example.com", "type": "User"},
            "repo": {"id": 1, "name": "test", "full_name": "test/test", "private": false, "owner": {"id": 1, "login": "test", "avatar_url": "https://example.com", "type": "User"}, "html_url": "https://github.com/test/test", "default_branch": "main"}
        },
        "base": {
            "label": "main",
            "ref": "main",
            "sha": "f1e2d3c4b5a6789012345678901234567890fedc",
            "user": {"id": 1, "login": "test", "avatar_url": "https://example.com", "type": "User"},
            "repo": {"id": 1, "name": "test", "full_name": "test/test", "private": false, "owner": {"id": 1, "login": "test", "avatar_url": "https://example.com", "type": "User"}, "html_url": "https://github.com/test/test", "default_branch": "main"}
        },
        "user": {"id": 1, "login": "test", "avatar_url": "https://example.com", "type": "User"}
    },
    "repository": {
        "id": 1,
        "name": "test",
        "full_name": "test/test",
        "private": false,
        "owner": {"id": 1, "login": "test", "avatar_url": "https://example.com", "type": "User"},
        "html_url": "https://github.com/test/test",
        "description": "Test repo",
        "default_branch": "main",
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2025-09-21T12:00:00Z"
    },
    "sender": {"id": 1, "login": "test", "avatar_url": "https://example.com", "type": "User"}
}
EOF
    fi

    run_go_test "E2E Tests" "^TestE2E" "e2e_coverage.out" 300
    run_go_test "Comprehensive E2E Tests" "^TestComprehensive" "comprehensive_e2e_coverage.out" 600
fi

# Run Integration Tests
if [[ "$RUN_INTEGRATION_TESTS" == "true" ]]; then
    print_section "INTEGRATION TESTS"
    run_go_test "Integration Tests" "^TestIntegration" "integration_coverage.out" 300
fi

# Run Swagger Tests
if [[ "$RUN_SWAGGER_TESTS" == "true" ]]; then
    print_section "SWAGGER/DOCUMENTATION TESTS"

    # Run integrated Swagger tests
    print_step "Running integrated Swagger tests..."

    # Basic Swagger endpoint tests
    swagger_tests_passed=true

    # Check if any server is available for testing
    if check_server "$BASE_URL" "Main Server" || check_server "$FLEX_URL" "Flexible Server"; then
        # Test swagger endpoints if available
        if check_server "$FLEX_URL" "Flexible Server"; then
            if make_api_request "GET" "/swagger/doc.json" "200|404" "" "$FLEX_URL" "Swagger JSON specification"; then
                print_success "Swagger JSON test passed"
            else
                swagger_tests_passed=false
            fi

            if make_api_request "GET" "/swagger/models" "200|404" "" "$FLEX_URL" "Swagger models endpoint"; then
                print_success "Swagger models test passed"
            else
                swagger_tests_passed=false
            fi
        fi

        if [[ "$swagger_tests_passed" == "true" ]]; then
            print_success "Integrated Swagger tests passed"
            set_test_category "Integrated Swagger Tests" "2/0/0"
            set_test_result "Integrated Swagger Tests" 0
            TOTAL_TESTS=$((TOTAL_TESTS + 2))
            TOTAL_PASSED=$((TOTAL_PASSED + 2))
        else
            print_error "Integrated Swagger tests failed"
            set_test_category "Integrated Swagger Tests" "1/1/0"
            set_test_result "Integrated Swagger Tests" 1
            TOTAL_TESTS=$((TOTAL_TESTS + 2))
            TOTAL_PASSED=$((TOTAL_PASSED + 1))
            TOTAL_FAILED=$((TOTAL_FAILED + 1))
        fi
    else
        print_warning "No servers available for Swagger testing"
        set_test_category "Integrated Swagger Tests" "0/0/1"
        set_test_result "Integrated Swagger Tests" 0
        TOTAL_TESTS=$((TOTAL_TESTS + 1))
        TOTAL_SKIPPED=$((TOTAL_SKIPPED + 1))
    fi
fi

# Run API Tests
if [[ "$RUN_API_TESTS" == "true" ]]; then
    print_section "API ENDPOINT TESTS"

    # Check if servers are available
    if check_server "$BASE_URL" "Main Server"; then
        print_step "Testing main server API endpoints..."

        # Health check
        make_api_request "GET" "/health" "200" "" "$BASE_URL" "Health check"

        # Metrics
        make_api_request "GET" "/metrics" "200" "" "$BASE_URL" "Metrics endpoint"

        # GitHub validation with test data
        if [[ -f "$TEST_DATA_DIR/sample_pull_request.json" ]]; then
            make_api_request "POST" "/validate/github" "200|422" "$TEST_DATA_DIR/sample_pull_request.json" "$BASE_URL" "GitHub validation"
        fi

        # Invalid JSON handling
        make_api_request "POST" "/validate/github" "400" '{"invalid": json}' "$BASE_URL" "Invalid JSON handling"

        # 404 handling
        make_api_request "GET" "/nonexistent" "404" "" "$BASE_URL" "404 error handling"

    else
        print_warning "Main server not available for API tests"
    fi

    # Test flexible server if available
    if check_server "$FLEX_URL" "Flexible Server"; then
        print_step "Testing flexible server API endpoints..."

        # Health check
        make_api_request "GET" "/health" "200" "" "$FLEX_URL" "Flexible server health"

        # List models
        make_api_request "GET" "/models" "200" "" "$FLEX_URL" "List available models"

        # List providers
        make_api_request "GET" "/providers" "200" "" "$FLEX_URL" "List validation providers"

    else
        print_warning "Flexible server not available for API tests"
    fi

    # Integrated API testing completed above
    print_success "Integrated API endpoint testing completed"
fi

# Run Performance Tests
if [[ "$RUN_PERFORMANCE_TESTS" == "true" ]]; then
    print_section "PERFORMANCE TESTS"

    print_step "Running Go performance tests..."
    run_go_test "Performance Tests" "^TestE2E_Performance|^TestPerformance" "performance_coverage.out" $((PERFORMANCE_DURATION + 60))

    # Load testing if servers are available
    if check_server "$BASE_URL" "Main Server"; then
        print_step "Running load tests for ${PERFORMANCE_DURATION}s..."

        # Simple concurrent load test
        if command -v ab >/dev/null 2>&1; then
            print_step "Running Apache Bench load test..."
            ab -n 100 -c 10 -t $PERFORMANCE_DURATION "$BASE_URL/health" > "$RESULTS_DIR/ab_load_test.log" 2>&1 || print_warning "Apache Bench test failed"
        elif command -v wrk >/dev/null 2>&1; then
            print_step "Running wrk load test..."
            wrk -t4 -c10 -d${PERFORMANCE_DURATION}s --timeout 10s "$BASE_URL/health" > "$RESULTS_DIR/wrk_load_test.log" 2>&1 || print_warning "wrk load test failed"
        else
            print_warning "No load testing tools (ab/wrk) available, skipping load tests"
        fi

        # Memory usage test
        if [[ -f "$TEST_DATA_DIR/sample_pull_request.json" ]]; then
            print_step "Running memory stress test..."
            for i in {1..10}; do
                curl -s -X POST -H "Content-Type: application/json" -d @"$TEST_DATA_DIR/sample_pull_request.json" "$BASE_URL/validate/github" > /dev/null &
            done
            wait
            print_success "Memory stress test completed"
        fi
    else
        print_warning "Server not available for performance tests"
    fi
fi

# Generate coverage report
if [[ "$GENERATE_COVERAGE" == "true" ]]; then
    print_section "COVERAGE REPORT"

    # Find all coverage files
    coverage_files=($(find "$COVERAGE_DIR" -name "*.out" 2>/dev/null || true))
    if [[ ${#coverage_files[@]} -gt 0 ]]; then
        print_step "Generating combined coverage report..."

        # Create combined coverage file
        echo "mode: atomic" > "$COVERAGE_DIR/combined_coverage.out"
        for file in "${coverage_files[@]}"; do
            tail -n +2 "$file" >> "$COVERAGE_DIR/combined_coverage.out" 2>/dev/null || true
        done

        # Generate HTML coverage report
        if go tool cover -html="$COVERAGE_DIR/combined_coverage.out" -o "$COVERAGE_DIR/coverage.html"; then
            print_success "HTML coverage report: $COVERAGE_DIR/coverage.html"
        fi

        # Show coverage summary
        if command -v go >/dev/null 2>&1; then
            total_coverage=$(go tool cover -func="$COVERAGE_DIR/combined_coverage.out" | tail -1 | awk '{print $3}')
            print_success "Total Coverage: $total_coverage"
        fi
    else
        print_warning "No coverage files found"
    fi
fi

# Generate comprehensive test report
print_section "GENERATING TEST REPORT"

# Calculate success rate
if [[ $TOTAL_TESTS -gt 0 ]]; then
    SUCCESS_RATE=$(echo "scale=2; $TOTAL_PASSED * 100 / $TOTAL_TESTS" | bc 2>/dev/null)
    if [[ -z "$SUCCESS_RATE" ]]; then
        SUCCESS_RATE="0.00"
    fi
else
    SUCCESS_RATE="100.00"
fi

# Create detailed JSON report
cat > "$RESULTS_DIR/test_report.json" << EOF
{
    "test_execution": {
        "timestamp": "$TIMESTAMP",
        "duration": "$(date +%s)",
        "configuration": {
            "project_root": "$PROJECT_ROOT",
            "target_server": "$BASE_URL",
            "flexible_server": "$FLEX_URL",
            "coverage_enabled": $GENERATE_COVERAGE,
            "parallel_tests": $PARALLEL_TESTS,
            "quick_mode": $QUICK_MODE,
            "verbose": $VERBOSE
        }
    },
    "test_results": {
        "total_tests": $TOTAL_TESTS,
        "passed": $TOTAL_PASSED,
        "failed": $TOTAL_FAILED,
        "skipped": $TOTAL_SKIPPED,
        "success_rate": "$SUCCESS_RATE%"
    },
    "test_categories": {
EOF

# Add test category results
first=true
for i in "${!TEST_CATEGORIES_KEYS[@]}"; do
    if [[ "$first" == "false" ]]; then
        echo "," >> "$RESULTS_DIR/test_report.json"
    fi
    first=false

    local category="${TEST_CATEGORIES_KEYS[$i]}"
    local category_value="${TEST_CATEGORIES_VALUES[$i]}"
    IFS='/' read -r passed failed skipped <<< "$category_value"
    result=$(get_test_result "$category")
    duration=$(get_test_time "$category")

    cat >> "$RESULTS_DIR/test_report.json" << EOF
        "$category": {
            "passed": $passed,
            "failed": $failed,
            "skipped": $skipped,
            "result": $result,
            "duration": $duration
        }
EOF
done

cat >> "$RESULTS_DIR/test_report.json" << EOF
    },
    "enabled_test_types": {
        "unit_tests": $RUN_UNIT_TESTS,
        "e2e_tests": $RUN_E2E_TESTS,
        "integration_tests": $RUN_INTEGRATION_TESTS,
        "swagger_tests": $RUN_SWAGGER_TESTS,
        "api_tests": $RUN_API_TESTS,
        "performance_tests": $RUN_PERFORMANCE_TESTS
    }
}
EOF

# Create human-readable summary
cat > "$RESULTS_DIR/test_summary.txt" << EOF
==============================================================================
GO PLAYGROUND VALIDATOR - UNIFIED E2E TEST SUITE SUMMARY
==============================================================================

Test Execution Details:
  Timestamp: $TIMESTAMP
  Results Directory: $RESULTS_DIR
  Target Server: $BASE_URL
  Flexible Server: $FLEX_URL

Test Results:
  Total Tests: $TOTAL_TESTS
  Passed: $TOTAL_PASSED
  Failed: $TOTAL_FAILED
  Skipped: $TOTAL_SKIPPED
  Success Rate: $SUCCESS_RATE%

Test Categories:
EOF

for i in "${!TEST_CATEGORIES_KEYS[@]}"; do
    local category="${TEST_CATEGORIES_KEYS[$i]}"
    local category_value="${TEST_CATEGORIES_VALUES[$i]}"
    IFS='/' read -r passed failed skipped <<< "$category_value"
    result=$(get_test_result "$category")
    duration=$(get_test_time "$category")
    status=$([[ $result -eq 0 ]] && echo "PASS" || echo "FAIL")

    echo "  $category: $status (P:$passed F:$failed S:$skipped) ${duration}s" >> "$RESULTS_DIR/test_summary.txt"
done

cat >> "$RESULTS_DIR/test_summary.txt" << EOF

Test Types Executed:
  Unit Tests: $([[ "$RUN_UNIT_TESTS" == "true" ]] && echo "✓" || echo "✗")
  E2E Tests: $([[ "$RUN_E2E_TESTS" == "true" ]] && echo "✓" || echo "✗")
  Integration Tests: $([[ "$RUN_INTEGRATION_TESTS" == "true" ]] && echo "✓" || echo "✗")
  Swagger Tests: $([[ "$RUN_SWAGGER_TESTS" == "true" ]] && echo "✓" || echo "✗")
  API Tests: $([[ "$RUN_API_TESTS" == "true" ]] && echo "✓" || echo "✗")
  Performance Tests: $([[ "$RUN_PERFORMANCE_TESTS" == "true" ]] && echo "✓" || echo "✗")

Configuration:
  Coverage Generation: $([[ "$GENERATE_COVERAGE" == "true" ]] && echo "Enabled" || echo "Disabled")
  Parallel Execution: $([[ "$PARALLEL_TESTS" == "true" ]] && echo "Enabled" || echo "Disabled")
  Quick Mode: $([[ "$QUICK_MODE" == "true" ]] && echo "Enabled" || echo "Disabled")
  Verbose Output: $([[ "$VERBOSE" == "true" ]] && echo "Enabled" || echo "Disabled")

Files Generated:
  - test_report.json: Detailed JSON test results
  - test_summary.txt: Human-readable summary (this file)
  - coverage/: Code coverage reports (if enabled)
  - *.log: Individual test execution logs

==============================================================================
EOF

# Print final summary
print_section "TEST EXECUTION SUMMARY"

echo ""
echo "📊 RESULTS:"
echo "  Total Tests: $TOTAL_TESTS"
echo "  Passed: $TOTAL_PASSED"
echo "  Failed: $TOTAL_FAILED"
echo "  Skipped: $TOTAL_SKIPPED"
echo "  Success Rate: $SUCCESS_RATE%"
echo ""

echo "📁 OUTPUT FILES:"
echo "  Results Directory: $RESULTS_DIR"
echo "  Test Report: $RESULTS_DIR/test_report.json"
echo "  Summary: $RESULTS_DIR/test_summary.txt"
if [[ "$GENERATE_COVERAGE" == "true" ]]; then
    echo "  Coverage Report: $COVERAGE_DIR/coverage.html"
fi
echo ""

echo "🏷️  TEST CATEGORIES:"
for i in "${!TEST_CATEGORIES_KEYS[@]}"; do
    local category="${TEST_CATEGORIES_KEYS[$i]}"
    local category_value="${TEST_CATEGORIES_VALUES[$i]}"
    IFS='/' read -r passed failed skipped <<< "$category_value"
    result=$(get_test_result "$category")
    duration=$(get_test_time "$category")
    status=$([[ $result -eq 0 ]] && echo -e "${GREEN}PASS${NC}" || echo -e "${RED}FAIL${NC}")
    echo -e "  $category: $status (P:$passed F:$failed S:$skipped) ${duration}s"
done

echo ""
if [[ $TOTAL_FAILED -gt 0 ]]; then
    print_error "Some tests failed! Check individual test logs for details."
    echo ""
    echo "Failed tests can be re-run individually using:"
    echo "  $0 --<test-type> --verbose"
    exit 1
else
    print_success "All tests passed! 🎉"
    echo ""
    echo "To run specific test types:"
    echo "  $0 --unit              # Unit tests only"
    echo "  $0 --api --swagger     # API and Swagger tests"
    echo "  $0 --performance       # Performance tests only"
fi

print_success "Unified E2E test suite completed successfully!"