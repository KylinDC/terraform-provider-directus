.PHONY: help build test test-unit test-coverage test-acceptance test-acceptance-run \
	test-e2e test-e2e-basic-run test-e2e-comprehensive test-e2e-comprehensive-run test-e2e-quick test-all \
	docker-up docker-down docker-wait docker-logs docker-logs-tail docker-status docker-clean \
	setup e2e-setup clean install fmt lint check-mod-tidy deps all ci \
	release release-dry-run release-tag

# Default target
.DEFAULT_GOAL := help

# Variables
BINARY_NAME=terraform-provider-directus
BUILD_DIR=bin
PROVIDER_DIR=$(HOME)/.terraform.d/plugins/registry.terraform.io/kylindc/directus/0.1.0

# Detect OS and architecture
OS := $(shell uname -s | tr '[:upper:]' '[:lower:]')
ARCH := $(shell uname -m)

ifeq ($(OS),darwin)
	OS_NAME := darwin
endif
ifeq ($(OS),linux)
	OS_NAME := linux
endif

ifeq ($(ARCH),x86_64)
	ARCH_NAME := amd64
endif
ifeq ($(ARCH),aarch64)
	ARCH_NAME := arm64
endif
ifeq ($(ARCH),arm64)
	ARCH_NAME := arm64
endif

PROVIDER_PATH=$(PROVIDER_DIR)/$(OS_NAME)_$(ARCH_NAME)

# ==============================================================================
# Help
# ==============================================================================

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-30s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# ==============================================================================
# Build
# ==============================================================================

build: ## Build the provider binary
	@echo "Building provider..."
	@go build -o $(BINARY_NAME) .
	@echo "✓ Provider built: $(BINARY_NAME)"

install: build ## Build and install provider locally for Terraform
	@echo "Installing provider to $(PROVIDER_PATH)..."
	@mkdir -p $(PROVIDER_PATH)
	@cp $(BINARY_NAME) $(PROVIDER_PATH)/
	@chmod +x $(PROVIDER_PATH)/$(BINARY_NAME)
	@echo "✓ Provider installed"

# ==============================================================================
# Code Quality
# ==============================================================================

fmt: ## Format Go code
	@echo "Formatting code..."
	@go fmt ./...
	@echo "✓ Code formatted"

lint: ## Run linter (go vet)
	@echo "Running linter..."
	@go vet ./...
	@echo "✓ Linting complete"

check-mod-tidy: ## Verify go.mod and go.sum are tidy
	@echo "Checking go mod tidy..."
	@go mod tidy
	@git diff --exit-code go.mod go.sum
	@echo "✓ go.mod and go.sum are tidy"

deps: ## Download and tidy dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "✓ Dependencies updated"

# ==============================================================================
# Testing — CI-friendly targets (no docker/setup dependencies)
# ==============================================================================

test: test-unit ## Alias: run unit tests

test-unit: ## Run unit tests with race detection
	@echo "Running unit tests..."
	@go test ./... -v -race -count=1
	@echo "✓ Unit tests passed"

test-coverage: ## Run unit tests and generate coverage report
	@echo "Generating coverage report..."
	@go test ./... -v -race -coverprofile=coverage.out -covermode=atomic
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report generated: coverage.html"
	@go tool cover -func=coverage.out | grep total | awk '{print "  Total coverage: " $$3}'

test-acceptance-run: ## Run acceptance tests (assumes DIRECTUS_ENDPOINT and DIRECTUS_TOKEN are set)
	@echo "Running acceptance tests..."
	@TF_ACC=1 go test ./internal/provider/ -v -timeout 30m -run TestAcc -count=1
	@echo "✓ Acceptance tests passed"

test-e2e-basic-run: ## Run basic E2E tests (assumes Directus is running, .env exists, provider installed)
	@echo "Running basic E2E tests..."
	@./scripts/test-e2e.sh
	@echo "✓ Basic E2E tests passed"

test-e2e-comprehensive-run: ## Run comprehensive E2E tests (assumes Directus is running, .env exists, provider installed)
	@echo "Running comprehensive E2E tests..."
	@./scripts/test-e2e-comprehensive.sh
	@echo "✓ Comprehensive E2E tests passed"

# ==============================================================================
# Testing — Local targets (start docker, setup token, then run tests)
# ==============================================================================

test-acceptance: docker-up setup ## Run acceptance tests locally (starts Directus first)
	@echo "Running acceptance tests..."
	@. ./.env && \
		export DIRECTUS_ENDPOINT=$$TEST_DIRECTUS_ENDPOINT && \
		export DIRECTUS_TOKEN=$$TEST_DIRECTUS_TOKEN && \
		$(MAKE) test-acceptance-run
	@echo "✓ Acceptance tests passed"

