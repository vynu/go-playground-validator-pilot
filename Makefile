# Makefile for Go Playground Validator - Multi-platform Build & Deploy

# Configuration
DOCKER_IMAGE := go-playground-validator
DOCKER_TAG := latest

# Go Configuration
GO_MODULE := goplayground-data-validator
SRC_DIR := src
BIN_DIR := bin
BINARY_NAME := validator
MAIN_FILE := main.go


# Create bin directory if it doesn't exist
$(shell mkdir -p $(BIN_DIR))

# Test Configuration
TEST_TIMEOUT := 30s
COVERAGE_FILE := coverage.out
COVERAGE_HTML := coverage.html

# Output directories
TEST_RESULTS_DIR := test_results
LOGS_DIR := logs

# Default target
.DEFAULT_GOAL := help

# Colors for output
RED := \033[31m
GREEN := \033[32m
YELLOW := \033[33m
BLUE := \033[34m
RESET := \033[0m

##@ Go Module Commands

.PHONY: mod-download
mod-download: ## Download Go module dependencies
	@echo "$(BLUE)Downloading Go module dependencies...$(RESET)"
	cd $(SRC_DIR) && go mod download
	@echo "$(GREEN)✓ Dependencies downloaded$(RESET)"

.PHONY: mod-tidy
mod-tidy: ## Tidy Go module dependencies
	@echo "$(BLUE)Tidying Go module dependencies...$(RESET)"
	cd $(SRC_DIR) && go mod tidy
	@echo "$(GREEN)✓ Dependencies tidied$(RESET)"

.PHONY: mod-verify
mod-verify: ## Verify Go module dependencies
	@echo "$(BLUE)Verifying Go module dependencies...$(RESET)"
	cd $(SRC_DIR) && go mod verify
	@echo "$(GREEN)✓ Dependencies verified$(RESET)"


##@ Build Commands

.PHONY: build
build: mod-tidy ## Build binary for current platform
	@echo "$(BLUE)Building binary for current platform...$(RESET)"
	cd $(SRC_DIR) && go build -o ../$(BIN_DIR)/$(BINARY_NAME) $(MAIN_FILE)
	@echo "$(GREEN)✓ Binary built: $(BIN_DIR)/$(BINARY_NAME)$(RESET)"

.PHONY: build-linux
build-linux: mod-tidy ## Build for Linux
	@echo "$(BLUE)Building Linux binary...$(RESET)"
	cd $(SRC_DIR) && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags='-w -s' -o ../$(BIN_DIR)/$(BINARY_NAME)-linux $(MAIN_FILE)
	@echo "$(GREEN)✓ Linux binary built: $(BIN_DIR)/$(BINARY_NAME)-linux$(RESET)"

.PHONY: build-all
build-all: ## Build for all major platforms
	@echo "$(BLUE)Building for all platforms...$(RESET)"
	cd $(SRC_DIR) && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags='-w -s' -o ../$(BIN_DIR)/$(BINARY_NAME)-linux $(MAIN_FILE)
	cd $(SRC_DIR) && CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags='-w -s' -o ../$(BIN_DIR)/$(BINARY_NAME)-darwin $(MAIN_FILE)
	cd $(SRC_DIR) && CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags='-w -s' -o ../$(BIN_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_FILE)
	cd $(SRC_DIR) && CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags='-w -s' -o ../$(BIN_DIR)/$(BINARY_NAME)-windows.exe $(MAIN_FILE)
	@echo "$(GREEN)✓ All platform binaries built$(RESET)"

.PHONY: clean-binary
clean-binary: ## Clean built binaries
	@echo "$(YELLOW)Cleaning built binaries...$(RESET)"
	rm -rf $(BIN_DIR)
	@echo "$(GREEN)✓ Binaries cleaned$(RESET)"

##@ Docker Commands

.PHONY: docker-build
docker-build: ## Build Docker image (distroless)
	@echo "$(BLUE)Building Docker image...$(RESET)"
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "$(GREEN)✓ Docker build completed: $(DOCKER_IMAGE):$(DOCKER_TAG)$(RESET)"

.PHONY: docker-build-alpine
docker-build-alpine: ## Build Alpine image
	@echo "$(BLUE)Building Alpine Docker image...$(RESET)"
	docker build --target alpine -t $(DOCKER_IMAGE):alpine .
	@echo "$(GREEN)✓ Alpine Docker build completed$(RESET)"

##@ Unit Testing Commands

# Internal helper function for tests
define run_tests
	@echo "$(BLUE)Running $(1) tests...$(RESET)"
	cd $(SRC_DIR) && go test -timeout $(TEST_TIMEOUT) $(2) $(3)
	@echo "$(GREEN)✓ $(1) tests completed$(RESET)"
endef

.PHONY: test
test: mod-tidy ## Run all unit tests
	$(call run_tests,all unit,-v,./...)

.PHONY: test-coverage
test-coverage: mod-tidy ## Run tests with coverage report
	@echo "$(BLUE)Running tests with coverage...$(RESET)"
	@mkdir -p coverage
	cd $(SRC_DIR) && go test -timeout $(TEST_TIMEOUT) -coverprofile=../$(COVERAGE_FILE) -covermode=atomic ./...
	cd $(SRC_DIR) && go tool cover -html=../$(COVERAGE_FILE) -o ../$(COVERAGE_HTML)
	@echo "$(GREEN)✓ Coverage report generated: $(COVERAGE_HTML)$(RESET)"

.PHONY: test-race
test-race: mod-tidy ## Run tests with race detection
	$(call run_tests,race detection,-race -v,./...)

.PHONY: test-short
test-short: mod-tidy ## Run short tests only
	$(call run_tests,short,-short -v,./...)

.PHONY: test-models
test-models: mod-tidy ## Run model-specific tests
	$(call run_tests,model,-v,./models/...)

.PHONY: test-validations
test-validations: mod-tidy ## Run validation-specific tests
	$(call run_tests,validation,-v,./validations/...)

.PHONY: test-registry
test-registry: mod-tidy ## Run registry-specific tests
	$(call run_tests,registry,-v,./registry/...)

.PHONY: test-function
test-function: ## Run specific test function (usage: make test-function FUNC=TestFunctionName)
	@if [ -z "$(FUNC)" ]; then \
		echo "$(RED)Error: FUNC parameter required. Usage: make test-function FUNC=TestFunctionName$(RESET)"; \
		exit 1; \
	fi
	@echo "$(BLUE)Running test function: $(FUNC)...$(RESET)"
	cd $(SRC_DIR) && go test -timeout $(TEST_TIMEOUT) -v -run $(FUNC) ./...
	@echo "$(GREEN)✓ Test function $(FUNC) completed$(RESET)"

.PHONY: test-package
test-package: ## Run tests for specific package (usage: make test-package PKG=models)
	@if [ -z "$(PKG)" ]; then \
		echo "$(RED)Error: PKG parameter required. Usage: make test-package PKG=models$(RESET)"; \
		exit 1; \
	fi
	@echo "$(BLUE)Running tests for package: $(PKG)...$(RESET)"
	cd $(SRC_DIR) && go test -timeout $(TEST_TIMEOUT) -v ./$(PKG)/...
	@echo "$(GREEN)✓ Package $(PKG) tests completed$(RESET)"

.PHONY: benchmark
benchmark: mod-tidy ## Run benchmarks
	@echo "$(BLUE)Running benchmarks...$(RESET)"
	cd $(SRC_DIR) && go test -bench=. -benchmem -timeout $(TEST_TIMEOUT) ./...
	@echo "$(GREEN)✓ Benchmarks completed$(RESET)"

.PHONY: benchmark-package
benchmark-package: ## Run benchmarks for specific package (usage: make benchmark-package PKG=validations)
	@if [ -z "$(PKG)" ]; then \
		echo "$(RED)Error: PKG parameter required. Usage: make benchmark-package PKG=validations$(RESET)"; \
		exit 1; \
	fi
	@echo "$(BLUE)Running benchmarks for package: $(PKG)...$(RESET)"
	cd $(SRC_DIR) && go test -bench=. -benchmem -timeout $(TEST_TIMEOUT) ./$(PKG)/...
	@echo "$(GREEN)✓ Package $(PKG) benchmarks completed$(RESET)"

##@ E2E Testing Commands

.PHONY: test-e2e
test-e2e: build ## Run comprehensive E2E test suite (all validation scenarios including Phase 2)
	@echo "$(BLUE)Running E2E test suite (unit + validation + array + threshold + batch)...$(RESET)"
	chmod +x ./e2e_test_suite.sh
	PORT=8086 ./e2e_test_suite.sh
	@echo "$(GREEN)✓ E2E tests completed$(RESET)"

.PHONY: test-all
test-all: test test-race test-e2e ## Run all tests (unit + race + E2E)
	@echo "$(GREEN)✓ All tests completed successfully$(RESET)"

##@ Docker Development Commands

.PHONY: docker-dev
docker-dev: ## Start development environment
	@echo "$(BLUE)Starting Docker development environment...$(RESET)"
	docker-compose up -d validator-debug
	@echo "$(GREEN)✓ Development server running on http://localhost:8081$(RESET)"

.PHONY: docker-dev-logs
docker-dev-logs: ## Follow development logs
	docker-compose logs -f validator-debug

.PHONY: docker-dev-shell
docker-dev-shell: ## Get shell access to debug container
	docker-compose exec validator-debug /busybox/sh

##@ Docker Production Commands

.PHONY: docker-up
docker-up: docker-build ## Start production environment
	@echo "$(BLUE)Starting Docker production environment...$(RESET)"
	docker-compose up -d validator
	@echo "$(GREEN)✓ Production server running on http://localhost:8080$(RESET)"

.PHONY: docker-up-full
docker-up-full: docker-build ## Start full stack (with monitoring)
	@echo "$(BLUE)Starting full Docker production stack...$(RESET)"
	docker-compose --profile production up -d
	@echo "$(GREEN)✓ Full stack running:$(RESET)"
	@echo "  - Validator: http://localhost:8080"
	@echo "  - Traefik Dashboard: http://localhost:8090"
	@echo "  - Prometheus: http://localhost:9090"
	@echo "  - Grafana: http://localhost:3000 (admin/admin)"

.PHONY: docker-down
docker-down: ## Stop all Docker services
	@echo "$(YELLOW)Stopping all Docker services...$(RESET)"
	docker-compose down
	@echo "$(GREEN)✓ All Docker services stopped$(RESET)"

.PHONY: docker-down-volumes
docker-down-volumes: ## Stop all Docker services and remove volumes
	@echo "$(RED)Stopping all Docker services and removing volumes...$(RESET)"
	docker-compose down -v
	@echo "$(GREEN)✓ All Docker services stopped and volumes removed$(RESET)"

##@ Docker Run Commands (with Port Checking)

.PHONY: check-port
check-port: ## Check if port is available (usage: make check-port PORT=8080)
	@if [ -z "$(PORT)" ]; then \
		echo "$(RED)Error: PORT parameter required. Usage: make check-port PORT=8080$(RESET)"; \
		exit 1; \
	fi
	@echo "$(BLUE)Checking if port $(PORT) is available...$(RESET)"
	@if lsof -Pi :$(PORT) -sTCP:LISTEN -t >/dev/null 2>&1; then \
		echo "$(RED)Port $(PORT) is already in use!$(RESET)"; \
		echo "$(YELLOW)Process using port $(PORT):$(RESET)"; \
		lsof -Pi :$(PORT) -sTCP:LISTEN | head -10; \
		echo ""; \
		echo "$(BLUE)To kill the process, run:$(RESET)"; \
		echo "  make kill-port PORT=$(PORT)"; \
		echo ""; \
		echo "$(BLUE)Or manually kill with PID:$(RESET)"; \
		PID=$$(lsof -Pi :$(PORT) -sTCP:LISTEN -t | head -1); \
		echo "  kill $$PID"; \
		exit 1; \
	else \
		echo "$(GREEN)✓ Port $(PORT) is available$(RESET)"; \
	fi

.PHONY: kill-port
kill-port: ## Kill process using specified port (usage: make kill-port PORT=8080)
	@if [ -z "$(PORT)" ]; then \
		echo "$(RED)Error: PORT parameter required. Usage: make kill-port PORT=8080$(RESET)"; \
		exit 1; \
	fi
	@echo "$(BLUE)Killing process on port $(PORT)...$(RESET)"
	@if lsof -Pi :$(PORT) -sTCP:LISTEN -t >/dev/null 2>&1; then \
		echo "$(YELLOW)Process found on port $(PORT):$(RESET)"; \
		lsof -Pi :$(PORT) -sTCP:LISTEN; \
		echo ""; \
		PID=$$(lsof -Pi :$(PORT) -sTCP:LISTEN -t | head -1); \
		echo "$(BLUE)Killing PID: $$PID$(RESET)"; \
		kill $$PID 2>/dev/null || echo "$(YELLOW)Process may have already terminated$(RESET)"; \
		sleep 1; \
		if lsof -Pi :$(PORT) -sTCP:LISTEN -t >/dev/null 2>&1; then \
			echo "$(YELLOW)Process still running, force killing...$(RESET)"; \
			kill -9 $$PID 2>/dev/null || true; \
		fi; \
		echo "$(GREEN)✓ Process on port $(PORT) terminated$(RESET)"; \
	else \
		echo "$(GREEN)✓ No process found on port $(PORT)$(RESET)"; \
	fi

.PHONY: docker-run-distroless
docker-run-distroless: docker-build ## Run distroless container with port checking
	@echo "$(BLUE)Starting distroless container...$(RESET)"
	@$(MAKE) check-port PORT=8080 || { \
		echo "$(YELLOW)Port 8080 is occupied. Do you want to kill the process? (y/N)$(RESET)"; \
		read -p "Enter choice: " choice; \
		case "$$choice" in \
			[Yy]* ) $(MAKE) kill-port PORT=8080 && sleep 2 ;; \
			* ) echo "$(RED)Aborted. Please free port 8080 and try again$(RESET)"; exit 1 ;; \
		esac; \
	}
	@echo "$(BLUE)Running distroless container on port 8080...$(RESET)"
	docker run -d --name validator-distroless -p 8080:8080 $(DOCKER_IMAGE):$(DOCKER_TAG)
	@echo "$(GREEN)✓ Distroless container started: http://localhost:8080$(RESET)"
	@echo "$(BLUE)Health check:$(RESET)"
	@sleep 3 && curl -s http://localhost:8080/health | jq . || echo "$(YELLOW)Service starting up...$(RESET)"
	@echo ""
	@echo "$(BLUE)To stop the container:$(RESET) docker stop validator-distroless && docker rm validator-distroless"

.PHONY: docker-run-alpine
docker-run-alpine: docker-build-alpine ## Run Alpine container with port checking
	@echo "$(BLUE)Starting Alpine container...$(RESET)"
	@$(MAKE) check-port PORT=8081 || { \
		echo "$(YELLOW)Port 8081 is occupied. Do you want to kill the process? (y/N)$(RESET)"; \
		read -p "Enter choice: " choice; \
		case "$$choice" in \
			[Yy]* ) $(MAKE) kill-port PORT=8081 && sleep 2 ;; \
			* ) echo "$(RED)Aborted. Please free port 8081 and try again$(RESET)"; exit 1 ;; \
		esac; \
	}
	@echo "$(BLUE)Running Alpine container on port 8081...$(RESET)"
	docker run -d --name validator-alpine -p 8081:8080 $(DOCKER_IMAGE):alpine
	@echo "$(GREEN)✓ Alpine container started: http://localhost:8081$(RESET)"
	@echo "$(BLUE)Health check:$(RESET)"
	@sleep 3 && curl -s http://localhost:8081/health | jq . || echo "$(YELLOW)Service starting up...$(RESET)"
	@echo ""
	@echo "$(BLUE)To stop the container:$(RESET) docker stop validator-alpine && docker rm validator-alpine"

.PHONY: docker-run-custom
docker-run-custom: ## Run container on custom port with port checking (usage: make docker-run-custom PORT=9090 IMAGE=distroless)
	@if [ -z "$(PORT)" ]; then \
		echo "$(RED)Error: PORT parameter required. Usage: make docker-run-custom PORT=9090 IMAGE=distroless$(RESET)"; \
		exit 1; \
	fi
	@IMAGE_TYPE=$${IMAGE:-distroless}; \
	CONTAINER_NAME="validator-$$IMAGE_TYPE-$(PORT)"; \
	echo "$(BLUE)Starting $$IMAGE_TYPE container on port $(PORT)...$(RESET)"; \
	$(MAKE) check-port PORT=$(PORT) || { \
		echo "$(YELLOW)Port $(PORT) is occupied. Do you want to kill the process? (y/N)$(RESET)"; \
		read -p "Enter choice: " choice; \
		case "$$choice" in \
			[Yy]* ) $(MAKE) kill-port PORT=$(PORT) && sleep 2 ;; \
			* ) echo "$(RED)Aborted. Please free port $(PORT) and try again$(RESET)"; exit 1 ;; \
		esac; \
	}; \
	echo "$(BLUE)Running $$IMAGE_TYPE container: $$CONTAINER_NAME...$(RESET)"; \
	docker run -d --name $$CONTAINER_NAME -p $(PORT):8080 $(DOCKER_IMAGE):$$IMAGE_TYPE; \
	echo "$(GREEN)✓ Container started: http://localhost:$(PORT)$(RESET)"; \
	echo "$(BLUE)Health check:$(RESET)"; \
	sleep 3 && curl -s http://localhost:$(PORT)/health | jq . || echo "$(YELLOW)Service starting up...$(RESET)"; \
	echo ""; \
	echo "$(BLUE)To stop the container:$(RESET) docker stop $$CONTAINER_NAME && docker rm $$CONTAINER_NAME"

