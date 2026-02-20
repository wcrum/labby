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
up: ## Start all services with docker-compose
	@echo "Starting all services..."
	docker-compose up -d
	@echo "All services started successfully!"
	@echo "Services available at:"
	@echo "  - Main App: http://localhost:8080"
	@echo "  - Dex OIDC: http://localhost:5556"
	@echo "  - pgAdmin: http://localhost:8081"
	@echo "  - PostgreSQL: localhost:5432"

.PHONY: up-logs
up-logs: ## Start all services with logs
	@echo "Starting all services with logs..."
	docker-compose up

.PHONY: down
down: ## Stop all services
	@echo "Stopping all services..."
	docker-compose down
	@echo "All services stopped!"

.PHONY: down-volumes
down-volumes: ## Stop all services and delete volumes
	@echo "Stopping all services and deleting volumes..."
	docker-compose down -v
	@echo "All services stopped and volumes deleted!"

.PHONY: logs
logs: ## Show logs for all services
	docker-compose logs -f

.PHONY: logs-app
logs-app: ## Show logs for main application
	docker-compose logs -f spectro-lab

.PHONY: logs-db
logs-db: ## Show logs for database
	docker-compose logs -f postgres

.PHONY: logs-dex
logs-dex: ## Show logs for Dex OIDC server
	docker-compose logs -f dex

.PHONY: dev-with-services
dev-with-services: up ## Start development environment with all services
	@echo "Starting development environment with all services..."
	@echo "Services:"
	@echo "  - Dex OIDC Server: http://localhost:5556"
	@echo "  - PostgreSQL Database: localhost:5432"
	@echo "  - pgAdmin: http://localhost:8081"
	@echo "  - Backend API: http://localhost:8080"
	@echo "  - Frontend: http://localhost:3000"
	@echo ""
	@echo "Test users (OIDC):"
	@echo "  - admin@spectrocloud.com / password (admin role)"
	@echo "  - test@test.com / password (example-org)"
	@echo ""
	@echo "Press Ctrl+C to stop all services"
	@echo ""
	@trap 'make down' INT TERM; \
	cd $(BACKEND_DIR) && go run cmd/server/main.go & \
	BACKEND_PID=$$!; \
	cd .. && pnpm dev & \
	FRONTEND_PID=$$!; \
	wait $$BACKEND_PID $$FRONTEND_PID

# Health check
.PHONY: health
health: ## Check application health
	@echo "Checking application health..."
	curl -f http://localhost:$(PORT)/health || echo "Application not running"
