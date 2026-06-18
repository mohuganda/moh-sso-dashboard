# ============================================================
#                 📦 Project Variables
# ============================================================
APP_NAME = moh-sso-dashboard
BIN_DIR = bin
MAIN_FILE = ./cmd/server/main.go
GO_FILES = $(shell find . -type f -name '*.go' -not -path "./vendor/*")


VERSION := $(shell git describe --tags --always --dirty)
COMMIT := $(shell git rev-parse --short HEAD)
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := -s -w \

	-X 'github.com/mohuganda/moh-sso-dashboard/internal/version.Version=$(VERSION)' \
	-X 'github.com/mohuganda/moh-sso-dashboard/internal/version.Commit=$(COMMIT)' \
	-X 'github.com/mohuganda/moh-sso-dashboard/internal/version.BuildTime=$(BUILD_TIME)'

# ============================================================
#                 🔧 Configuration
# ============================================================
DOCKER_CONTAINER_NAME := backend-db
DB_NAME               := sso_dashboard
DB_URL                := postgresql://postgres:postgres@localhost:5432/$(DB_NAME)?sslmode=disable
MIGRATE_PATH          := internal/db/migrations
DOCKER_COMPOSE        := docker-compose.yml

# ============================================================
#                 🔨 Go Commands
# ============================================================
GO       = go
GOTEST   = $(GO) test
GOTIDY   = $(GO) mod tidy
GOINSTALL= $(GO) install
GOBUILD  = $(GO) build
GORUN    = $(GO) run
GOFMT    = gofmt
GOLINT   = golangci-lint run

.DEFAULT_GOAL := help

# ============================================================
#        🛠️ General Utility Commands
# ============================================================

help: ## Show available commands
	@echo ""
	@echo "📘 Available Commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?##' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""

tidy: ## Clean and tidy up go modules
	$(GOTIDY)

fmt: ## Format Go files
	@echo "Formatting code..."
	@$(GOFMT) -s -w $(GO_FILES)

lint: ## Run static analysis
	@echo "Running linter..."
	@$(GOLINT) || true

test: ## Run all tests
	@echo "Running tests..."
	@$(GOTEST) ./... -v

deps: ## Install go dependencies
	@echo "📦 Installing dependencies..."
	$(GO) mod download

verify: fmt lint test ## Run code checks before commit
	@echo "✅ All checks passed!"

# ============================================================
#        🚀 Build & Run
# ============================================================

build: ## Build the Go binary
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) -o $(BIN_DIR)/$(APP_NAME) $(MAIN_FILE)
	@echo "✅ Build complete: $(BIN_DIR)/$(APP_NAME)"

run: ## Run the Go application
	@echo "🚀 Starting $(APP_NAME)..."
	$(GORUN) $(MAIN_FILE)

clean: ## Remove binaries and temp files
	@echo "🧹 Cleaning up..."
	@rm -rf $(BIN_DIR)


version:
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Build Time: $(BUILD_TIME)"

# ============================================================
#        🐳 Docker Commands
# ============================================================

postgres: ## Run Postgres container
	docker run --name $(DOCKER_CONTAINER_NAME) \
		-p 5432:5432 \
		-e POSTGRES_USER=sitoowa \
		-e POSTGRES_PASSWORD=sitoowa123 \
		-d postgres:14.4-alpine

createdb: ## Create database inside container
	docker exec -it $(DOCKER_CONTAINER_NAME) \
		createdb --username=sitoowa --owner=sitoowa $(DB_NAME)

backupdb: ## Backup the database to backup.sql
	docker exec -it $(DOCKER_CONTAINER_NAME) \
		pg_dump $(DB_NAME) > backup.sql

restoredb: ## Restore database from backup.sql
	docker exec -i $(DOCKER_CONTAINER_NAME) \
		psql $(DB_NAME) < backup.sql

docker-up: ## Start docker-compose services
	@echo "🟢 Starting Docker containers..."
	docker compose -f $(DOCKER_COMPOSE) up -d

docker-down: ## Stop docker-compose services
	@echo "🔴 Stopping Docker containers..."
	docker compose -f $(DOCKER_COMPOSE) down

docker-logs: ## View logs from docker containers
	docker compose logs -f

docker-restart: docker-down docker-up ## Restart docker services

# ============================================================
#        🗄️ Database Migrations
# ============================================================

create_migration: ## Create a new migration file: make create_migration name=initial_schema
	migrate create -ext sql -dir "$(MIGRATE_PATH)" -seq $(name)

migrateup: ## Apply all up migrations
	migrate --path "$(MIGRATE_PATH)" --database "$(DB_URL)" --verbose up

migrateup1: ## Apply one up migration
	migrate --path "$(MIGRATE_PATH)" --database "$(DB_URL)" --verbose up 1

migratedown: ## Rollback all migrations
	migrate --path "$(MIGRATE_PATH)" --database "$(DB_URL)" --verbose down

migratedown1: ## Rollback one migration
	migrate --path "$(MIGRATE_PATH)" --database "$(DB_URL)" --verbose down 1

# ============================================================
# Make everything phony
# ============================================================
.PHONY: help tidy fmt lint test build run clean version \
        postgres createdb backupdb restoredb \
        docker-up docker-down docker-logs docker-restart \
        create_migration migrateup migrateup1 migratedown migratedown1 \
        deps verify