.PHONY: docker-stop-all-validators
docker-stop-all-validators: ## Stop and remove all validator containers
	@echo "$(BLUE)Stopping all validator containers...$(RESET)"
	@docker ps -a --filter "name=validator-" --format "{{.Names}}" | while read container; do \
		if [ ! -z "$$container" ]; then \
			echo "$(YELLOW)Stopping: $$container$(RESET)"; \
			docker stop $$container 2>/dev/null || true; \
			docker rm $$container 2>/dev/null || true; \
		fi; \
	done
	@echo "$(GREEN)✓ All validator containers stopped and removed$(RESET)"

.PHONY: docker-ps-validators
docker-ps-validators: ## Show all running validator containers
	@echo "$(BLUE)Running validator containers:$(RESET)"
	@docker ps --filter "name=validator-" --format "table {{.Names}}\t{{.Image}}\t{{.Status}}\t{{.Ports}}" || echo "$(YELLOW)No validator containers running$(RESET)"

##@ Docker Testing Commands

.PHONY: docker-test
docker-test: ## Run tests in Docker container
	@echo "$(BLUE)Running tests in Docker container...$(RESET)"
	docker-compose --profile test run --rm test-runner
	@echo "$(GREEN)✓ Docker tests completed$(RESET)"

.PHONY: docker-test-build
docker-test-build: ## Test Docker build process only
	@echo "$(BLUE)Testing Docker build process...$(RESET)"
	docker build --target builder -t $(DOCKER_IMAGE):test-build .
	@echo "$(GREEN)✓ Docker build test completed$(RESET)"

