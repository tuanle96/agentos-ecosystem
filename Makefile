# AgentOS Ecosystem - Makefile
# Go-based backend services development workflow

.PHONY: help build test clean dev docker-up docker-down migrate lint format deps setup

# Default target
help: ## Show this help message
	@echo "AgentOS Ecosystem - Development Commands"
	@echo "========================================"
	@echo "setup           - Setup development environment"
	@echo "dev             - Start development environment"
	@echo "dev-services    - Start infrastructure services only"
	@echo "build           - Build all Go services"
	@echo "test            - Run all tests"
	@echo "clean           - Clean build artifacts"
	@echo "docker-up       - Start all services with Docker"
	@echo "docker-down     - Stop all Docker services"

# Development
setup: ## Setup development environment
	@echo "Setting up AgentOS development environment..."
	go mod tidy
	@echo "Environment setup complete!"

dev: ## Start development environment with hot reload
	@echo "Starting AgentOS development environment..."
	docker-compose up -d postgres redis nats elasticsearch minio
	@echo "Infrastructure services started!"

dev-services: ## Start only infrastructure services
	@echo "Starting infrastructure services..."
	docker-compose up -d postgres redis nats elasticsearch minio prometheus grafana jaeger

dev-stop: ## Stop development environment
	@echo "Stopping development environment..."
	docker-compose down

# Build
build: ## Build all Go services
	@echo "Building all Go services..."
	go build ./...

build-core-api: ## Build core API service
	@echo "Building core API service..."
	@if [ -d "services/core-api" ]; then cd services/core-api && go build -o bin/core-api ./...; fi

build-agent-engine: ## Build agent engine service
	@echo "Building agent engine service..."
	@if [ -d "services/agent-engine" ]; then cd services/agent-engine && go build -o bin/agent-engine ./...; fi

build-memory-service: ## Build memory service
	@echo "Building memory service..."
	@if [ -d "services/memory-service" ]; then cd services/memory-service && go build -o bin/memory-service ./...; fi

build-tool-registry: ## Build tool registry service
	@echo "Building tool registry service..."
	@if [ -d "services/tool-registry" ]; then cd services/tool-registry && go build -o bin/tool-registry ./...; fi

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
	@if [ -f coverage.out ]; then go tool cover -html=coverage.out -o coverage.html; fi

# Docker
docker-up: ## Start all services with Docker
	@echo "Starting all services with Docker..."
	docker-compose up -d

docker-down: ## Stop all Docker services
	@echo "Stopping all Docker services..."
	docker-compose down

docker-logs: ## Show Docker logs
	@echo "Showing Docker logs..."
	docker-compose logs -f

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

docker-build: ## Build Docker images
	@echo "Building Docker images..."
	docker-compose build

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
	@echo "# AgentOS Environment Configuration" > .env.example
	@echo "GO_ENV=development" >> .env.example
	@echo "DATABASE_URL=postgres://agentos:agentos_dev_password@localhost:5432/agentos_dev?sslmode=disable" >> .env.example
	@echo "REDIS_URL=redis://localhost:6379/0" >> .env.example
	@echo "NATS_URL=nats://localhost:4222" >> .env.example
	@echo "JWT_SECRET=your-jwt-secret-here" >> .env.example
	@echo "OPENAI_API_KEY=your-openai-api-key" >> .env.example

# Setup
setup-full: deps install-tools env-example ## Full setup with tools installation
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

# Week 4 Advanced Memory System Testing
test-week4: ## Run all Week 4 memory system tests
	@echo "🧠 Running Week 4 Advanced Memory System Tests..."
	chmod +x scripts/run_week4_tests.sh
	./scripts/run_week4_tests.sh

test-week4-go: ## Run Week 4 Go memory tests
	@echo "🔧 Running Week 4 Go Memory Tests..."
	cd core/api && go test -v -race ./tests/memory_handlers_unit_test.go ./tests/setup_test.go
	cd core/api && go test -v -race ./tests/memory_integration_test.go ./tests/setup_test.go

test-week4-python: ## Run Week 4 Python memory tests
	@echo "🐍 Running Week 4 Python Memory Tests..."
	cd core/ai-worker && python -m pytest tests/test_mem0_memory_engine.py tests/test_framework_adapters.py -v --tb=short

test-week4-mem0: ## Run Week 4 mem0 integration tests
	@echo "🧠 Running Week 4 mem0 Integration Tests..."
	cd core/api && go test -v -race ./tests/week4_mem0_integration_test.go ./tests/setup_test.go

test-week4-coverage: ## Generate Week 4 coverage reports
	@echo "📊 Generating Week 4 Coverage Reports..."
	chmod +x scripts/run_week4_tests.sh
	./scripts/run_week4_tests.sh --coverage

test-week4-performance: ## Run Week 4 performance tests
	@echo "⚡ Running Week 4 Performance Tests..."
	cd core/api && go test -bench=. -benchmem ./tests/memory_handlers_unit_test.go ./tests/setup_test.go
	cd core/ai-worker && python -m pytest tests/test_mem0_memory_engine.py::TestMem0MemoryEnginePerformance -v --benchmark-only

test-memory-unit: ## Run memory unit tests
	@echo "🧪 Running Memory Unit Tests..."
	cd core/api && go test -v ./tests/memory_handlers_unit_test.go ./tests/setup_test.go
	cd core/ai-worker && python -m pytest tests/test_mem0_memory_engine.py -v --tb=short

test-memory-integration: ## Run memory integration tests
	@echo "🔗 Running Memory Integration Tests..."
	cd core/api && go test -v ./tests/memory_integration_test.go ./tests/setup_test.go

test-framework-adapters: ## Run framework adapter tests
	@echo "🔌 Running Framework Adapter Tests..."
	cd core/ai-worker && python -m pytest tests/test_framework_adapters.py -v --tb=short

test-memory-benchmarks: ## Run memory system benchmarks
	@echo "📈 Running Memory System Benchmarks..."
	cd core/api && go test -bench=BenchmarkMemoryOperations -benchmem ./tests/memory_handlers_unit_test.go ./tests/setup_test.go
	cd core/api && go test -bench=BenchmarkMemoryOperations -benchmem ./tests/memory_integration_test.go ./tests/setup_test.go