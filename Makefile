# Makefile for cruise-price-compare
# Usage: make <target>

.PHONY: all proto-gen build test lint clean help dev run migrate web-dev web-build

# Default target
all: proto-gen build test lint

# =============================================================================
# Variables
# =============================================================================
GO := go
PROTOC := protoc
NPM := npm

PROTO_DIR := api/proto
PROTO_GO_OUT := api/gen/go
PROTO_TS_OUT := api/gen/ts

BINARY_NAME := cruise-server
WORKER_BINARY := cruise-worker
MIGRATE_BINARY := cruise-migrate

BUILD_DIR := bin
WEB_DIR := web

# =============================================================================
# Protobuf Generation
# =============================================================================
proto-gen: proto-go proto-ts

proto-go:
	@echo "Generating Go code from proto files..."
	@mkdir -p $(PROTO_GO_OUT)
	$(PROTOC) \
		--proto_path=$(PROTO_DIR) \
		--go_out=$(PROTO_GO_OUT) \
		--go_opt=paths=source_relative \
		$(PROTO_DIR)/*.proto
	@echo "Go proto generation complete."

proto-ts:
	@echo "Generating TypeScript types from proto files..."
	@mkdir -p $(PROTO_TS_OUT)
	@cd $(WEB_DIR) && $(NPM) run proto:gen || echo "Run 'npm install' in web/ first"
	@echo "TypeScript proto generation complete."

# =============================================================================
# Build Targets
# =============================================================================
build: build-server build-worker build-migrate

build-server:
	@echo "Building server..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/server
	@echo "Server built: $(BUILD_DIR)/$(BINARY_NAME)"

build-worker:
	@echo "Building worker..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build -o $(BUILD_DIR)/$(WORKER_BINARY) ./cmd/worker
	@echo "Worker built: $(BUILD_DIR)/$(WORKER_BINARY)"

build-migrate:
	@echo "Building migrate tool..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build -o $(BUILD_DIR)/$(MIGRATE_BINARY) ./cmd/migrate
	@echo "Migrate tool built: $(BUILD_DIR)/$(MIGRATE_BINARY)"

# =============================================================================
# Development
# =============================================================================
dev: 
	@echo "Starting development server..."
	$(GO) run ./cmd/server

run: dev

worker:
	@echo "Starting worker..."
	$(GO) run ./cmd/worker

migrate:
	@echo "Running migrations..."
	$(GO) run ./cmd/migrate

migrate-down:
	@echo "Rolling back migrations..."
	$(GO) run ./cmd/migrate -down

# =============================================================================
# Frontend
# =============================================================================
web-dev:
	@echo "Starting frontend dev server..."
	@cd $(WEB_DIR) && $(NPM) run dev

web-build:
	@echo "Building frontend..."
	@cd $(WEB_DIR) && $(NPM) run build

web-install:
	@echo "Installing frontend dependencies..."
	@cd $(WEB_DIR) && $(NPM) install

# =============================================================================
# Testing
# =============================================================================
test:
	@echo "Running tests..."
	$(GO) test -v -race -cover ./...

test-coverage:
	@echo "Running tests with coverage..."
	$(GO) test -v -race -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

test-web:
	@echo "Running frontend tests..."
	@cd $(WEB_DIR) && $(NPM) run test

# =============================================================================
# Linting
# =============================================================================
lint: lint-go lint-web

lint-go:
	@echo "Linting Go code..."
	golangci-lint run ./...

lint-web:
	@echo "Linting frontend code..."
	@cd $(WEB_DIR) && $(NPM) run lint

fmt:
	@echo "Formatting Go code..."
	$(GO) fmt ./...
	goimports -w .

fmt-web:
	@echo "Formatting frontend code..."
	@cd $(WEB_DIR) && $(NPM) run format

# =============================================================================
# Docker
# =============================================================================
docker-up:
	@echo "Starting Docker services..."
	docker-compose up -d

docker-down:
	@echo "Stopping Docker services..."
	docker-compose down

docker-logs:
	docker-compose logs -f

# =============================================================================
# Clean
# =============================================================================
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -rf $(PROTO_GO_OUT)/*.pb.go
	rm -rf $(PROTO_TS_OUT)/*.ts
	rm -f coverage.out coverage.html
	@echo "Clean complete."

# =============================================================================
# Help
# =============================================================================
help:
	@echo "Cruise Price Compare - Available targets:"
	@echo ""
	@echo "  proto-gen      Generate code from proto files (Go + TS)"
	@echo "  proto-go       Generate Go code from proto files"
	@echo "  proto-ts       Generate TypeScript types from proto files"
	@echo ""
	@echo "  build          Build all binaries"
	@echo "  build-server   Build API server"
	@echo "  build-worker   Build job worker"
	@echo "  build-migrate  Build migration tool"
	@echo ""
	@echo "  dev/run        Run API server in dev mode"
	@echo "  worker         Run job worker"
	@echo "  migrate        Run database migrations"
	@echo "  migrate-down   Rollback migrations"
	@echo ""
	@echo "  web-dev        Start frontend dev server"
	@echo "  web-build      Build frontend for production"
	@echo "  web-install    Install frontend dependencies"
	@echo ""
	@echo "  test           Run all Go tests"
	@echo "  test-coverage  Run tests with coverage report"
	@echo "  test-web       Run frontend tests"
	@echo ""
	@echo "  lint           Run all linters"
	@echo "  lint-go        Run Go linter"
	@echo "  lint-web       Run frontend linter"
	@echo "  fmt            Format Go code"
	@echo "  fmt-web        Format frontend code"
	@echo ""
	@echo "  docker-up      Start Docker services"
	@echo "  docker-down    Stop Docker services"
	@echo "  docker-logs    View Docker logs"
	@echo ""
	@echo "  clean          Remove build artifacts"
	@echo "  help           Show this help message"
