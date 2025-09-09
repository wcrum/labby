# Spectro Lab Makefile

# Variables
APP_NAME = spectro-lab
DOCKER_IMAGE = spectro-lab
PORT = 8080
BACKEND_DIR = backend

# Default target
.DEFAULT_GOAL := help

# Help target
.PHONY: help
help: ## Show this help message
	@echo "Spectro Lab - Available targets:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Build targets
.PHONY: build-docker
build-docker: ## Build Docker image
	@echo "Building $(DOCKER_IMAGE) Docker image..."
	docker build -t $(DOCKER_IMAGE) .
	@echo "Docker image built successfully!"

.PHONY: build-local
build-local: ## Build local binary
	@echo "Building local binary..."
	cd . && pnpm run build
	mkdir -p $(BACKEND_DIR)/static
	cp -r out/* $(BACKEND_DIR)/static/
	cp public/*.png $(BACKEND_DIR)/static/
	cd $(BACKEND_DIR) && go build -o server cmd/server/main.go
	@echo "Local build completed!"

# Development targets
.PHONY: dev
dev: ## Start development environment
	@echo "Starting development environment..."
	cd . && pnpm dev

.PHONY: dev-backend
dev-backend: ## Start backend in development mode
	@echo "Starting backend in development mode..."
	cd $(BACKEND_DIR) && go run cmd/server/main.go

# Testing targets
.PHONY: test
test: ## Run all tests
	@echo "Running tests..."
	cd $(BACKEND_DIR) && go test ./...
	cd . && pnpm test

.PHONY: test-api
test-api: ## Test API endpoints
	@echo "Testing API endpoints..."
	cd $(BACKEND_DIR) && ./test_api.sh

# Documentation
.PHONY: swagger
swagger: ## Generate Swagger documentation
	@echo "Generating Swagger documentation..."
	cd $(BACKEND_DIR) && ./generate-swagger.sh

# Dependencies
.PHONY: deps
deps: ## Install all dependencies
	@echo "Installing dependencies..."
	cd $(BACKEND_DIR) && go mod download
	cd . && pnpm install

# Cleanup
.PHONY: clean
clean: ## Clean all build artifacts
	@echo "Cleaning build artifacts..."
	rm -f $(BACKEND_DIR)/server
	rm -rf out
	rm -rf .next
	docker rmi $(DOCKER_IMAGE) || true

# Docker Compose
.PHONY: up
up: ## Start with docker-compose
	docker-compose up

.PHONY: down
down: ## Stop docker-compose
	docker-compose down

# Database targets
.PHONY: db-up
db-up: ## Start database with docker-compose
	@echo "Starting database..."
	cd $(BACKEND_DIR) && docker-compose -f docker-compose.db.yml up -d
	@echo "Database started successfully!"

.PHONY: db-down
db-down: ## Stop database and delete volumes
	@echo "Stopping database and deleting volumes..."
	cd $(BACKEND_DIR) && docker-compose -f docker-compose.db.yml down -v
	@echo "Database stopped and volumes deleted!"

# Health check
.PHONY: health
health: ## Check application health
	@echo "Checking application health..."
	curl -f http://localhost:$(PORT)/health || echo "Application not running"