.PHONY: docker-test-e2e
docker-test-e2e: docker-build ## Run E2E tests against Docker container
	@echo "$(BLUE)Running E2E tests against Docker container...$(RESET)"
	@echo "$(BLUE)Starting container on port 8087...$(RESET)"
	@$(MAKE) kill-port PORT=8087 2>/dev/null || true
	@docker run -d --name validator-e2e-test -p 8087:8080 $(DOCKER_IMAGE):$(DOCKER_TAG)
	@echo "$(BLUE)Waiting for container to be ready...$(RESET)"
	@sleep 5
	@echo "$(BLUE)Running E2E tests...$(RESET)"
	@VALIDATOR_URL=http://localhost:8087 TEST_MODE=docker ./e2e_test_suite.sh || (docker stop validator-e2e-test && docker rm validator-e2e-test && exit 1)
	@echo "$(BLUE)Cleaning up test container...$(RESET)"
	@docker stop validator-e2e-test
	@docker rm validator-e2e-test
	@echo "$(GREEN)✓ Docker E2E tests completed$(RESET)"

.PHONY: docker-test-e2e-alpine
docker-test-e2e-alpine: docker-build-alpine ## Run E2E tests against Alpine Docker container
	@echo "$(BLUE)Running E2E tests against Alpine Docker container...$(RESET)"
	@echo "$(BLUE)Starting Alpine container on port 8087...$(RESET)"
	@$(MAKE) kill-port PORT=8087 2>/dev/null || true
	@docker run -d --name validator-e2e-test-alpine -p 8087:8080 $(DOCKER_IMAGE):alpine
	@echo "$(BLUE)Waiting for container to be ready...$(RESET)"
	@sleep 5
	@echo "$(BLUE)Running E2E tests...$(RESET)"
	@VALIDATOR_URL=http://localhost:8087 TEST_MODE=docker ./e2e_test_suite.sh || (docker stop validator-e2e-test-alpine && docker rm validator-e2e-test-alpine && exit 1)
	@echo "$(BLUE)Cleaning up test container...$(RESET)"
	@docker stop validator-e2e-test-alpine
	@docker rm validator-e2e-test-alpine
	@echo "$(GREEN)✓ Docker E2E tests (Alpine) completed$(RESET)"

