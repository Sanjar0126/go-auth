# go-fr-project/Makefile

.PHONY: build clean test run run-api run-worker migrate-up migrate-down help

# Go build flags
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
BUILD_DIR ?= ./build
BIN_DIR ?= $(BUILD_DIR)/bin

# Database settings
DB_URL ?= postgresql://sanjar:npg_oMOWeyCXh7a0@ep-young-glitter-a2q6k58c-pooler.eu-central-1.aws.neon.tech/my-db?sslmode=require

# Version info
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags
GO_BUILD_FLAGS = -ldflags "-X main.version=$(VERSION) -X main.gitCommit=$(GIT_COMMIT) -X main.buildDate=$(BUILD_DATE)"

# Packages
API_PACKAGE = ./cmd/api
WORKER_PACKAGE = ./cmd/worker
MIGRATIONS_PACKAGE = ./cmd/migrations

help:
	@echo "Usage:"
	@echo "  make build         - Build all binaries"
	@echo "  make clean         - Remove build artifacts"
	@echo "  make test          - Run tests"
	@echo "  make run-api       - Run API server"
	@echo "  make run-worker    - Run background worker"
	@echo "  make migrate-up    - Run database migrations (up)"
	@echo "  make migrate-down  - Rollback database migrations"
	@echo "  make help          - Show this help"

build:
	@echo "Building binaries..."
	@mkdir -p $(BIN_DIR)
	@echo "Building API server..."
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(GO_BUILD_FLAGS) -o $(BIN_DIR)/api $(API_PACKAGE)
	@echo "Building worker..."
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(GO_BUILD_FLAGS) -o $(BIN_DIR)/worker $(WORKER_PACKAGE)
	@echo "Building migrations tool..."
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(GO_BUILD_FLAGS) -o $(BIN_DIR)/migrations $(MIGRATIONS_PACKAGE)
	@echo "Done."

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@go clean
	@echo "Done."

test:
	@echo "Running tests..."
	@go test -v ./...

run-api:
	@echo "Running API server..."
	@go run $(API_PACKAGE) --db-url="$(DB_URL)"

run-worker:
	@echo "Running background worker..."
	@go run $(WORKER_PACKAGE) --db-url="$(DB_URL)"

migrate-up:
	@echo "Running database migrations (up)..."
	@go run $(MIGRATIONS_PACKAGE) --db-url="$(DB_URL)" --command=up

migrate-down:
	@echo "Rolling back database migrations..."
	@go run $(MIGRATIONS_PACKAGE) --db-url="$(DB_URL)" --command=down

# Create a new migration file
create-migration:
	@read -p "Enter migration name: " name; \
	timestamp=$$(date +%Y%m%d%H%M%S); \
	mkdir -p migrations; \
	touch migrations/$${timestamp}_$${name}.up.sql; \
	touch migrations/$${timestamp}_$${name}.down.sql; \
	echo "Created new migration: migrations/$${timestamp}_$${name}.up.sql and migrations/$${timestamp}_$${name}.down.sql"