test-e2e: docker-up setup install ## Run basic E2E tests locally (starts Directus first)
	@./scripts/test-e2e.sh || ($(MAKE) docker-logs-tail && exit 1)
	@echo "✓ Basic E2E tests passed"

test-e2e-comprehensive: docker-up setup install ## Run comprehensive E2E tests locally (starts Directus first)
	@./scripts/test-e2e-comprehensive.sh || ($(MAKE) docker-logs-tail && exit 1)
	@echo "✓ Comprehensive E2E tests passed"

test-e2e-quick: install ## Run comprehensive E2E tests (assumes Directus already running)
	@echo "Running E2E tests (quick mode - no docker restart)..."
	@./scripts/test-e2e-comprehensive.sh
	@echo "✓ E2E tests passed"

test-all: test-unit test-acceptance test-e2e ## Run unit, acceptance, and basic E2E tests

# ==============================================================================
# Docker
# ==============================================================================

docker-up: ## Start Directus with Docker
	@echo "Starting Directus..."
	@docker compose up -d
	@echo "✓ Directus starting... (waiting for health checks)"
	@$(MAKE) docker-wait

docker-wait: ## Wait for Directus to be healthy
	@echo "Waiting for services to be healthy..."
	@max_attempts=60; \
	attempt=0; \
	while [ $$attempt -lt $$max_attempts ]; do \
		postgres_health=$$(docker inspect --format='{{.State.Health.Status}}' directus-db 2>/dev/null || echo "not running"); \
		directus_health=$$(docker inspect --format='{{.State.Health.Status}}' directus-cms 2>/dev/null || echo "not running"); \
		if [ "$$postgres_health" = "healthy" ] && [ "$$directus_health" = "healthy" ]; then \
			echo "✓ All services are healthy!"; \
			docker compose ps; \
			exit 0; \
		fi; \
		attempt=$$((attempt + 1)); \
		printf "\r[%d/%d] PostgreSQL: %-12s | Directus: %-12s" $$attempt $$max_attempts "$$postgres_health" "$$directus_health"; \
		sleep 2; \
	done; \
	echo ""; \
	echo "⚠️  Timeout waiting for services"; \
	docker compose ps; \
	docker compose logs --tail=20 directus; \
	exit 1

docker-down: ## Stop Directus
	@echo "Stopping Directus..."
	@docker compose down
	@echo "✓ Directus stopped"

docker-logs: ## View Directus logs (follow mode)
	@docker compose logs -f directus

docker-logs-tail: ## View recent Directus logs (non-blocking)
	@docker compose logs --tail=50 directus

docker-status: ## Check Docker services status
	@docker compose ps

docker-clean: docker-down ## Stop and remove all data
	@echo "Removing volumes..."
	@docker compose down -v
	@echo "✓ All data removed"

# ==============================================================================
# Environment Setup
# ==============================================================================

setup: ## Setup Directus and create test token (writes .env)
	@echo "Setting up Directus test environment..."
	@chmod +x scripts/setup-directus.sh
	@./scripts/setup-directus.sh
	@echo "✓ Setup complete - .env file created with access token"

e2e-setup: docker-clean docker-up setup ## Fresh E2E environment setup
	@echo "✓ Fresh E2E environment ready!"

# ==============================================================================
# Release
# ==============================================================================

release-dry-run: ## Build release artifacts locally (no publish)
	@echo "Running GoReleaser dry-run..."
	@goreleaser release --snapshot --clean
	@echo "✓ Dry-run complete — artifacts in dist/"

release: lint test-unit ## Build and publish release via GoReleaser (requires GITHUB_TOKEN)
	@echo "Running GoReleaser release..."
	@goreleaser release --clean
	@echo "✓ Release published"

release-tag: ## Create and push a version tag (usage: make release-tag VERSION=0.2.0)
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: VERSION is required. Usage: make release-tag VERSION=0.2.0"; \
		exit 1; \
	fi
	@if ! echo "$(VERSION)" | grep -qE '^[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.]+)?$$'; then \
		echo "Error: VERSION must be valid semver (e.g. 0.2.0 or 1.0.0-beta1)"; \
		exit 1; \
	fi
	@echo "Creating tag v$(VERSION)..."
	@git tag -a "v$(VERSION)" -m "v$(VERSION)"
	@git push origin "v$(VERSION)"
	@echo "✓ Tag v$(VERSION) pushed — release workflow will run on GitHub Actions"

# ==============================================================================
# Housekeeping
# ==============================================================================

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@rm -rf $(BUILD_DIR)
	@rm -rf test-e2e test-e2e-comprehensive
	@rm -f coverage.out coverage.html
	@echo "✓ Cleaned"

# ==============================================================================
# Composite Targets
# ==============================================================================

all: fmt lint build test ## Run fmt, lint, build, and unit tests

ci: lint check-mod-tidy test-coverage build ## CI pipeline: lint, tidy check, test with coverage, build
