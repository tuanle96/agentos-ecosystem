# Development Tools

> Development and deployment tools for AgentOS ecosystem

## Overview

Collection of tools and scripts for development, testing, building, and deploying the AgentOS ecosystem.

## Directory Structure

```
tools/
├── build-scripts/   # [PUBLIC] Build automation
├── deployment/      # [PRIVATE] Deployment scripts
├── testing/         # [PUBLIC] Testing utilities
├── ci-cd/          # [PRIVATE] CI/CD configurations
└── README.md       # This file
```

## Public Tools (Open Source)

### Build Scripts
- **build-all.sh**: Build all services and packages
- **test-all.sh**: Run comprehensive test suite
- **lint-all.sh**: Code quality checks
- **format-all.sh**: Code formatting

### Testing Tools
- **integration-tests/**: Integration test suite
- **load-tests/**: Performance and load testing
- **e2e-tests/**: End-to-end testing
- **test-data/**: Test data generators

## Private Tools (Internal)

### Deployment Scripts
- **deploy-prod.sh**: Production deployment
- **deploy-staging.sh**: Staging deployment
- **rollback.sh**: Rollback deployments
- **health-check.sh**: Post-deployment health checks

### CI/CD Configurations
- **github-actions/**: GitHub Actions workflows
- **jenkins/**: Jenkins pipeline configurations
- **docker/**: Docker build configurations
- **security/**: Security scanning tools

## Usage

### Development

```bash
# Build everything
./tools/build-scripts/build-all.sh

# Run tests
./tools/testing/run-tests.sh

# Format code
./tools/build-scripts/format-all.sh

# Lint code
./tools/build-scripts/lint-all.sh
```

### Testing

```bash
# Unit tests
./tools/testing/unit-tests.sh

# Integration tests
./tools/testing/integration-tests.sh

# Load tests
./tools/testing/load-tests.sh

# End-to-end tests
./tools/testing/e2e-tests.sh
```

### Deployment (Internal)

```bash
# Deploy to staging
./tools/deployment/deploy-staging.sh

# Deploy to production
./tools/deployment/deploy-prod.sh

# Health check
./tools/deployment/health-check.sh

# Rollback if needed
./tools/deployment/rollback.sh
```

## Tool Categories

### Build Tools
- **Go**: Build Go services
- **Node.js**: Build frontend packages
- **Docker**: Container builds
- **Documentation**: Generate docs

### Testing Tools
- **Unit Testing**: Service-level tests
- **Integration Testing**: Cross-service tests
- **Load Testing**: Performance validation
- **Security Testing**: Vulnerability scanning

### Quality Tools
- **Linting**: Code quality checks
- **Formatting**: Code style enforcement
- **Security Scanning**: Dependency vulnerabilities
- **License Checking**: License compliance

### Deployment Tools
- **Environment Setup**: Infrastructure provisioning
- **Application Deployment**: Service deployment
- **Database Migration**: Schema updates
- **Monitoring Setup**: Observability configuration

## Configuration

Tool configurations in `tools/config/`:

```yaml
# tools/config/build.yml
build:
  go_version: "1.21"
  node_version: "18"
  docker_registry: "registry.agentos.ai"
  
testing:
  timeout: "30m"
  parallel: true
  coverage_threshold: 80
  
deployment:
  environments:
    - dev
    - staging
    - prod
  health_check_timeout: "5m"
```