.PHONY: docker-test-compose
docker-test-compose: ## Run E2E tests using docker-compose
	@echo "$(BLUE)Running E2E tests using docker-compose...$(RESET)"
	@docker-compose up -d validator
	@echo "$(BLUE)Waiting for service to be ready...$(RESET)"
	@sleep 5
	@echo "$(BLUE)Running E2E tests...$(RESET)"
	@VALIDATOR_URL=http://localhost:8080 TEST_MODE=docker ./e2e_test_suite.sh || (docker-compose down && exit 1)
	@docker-compose down
	@echo "$(GREEN)✓ Docker compose E2E tests completed$(RESET)"

.PHONY: docker-benchmark
docker-benchmark: ## Run benchmarks in Docker
	@echo "$(BLUE)Running benchmarks in Docker...$(RESET)"
	docker run --rm $(DOCKER_IMAGE):test-build go test -bench=. -benchmem ./...

##@ Health & Monitoring Commands

.PHONY: health
health: ## Check health of running services
	@echo "$(BLUE)Checking service health...$(RESET)"
	@docker-compose ps
	@echo "\n$(BLUE)Health check results:$(RESET)"
	@curl -s http://localhost:8080/health | jq . || echo "Service not responding"

.PHONY: logs
logs: ## View logs from all services
	docker-compose logs -f

