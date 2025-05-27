# AgentOS Ecosystem - Makefile
# Go-based backend services development workflow

.PHONY: help build test clean dev docker-up docker-down migrate lint format deps

# Default target
help: ## Show this help message
	@echo "AgentOS Ecosystem - Development Commands"
	@echo "========================================"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development
dev: ## Start development environment with hot reload
	@echo "Starting AgentOS development environment..."
	docker-compose up -d postgres redis nats elasticsearch minio
	@echo "Waiting for services to be ready..."
	sleep 10
	@echo "Starting Go services with hot reload..."
	air -c .air.toml

dev-services: ## Start only infrastructure services
	@echo "Starting infrastructure services..."
	docker-compose up -d postgres redis nats elasticsearch minio prometheus grafana jaeger

dev-stop: ## Stop development environment
	@echo "Stopping development environment..."
	docker-compose down

# Build
build: ## Build all Go services
	@echo "Building all Go services..."
	@for service in services/*/; do \
		if [ -f "$$service/main.go" ]; then \
			echo "Building $$service..."; \
			cd "$$service" && go build -o bin/service ./... && cd ../..; \
		fi \
	done

build-core-api: ## Build core API service
	@echo "Building core API service..."
	cd services/core-api && go build -o bin/core-api ./...

build-agent-engine: ## Build agent engine service
	@echo "Building agent engine service..."
	cd services/agent-engine && go build -o bin/agent-engine ./...

build-memory-service: ## Build memory service
	@echo "Building memory service..."
	cd services/memory-service && go build -o bin/memory-service ./...

build-tool-registry: ## Build tool registry service
	@echo "Building tool registry service..."
	cd services/tool-registry && go build -o bin/tool-registry ./...

# Testing
test: ## Run all tests
	@echo "Running all tests..."
	go test ./...

test-verbose: ## Run tests with verbose output
	@echo "Running tests with verbose output..."
	go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-core-api: ## Test core API service
	@echo "Testing core API service..."
	cd services/core-api && go test ./...

test-agent-engine: ## Test agent engine service
	@echo "Testing agent engine service..."
	cd services/agent-engine && go test ./...

test-memory-service: ## Test memory service
	@echo "Testing memory service..."
	cd services/memory-service && go test ./...

test-tool-registry: ## Test tool registry service
	@echo "Testing tool registry service..."
	cd services/tool-registry && go test ./...

# Code Quality
lint: ## Run linter
	@echo "Running linter..."
	golangci-lint run ./...

format: ## Format Go code
	@echo "Formatting Go code..."
	gofmt -s -w .
	goimports -w .

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

# Dependencies
deps: ## Download and tidy dependencies
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

deps-update: ## Update dependencies
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

deps-vendor: ## Vendor dependencies
	@echo "Vendoring dependencies..."
	go mod vendor

# Database
migrate-up: ## Run database migrations up
	@echo "Running database migrations up..."
	migrate -path migrations -database "postgres://agentos:agentos_dev_password@localhost:5432/agentos_dev?sslmode=disable" up

migrate-down: ## Run database migrations down
	@echo "Running database migrations down..."
	migrate -path migrations -database "postgres://agentos:agentos_dev_password@localhost:5432/agentos_dev?sslmode=disable" down

migrate-create: ## Create new migration (usage: make migrate-create name=migration_name)
	@echo "Creating new migration: $(name)"
	migrate create -ext sql -dir migrations $(name)

# Docker
docker-build: ## Build Docker images
	@echo "Building Docker images..."
	docker-compose build

docker-up: ## Start all services with Docker
	@echo "Starting all services with Docker..."
	docker-compose up -d

docker-down: ## Stop all Docker services
	@echo "Stopping all Docker services..."
	docker-compose down

docker-logs: ## Show Docker logs
	@echo "Showing Docker logs..."
	docker-compose logs -f

docker-clean: ## Clean Docker containers and volumes
	@echo "Cleaning Docker containers and volumes..."
	docker-compose down -v
	docker system prune -f

# Monitoring
logs-core-api: ## Show core API logs
	docker-compose logs -f core-api

logs-agent-engine: ## Show agent engine logs
	docker-compose logs -f agent-engine

logs-memory-service: ## Show memory service logs
	docker-compose logs -f memory-service

logs-tool-registry: ## Show tool registry logs
	docker-compose logs -f tool-registry

# Tools Installation
install-tools: ## Install development tools
	@echo "Installing development tools..."
	go install github.com/air-verse/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/swaggo/swag/cmd/swag@latest

# API Documentation
swagger: ## Generate Swagger documentation
	@echo "Generating Swagger documentation..."
	@for service in services/*/; do \
		if [ -f "$$service/main.go" ]; then \
			echo "Generating docs for $$service..."; \
			cd "$$service" && swag init && cd ../..; \
		fi \
	done

# Cleanup
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@find . -name "bin" -type d -exec rm -rf {} +
	@find . -name "*.out" -delete
	@find . -name "*.html" -delete
	go clean ./...

clean-all: clean docker-clean ## Clean everything including Docker

# Production
build-prod: ## Build for production
	@echo "Building for production..."
	@for service in services/*/; do \
		if [ -f "$$service/main.go" ]; then \
			echo "Building $$service for production..."; \
			cd "$$service" && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/service ./... && cd ../..; \
		fi \
	done

# Health Checks
health: ## Check service health
	@echo "Checking service health..."
	@curl -f http://localhost:8000/health || echo "Core API: DOWN"
	@curl -f http://localhost:8001/health || echo "Agent Engine: DOWN"
	@curl -f http://localhost:8002/health || echo "Memory Service: DOWN"
	@curl -f http://localhost:8003/health || echo "Tool Registry: DOWN"

# Environment
env-example: ## Create .env.example file
	@echo "Creating .env.example file..."
	@cat > .env.example << 'EOF'
# AgentOS Environment Configuration
GO_ENV=development
DATABASE_URL=postgres://agentos:agentos_dev_password@localhost:5432/agentos_dev?sslmode=disable
REDIS_URL=redis://localhost:6379/0
NATS_URL=nats://localhost:4222
ELASTICSEARCH_URL=http://localhost:9200
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=agentos
MINIO_SECRET_KEY=agentos_dev_password
JWT_SECRET=your-jwt-secret-here
OPENAI_API_KEY=your-openai-api-key
ANTHROPIC_API_KEY=your-anthropic-api-key
PINECONE_API_KEY=your-pinecone-api-key
EOF

# Setup
setup: deps install-tools env-example ## Setup development environment
	@echo "Setting up AgentOS development environment..."
	@echo "1. Dependencies downloaded"
	@echo "2. Development tools installed"
	@echo "3. Environment example created"
	@echo ""
	@echo "Next steps:"
	@echo "1. Copy .env.example to .env and configure your API keys"
	@echo "2. Run 'make dev-services' to start infrastructure"
	@echo "3. Run 'make migrate-up' to setup database"
	@echo "4. Run 'make dev' to start development environment"