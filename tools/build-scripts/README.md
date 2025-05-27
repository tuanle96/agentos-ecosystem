# Build Scripts

> Automated build scripts for AgentOS hybrid ecosystem (Go + Python)

## Overview

Collection of build automation scripts for compiling, testing, and packaging AgentOS hybrid components including Go backend services and Python AI workers.

## Scripts

### Core Build Scripts

- **build-all.sh**: Build all services and packages (Go + Python + Frontend)
- **build-go-services.sh**: Build Go backend services
- **build-python-workers.sh**: Build Python AI workers
- **build-frontend.sh**: Build frontend packages
- **build-docker.sh**: Build Docker images for hybrid architecture
- **clean-all.sh**: Clean build artifacts

### Quality Scripts

- **lint-all.sh**: Run linters on all code (Go + Python + JavaScript)
- **format-all.sh**: Format all code (gofmt + black + prettier)
- **test-all.sh**: Run all tests (Go + Python + Frontend)
- **test-go.sh**: Run Go service tests
- **test-python.sh**: Run Python AI worker tests
- **test-integration.sh**: Run Go ↔ Python integration tests
- **security-scan.sh**: Security vulnerability scanning

### Utility Scripts

- **deps-update.sh**: Update dependencies
- **docs-generate.sh**: Generate documentation
- **version-bump.sh**: Bump version numbers
- **changelog-generate.sh**: Generate changelog

## Usage

### **Hybrid Development Workflow**

```bash
# Make scripts executable
chmod +x tools/build-scripts/*.sh

# Build everything (Go + Python + Frontend)
./tools/build-scripts/build-all.sh

# Build specific components
./tools/build-scripts/build-go-services.sh      # Go backend services
./tools/build-scripts/build-python-workers.sh   # Python AI workers
./tools/build-scripts/build-frontend.sh         # React frontend

# Development workflow
make dev-services                                # Start infrastructure
./tools/build-scripts/build-go-services.sh      # Build Go services
./tools/build-scripts/build-python-workers.sh   # Setup Python workers
make dev                                         # Start development

# Quality checks
./tools/build-scripts/lint-all.sh               # Lint Go + Python + JS
./tools/build-scripts/test-all.sh               # Test all components
./tools/build-scripts/test-integration.sh       # Test Go ↔ Python integration

# Clean up
./tools/build-scripts/clean-all.sh
```

## Configuration

Build configuration in `build-config.yml`:

```yaml
# Go Backend Services
go:
  version: "1.21"
  services:
    - core-api
    - agent-engine
    - memory-service
    - tool-registry
    - auth-service
    - billing-service
    - notification-service

# Python AI Workers
python:
  version: "3.11"
  workers:
    - langchain-worker
    - crewai-worker
    - swarms-worker
    - autogen-worker
    - embedding-worker
    - model-worker

# Frontend Packages
node:
  version: "18"
  packages:
    - packages/core
    - packages/ui-components
    - packages/api-client
    - packages/testing

# Docker Configuration
docker:
  registry: "registry.agentos.ai"
  tag_latest: true
  platforms:
    - linux/amd64
    - linux/arm64

  # Go services (small images)
  go_base_image: "alpine:3.18"

  # Python workers (larger images)
  python_base_image: "python:3.11-slim"
```

## Build Targets

### Development
- **Go Services**: Fast builds with debug symbols, hot reload with Air
- **Python Workers**: Virtual environments, development dependencies
- **Frontend**: Development server with hot reload
- **Integration**: Local Docker Compose for full stack testing

### Production
- **Go Services**: Optimized static binaries, minimal Alpine images
- **Python Workers**: Production dependencies only, security hardening
- **Frontend**: Minified assets, CDN optimization
- **Deployment**: Multi-stage Docker builds, Kubernetes manifests

### Testing
- **Go Services**: Test coverage reports, benchmark results
- **Python Workers**: Pytest with coverage, AI model testing
- **Integration**: End-to-end testing of Go ↔ Python communication
- **Performance**: Load testing hybrid architecture

## Hybrid Architecture Benefits

### Build Performance
- **Go Services**: Sub-second compilation, single binary output
- **Python Workers**: Isolated environments, parallel builds
- **Container Size**: 20MB Go images vs 200MB Python images
- **Deployment Speed**: Fast Go service startup, cached Python dependencies