.PHONY: logs-validator
logs-validator: ## View validator logs only
	docker-compose logs -f validator

.PHONY: stats
stats: ## Show container resource usage
	docker stats --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}\t{{.BlockIO}}"


##@ Utility Commands

.PHONY: size
size: ## Show image sizes
	@echo "$(BLUE)Image sizes:$(RESET)"
	@docker images $(DOCKER_IMAGE) --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}\t{{.CreatedAt}}"

.PHONY: inspect
inspect: ## Inspect the main image
	@echo "$(BLUE)Image inspection:$(RESET)"
	docker inspect $(DOCKER_IMAGE):$(DOCKER_TAG) | jq '.[]'

.PHONY: security
security: ## Run security scan (requires docker scan or trivy)
	@echo "$(BLUE)Running security scan...$(RESET)"
	@if command -v trivy >/dev/null 2>&1; then \
		trivy image $(DOCKER_IMAGE):$(DOCKER_TAG); \
	elif docker scan --help >/dev/null 2>&1; then \
		docker scan $(DOCKER_IMAGE):$(DOCKER_TAG); \
	else \
		echo "$(YELLOW)No security scanner found. Install trivy or enable docker scan$(RESET)"; \
	fi

##@ Development Workflow Commands

.PHONY: setup
setup: mod-download mod-tidy ## Setup development environment
	@echo "$(GREEN)✓ Development environment setup completed$(RESET)"

.PHONY: dev-run
dev-run: build ## Run the application locally
	@echo "$(BLUE)Starting local development server...$(RESET)"
	PORT=8080 ./$(BIN_DIR)/$(BINARY_NAME)

.PHONY: dev-test
dev-test: test-short test-coverage ## Quick development test cycle
	@echo "$(GREEN)✓ Development test cycle completed$(RESET)"

.PHONY: ci
ci: mod-verify test test-race build ## CI/CD pipeline simulation
	@echo "$(GREEN)✓ CI pipeline simulation completed$(RESET)"

.PHONY: release
release: test-all build-all docker-build security ## Full release preparation
	@echo "$(GREEN)✓ Release preparation completed$(RESET)"

##@ Clean Commands

.PHONY: clean
clean: ## Clean binaries and test artifacts
	@echo "$(YELLOW)Cleaning build artifacts...$(RESET)"
	rm -rf $(BIN_DIR) coverage coverage*.* test_results*
	find . -name "*.test" -type f -delete 2>/dev/null || true
	@echo "$(GREEN)✓ Build artifacts cleaned$(RESET)"

.PHONY: clean-docker
clean-docker: ## Clean Docker artifacts for this project only
	@echo "$(YELLOW)Cleaning project Docker artifacts...$(RESET)"
	docker-compose down --volumes 2>/dev/null || true
	@echo "$(BLUE)Removing project images...$(RESET)"
	docker rmi $(DOCKER_IMAGE):$(DOCKER_TAG) 2>/dev/null || true
	docker rmi $(DOCKER_IMAGE):alpine 2>/dev/null || true
	docker rmi $(DOCKER_IMAGE):test-build 2>/dev/null || true
	@echo "$(BLUE)Stopping and removing project containers...$(RESET)"
	@docker ps -a --filter "name=validator-" --format "{{.Names}}" | while read container; do \
		if [ ! -z "$$container" ]; then \
			echo "$(YELLOW)  Removing: $$container$(RESET)"; \
			docker stop $$container 2>/dev/null || true; \
			docker rm $$container 2>/dev/null || true; \
		fi; \
	done
	@echo "$(GREEN)✓ Project Docker artifacts cleaned$(RESET)"

.PHONY: clean-all
clean-all: clean clean-docker ## Clean everything
	@echo "$(YELLOW)Deep cleaning...$(RESET)"
	cd $(SRC_DIR) && go clean -cache -testcache 2>/dev/null || true
	find . -name ".DS_Store" -type f -delete 2>/dev/null || true
	@echo "$(GREEN)✓ All artifacts cleaned$(RESET)"

##@ Help

.PHONY: help
help: ## Display this help message
	@awk 'BEGIN {FS = ":.*##"; printf "\n$(BLUE)Go Playground Validator - Development & Docker Operations$(RESET)\n\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  $(GREEN)%-20s$(RESET) %s\n", $$1, $$2 } /^##@/ { printf "\n$(YELLOW)%s$(RESET)\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
	@echo ""
	@echo "$(BLUE)Quick Start Examples:$(RESET)"
	@echo "  $(GREEN)Development:$(RESET)"
	@echo "    make setup               # Setup development environment"
	@echo "    make dev-run             # Run application locally"
	@echo "    make test                # Run unit tests"
	@echo "    make test-e2e            # Run E2E tests"
	@echo ""
	@echo "  $(GREEN)Docker Development:$(RESET)"
	@echo "    make docker-dev          # Start Docker dev environment"
	@echo "    make docker-dev-logs     # View development logs"
	@echo "    make docker-dev-shell    # Access debug container"
	@echo ""
	@echo "  $(GREEN)Production:$(RESET)"
	@echo "    make docker-build        # Build production image"
	@echo "    make docker-up           # Start production stack"
	@echo "    make docker-up-full      # Start with monitoring"
	@echo ""
	@echo "  $(GREEN)Docker Run (with Port Checking):$(RESET)"
	@echo "    make docker-run-distroless      # Run distroless on port 8080"
	@echo "    make docker-run-alpine          # Run Alpine on port 8081"
	@echo "    make docker-run-custom PORT=9090 IMAGE=distroless  # Custom port"
	@echo "    make docker-ps-validators       # Show running containers"
	@echo "    make docker-stop-all-validators # Stop all containers"
	@echo ""
	@echo "  $(GREEN)Testing:$(RESET)"
	@echo "    make test-models         # Test specific models"
	@echo "    make test-function FUNC=TestName  # Test specific function"
	@echo "    make test-package PKG=validations # Test specific package"
	@echo "    make test-all            # Run all tests"
	@echo ""
	@echo "  $(GREEN)Utilities:$(RESET)"
	@echo "    make health              # Check service health"
	@echo "    make security            # Run security scan"
	@echo "    make clean-all           # Clean everything"